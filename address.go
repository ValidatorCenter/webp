package main

import (
	"encoding/json"
	"fmt"
	"math"
	"strconv"

	"github.com/go-macaron/session"
	"gopkg.in/macaron.v1"

	s "github.com/ValidatorCenter/prs3r/strc"
)

// Страница с информацией об одном адресе
func hndAddressOneInfo(ctx *macaron.Context, sess session.Store) {
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

	// Пагинаципя - определение страницы списка показа транзакций Адреса кошелька
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
	skipTrns := pageNmbrNow * 50

	nmbrAddrs := ctx.Params(":number")
	totalBlock := srchTrxAddrsAmnt(dbSQL, nmbrAddrs) // Определение сколько всего транзакций данного Адреса

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

	//..........................................................................
	trnM := srchTrxAddrsSql(dbSQL, nmbrAddrs, skipTrns) // получаем из БД список транзакций Адреса по заданной странице из 50 строк

	// перерабатываем
	trnM1 := []Trans1{}
	for _, v := range trnM {
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

		// Используется в Trx продажи и покупке монет
		t0.TxReturn = v.Tags.TxReturn
		t0.TxSellAmount = v.Tags.TxSellAmount

		switch t0.Type {
		case 1:
			t0.TypeTxt = "Send"
			dt0 := s.Tx1SendData{}
			jsonBytes, _ := json.Marshal(v.Data)
			json.Unmarshal(jsonBytes, &dt0)
			t0.Data = dt0
			t0.Amount = dt0.Value
		case 2:
			t0.TypeTxt = "SellCoin"
			dt0 := s.Tx2SellCoinData{}
			jsonBytes, _ := json.Marshal(v.Data)
			json.Unmarshal(jsonBytes, &dt0)
			t0.Data = dt0
			t0.Amount = dt0.ValueToSell
		case 3:
			t0.TypeTxt = "SellAllCoin"
			dt0 := s.Tx3SellAllCoinData{}
			jsonBytes, _ := json.Marshal(v.Data)
			json.Unmarshal(jsonBytes, &dt0)
			t0.Data = dt0
			t0.Amount = 0 //dt0....
		case 4:
			t0.TypeTxt = "BuyCoin"
			dt0 := s.Tx4BuyCoinData{}
			jsonBytes, _ := json.Marshal(v.Data)
			json.Unmarshal(jsonBytes, &dt0)
			t0.Data = dt0
			t0.Amount = dt0.ValueToBuy
		case 5:
			t0.TypeTxt = "CreateCoin"
			dt0 := s.Tx5CreateCoinData{}
			jsonBytes, _ := json.Marshal(v.Data)
			json.Unmarshal(jsonBytes, &dt0)
			t0.Data = dt0
			t0.Amount = dt0.InitialReserve
		case 6:
			t0.TypeTxt = "DeclareCandidacy"
			dt0 := s.Tx6DeclareCandidacyData{}
			jsonBytes, _ := json.Marshal(v.Data)
			json.Unmarshal(jsonBytes, &dt0)
			t0.Data = dt0
			t0.Amount = dt0.Stake
		case 7:
			t0.TypeTxt = "Delegate"
			dt0 := s.Tx7DelegateDate{}
			jsonBytes, _ := json.Marshal(v.Data)
			json.Unmarshal(jsonBytes, &dt0)
			t0.Data = dt0
			t0.Amount = dt0.Stake
		case 8:
			t0.TypeTxt = "Unbond"
			dt0 := s.Tx8UnbondData{}
			jsonBytes, _ := json.Marshal(v.Data)
			json.Unmarshal(jsonBytes, &dt0)
			t0.Data = dt0
			t0.Amount = dt0.Value
		case 9:
			t0.TypeTxt = "RedeemCheck"
			dt0 := s.Tx9RedeemCheckData{}
			jsonBytes, _ := json.Marshal(v.Data)
			json.Unmarshal(jsonBytes, &dt0)
			t0.Data = dt0
			t0.Amount = 0 //dt0...
		case 10:
			t0.TypeTxt = "SetCandidateOn"
			dt0 := s.Tx10SetCandidateOnData{}
			jsonBytes, _ := json.Marshal(v.Data)
			json.Unmarshal(jsonBytes, &dt0)
			t0.Data = dt0
			t0.Amount = 0 //dt0...
		case 11:
			t0.TypeTxt = "SetCandidateOff"
			dt0 := s.Tx11SetCandidateOffData{}
			jsonBytes, _ := json.Marshal(v.Data)
			json.Unmarshal(jsonBytes, &dt0)
			t0.Data = dt0
			t0.Amount = 0 //dt0...
		case 12:
			t0.TypeTxt = "CreateMultisig"
			dt0 := s.Tx12CreateMultisigData{}
			jsonBytes, _ := json.Marshal(v.Data)
			json.Unmarshal(jsonBytes, &dt0)
			t0.Data = dt0
			t0.Amount = 0 //dt0...
		}

		trnM1 = append(trnM1, t0)
	}

	//Список монет и общая сумма в базовой монете
	coinsAddrs, totalAmnt, _ := sdk.GetAddress(nmbrAddrs)

	totalReward := float32(0.0)
	totalDeleg := float32(0.0)
	//разбор что и куда делегировал
	//1 - в какого валидатора сколько делегировал
	//nodesDeleg := []s.NodeExt{} // УДАЛИТЬ!
	allDeleg := []Delegate{}

	allDeleg = srchNodeAddrsSql(dbSQL, nmbrAddrs)
	for iS, _ := range allDeleg {
		totalDeleg += allDeleg[iS].ValueBip
	}

	//2 - прибыль от какого валидатора и сколько
	rE := srchRewardAddrsSql(dbSQL, nmbrAddrs)

	for _, vReStake := range rE {
		totalReward += vReStake.Amnt
	}

	//..........................................................................

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
	ctx.Data["Title"] = "Address"

	ctx.Data["AllCoins"] = coinsAddrs
	ctx.Data["TotalAmntInBaseCoin"] = totalAmnt

	ctx.Data["AllTrns"] = trnM1
	ctx.Data["AllDeleg"] = allDeleg
	ctx.Data["AllReward"] = rE
	ctx.Data["TotalBlock"] = totalBlock
	ctx.Data["Address"] = nmbrAddrs

	ctx.Data["TotalReward"] = totalReward
	ctx.Data["TotalDelegate"] = totalDeleg

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
	ctx.HTML(200, "address_info")
}
