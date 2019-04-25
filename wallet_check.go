package main

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/go-macaron/session"
	"gopkg.in/macaron.v1"

	ms "github.com/ValidatorCenter/minter-go-sdk"
)

// Функция создания чека
func CreateCheck(sess session.Store, coin string, value float32, pswrd string, nonce uint64) (string, error) {
	sdk.AccAddress = sess.Get("login").(string)
	sdk.AccPrivateKey = sess.Get("priv_k").(string)

	chDt := ms.TxCreateCkeckData{
		Coin:     coin,
		Stake:    value,
		Password: pswrd,
		Nonce:    nonce,
	}
	resCheck, err := sdk.TxCreateCheck(&chDt)
	if err != nil {
		return "", err
	}
	return resCheck, nil
}

// Функция обналичивания чека
func RedeemCheck(sess session.Store, check string, pswrd string) (string, error) {
	sdk.AccAddress = sess.Get("login").(string)
	sdk.AccPrivateKey = sess.Get("priv_k").(string)

	rchDt := ms.TxRedeemCheckData{
		Check:    check,
		Password: pswrd,
		GasCoin:  ms.GetBaseCoin(),
		GasPrice: 1,
	}

	resHash, err := sdk.TxRedeemCheck(&rchDt)
	if err != nil {
		return "", err
	}
	return resHash, nil
}

// СТРАНИЦА: управления Чеками
func hndWalletTxChecks(ctx *macaron.Context, sess session.Store) {
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

	//...............................................................
	newCheckData := ""
	// если был POST!
	ctx.Req.ParseForm()
	ctx.Resp.WriteHeader(http.StatusOK)

	typeAct := ctx.Req.PostFormValue("typeAct")

	if typeAct == "REDEEM" {
		inputCheck := ctx.Req.PostFormValue("inputCheck")
		inputPswrd := ctx.Req.PostFormValue("inputPswrd")

		resHash, err := RedeemCheck(sess, inputCheck, inputPswrd)
		if err != nil {
			alertAct = "Redeem check"
			alertMsg = fmt.Sprintf("%s-%s", err, resHash)
			alertType = "danger" //danger warning
		} else {
			alertAct = "Redeem check"
			alertMsg = resHash
			alertType = "success" //danger warning
		}
	} else if typeAct == "NEWCHECK" {
		inputNonce := ctx.Req.PostFormValue("inputNonce")
		inputAmnt := ctx.Req.PostFormValue("inputAmnt")
		inputCoin := ctx.Req.PostFormValue("inputCoin")
		inputPswrd := ctx.Req.PostFormValue("inputPswrd")
		// TODO: в SDK реализовать!
		//inputLiveBlock := ctx.Req.PostFormValue("inputLiveBlock")

		inputAmnt_f32, err := strconv.ParseFloat(inputAmnt, 32)
		if err != nil {
			alertAct = "New check(a)"
			alertMsg = fmt.Sprintf("%s", err)
			alertType = "danger" //danger warning
		} else {
			inputNonce_i, err := strconv.Atoi(inputNonce)
			if err != nil {
				alertAct = "New check(i)"
				alertMsg = fmt.Sprintf("%s", err)
				alertType = "danger" //danger warning
			} else {
				fmt.Println(inputNonce_i)
				resCheck, err := CreateCheck(sess, inputCoin, float32(inputAmnt_f32), inputPswrd, uint64(inputNonce_i))
				if err != nil {
					alertAct = "New check"
					alertMsg = fmt.Sprintf("%s-%s", err, resCheck)
					alertType = "danger" //danger warning
				} else {
					alertAct = "New check"
					alertMsg = resCheck
					alertType = "success" //danger warning
					newCheckData = resCheck
				}
			}
		}
	}
	//...............................................................

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

	ctx.Data["NewCheckData"] = newCheckData

	ctx.Data["Title"] = "Checks"
	ctx.Data["UsrAuth"] = auth
	ctx.Data["UsrName"] = usrName
	ctx.Data["UsrAddress"] = idUsr
	ctx.Data["AlertAct"] = alertAct
	ctx.Data["AlertMsg"] = alertMsg
	ctx.Data["AlertType"] = alertType
	ctx.HTML(200, "tx_checks")
}
