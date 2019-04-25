package main

import (
	"fmt"

	ms "github.com/ValidatorCenter/minter-go-sdk"
	s "github.com/ValidatorCenter/prs3r/strc"

	// база данных на SQL
	"github.com/jmoiron/sqlx"
	_ "github.com/kshvakov/clickhouse"
)

// Получение блоков из деапозона высот(ID), отсортированные по времени
func srchBlockN(db *sqlx.DB, minBlock uint32, maxBlock uint32) []s.BlockResponse2 {
	blsM := []s.BlockResponse2{}

	// Получаем блоки с конца(от последнего блока в системе), а не с начала!!!
	err := db.Select(&blsM, fmt.Sprintf(`
		SELECT b.height_i32, b.time, b.num_txs_i32, b.block_reward_f32, b.size_i32, b.proposer, count(bv.pub_key) as valid_amnt
		FROM blocks b 
		LEFT JOIN block_valid bv 
		ON b.height_i32 = bv.height_i32 
		WHERE b.height_i32 >= %d AND b.height_i32 <= %d 
		GROUP BY b.height_i32,b.time, b.num_txs_i32, b.block_reward_f32, b.size_i32, b.proposer
		ORDER BY b.height_i32 DESC
	`, maxBlock, minBlock))
	if err != nil {
		log("ERR", fmt.Sprint("[block_sql.go] srchBlockN(Select) - [", minBlock, "-", maxBlock, "] ERR:", err), "")
		panic(err) //dbg
	}

	return blsM
}

// Количество блоков загружено в базу
func srchBlockAmnt(db *sqlx.DB) int {
	// TODO: количество загруженных блоков можно брать из Redis!

	iAmnt := 0
	err := db.Get(&iAmnt, "SELECT count() FROM blocks")
	if err != nil {
		log("ERR", fmt.Sprint("[block_sql.go] srchBlockAmnt(Get) - ERR:", err), "")
		panic(err) //dbg
	}

	return iAmnt
}

// Получение блока(минимум), по его высоте(ID)
func srchBlockMin(db *sqlx.DB, height uint32) s.BlockResponse2 {
	var err error
	bls := s.BlockResponse2{}

	trxB1 := []s.TransResponseMin{}
	vldB1 := []ms.BlockValidatorsResponse{}
	evnB1 := []ms.BlockEventsResponse{}
	evnB1sql := []s.BlockEvent{}

	// TODO:
	// Тащим - список транзакций... хотя их hash имеется в блоке, но нужно более подробная инфа о них
	// Тащим - список валидаторов
	// Тащим - список событий блока
	// Upd.: структуры ms.* не совсем подходящие для моего SQL запроса ((((

	// TODO: Делема -> получать всё в одном запросе или разбить на несколько запросов к БД(SQL)???
	// TODO: Можно 4 параллельных запроса!!1

	err = db.Get(&bls, fmt.Sprintf(`
		SELECT b.*
		FROM blocks b
		WHERE b.height_i32 = %d
	`, height))
	if err != nil {
		if err.Error() == "sql: no rows in result set" {
			log("WRN", fmt.Sprint("[block_sql.go] srchBlock(Get->blocks) - [block №", height, "] ERR:", err), "")
		} else {
			log("ERR", fmt.Sprint("[block_sql.go] srchBlock(Get->blocks) - [block №", height, "] ERR:", err), "")
			panic(err) //dbg
		}
		return bls
	}

	err = db.Select(&trxB1, fmt.Sprintf(`
		SELECT b.hash, b.from_adrs, b.type, b.amount_bip_f32
		FROM trx b
		WHERE b.height_i32 = %d
	`, height))
	if err != nil {
		if err.Error() == "sql: no rows in result set" {
			log("WRN", fmt.Sprint("[block_sql.go] srchBlock(Select->trx) - [block №", height, "] ERR:", err), "")
		} else {
			log("ERR", fmt.Sprint("[block_sql.go] srchBlock(Select->trx) - [block №", height, "] ERR:", err), "")
			panic(err) //dbg
		}
		return bls
	}

	err = db.Select(&evnB1sql, fmt.Sprintf(`
		SELECT b.*
		FROM block_event b
		WHERE b.height_i32 = %d
	`, height))
	if err != nil {
		if err.Error() == "sql: no rows in result set" {
			log("WRN", fmt.Sprint("[block_sql.go] srchBlock(Select->block_evnt) - [block №", height, "] ERR:", err), "")
		} else {
			log("ERR", fmt.Sprint("[block_sql.go] srchBlock(Select->block_evnt) - [block №", height, "] ERR:", err), "")
			panic(err) //dbg
		}
		return bls
	}
	//FIXME: рефакторинг: цикла, который ниже. А точнее - нужно переделать структуру s.BlockResponse2{}
	for iEv, _ := range evnB1sql {
		evnB1 = append(evnB1, ms.BlockEventsResponse{
			Type: evnB1sql[iEv].Type,
			Value: ms.EventValueData{
				Role:            evnB1sql[iEv].Role,
				Address:         evnB1sql[iEv].Address,
				Amount:          evnB1sql[iEv].Amount,
				Coin:            evnB1sql[iEv].Coin,
				ValidatorPubKey: evnB1sql[iEv].ValidatorPubKey,
			},
		})
	}

	err = db.Select(&vldB1, fmt.Sprintf(`
		SELECT b.pub_key, b.signed
		FROM block_valid b
		WHERE b.height_i32 = %d
	`, height))
	if err != nil {
		if err.Error() == "sql: no rows in result set" {
			log("WRN", fmt.Sprint("[block_sql.go] srchBlock(Select->block_valid) - [block №", height, "] ERR:", err), "")
		} else {
			log("ERR", fmt.Sprint("[block_sql.go] srchBlock(Select->block_valid) - [block №", height, "] ERR:", err), "")
			panic(err) //dbg
		}
		return bls
	}

	bls.Transactions = trxB1
	bls.Validators = vldB1
	bls.Events = evnB1

	return bls
}
