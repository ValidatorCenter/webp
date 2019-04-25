package main

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/go-macaron/session"
	"gopkg.in/macaron.v1"

	ms "github.com/ValidatorCenter/minter-go-sdk"
)

// Функция отправки монет
func SendCoin(sess session.Store, toAddrs string, sndCoin string, feeCoin string, valueBuy float32) (string, error) {
	sdk.AccPrivateKey = sess.Get("priv_k").(string)
	//sdk.AccAddress,_ = ms.GetAddressPrivateKey(sdk.AccPrivateKey) // другой вариант ниже
	sdk.AccAddress = sess.Get("login").(string)

	sndDt := ms.TxSendCoinData{
		Coin:      sndCoin,
		ToAddress: toAddrs,
		Value:     valueBuy,
		// Gas
		GasCoin:  feeCoin,
		GasPrice: 1,
	}

	resHash, err := sdk.TxSendCoin(&sndDt)
	if err != nil {
		return "", err
	}
	return resHash, nil
}

// СТРАНИЦА: отправки монет из личного кабинета
func hndWalletTxSend(ctx *macaron.Context, sess session.Store) {
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

	inputAddress := ctx.Req.PostFormValue("inputAddress")
	inputAmnt := ctx.Req.PostFormValue("inputAmnt")
	inputCoin := ctx.Req.PostFormValue("inputCoin")
	//inputMsg := ctx.Req.PostFormValue("inputMsg")
	inputFeeCoin := ctx.Req.PostFormValue("inputFeeCoin")
	typeAct := ctx.Req.PostFormValue("typeAct")

	if typeAct != "" {

		inputAmnt_f32, err := strconv.ParseFloat(inputAmnt, 32)
		if err != nil {
			alertAct = "Send"
			alertMsg = fmt.Sprintf("%s", err)
			alertType = "danger" //danger warning
		} else {
			fmt.Println("SEND", inputAddress, inputCoin, inputAmnt_f32)
			hashTr, err := SendCoin(sess, inputAddress, inputCoin, inputFeeCoin, float32(inputAmnt_f32))
			if err != nil {
				alertAct = "Send"
				alertMsg = fmt.Sprintf("%s-%s", err, hashTr)
				alertType = "danger" //danger warning
			} else {
				fmt.Println("....SEND:", hashTr)
				alertAct = "Send"
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

	ctx.Data["Title"] = "SendCoin"
	ctx.Data["UsrAuth"] = auth
	ctx.Data["UsrName"] = usrName
	ctx.Data["UsrAddress"] = idUsr
	ctx.Data["AlertAct"] = alertAct
	ctx.Data["AlertMsg"] = alertMsg
	ctx.Data["AlertType"] = alertType
	ctx.HTML(200, "tx_send")
}
