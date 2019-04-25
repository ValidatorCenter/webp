package main

import (
	"fmt"

	s "github.com/ValidatorCenter/prs3r/strc"

	// база данных на SQL
	"github.com/jmoiron/sqlx"
	_ "github.com/kshvakov/clickhouse"
)

// Получить данные о всех монетах в базе
func srchCoin(db *sqlx.DB) []s.CoinMarketCapData {
	allCoins := []s.CoinMarketCapData{}

	err := db.Select(&allCoins, `
		SELECT *
		FROM coins
	`)
	if err != nil {
		log("ERR", fmt.Sprint("[coin_sql.go] srchCoin(Select) - ERR:", err), "")
		panic(err) //dbg
	}

	return allCoins
}

// Получить транзакции для монеты (взята из CMC0)
// TODO: нужно будет добавить период поиска транзакция и вариант свертывания по дате(день, месяц, час...)
func srchCoinTrxSql(db *sqlx.DB, coinDt *s.CoinMarketCapData) error {
	coinTrxs := []s.CoinActionpData{}
	/*err := db.Select(&coinTrxs, fmt.Sprintf(`
	SELECT
		b.time AS time,
		p.hash AS hash,
		p.type AS type,
		ps.coin_to_sell AS coin_to_sell,
		ps.coin_to_buy AS coin_to_buy,
		ps.value_to_sell_f32 AS value_to_sell_f32,
		ps.value_to_buy_f32 AS value_to_buy_f32
	FROM trx p
	LEFT OUTER JOIN trx_data ps ON p.hash = ps.hash
	LEFT OUTER JOIN blocks b ON p.height_i32= b.height_i32
	WHERE (p.type = %d OR p.type = %d OR p.type = %d) AND (ps.coin_to_sell = '%s' OR ps.coin_to_buy = '%s')
	`, ms.TX_SellAllCoinData, ms.TX_SellCoinData, ms.TX_BuyCoinData, coinDt.CoinSymbol, coinDt.CoinSymbol))
	*/
	err := db.Select(&coinTrxs, fmt.Sprintf(`
	SELECT *
	FROM coin_trx
	WHERE (coin_to_sell = '%s' OR coin_to_buy = '%s')
	`, coinDt.CoinSymbol, coinDt.CoinSymbol))

	if err != nil {
		return err
	}

	coinDt.Transactions = coinTrxs
	return err
}

// Получить информацию об одной монете по Тикеру из базы (с транзакциями для графика)
func srchCoin1(db *sqlx.DB, ticker string) s.CoinMarketCapData {
	allCoins := s.CoinMarketCapData{}

	err := db.Get(&allCoins, fmt.Sprintf(`
		SELECT * 
		FROM coins
		WHERE symbol = '%s'
	`, ticker))

	if err != nil {
		if err.Error() == "sql: no rows in result set" {
			log("WRN", fmt.Sprint("[coin_sql.go] srchCoin1(Get) - ", err), "")
		} else {
			log("ERR", fmt.Sprint("[coin_sql.go] srchCoin1(Get) - ", err), "")
			panic(err)
		}
		return allCoins
	}

	// TODO: 1 - Получение доп.инфы из Redis! будет не этой функции, хотя не совсем правильно!

	// 2 - Получение транзакций монеты из SQL
	err = srchCoinTrxSql(db, &allCoins)
	if err != nil {
		log("ERR", fmt.Sprint("[coin_sql.go] srchCoin1(srchCoinTrxSql) - ERR:", err), "")
		panic(err) //dbg
	}

	return allCoins
}

// Поиск/проверка монеты по её создателю
// TODO: Переделать на возврат true/false после того как инф. будет в Redis!!!
func srchCoinCreator(db *sqlx.DB, ticker string, address string) s.CoinMarketCapData {
	vC := s.CoinMarketCapData{}
	err := db.Get(&vC, fmt.Sprintf(`
		SELECT * FROM coins
		WHERE symbol = '%s' AND creator = '%s'
	`, ticker, address))
	if err != nil {
		log("ERR", fmt.Sprint("[coin_sql.go] srchCoinCreator(Get) - ERR:", err), "")
		panic(err) //dbg
		return vC
	}
	return vC
}
