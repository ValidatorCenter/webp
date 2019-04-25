package main

import (
	"encoding/json"
	"fmt"
	"time"

	ms "github.com/ValidatorCenter/minter-go-sdk"
	s "github.com/ValidatorCenter/prs3r/strc"

	// база данных на SQL
	"github.com/jmoiron/sqlx"
	_ "github.com/kshvakov/clickhouse"
)

// FIXME: Универсальная структура транзакции для получения из "SELECT * FROM trx LEFT JOIN trx_data ON trx.hash = trx_data.hash"
// взята из cmc0 СДЕЛАТЬ ВНЕШНЕЙ В cmc0!!!!!!!!!!!!!!!!!!!!!!!!!!!!!
type TrxS struct {
	//trx:
	Hash         string    `json:"hash" db:"hash"`
	RawTx        string    `json:"raw_tx" db:"raw_tx"`
	Height       int32     `json:"height_i32" db:"height_i32"`
	Index        int32     `json:"index_i32" db:"index_i32"`
	FromAdrs     string    `json:"from_adrs" db:"from_adrs"`
	Nonce        int32     `json:"nonce_i32" db:"nonce_i32"`
	GasPrice     int32     `json:"gas_price_i32" db:"gas_price_i32"`
	GasCoin      string    `json:"gas_coin" db:"gas_coin"`
	GasUsed      int32     `json:"gas_used_i32" db:"gas_used_i32"`
	Type         int       `json:"type" db:"type"`
	Amount       float32   `json:"amount_bip_f32" db:"amount_bip_f32"`
	Payload      string    `json:"payload" db:"payload"`
	TagsReturn   float32   `json:"tags_return" db:"tags_return"`
	TagsSellAmnt float32   `json:"tags_sellamnt" db:"tags_sellamnt"`
	Code         int32     `json:"code" db:"code"`
	Log          string    `json:"log" db:"log"`
	UpdYCH       time.Time `json:"-" db:"updated_date"` // ClickHouse::UpdateDate
	//trx_data:
	Hash_Dt        string    `json:"-" db:"trx_data.hash"`
	UpdYCH_Dt      time.Time `json:"-" db:"trx_data.updated_date"` // ClickHouse::UpdateDate
	Coin           string    `json:"coin" db:"coin"`
	ToAdrs         string    `json:"to_adrs" db:"to_adrs"`
	Value          float32   `json:"value_f32" db:"value_f32"`
	CoinToSell     string    `json:"coin_to_sell" db:"coin_to_sell"`
	CoinToBuy      string    `json:"coin_to_buy" db:"coin_to_buy"`
	ValueToSell    float32   `json:"value_to_sell_f32" db:"value_to_sell_f32"`
	ValueToBuy     float32   `json:"value_to_buy_f32" db:"value_to_buy_f32"`
	Name           string    `json:"name" db:"name"`
	Symbol         string    `json:"symbol" db:"symbol"`
	CRR            int32     `json:"constant_reserve_ratio" db:"constant_reserve_ratio"`
	InitialAmount  float32   `json:"initial_amount_f32" db:"initial_amount_f32"`
	InitialReserve float32   `json:"initial_reserve_f32" db:"initial_reserve_f32"`
	Address        string    `json:"address" db:"address"`
	PubKey         string    `json:"pub_key" db:"pub_key"`
	Commission     int32     `json:"commission" db:"commission"`
	Stake          float32   `json:"stake_f32" db:"stake_f32"`
	RawCheck       string    `json:"raw_check" db:"raw_check"`
	Proof          string    `json:"proof" db:"proof"`
	Coin13s        string    `json:"-" db:"coin_13as"` //???
	Coin13a        []string  `json:"coin_13a" db:"coin_13a"`
	To13s          string    `json:"-" db:"to_13as"` //???
	To13a          []string  `json:"to_13a" db:"to_13a"`
	Value13s       string    `json:"-" db:"value_f32_13as"` //???
	Value13a       []float32 `json:"value_f32_13a" db:"value_f32_13a"`
}

// Все транзакции в диапозоне, по порядку (отсортированы по дате)
func srchTrxList(db *sqlx.DB, skipTrns uint32) []ms.TransResponse {
	trnM := []ms.TransResponse{}
	dt := []TrxS{}
	//ASC (по возрастанию) или DESC (по убыванию)
	//LIMIT n, m позволяет выбрать из результата первые m строк после пропуска первых n строк.
	//	n и m должны быть неотрицательными целыми числами.
	//  https://clickhouse.yandex/docs/ru/query_language/select/#sektsiia-limit
	err := db.Select(&dt, fmt.Sprintf(`
		SELECT * FROM trx 
		LEFT JOIN trx_data 
		ON trx.hash = trx_data.hash 
		ORDER BY trx.height_i32 DESC, trx.updated_date DESC 
		LIMIT %d, 50
	`, skipTrns))
	if err != nil {
		log("ERR", fmt.Sprint("[transaction_sql.go] srchTrxList(Select) - ERR:", err), "")
		panic(err) //dbg
	}

	for _, st1 := range dt {
		//+ Код может 'тут' и не пригодится, но он универсальный и подойдет в другом месте
		if len(st1.Coin13s) > 0 {
			json.Unmarshal([]byte(st1.Coin13s), &st1.Coin13a)
		}
		if len(st1.To13s) > 0 {
			json.Unmarshal([]byte(st1.To13s), &st1.To13a)
		}
		if len(st1.Value13s) > 0 {
			json.Unmarshal([]byte(st1.Value13s), &st1.Value13a)
		}
		//-

		// FIXME: Рефакторинг, надо попробывать dt0 := interface{} и потом в case присваивать s.Tx...Data{}!
		switch st1.Type {
		case ms.TX_SendData: //1
			dt0 := s.Tx1SendData{}
			jsonBytes, _ := json.Marshal(st1)
			json.Unmarshal(jsonBytes, &dt0)
			trnM = append(trnM, ms.TransResponse{
				Hash:     st1.Hash,
				RawTx:    st1.RawTx,
				Height:   int(st1.Height),
				Index:    int(st1.Index),
				From:     st1.FromAdrs,
				Nonce:    int(st1.Nonce),
				GasPrice: int(st1.GasPrice),
				GasCoin:  st1.GasCoin,
				GasUsed:  int(st1.GasUsed),
				Type:     st1.Type,
				Payload:  st1.Payload,
				Code:     int(st1.Code),
				Log:      st1.Log,
				Data:     dt0,
			})
		case ms.TX_SellCoinData: //2
			dt0 := s.Tx2SellCoinData{}
			jsonBytes, _ := json.Marshal(st1)
			json.Unmarshal(jsonBytes, &dt0)
			trnM = append(trnM, ms.TransResponse{
				Hash:     st1.Hash,
				RawTx:    st1.RawTx,
				Height:   int(st1.Height),
				Index:    int(st1.Index),
				From:     st1.FromAdrs,
				Nonce:    int(st1.Nonce),
				GasPrice: int(st1.GasPrice),
				GasCoin:  st1.GasCoin,
				GasUsed:  int(st1.GasUsed),
				Type:     st1.Type,
				Payload:  st1.Payload,
				Code:     int(st1.Code),
				Log:      st1.Log,
				Data:     dt0,
			})
		case ms.TX_SellAllCoinData: //3
			dt0 := s.Tx3SellAllCoinData{}
			jsonBytes, _ := json.Marshal(st1)
			json.Unmarshal(jsonBytes, &dt0)
			trnM = append(trnM, ms.TransResponse{
				Hash:     st1.Hash,
				RawTx:    st1.RawTx,
				Height:   int(st1.Height),
				Index:    int(st1.Index),
				From:     st1.FromAdrs,
				Nonce:    int(st1.Nonce),
				GasPrice: int(st1.GasPrice),
				GasCoin:  st1.GasCoin,
				GasUsed:  int(st1.GasUsed),
				Type:     st1.Type,
				Payload:  st1.Payload,
				Code:     int(st1.Code),
				Log:      st1.Log,
				Data:     dt0,
			})
		case ms.TX_BuyCoinData: //4
			dt0 := s.Tx4BuyCoinData{}
			jsonBytes, _ := json.Marshal(st1)
			json.Unmarshal(jsonBytes, &dt0)
			trnM = append(trnM, ms.TransResponse{
				Hash:     st1.Hash,
				RawTx:    st1.RawTx,
				Height:   int(st1.Height),
				Index:    int(st1.Index),
				From:     st1.FromAdrs,
				Nonce:    int(st1.Nonce),
				GasPrice: int(st1.GasPrice),
				GasCoin:  st1.GasCoin,
				GasUsed:  int(st1.GasUsed),
				Type:     st1.Type,
				Payload:  st1.Payload,
				Code:     int(st1.Code),
				Log:      st1.Log,
				Data:     dt0,
			})
		case ms.TX_CreateCoinData: //5
			dt0 := s.Tx5CreateCoinData{}
			jsonBytes, _ := json.Marshal(st1)
			json.Unmarshal(jsonBytes, &dt0)
			trnM = append(trnM, ms.TransResponse{
				Hash:     st1.Hash,
				RawTx:    st1.RawTx,
				Height:   int(st1.Height),
				Index:    int(st1.Index),
				From:     st1.FromAdrs,
				Nonce:    int(st1.Nonce),
				GasPrice: int(st1.GasPrice),
				GasCoin:  st1.GasCoin,
				GasUsed:  int(st1.GasUsed),
				Type:     st1.Type,
				Payload:  st1.Payload,
				Code:     int(st1.Code),
				Log:      st1.Log,
				Data:     dt0,
			})
		case ms.TX_DeclareCandidacyData: //6
			dt0 := s.Tx6DeclareCandidacyData{}
			jsonBytes, _ := json.Marshal(st1)
			json.Unmarshal(jsonBytes, &dt0)
			trnM = append(trnM, ms.TransResponse{
				Hash:     st1.Hash,
				RawTx:    st1.RawTx,
				Height:   int(st1.Height),
				Index:    int(st1.Index),
				From:     st1.FromAdrs,
				Nonce:    int(st1.Nonce),
				GasPrice: int(st1.GasPrice),
				GasCoin:  st1.GasCoin,
				GasUsed:  int(st1.GasUsed),
				Type:     st1.Type,
				Payload:  st1.Payload,
				Code:     int(st1.Code),
				Log:      st1.Log,
				Data:     dt0,
			})
		case ms.TX_DelegateDate: //7
			dt0 := s.Tx7DelegateDate{}
			jsonBytes, _ := json.Marshal(st1)
			json.Unmarshal(jsonBytes, &dt0)
			trnM = append(trnM, ms.TransResponse{
				Hash:     st1.Hash,
				RawTx:    st1.RawTx,
				Height:   int(st1.Height),
				Index:    int(st1.Index),
				From:     st1.FromAdrs,
				Nonce:    int(st1.Nonce),
				GasPrice: int(st1.GasPrice),
				GasCoin:  st1.GasCoin,
				GasUsed:  int(st1.GasUsed),
				Type:     st1.Type,
				Payload:  st1.Payload,
				Code:     int(st1.Code),
				Log:      st1.Log,
				Data:     dt0,
			})
		case ms.TX_UnbondData: //8
			dt0 := s.Tx8UnbondData{}
			jsonBytes, _ := json.Marshal(st1)
			json.Unmarshal(jsonBytes, &dt0)
			trnM = append(trnM, ms.TransResponse{
				Hash:     st1.Hash,
				RawTx:    st1.RawTx,
				Height:   int(st1.Height),
				Index:    int(st1.Index),
				From:     st1.FromAdrs,
				Nonce:    int(st1.Nonce),
				GasPrice: int(st1.GasPrice),
				GasCoin:  st1.GasCoin,
				GasUsed:  int(st1.GasUsed),
				Type:     st1.Type,
				Payload:  st1.Payload,
				Code:     int(st1.Code),
				Log:      st1.Log,
				Data:     dt0,
			})
		case ms.TX_RedeemCheckData: //9
			dt0 := s.Tx9RedeemCheckData{}
			jsonBytes, _ := json.Marshal(st1)
			json.Unmarshal(jsonBytes, &dt0)
			trnM = append(trnM, ms.TransResponse{
				Hash:     st1.Hash,
				RawTx:    st1.RawTx,
				Height:   int(st1.Height),
				Index:    int(st1.Index),
				From:     st1.FromAdrs,
				Nonce:    int(st1.Nonce),
				GasPrice: int(st1.GasPrice),
				GasCoin:  st1.GasCoin,
				GasUsed:  int(st1.GasUsed),
				Type:     st1.Type,
				Payload:  st1.Payload,
				Code:     int(st1.Code),
				Log:      st1.Log,
				Data:     dt0,
			})
		case ms.TX_SetCandidateOnData: //10
			dt0 := s.Tx10SetCandidateOnData{}
			jsonBytes, _ := json.Marshal(st1)
			json.Unmarshal(jsonBytes, &dt0)
			trnM = append(trnM, ms.TransResponse{
				Hash:     st1.Hash,
				RawTx:    st1.RawTx,
				Height:   int(st1.Height),
				Index:    int(st1.Index),
				From:     st1.FromAdrs,
				Nonce:    int(st1.Nonce),
				GasPrice: int(st1.GasPrice),
				GasCoin:  st1.GasCoin,
				GasUsed:  int(st1.GasUsed),
				Type:     st1.Type,
				Payload:  st1.Payload,
				Code:     int(st1.Code),
				Log:      st1.Log,
				Data:     dt0,
			})
		case ms.TX_SetCandidateOffData: //11
			dt0 := s.Tx11SetCandidateOffData{}
			jsonBytes, _ := json.Marshal(st1)
			json.Unmarshal(jsonBytes, &dt0)
			trnM = append(trnM, ms.TransResponse{
				Hash:     st1.Hash,
				RawTx:    st1.RawTx,
				Height:   int(st1.Height),
				Index:    int(st1.Index),
				From:     st1.FromAdrs,
				Nonce:    int(st1.Nonce),
				GasPrice: int(st1.GasPrice),
				GasCoin:  st1.GasCoin,
				GasUsed:  int(st1.GasUsed),
				Type:     st1.Type,
				Payload:  st1.Payload,
				Code:     int(st1.Code),
				Log:      st1.Log,
				Data:     dt0,
			})
		case ms.TX_CreateMultisigData: //12
			// TODO: Реализовать
			//dt0 := s.Tx12CreateMultisigData{}
			log("WRN", "[transaction_sql.go] srchTrxHash(ms.TX_CreateMultisigData) - НЕ РЕАЛИЗОВАН!", "")
		case ms.TX_MultisendData: //13
			dt0 := s.Tx13MultisendData{}
			jsonBytes, _ := json.Marshal(st1)
			json.Unmarshal(jsonBytes, &dt0)
			trnM = append(trnM, ms.TransResponse{
				Hash:     st1.Hash,
				RawTx:    st1.RawTx,
				Height:   int(st1.Height),
				Index:    int(st1.Index),
				From:     st1.FromAdrs,
				Nonce:    int(st1.Nonce),
				GasPrice: int(st1.GasPrice),
				GasCoin:  st1.GasCoin,
				GasUsed:  int(st1.GasUsed),
				Type:     st1.Type,
				Payload:  st1.Payload,
				Code:     int(st1.Code),
				Log:      st1.Log,
				Data:     dt0,
			})
		case ms.TX_EditCandidateData: //14
			// TODO: реализовать
			log("WRN", "[transaction_sql.go] srchTrxHash(ms.TX_EditCandidateData) - НЕ РЕАЛИЗОВАН!", "")
		default:
			log("ERR", fmt.Sprint("[transaction_sql.go] srchTrxHash(...) - неизвестный статус st1.Type - ", st1.Type), "")
		}
	}

	return trnM
}

// Количество транзакций загружено в базу
func srchTrxAmnt(db *sqlx.DB) int {
	// TODO: количество загруженных блоков можно брать из Redis!

	iAmnt := 0
	// Если в запросе не перечислено ни одного столбца (например, SELECT count() FROM t), то из таблицы всё равно вынимается один какой-нибудь столбец (предпочитается самый маленький), для того, чтобы можно было хотя бы посчитать количество строк.
	err := db.Get(&iAmnt, "SELECT count() FROM trx")
	if err != nil {
		log("ERR", fmt.Sprint("[transaction_sql.go] srchTrxAmnt(Get) - ERR:", err), "")
		panic(err) //dbg
	}

	return iAmnt
}

// Получить транзакцию по Хэшу
func srchTrxHash(db *sqlx.DB, hashTrans string) ms.TransResponse {
	v := ms.TransResponse{}
	st1 := TrxS{}
	err := db.Get(&st1, fmt.Sprintf(`
		SELECT * 
		FROM trx 
		LEFT JOIN trx_data ON trx.hash = trx_data.hash 
		WHERE trx.hash = '%s'
	`, hashTrans))
	if err != nil {
		if err.Error() == "sql: no rows in result set" {
			log("WRN", fmt.Sprint("[transaction_sql.go] srchTrxHash(Get) - [hash №", hashTrans, "] ERR:", err), "")
		} else {
			log("ERR", fmt.Sprint("[transaction_sql.go] srchTrxHash(Get) - [hash №", hashTrans, "] ERR:", err), "")
			panic(err) //dbg
		}
		return v
	}

	// Теперь разбор
	//+ Код может 'тут' и не пригодится, но он универсальный и подойдет в другом месте
	if len(st1.Coin13s) > 0 {
		json.Unmarshal([]byte(st1.Coin13s), &st1.Coin13a)
	}
	if len(st1.To13s) > 0 {
		json.Unmarshal([]byte(st1.To13s), &st1.To13a)
	}
	if len(st1.Value13s) > 0 {
		json.Unmarshal([]byte(st1.Value13s), &st1.Value13a)
	}
	//-

	// FIXME: Рефакторинг, надо попробывать dt0 := interface{} и потом в case присваивать s.Tx...Data{}!
	switch st1.Type {
	case ms.TX_SendData: //1
		dt0 := s.Tx1SendData{}
		jsonBytes, _ := json.Marshal(st1)
		json.Unmarshal(jsonBytes, &dt0)
		v = ms.TransResponse{
			Hash:     st1.Hash,
			RawTx:    st1.RawTx,
			Height:   int(st1.Height),
			Index:    int(st1.Index),
			From:     st1.FromAdrs,
			Nonce:    int(st1.Nonce),
			GasPrice: int(st1.GasPrice),
			GasCoin:  st1.GasCoin,
			GasUsed:  int(st1.GasUsed),
			Type:     st1.Type,
			Payload:  st1.Payload,
			Code:     int(st1.Code),
			Log:      st1.Log,
			Data:     dt0,
		}
	case ms.TX_SellCoinData: //2
		dt0 := s.Tx2SellCoinData{}
		jsonBytes, _ := json.Marshal(st1)
		json.Unmarshal(jsonBytes, &dt0)
		v = ms.TransResponse{
			Hash:     st1.Hash,
			RawTx:    st1.RawTx,
			Height:   int(st1.Height),
			Index:    int(st1.Index),
			From:     st1.FromAdrs,
			Nonce:    int(st1.Nonce),
			GasPrice: int(st1.GasPrice),
			GasCoin:  st1.GasCoin,
			GasUsed:  int(st1.GasUsed),
			Type:     st1.Type,
			Payload:  st1.Payload,
			Code:     int(st1.Code),
			Log:      st1.Log,
			Data:     dt0,
		}
	case ms.TX_SellAllCoinData: //3
		dt0 := s.Tx3SellAllCoinData{}
		jsonBytes, _ := json.Marshal(st1)
		json.Unmarshal(jsonBytes, &dt0)
		v = ms.TransResponse{
			Hash:     st1.Hash,
			RawTx:    st1.RawTx,
			Height:   int(st1.Height),
			Index:    int(st1.Index),
			From:     st1.FromAdrs,
			Nonce:    int(st1.Nonce),
			GasPrice: int(st1.GasPrice),
			GasCoin:  st1.GasCoin,
			GasUsed:  int(st1.GasUsed),
			Type:     st1.Type,
			Payload:  st1.Payload,
			Code:     int(st1.Code),
			Log:      st1.Log,
			Data:     dt0,
		}
	case ms.TX_BuyCoinData: //4
		dt0 := s.Tx4BuyCoinData{}
		jsonBytes, _ := json.Marshal(st1)
		json.Unmarshal(jsonBytes, &dt0)
		v = ms.TransResponse{
			Hash:     st1.Hash,
			RawTx:    st1.RawTx,
			Height:   int(st1.Height),
			Index:    int(st1.Index),
			From:     st1.FromAdrs,
			Nonce:    int(st1.Nonce),
			GasPrice: int(st1.GasPrice),
			GasCoin:  st1.GasCoin,
			GasUsed:  int(st1.GasUsed),
			Type:     st1.Type,
			Payload:  st1.Payload,
			Code:     int(st1.Code),
			Log:      st1.Log,
			Data:     dt0,
		}
	case ms.TX_CreateCoinData: //5
		dt0 := s.Tx5CreateCoinData{}
		jsonBytes, _ := json.Marshal(st1)
		json.Unmarshal(jsonBytes, &dt0)
		v = ms.TransResponse{
			Hash:     st1.Hash,
			RawTx:    st1.RawTx,
			Height:   int(st1.Height),
			Index:    int(st1.Index),
			From:     st1.FromAdrs,
			Nonce:    int(st1.Nonce),
			GasPrice: int(st1.GasPrice),
			GasCoin:  st1.GasCoin,
			GasUsed:  int(st1.GasUsed),
			Type:     st1.Type,
			Payload:  st1.Payload,
			Code:     int(st1.Code),
			Log:      st1.Log,
			Data:     dt0,
		}
	case ms.TX_DeclareCandidacyData: //6
		dt0 := s.Tx6DeclareCandidacyData{}
		jsonBytes, _ := json.Marshal(st1)
		json.Unmarshal(jsonBytes, &dt0)
		v = ms.TransResponse{
			Hash:     st1.Hash,
			RawTx:    st1.RawTx,
			Height:   int(st1.Height),
			Index:    int(st1.Index),
			From:     st1.FromAdrs,
			Nonce:    int(st1.Nonce),
			GasPrice: int(st1.GasPrice),
			GasCoin:  st1.GasCoin,
			GasUsed:  int(st1.GasUsed),
			Type:     st1.Type,
			Payload:  st1.Payload,
			Code:     int(st1.Code),
			Log:      st1.Log,
			Data:     dt0,
		}
	case ms.TX_DelegateDate: //7
		dt0 := s.Tx7DelegateDate{}
		jsonBytes, _ := json.Marshal(st1)
		json.Unmarshal(jsonBytes, &dt0)
		v = ms.TransResponse{
			Hash:     st1.Hash,
			RawTx:    st1.RawTx,
			Height:   int(st1.Height),
			Index:    int(st1.Index),
			From:     st1.FromAdrs,
			Nonce:    int(st1.Nonce),
			GasPrice: int(st1.GasPrice),
			GasCoin:  st1.GasCoin,
			GasUsed:  int(st1.GasUsed),
			Type:     st1.Type,
			Payload:  st1.Payload,
			Code:     int(st1.Code),
			Log:      st1.Log,
			Data:     dt0,
		}
	case ms.TX_UnbondData: //8
		dt0 := s.Tx8UnbondData{}
		jsonBytes, _ := json.Marshal(st1)
		json.Unmarshal(jsonBytes, &dt0)
		v = ms.TransResponse{
			Hash:     st1.Hash,
			RawTx:    st1.RawTx,
			Height:   int(st1.Height),
			Index:    int(st1.Index),
			From:     st1.FromAdrs,
			Nonce:    int(st1.Nonce),
			GasPrice: int(st1.GasPrice),
			GasCoin:  st1.GasCoin,
			GasUsed:  int(st1.GasUsed),
			Type:     st1.Type,
			Payload:  st1.Payload,
			Code:     int(st1.Code),
			Log:      st1.Log,
			Data:     dt0,
		}
	case ms.TX_RedeemCheckData: //9
		dt0 := s.Tx9RedeemCheckData{}
		jsonBytes, _ := json.Marshal(st1)
		json.Unmarshal(jsonBytes, &dt0)
		v = ms.TransResponse{
			Hash:     st1.Hash,
			RawTx:    st1.RawTx,
			Height:   int(st1.Height),
			Index:    int(st1.Index),
			From:     st1.FromAdrs,
			Nonce:    int(st1.Nonce),
			GasPrice: int(st1.GasPrice),
			GasCoin:  st1.GasCoin,
			GasUsed:  int(st1.GasUsed),
			Type:     st1.Type,
			Payload:  st1.Payload,
			Code:     int(st1.Code),
			Log:      st1.Log,
			Data:     dt0,
		}
	case ms.TX_SetCandidateOnData: //10
		dt0 := s.Tx10SetCandidateOnData{}
		jsonBytes, _ := json.Marshal(st1)
		json.Unmarshal(jsonBytes, &dt0)
		v = ms.TransResponse{
			Hash:     st1.Hash,
			RawTx:    st1.RawTx,
			Height:   int(st1.Height),
			Index:    int(st1.Index),
			From:     st1.FromAdrs,
			Nonce:    int(st1.Nonce),
			GasPrice: int(st1.GasPrice),
			GasCoin:  st1.GasCoin,
			GasUsed:  int(st1.GasUsed),
			Type:     st1.Type,
			Payload:  st1.Payload,
			Code:     int(st1.Code),
			Log:      st1.Log,
			Data:     dt0,
		}
	case ms.TX_SetCandidateOffData: //11
		dt0 := s.Tx11SetCandidateOffData{}
		jsonBytes, _ := json.Marshal(st1)
		json.Unmarshal(jsonBytes, &dt0)
		v = ms.TransResponse{
			Hash:     st1.Hash,
			RawTx:    st1.RawTx,
			Height:   int(st1.Height),
			Index:    int(st1.Index),
			From:     st1.FromAdrs,
			Nonce:    int(st1.Nonce),
			GasPrice: int(st1.GasPrice),
			GasCoin:  st1.GasCoin,
			GasUsed:  int(st1.GasUsed),
			Type:     st1.Type,
			Payload:  st1.Payload,
			Code:     int(st1.Code),
			Log:      st1.Log,
			Data:     dt0,
		}
	case ms.TX_CreateMultisigData: //12
		// TODO: Реализовать
		//dt0 := s.Tx12CreateMultisigData{}
		log("WRN", "[transaction_sql.go] srchTrxHash(ms.TX_CreateMultisigData) - НЕ РЕАЛИЗОВАН!", "")
	case ms.TX_MultisendData: //13
		dt0 := s.Tx13MultisendData{}
		jsonBytes, _ := json.Marshal(st1)
		json.Unmarshal(jsonBytes, &dt0)
		v = ms.TransResponse{
			Hash:     st1.Hash,
			RawTx:    st1.RawTx,
			Height:   int(st1.Height),
			Index:    int(st1.Index),
			From:     st1.FromAdrs,
			Nonce:    int(st1.Nonce),
			GasPrice: int(st1.GasPrice),
			GasCoin:  st1.GasCoin,
			GasUsed:  int(st1.GasUsed),
			Type:     st1.Type,
			Payload:  st1.Payload,
			Code:     int(st1.Code),
			Log:      st1.Log,
			Data:     dt0,
		}
	case ms.TX_EditCandidateData: //14
		// TODO: реализовать
		log("WRN", "[transaction_sql.go] srchTrxHash(ms.TX_EditCandidateData) - НЕ РЕАЛИЗОВАН!", "")
	default:
		log("ERR", fmt.Sprint("[transaction_sql.go] srchTrxHash(...) - неизвестный статус st1.Type - ", st1.Type), "")
	}

	return v
}

// Количество транзакций загружено в базу определенного Адреса-кошелька
func srchTrxAddrsAmnt(db *sqlx.DB, addrs string) int {
	iAmnt := 0
	err := db.Get(&iAmnt, fmt.Sprintf("SELECT count() FROM trx WHERE from_adrs = '%s'", addrs))
	if err != nil {
		log("ERR", fmt.Sprint("[transaction_sql.go] srchTrxAddrsAmnt(Get) - ERR:", err), "")
		panic(err) //dbg
	}

	return iAmnt
}

// Поиск транзакции по Адресу("from": nmbrAddrs), в обратном порядке, за раз 50шт, с пагинацией(skipTrns)
func srchTrxAddrsSql(db *sqlx.DB, addrs string, skipTrns int) []ms.TransResponse {
	//trnsTbl.Find(bson.M{"from": nmbrAddrs}).Sort("-$natural").Skip(skipTrns).Limit(50).All(&trnM)
	trnM := []ms.TransResponse{}
	dt := []TrxS{}
	err := db.Select(&dt, fmt.Sprintf("SELECT * FROM trx LEFT JOIN trx_data ON trx.hash = trx_data.hash WHERE trx.from_adrs = '%s'", addrs))
	if err != nil {
		log("ERR", fmt.Sprint("[transaction_sql.go] srchTrxAddrsSql(Select) - [addrs №", addrs, "] ERR:", err), "")
		panic(err) //dbg
	}

	for _, st1 := range dt {
		//+ Код может 'тут' и не пригодится, но он универсальный и подойдет в другом месте
		if len(st1.Coin13s) > 0 {
			json.Unmarshal([]byte(st1.Coin13s), &st1.Coin13a)
		}
		if len(st1.To13s) > 0 {
			json.Unmarshal([]byte(st1.To13s), &st1.To13a)
		}
		if len(st1.Value13s) > 0 {
			json.Unmarshal([]byte(st1.Value13s), &st1.Value13a)
		}
		//-

		// FIXME: Рефакторинг, надо попробывать dt0 := interface{} и потом в case присваивать s.Tx...Data{}!
		switch st1.Type {
		case ms.TX_SendData: //1
			dt0 := s.Tx1SendData{}
			jsonBytes, _ := json.Marshal(st1)
			json.Unmarshal(jsonBytes, &dt0)
			trnM = append(trnM, ms.TransResponse{
				Hash:     st1.Hash,
				RawTx:    st1.RawTx,
				Height:   int(st1.Height),
				Index:    int(st1.Index),
				From:     st1.FromAdrs,
				Nonce:    int(st1.Nonce),
				GasPrice: int(st1.GasPrice),
				GasCoin:  st1.GasCoin,
				GasUsed:  int(st1.GasUsed),
				Type:     st1.Type,
				Payload:  st1.Payload,
				Code:     int(st1.Code),
				Log:      st1.Log,
				Data:     dt0,
			})
		case ms.TX_SellCoinData: //2
			dt0 := s.Tx2SellCoinData{}
			jsonBytes, _ := json.Marshal(st1)
			json.Unmarshal(jsonBytes, &dt0)
			dTR := ms.TransResponse{
				Hash:     st1.Hash,
				RawTx:    st1.RawTx,
				Height:   int(st1.Height),
				Index:    int(st1.Index),
				From:     st1.FromAdrs,
				Nonce:    int(st1.Nonce),
				GasPrice: int(st1.GasPrice),
				GasCoin:  st1.GasCoin,
				GasUsed:  int(st1.GasUsed),
				Type:     st1.Type,
				Payload:  st1.Payload,
				Code:     int(st1.Code),
				Log:      st1.Log,
				Data:     dt0,
			}
			dTR.Tags.TxReturn = st1.TagsReturn
			trnM = append(trnM, dTR)
		case ms.TX_SellAllCoinData: //3
			dt0 := s.Tx3SellAllCoinData{}
			jsonBytes, _ := json.Marshal(st1)
			json.Unmarshal(jsonBytes, &dt0)
			dTR := ms.TransResponse{
				Hash:     st1.Hash,
				RawTx:    st1.RawTx,
				Height:   int(st1.Height),
				Index:    int(st1.Index),
				From:     st1.FromAdrs,
				Nonce:    int(st1.Nonce),
				GasPrice: int(st1.GasPrice),
				GasCoin:  st1.GasCoin,
				GasUsed:  int(st1.GasUsed),
				Type:     st1.Type,
				Payload:  st1.Payload,
				Code:     int(st1.Code),
				Log:      st1.Log,
				Data:     dt0,
			}
			dTR.Tags.TxReturn = st1.TagsReturn
			trnM = append(trnM, dTR)
		case ms.TX_BuyCoinData: //4
			dt0 := s.Tx4BuyCoinData{}
			jsonBytes, _ := json.Marshal(st1)
			json.Unmarshal(jsonBytes, &dt0)
			dTR := ms.TransResponse{
				Hash:     st1.Hash,
				RawTx:    st1.RawTx,
				Height:   int(st1.Height),
				Index:    int(st1.Index),
				From:     st1.FromAdrs,
				Nonce:    int(st1.Nonce),
				GasPrice: int(st1.GasPrice),
				GasCoin:  st1.GasCoin,
				GasUsed:  int(st1.GasUsed),
				Type:     st1.Type,
				Payload:  st1.Payload,
				Code:     int(st1.Code),
				Log:      st1.Log,
				Data:     dt0,
			}
			dTR.Tags.TxReturn = st1.TagsReturn
			trnM = append(trnM, dTR)
		case ms.TX_CreateCoinData: //5
			dt0 := s.Tx5CreateCoinData{}
			jsonBytes, _ := json.Marshal(st1)
			json.Unmarshal(jsonBytes, &dt0)
			trnM = append(trnM, ms.TransResponse{
				Hash:     st1.Hash,
				RawTx:    st1.RawTx,
				Height:   int(st1.Height),
				Index:    int(st1.Index),
				From:     st1.FromAdrs,
				Nonce:    int(st1.Nonce),
				GasPrice: int(st1.GasPrice),
				GasCoin:  st1.GasCoin,
				GasUsed:  int(st1.GasUsed),
				Type:     st1.Type,
				Payload:  st1.Payload,
				Code:     int(st1.Code),
				Log:      st1.Log,
				Data:     dt0,
			})
		case ms.TX_DeclareCandidacyData: //6
			dt0 := s.Tx6DeclareCandidacyData{}
			jsonBytes, _ := json.Marshal(st1)
			json.Unmarshal(jsonBytes, &dt0)
			trnM = append(trnM, ms.TransResponse{
				Hash:     st1.Hash,
				RawTx:    st1.RawTx,
				Height:   int(st1.Height),
				Index:    int(st1.Index),
				From:     st1.FromAdrs,
				Nonce:    int(st1.Nonce),
				GasPrice: int(st1.GasPrice),
				GasCoin:  st1.GasCoin,
				GasUsed:  int(st1.GasUsed),
				Type:     st1.Type,
				Payload:  st1.Payload,
				Code:     int(st1.Code),
				Log:      st1.Log,
				Data:     dt0,
			})
		case ms.TX_DelegateDate: //7
			dt0 := s.Tx7DelegateDate{}
			jsonBytes, _ := json.Marshal(st1)
			json.Unmarshal(jsonBytes, &dt0)
			trnM = append(trnM, ms.TransResponse{
				Hash:     st1.Hash,
				RawTx:    st1.RawTx,
				Height:   int(st1.Height),
				Index:    int(st1.Index),
				From:     st1.FromAdrs,
				Nonce:    int(st1.Nonce),
				GasPrice: int(st1.GasPrice),
				GasCoin:  st1.GasCoin,
				GasUsed:  int(st1.GasUsed),
				Type:     st1.Type,
				Payload:  st1.Payload,
				Code:     int(st1.Code),
				Log:      st1.Log,
				Data:     dt0,
			})
		case ms.TX_UnbondData: //8
			dt0 := s.Tx8UnbondData{}
			jsonBytes, _ := json.Marshal(st1)
			json.Unmarshal(jsonBytes, &dt0)
			trnM = append(trnM, ms.TransResponse{
				Hash:     st1.Hash,
				RawTx:    st1.RawTx,
				Height:   int(st1.Height),
				Index:    int(st1.Index),
				From:     st1.FromAdrs,
				Nonce:    int(st1.Nonce),
				GasPrice: int(st1.GasPrice),
				GasCoin:  st1.GasCoin,
				GasUsed:  int(st1.GasUsed),
				Type:     st1.Type,
				Payload:  st1.Payload,
				Code:     int(st1.Code),
				Log:      st1.Log,
				Data:     dt0,
			})
		case ms.TX_RedeemCheckData: //9
			dt0 := s.Tx9RedeemCheckData{}
			jsonBytes, _ := json.Marshal(st1)
			json.Unmarshal(jsonBytes, &dt0)
			trnM = append(trnM, ms.TransResponse{
				Hash:     st1.Hash,
				RawTx:    st1.RawTx,
				Height:   int(st1.Height),
				Index:    int(st1.Index),
				From:     st1.FromAdrs,
				Nonce:    int(st1.Nonce),
				GasPrice: int(st1.GasPrice),
				GasCoin:  st1.GasCoin,
				GasUsed:  int(st1.GasUsed),
				Type:     st1.Type,
				Payload:  st1.Payload,
				Code:     int(st1.Code),
				Log:      st1.Log,
				Data:     dt0,
			})
		case ms.TX_SetCandidateOnData: //10
			dt0 := s.Tx10SetCandidateOnData{}
			jsonBytes, _ := json.Marshal(st1)
			json.Unmarshal(jsonBytes, &dt0)
			trnM = append(trnM, ms.TransResponse{
				Hash:     st1.Hash,
				RawTx:    st1.RawTx,
				Height:   int(st1.Height),
				Index:    int(st1.Index),
				From:     st1.FromAdrs,
				Nonce:    int(st1.Nonce),
				GasPrice: int(st1.GasPrice),
				GasCoin:  st1.GasCoin,
				GasUsed:  int(st1.GasUsed),
				Type:     st1.Type,
				Payload:  st1.Payload,
				Code:     int(st1.Code),
				Log:      st1.Log,
				Data:     dt0,
			})
		case ms.TX_SetCandidateOffData: //11
			dt0 := s.Tx11SetCandidateOffData{}
			jsonBytes, _ := json.Marshal(st1)
			json.Unmarshal(jsonBytes, &dt0)
			trnM = append(trnM, ms.TransResponse{
				Hash:     st1.Hash,
				RawTx:    st1.RawTx,
				Height:   int(st1.Height),
				Index:    int(st1.Index),
				From:     st1.FromAdrs,
				Nonce:    int(st1.Nonce),
				GasPrice: int(st1.GasPrice),
				GasCoin:  st1.GasCoin,
				GasUsed:  int(st1.GasUsed),
				Type:     st1.Type,
				Payload:  st1.Payload,
				Code:     int(st1.Code),
				Log:      st1.Log,
				Data:     dt0,
			})
		case ms.TX_CreateMultisigData: //12
			// TODO: Реализовать
			//dt0 := s.Tx12CreateMultisigData{}
			log("WRN", "[sql_trx.go] srchTrxHash(ms.TX_CreateMultisigData) - НЕ РЕАЛИЗОВАН!", "")
		case ms.TX_MultisendData: //13
			dt0 := s.Tx13MultisendData{}
			jsonBytes, _ := json.Marshal(st1)
			json.Unmarshal(jsonBytes, &dt0)
			trnM = append(trnM, ms.TransResponse{
				Hash:     st1.Hash,
				RawTx:    st1.RawTx,
				Height:   int(st1.Height),
				Index:    int(st1.Index),
				From:     st1.FromAdrs,
				Nonce:    int(st1.Nonce),
				GasPrice: int(st1.GasPrice),
				GasCoin:  st1.GasCoin,
				GasUsed:  int(st1.GasUsed),
				Type:     st1.Type,
				Payload:  st1.Payload,
				Code:     int(st1.Code),
				Log:      st1.Log,
				Data:     dt0,
			})
		case ms.TX_EditCandidateData: //14
			// TODO: реализовать
			log("WRN", "[sql_trx.go] srchTrxHash(ms.TX_EditCandidateData) - НЕ РЕАЛИЗОВАН!", "")
		default:
			log("ERR", fmt.Sprint("[sql_trx.go] srchTrxHash(...) - неизвестный статус st1.Type - ", st1.Type), "")
		}
	}

	return trnM
}
