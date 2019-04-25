package main

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/go-macaron/session"
	"gopkg.in/macaron.v1"

	ms "github.com/ValidatorCenter/minter-go-sdk"
)

// Функция создание монет
func CreateCoin(sess session.Store, nameCoin string, tickerCoin string, initAmnt int64, initResrv int64, crr uint, feeCoin string) (string, error) {
	sdk.AccPrivateKey = sess.Get("priv_k").(string)
	//sdk.AccAddress,_ = ms.GetAddressPrivateKey(sdk.AccPrivateKey) // другой вариант ниже
	sdk.AccAddress = sess.Get("login").(string)

	creatDt := ms.TxCreateCoinData{
		Name:                 nameCoin,
		Symbol:               tickerCoin,
		InitialAmount:        initAmnt,
		InitialReserve:       initResrv,
		ConstantReserveRatio: crr,
		// Gas
		GasCoin:  feeCoin,
		GasPrice: 1,
	}

	resHash, err := sdk.TxCreateCoin(&creatDt)
	if err != nil {
		return "", err
	}
	return resHash, nil
}

// СТРАНИЦА: создания монет
func hndWalletTxCoiner(ctx *macaron.Context, sess session.Store) {
	var alertMsg, alertType, alertAct string
	var auth bool = false
	var usrName, idUsr string

	if sess.Get("login") != nil {
		auth = true
		idUsr = sess.Get("login").(string)
		usrName = fmt.Sprintf("%s...%s", idUsr[:6], idUsr[len(idUsr)-4:len(idUsr)])
	} else {
		// Редирект на главную или страницу авторизации
		ctx.Redirect("/coins")
		return
	}

	// если был POST!
	ctx.Req.ParseForm()
	ctx.Resp.WriteHeader(http.StatusOK)

	inputCoinName := ctx.Req.PostFormValue("inputCoinName")
	inputTicker := ctx.Req.PostFormValue("inputTicker")
	inputInitAmnt := ctx.Req.PostFormValue("inputInitAmnt")
	inputInitResrv := ctx.Req.PostFormValue("inputInitResrv")
	inputCRR := ctx.Req.PostFormValue("inputCRR")
	//inputMsg := ctx.Req.PostFormValue("inputMsg")
	inputFeeCoin := ctx.Req.PostFormValue("inputFeeCoin")
	typeAct := ctx.Req.PostFormValue("typeAct")

	if typeAct != "" {
		goodStep := true
		inputInitAmnt_i64, err := strconv.ParseFloat(inputInitAmnt, 32)
		if err != nil {
			alertAct = "Coiner"
			alertMsg = fmt.Sprintf("%s", err)
			alertType = "danger" //danger warning
			goodStep = false
		}
		inputInitResrv_i64, err := strconv.ParseFloat(inputInitResrv, 32)
		if err != nil {
			alertAct = "Coiner"
			alertMsg = fmt.Sprintf("%s", err)
			alertType = "danger" //danger warning
			goodStep = false
		}
		inputCRR_ui, err := strconv.Atoi(inputCRR)
		if err != nil {
			alertAct = "Coiner"
			alertMsg = fmt.Sprintf("%s", err)
			alertType = "danger" //danger warning
			goodStep = false
		}

		if goodStep {
			fmt.Println("COINER", inputTicker, inputCRR_ui)
			hashTr, err := CreateCoin(sess, inputCoinName, inputTicker, int64(inputInitAmnt_i64), int64(inputInitResrv_i64), uint(inputCRR_ui), inputFeeCoin)
			if err != nil {
				alertAct = "Coiner"
				alertMsg = fmt.Sprintf("%s-%s", err, hashTr)
				alertType = "danger" //danger warning
			} else {
				fmt.Println("....COINER:", hashTr)
				alertAct = "Coiner"
				alertMsg = hashTr
				alertType = "success" //danger warning
			}
		}
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

	// получаем список
	coinsAddrs, totalAmnt, _ := sdk.GetAddress(idUsr)
	ctx.Data["AllCoins"] = coinsAddrs
	ctx.Data["TotalAmntInBaseCoin"] = totalAmnt

	ctx.Data["Title"] = "Coiner"
	ctx.Data["UsrAuth"] = auth
	ctx.Data["UsrName"] = usrName
	ctx.Data["UsrAddress"] = idUsr
	ctx.Data["AlertAct"] = alertAct
	ctx.Data["AlertMsg"] = alertMsg
	ctx.Data["AlertType"] = alertType
	ctx.HTML(200, "tx_coiner")
}
