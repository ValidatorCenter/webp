package main

import (
	"encoding/json"
	"fmt"
	"math"
	"strconv"

	"github.com/go-macaron/session"
	"gopkg.in/macaron.v1"

	ms "github.com/ValidatorCenter/minter-go-sdk"
	s "github.com/ValidatorCenter/prs3r/strc"
)

type Trans1 struct {
	Status   string
	Hash     string
	HashMin  string
	Height   int
	Index    int
	From     string
	FromMin  string
	Nonce    int
	GasPrice int
	GasCoin  string
	GasUsed  int
	GasFee   float32
	Type     int
	TypeTxt  string
	Data     interface{}
	Payload  string
	Code     int
	Log      string
	Amount   float32
	//------------------------------------------------
	//Tags     tagKeyValue2 `json:"tags" gorm:"tags"` // TODO: нет необходимости в нём, пока из Покупки/Продажи результат обмена tx.return не вынесут на уровень выше
	TxCoinToBuy  string
	TxCoinToSell string
	TxFrom       string
	TxReturn     float32
	TxSellAmount float32
}

// Преобразование из одной структуры в другую
func TransResp2Trans1(v ms.TransResponse) Trans1 {
	t0 := Trans1{}
	t0.Hash = v.Hash
	t0.HashMin = getMinString(v.Hash)
	t0.Height = v.Height
	t0.Index = v.Index
	t0.From = v.From
	t0.FromMin = getMinString(v.From)
	t0.Nonce = v.Nonce
	t0.GasPrice = v.GasPrice
	t0.GasCoin = v.GasCoin
	t0.GasUsed = v.GasUsed
	t0.GasFee = float32(v.GasUsed) / 1000 // комиссия
	t0.Type = v.Type
	t0.Payload = v.Payload
	t0.Code = v.Code
	t0.Log = v.Log
	if v.Code == 0 {
		t0.Status = "Success"
	} else {
		t0.Status = "Failure" // TODO: можно и лог! v.Log
	}
	t0.TxCoinToBuy = v.Tags.TxCoinToBuy
	t0.TxCoinToSell = v.Tags.TxCoinToSell
	t0.TxFrom = v.Tags.TxFrom
	t0.TxReturn = v.Tags.TxReturn
	t0.TxSellAmount = v.Tags.TxSellAmount

	switch t0.Type {
	case ms.TX_SendData: //1
		t0.TypeTxt = "Send"
		dt0 := s.Tx1SendData{}
		jsonBytes, _ := json.Marshal(v.Data)
		json.Unmarshal(jsonBytes, &dt0)
		t0.Data = dt0
		t0.Amount = dt0.Value
	case ms.TX_SellCoinData: //2
		t0.TypeTxt = "SellCoin"
		dt0 := s.Tx2SellCoinData{}
		jsonBytes, _ := json.Marshal(v.Data)
		json.Unmarshal(jsonBytes, &dt0)
		t0.Data = dt0
		t0.Amount = dt0.ValueToSell
	case ms.TX_SellAllCoinData: //3
		t0.TypeTxt = "SellAllCoin"
		dt0 := s.Tx3SellAllCoinData{}
		jsonBytes, _ := json.Marshal(v.Data)
		json.Unmarshal(jsonBytes, &dt0)
		t0.Data = dt0
		t0.Amount = 0 //dt0....
	case ms.TX_BuyCoinData: //4
		t0.TypeTxt = "BuyCoin"
		dt0 := s.Tx4BuyCoinData{}
		jsonBytes, _ := json.Marshal(v.Data)
		json.Unmarshal(jsonBytes, &dt0)
		t0.Data = dt0
		t0.Amount = dt0.ValueToBuy
	case ms.TX_CreateCoinData: //5
		t0.TypeTxt = "CreateCoin"
		dt0 := s.Tx5CreateCoinData{}
		jsonBytes, _ := json.Marshal(v.Data)
		json.Unmarshal(jsonBytes, &dt0)
		t0.Data = dt0
		t0.Amount = dt0.InitialReserve
	case ms.TX_DeclareCandidacyData: //6
		t0.TypeTxt = "DeclareCandidacy"
		dt0 := s.Tx6DeclareCandidacyData{}
		jsonBytes, _ := json.Marshal(v.Data)
		json.Unmarshal(jsonBytes, &dt0)
		t0.Data = dt0
		t0.Amount = dt0.Stake
	case ms.TX_DelegateDate: //7
		t0.TypeTxt = "Delegate"
		dt0 := s.Tx7DelegateDate{}
		jsonBytes, _ := json.Marshal(v.Data)
		json.Unmarshal(jsonBytes, &dt0)
		t0.Data = dt0
		t0.Amount = dt0.Stake
	case ms.TX_UnbondData: //8
		t0.TypeTxt = "Unbond"
		dt0 := s.Tx8UnbondData{}
		jsonBytes, _ := json.Marshal(v.Data)
		json.Unmarshal(jsonBytes, &dt0)
		t0.Data = dt0
		t0.Amount = dt0.Value
	case ms.TX_RedeemCheckData: //9
		t0.TypeTxt = "RedeemCheck"
		dt0 := s.Tx9RedeemCheckData{}
		jsonBytes, _ := json.Marshal(v.Data)
		json.Unmarshal(jsonBytes, &dt0)
		t0.Data = dt0
		t0.Amount = 0 //dt0...
	case ms.TX_SetCandidateOnData: //10
		t0.TypeTxt = "SetCandidateOn"
		dt0 := s.Tx10SetCandidateOnData{}
		jsonBytes, _ := json.Marshal(v.Data)
		json.Unmarshal(jsonBytes, &dt0)
		t0.Data = dt0
		t0.Amount = 0 //dt0...
	case ms.TX_SetCandidateOffData: //11
		t0.TypeTxt = "SetCandidateOff"
		dt0 := s.Tx11SetCandidateOffData{}
		jsonBytes, _ := json.Marshal(v.Data)
		json.Unmarshal(jsonBytes, &dt0)
		t0.Data = dt0
		t0.Amount = 0 //dt0...
	case ms.TX_CreateMultisigData: //12
		t0.TypeTxt = "CreateMultisig"
		dt0 := s.Tx12CreateMultisigData{}
		jsonBytes, _ := json.Marshal(v.Data)
		json.Unmarshal(jsonBytes, &dt0)
		t0.Data = dt0
		t0.Amount = 0 //dt0...
	case ms.TX_MultisendData: //13
		// FIXME: Нет кода и идей по хранению в базе DB!!! Исправить это!
		t0.TypeTxt = "Multisend"
		dt0 := s.Tx13MultisendData{}
		jsonBytes, _ := json.Marshal(v.Data)
		json.Unmarshal(jsonBytes, &dt0)
		t0.Data = dt0
		for iLs, _ := range dt0.List {
			t0.Amount += dt0.List[iLs].Value
		}
	case ms.TX_EditCandidateData: //14
		// TODO: Реализовать
	}
	return t0
}

// СТРАНИЦА: список транзакций по убыванию
func hndTransactionsInfo(ctx *macaron.Context, sess session.Store) {
	var alertMsg, alertType, alertAct string
	var auth bool = false
	var usrName, idUsr string
	var err error

	// SESSION
	if sess.Get("login") != nil {
		auth = true
		idUsr = sess.Get("login").(string)
		usrName = fmt.Sprintf("%s...%s", idUsr[:6], idUsr[len(idUsr)-4:len(idUsr)])
	}

	// Пагинация
	pageNmbrNow := 0
	if ctx.Params(":pgn") != "" {
		pageNmbrNowTxt := ctx.Params(":pgn")
		pageNmbrNow, err = strconv.Atoi(pageNmbrNowTxt)
		if err != nil {
			pageNmbrNow = 0
		}
	}
	if pageNmbrNow < 0 {
		pageNmbrNow = 0
	}

	totalBlock := srchTrxAmnt(dbSQL) // количество транзакций уже загружено в БД

	skipTrns := pageNmbrNow * 50

	// Расчет кнопок навигации
	BtnL := 0
	if pageNmbrNow != 0 {
		BtnL = pageNmbrNow - 1
	}
	BtnR := 0
	BtnRR := 0
	BtnR = pageNmbrNow + 1
	BtnRR = int(math.Ceil(float64(totalBlock)/50) - 1)
	if pageNmbrNow == BtnRR {
		BtnR = 0
		BtnRR = 0
	}

	trnM := srchTrxList(dbSQL, uint32(skipTrns)) // получаем из БД список транзакций по заданной странице из 50 строк

	// Перерабатываем данные полученные из БД по транзакциям
	trnM1 := []Trans1{}
	for _, v := range trnM {
		// перерабатываем
		t0 := TransResp2Trans1(v)
		trnM1 = append(trnM1, t0)
	}

	// Последний синхронизированный блок
	ResultNetwork, _ := sdk.GetStatus()
	statusMDB := srchSysSql(dbSys)

	// Инф. о синхронизации БД с БлокЧейном:
	ctx.Data["LastSync"] = statusMDB.LatestBlockSave
	ctx.Data["Current"] = ResultNetwork.LatestBlockHeight
	if sdk.ChainMainnet {
		ctx.Data["ChainNet"] = "mainnet"
	} else {
		ctx.Data["ChainNet"] = "testnet"
	}

	// Заголовок страницы:
	ctx.Data["Title"] = "Transaction"

	ctx.Data["AllTrns"] = trnM1
	ctx.Data["TotalBlock"] = totalBlock
	ctx.Data["CoinMinter"] = CoinMinter // Базовая монета систм.

	// Кнопки навигации по страницам:
	ctx.Data["BtnNow"] = pageNmbrNow
	ctx.Data["BtnL"] = BtnL
	ctx.Data["BtnR"] = BtnR
	ctx.Data["BtnRR"] = BtnRR

	// Пользователь:
	ctx.Data["UsrAuth"] = auth
	ctx.Data["UsrName"] = usrName
	ctx.Data["UsrAddress"] = idUsr

	// Инф.сообщения от системы:
	ctx.Data["AlertAct"] = alertAct
	ctx.Data["AlertMsg"] = alertMsg
	ctx.Data["AlertType"] = alertType

	// Вывод страницы:
	ctx.HTML(200, "transactions_info")
}

// Получаем данные по 1(одной) транзакции для визуализации
func GetTransactionOneInfo(hashTrans string) Trans1 {
	// Получение транзакции с БД
	v := srchTrxHash(dbSQL, hashTrans)

	//t0 := Trans1{}

	// проверка что транзакция найдена!!!
	if v.Hash == "" {
		//return t0
		return Trans1{}
	}

	// перерабатываем и возвращаем
	//t0 = TransResp2Trans1(v)
	return TransResp2Trans1(v)
}

// СТРАНИЦА: одна транзакция
func hndTransactionOneInfo(ctx *macaron.Context, sess session.Store) {
	var alertMsg, alertType, alertAct string
	var auth bool = false
	var usrName, idUsr string

	// SESSION
	if sess.Get("login") != nil {
		auth = true
		idUsr = sess.Get("login").(string)
		usrName = fmt.Sprintf("%s...%s", idUsr[:6], idUsr[len(idUsr)-4:len(idUsr)])
	}

	hashTransMt := ctx.Params(":number")

	t0 := GetTransactionOneInfo(hashTransMt)

	// проверка что транзакция найдена!!! иначе на страницу 404
	if t0.Hash == "" {
		page404(ctx)
		return
	}

	// Последний синхронизированный блок
	ResultNetwork, _ := sdk.GetStatus()
	statusMDB := srchSysSql(dbSys)

	// Инф. о синхронизации БД с БлокЧейном:
	ctx.Data["LastSync"] = statusMDB.LatestBlockSave
	ctx.Data["Current"] = ResultNetwork.LatestBlockHeight
	if sdk.ChainMainnet {
		ctx.Data["ChainNet"] = "mainnet"
	} else {
		ctx.Data["ChainNet"] = "testnet"
	}

	// Заголовок страницы:
	ctx.Data["Title"] = "Transaction"

	ctx.Data["OneTrns"] = t0
	ctx.Data["CoinMinter"] = CoinMinter // Базовая монета систм.

	// Пользователь:
	ctx.Data["UsrAuth"] = auth
	ctx.Data["UsrName"] = usrName
	ctx.Data["UsrAddress"] = idUsr

	// Инф.сообщения от системы:
	ctx.Data["AlertAct"] = alertAct
	ctx.Data["AlertMsg"] = alertMsg
	ctx.Data["AlertType"] = alertType

	// Вывод страницы:
	ctx.HTML(200, "transaction_info")
}

// Структура возврата транзакции в виде JSON
type RetJSONTrx struct {
	Status bool   `json:"status"`
	Trx    Trans1 `json:"transaction"`
	ErrMsg string `json:"err_msg"`
}

// API: одна транзакция
func hndAPITransactionOneInfo(ctx *macaron.Context, sess session.Store) {
	retDt := RetJSONTrx{}

	retDt.Status = true // исполнен без ошибок
	retDt.ErrMsg = ""   // нет ошибок

	hashTransMt := ctx.Params(":number")

	t0 := GetTransactionOneInfo(hashTransMt)

	// проверка что транзакция найдена!!! иначе на страницу 404
	if t0.Hash == "" {
		retDt.Status = false                   // исполнен с ошибкой
		retDt.ErrMsg = "No search transaction" // текст ошибки
	}
	retDt.Trx = t0

	// возврат JSON данных
	ctx.JSON(200, &retDt)
}
