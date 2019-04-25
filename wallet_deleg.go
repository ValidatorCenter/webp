package main

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/go-macaron/session"
	"gopkg.in/macaron.v1"

	ms "github.com/ValidatorCenter/minter-go-sdk"
	s "github.com/ValidatorCenter/prs3r/strc"
)

// Функция делегирования
func Delegat(sess session.Store, delegCoin string, validator string, valueDeleg float32, feeCoin string) (string, error) {
	sdk.AccPrivateKey = sess.Get("priv_k").(string)
	//sdk.AccAddress,_ = ms.GetAddressPrivateKey(sdk.AccPrivateKey) // другой вариант ниже
	sdk.AccAddress = sess.Get("login").(string)

	delegDt := ms.TxDelegateData{
		Coin:   delegCoin,
		PubKey: validator,
		Stake:  valueDeleg,
		// Gas
		GasCoin:  feeCoin,
		GasPrice: 1,
	}

	resHash, err := sdk.TxDelegate(&delegDt)
	if err != nil {
		return "", err
	}
	return resHash, nil
}

// Функция отзыва делегированных
func UnDelegat(sess session.Store, delegCoin string, validator string, valueDeleg int64, feeCoin string) (string, error) {
	sdk.AccPrivateKey = sess.Get("priv_k").(string)
	//sdk.AccAddress,_ = ms.GetAddressPrivateKey(sdk.AccPrivateKey) // другой вариант ниже
	sdk.AccAddress = sess.Get("login").(string)

	unbDt := ms.TxUnbondData{
		Coin:   delegCoin,
		PubKey: validator,
		Value:  valueDeleg,
		// Gas
		GasCoin:  feeCoin,
		GasPrice: 1,
	}

	resHash, err := sdk.TxUnbond(&unbDt)
	if err != nil {
		return "", err
	}
	return resHash, nil
}

// Добавление нового правила делегирования
func AddAutodeleg(xPubKey, xAddress, xCoin string, xAmntPrc int) error {
	var err error

	vN := []s.AutodelegCfg{}
	vN = srchAutodeleg(dbSQL, xPubKey, xAddress, xCoin)

	if len(vN) > 0 {
		return errors.New("Dubl")
	} else {
		// добавляем
		newNUX := s.AutodelegCfg{}
		newNUX.Address = xAddress
		newNUX.PubKey = xPubKey
		newNUX.Coin = xCoin
		newNUX.WalletPrc = xAmntPrc

		err = addAutodelegSql(dbSQL, newNUX)
		if err != nil {
			return err
		}
	}
	return nil
}

// Сохранение изменений правила делегирования
func SavAutodeleg(xPubKey, xAddress, xCoin string, xAmntPrc int) error {
	var err error

	vN := []s.AutodelegCfg{}
	vN = srchAutodeleg(dbSQL, xPubKey, xAddress, xCoin)

	if len(vN) > 0 {
		// Изменение
		err = updAutodelegSql(dbSQL, s.AutodelegCfg{
			// Ищем:
			PubKey:  xPubKey,
			Address: xAddress,
			Coin:    xCoin,
			// Обновляем:
			WalletPrc: xAmntPrc,
		})
		if err != nil {
			return err
		}
	} else {
		return errors.New("NoData")
	}
	return nil
}

// Удаление правила делегирования
func DelAutodeleg(xPubKey, xAddress, xCoin string) error {
	var err error

	vN := []s.AutodelegCfg{}
	vN = srchAutodeleg(dbSQL, xPubKey, xAddress, xCoin)

	if len(vN) > 0 {
		// Изменение
		err = delAutodelegSql(dbSQL, xPubKey, xAddress, xCoin)
		if err != nil {
			return err
		}
	} else {
		return errors.New("NoData")
	}
	return nil
}

// СТРАНИЦА: делегирования монет из личного кабинета
func hndWalletTxDelegation(ctx *macaron.Context, sess session.Store) {
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
	inputAmnt := ctx.Req.PostFormValue("inputAmnt")
	inputCoin := ctx.Req.PostFormValue("inputCoin")
	//inputMsg := ctx.Req.PostFormValue("inputMsg")
	inputFeeCoin := ctx.Req.PostFormValue("inputFeeCoin")
	typeAct := ctx.Req.PostFormValue("typeAct")
	inputPrcDeleg := ctx.Req.PostFormValue("inputPrcDeleg")

	if typeAct != "" {
		if typeAct == "DELEG" {
			inputAmnt_f32, err := strconv.ParseFloat(inputAmnt, 32)
			if err != nil {
				alertAct = "Delegation"
				alertMsg = fmt.Sprintf("%s", err)
				alertType = "danger" //danger warning
			} else {
				fmt.Println("DELEG", inputPubKey, inputAmnt, inputCoin)
				hashTr, err := Delegat(sess, inputCoin, inputPubKey, float32(inputAmnt_f32), inputFeeCoin)
				if err != nil {
					alertAct = "Delegation"
					alertMsg = fmt.Sprintf("%s-%s", err, hashTr)
					alertType = "danger" //danger warning
				} else {
					fmt.Println("....DELEG:", hashTr)
					alertAct = "Delegation"
					alertMsg = hashTr
					alertType = "success" //danger warning
				}
			}

		} else if typeAct == "UNDELEG" {
			inputAmnt_i, err := strconv.Atoi(inputAmnt)
			if err != nil {
				alertAct = "UnDelegation"
				alertMsg = fmt.Sprintf("%s", err)
				alertType = "danger" //danger warning
			} else {
				fmt.Println("UNDELEG", inputPubKey, inputAmnt, inputCoin)
				hashTr, err := UnDelegat(sess, inputCoin, inputPubKey, int64(inputAmnt_i), inputFeeCoin)
				if err != nil {
					alertAct = "UnDelegation"
					alertMsg = fmt.Sprintf("%s-%s", err, hashTr)
					alertType = "danger" //danger warning
				} else {
					fmt.Println("....UNDELEG:", hashTr)
					alertAct = "UnDelegation"
					alertMsg = hashTr
					alertType = "success" //danger warning
				}
			}
		} else if typeAct == "ADELEG-ADD" {
			inputPrcDeleg_i, err := strconv.Atoi(inputPrcDeleg)
			if err != nil {
				alertAct = "AutoDelegation add"
				alertMsg = fmt.Sprintf("%s", err)
				alertType = "danger" //danger warning
			} else {
				err = AddAutodeleg(inputPubKey, idUsr, inputCoin, inputPrcDeleg_i)
				if err != nil {
					alertAct = "AutoDelegation add"
					alertMsg = fmt.Sprintf("%s", err)
					alertType = "danger" //danger warning
				} else {
					alertAct = "AutoDelegation add"
					alertMsg = ""
					alertType = "success" //danger warning
				}
			}
		} else if typeAct == "ADELEG-SAV" {
			inputPrcDeleg_i, err := strconv.Atoi(inputPrcDeleg)
			if err != nil {
				alertAct = "AutoDelegation save"
				alertMsg = fmt.Sprintf("%s", err)
				alertType = "danger" //danger warning
			} else {
				err = SavAutodeleg(inputPubKey, idUsr, inputCoin, inputPrcDeleg_i)
				if err != nil {
					alertAct = "AutoDelegation sav"
					alertMsg = fmt.Sprintf("%s", err)
					alertType = "danger" //danger warning
				} else {
					alertAct = "AutoDelegation sav"
					alertMsg = ""
					alertType = "success" //danger warning
				}
			}
		} else if typeAct == "ADELEG-DEL" {
			err := DelAutodeleg(inputPubKey, idUsr, inputCoin)
			if err != nil {
				alertAct = "AutoDelegation del"
				alertMsg = fmt.Sprintf("%s", err)
				alertType = "danger" //danger warning
			} else {
				alertAct = "AutoDelegation del"
				alertMsg = ""
				alertType = "success" //danger warning
			}
		}
	}

	vN := []s.AutodelegCfg{}
	vN = srchAutodelegAddress(dbSQL, idUsr)

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
	ctx.Data["AllAutoDeleg"] = vN

	ctx.Data["Title"] = "Delegation"
	ctx.Data["UsrAuth"] = auth
	ctx.Data["UsrName"] = usrName
	ctx.Data["UsrAddress"] = idUsr
	ctx.Data["AlertAct"] = alertAct
	ctx.Data["AlertMsg"] = alertMsg
	ctx.Data["AlertType"] = alertType
	ctx.HTML(200, "tx_deleg")
}

// API: получить JSON конфигурацию автоделегирования по адресу
func hndAPIAutoDelegAddress(ctx *macaron.Context) {
	nmbrAddrs := ctx.Params(":number")

	retDt := []s.AutodelegCfg{}
	retDt = srchAutodelegAddress(dbSQL, nmbrAddrs)

	// возврат JSON данных, если нет, то пустой массив
	ctx.JSON(200, &retDt)
}
