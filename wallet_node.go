package main

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/go-macaron/session"
	"gopkg.in/macaron.v1"

	ms "github.com/ValidatorCenter/minter-go-sdk"
	s "github.com/ValidatorCenter/prs3r/strc"
)

// Функция декларирования мастерноды
func DeclCandidacy(sess session.Store, validator string, comm uint, taxCoin string, valueStart int64, feeCoin string) (string, error) {
	sdk.AccPrivateKey = sess.Get("priv_k").(string)
	//sdk.AccAddress,_ = ms.GetAddressPrivateKey(sdk.AccPrivateKey) // другой вариант ниже
	sdk.AccAddress = sess.Get("login").(string)

	declDt := ms.TxDeclareCandidacyData{
		PubKey:     validator,
		Commission: comm,
		Coin:       taxCoin,
		Stake:      valueStart,
		// Gas
		GasCoin:  feeCoin,
		GasPrice: 1,
	}

	resHash, err := sdk.TxDeclareCandidacy(&declDt)
	if err != nil {
		return "", err
	}
	return resHash, nil
}

// Функция вкл/выкл мастерноды
func PowerCandidacy(sess session.Store, validator string, onoff bool, feeCoin string) (string, error) {
	sdk.AccPrivateKey = sess.Get("priv_k").(string)
	//sdk.AccAddress,_ = ms.GetAddressPrivateKey(sdk.AccPrivateKey) // другой вариант ниже
	sdk.AccAddress = sess.Get("login").(string)

	sndDt := ms.TxSetCandidateData{
		PubKey:   validator,
		Activate: onoff, //true-"on", false-"off"
		// Gas
		GasCoin:  feeCoin,
		GasPrice: 1,
	}

	resHash, err := sdk.TxSetCandidate(&sndDt)
	if err != nil {
		return "", err
	}
	return resHash, nil
}

// СТРАНИЦА: управления Мастернодой
func hndWalletTxMasternode(ctx *macaron.Context, sess session.Store) {
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

	inputPubKey := ctx.Req.PostFormValue("inputPubKey")
	//inputMsg := ctx.Req.PostFormValue("inputMsg")
	inputFeeCoin := ctx.Req.PostFormValue("inputFeeCoin")
	typeAct := ctx.Req.PostFormValue("typeAct")

	if typeAct != "" {
		if typeAct == "DECLARE" {
			// FIXME: Открыть в SDK -> inputAddrs для DeclCandidacy
			inputAmnt := ctx.Req.PostFormValue("inputAmnt")
			inputCommiss := ctx.Req.PostFormValue("inputCommiss")
			inputCoin := ctx.Req.PostFormValue("inputCoin")
			goodStep := true
			inputAmnt_i64, err := strconv.Atoi(inputAmnt)
			if err != nil {
				alertAct = "Declaration"
				alertMsg = fmt.Sprintf("%s", err)
				alertType = "danger" //danger warning
				goodStep = false
			}
			inputCommiss_ui, err := strconv.Atoi(inputCommiss)
			if err != nil {
				alertAct = "Declaration"
				alertMsg = fmt.Sprintf("%s", err)
				alertType = "danger" //danger warning
				goodStep = false
			}

			if goodStep {
				fmt.Println("DECLARE", inputPubKey, inputCommiss)
				hashTr, err := DeclCandidacy(sess, inputPubKey, uint(inputCommiss_ui), inputCoin, int64(inputAmnt_i64), inputFeeCoin)
				if err != nil {
					alertAct = "Declaration"
					alertMsg = fmt.Sprintf("%s-%s", err, hashTr)
					alertType = "danger" //danger warning
				} else {
					fmt.Println("....DECLARE:", hashTr)
					alertAct = "Declaration"
					alertMsg = hashTr
					alertType = "success" //danger warning
				}
			}

		} else if typeAct == "START" {
			fmt.Println("NODE-START", inputPubKey)
			hashTr, err := PowerCandidacy(sess, inputPubKey, true, inputFeeCoin)
			if err != nil {
				alertAct = "Masternode"
				alertMsg = fmt.Sprintf("%s-%s", err, hashTr)
				alertType = "danger" //danger warning
			} else {
				fmt.Println("....NODE-START:", hashTr)
				alertAct = "Masternode"
				alertMsg = hashTr
				alertType = "success" //danger warning
			}
		} else if typeAct == "STOP" {
			fmt.Println("NODE-STOP", inputPubKey)
			hashTr, err := PowerCandidacy(sess, inputPubKey, false, inputFeeCoin)
			if err != nil {
				alertAct = "Masternode"
				alertMsg = fmt.Sprintf("%s-%s", err, hashTr)
				alertType = "danger" //danger warning
			} else {
				fmt.Println("....NODE-STOP:", hashTr)
				alertAct = "Masternode"
				alertMsg = hashTr
				alertType = "success" //danger warning
			}
		}
	}

	// Список нод пользователя (с кнопками быстрого управления и статусом)
	nodeM := []s.NodeExt{}
	nodeM = srchNodeAddress(dbSQL, idUsr)

	for iM, _ := range nodeM {
		for _, dNB := range nodeM[iM].Blocks {
			if dNB.Type == "AbsentBlock" {
				nodeM[iM].AmnNoBlocks++
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
	ctx.Data["MyNodes"] = nodeM

	ctx.Data["Title"] = "Masternode"
	ctx.Data["UsrAuth"] = auth
	ctx.Data["UsrName"] = usrName
	ctx.Data["UsrAddress"] = idUsr
	ctx.Data["CoinMinter"] = CoinMinter
	ctx.Data["AlertAct"] = alertAct
	ctx.Data["AlertMsg"] = alertMsg
	ctx.Data["AlertType"] = alertType
	ctx.HTML(200, "tx_mnode")
}
