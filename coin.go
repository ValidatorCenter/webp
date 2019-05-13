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

// Структура показа инф. о монете в списке
type CoinView struct {
	ID          int
	Name        string
	Ticker      string
	CoinLogoImg string
	PriceBuy    string
	PriceBuyUSD string
	Volume24    string
	Change24    string
	Change24f32 float32
}

// Структура Японской-Свечки для графиков
type OneCoinTrans struct {
	Date   string
	Open   float32
	High   float32
	Low    float32
	Close  float32
	Volume float32
}

// Функция покупки монеты
func BuyCoin(sess session.Store, sellCoin string, buyCoin string, valueBuy float32) (string, error) {
	sdk.AccPrivateKey = sess.Get("priv_k").(string)
	//sdk.AccAddress,_ = ms.GetAddressPrivateKey(sdk.AccPrivateKey) // другой вариант ниже
	sdk.AccAddress = sess.Get("login").(string)

	buyDt := ms.TxBuyCoinData{
		CoinToSell: sellCoin,
		CoinToBuy:  buyCoin,
		ValueToBuy: valueBuy,
		// Gas
		GasCoin:  CoinMinter,
		GasPrice: 1,
	}

	resHash, err := sdk.TxBuyCoin(&buyDt)
	if err != nil {
		return "", err
	}
	return resHash, nil
}

// Функция продажи монеты
func SellCoin(sess session.Store, sellCoin string, buyCoin string, valueSell float32) (string, error) {
	sdk.AccPrivateKey = sess.Get("priv_k").(string)
	//sdk.AccAddress,_ = ms.GetAddressPrivateKey(sdk.AccPrivateKey) // другой вариант ниже
	sdk.AccAddress = sess.Get("login").(string)

	slDt := ms.TxSellCoinData{
		CoinToSell:  sellCoin,
		CoinToBuy:   buyCoin,
		ValueToSell: valueSell,
		GasCoin:     CoinMinter,
		GasPrice:    1,
	}

	resHash, err := sdk.TxSellCoin(&slDt)
	if err != nil {
		return "", err
	}
	return resHash, nil
}

// СТРАНИЦА: Общий список монет
func hndCoins(ctx *macaron.Context, sess session.Store) {
	var alertMsg, alertType, alertAct string
	var auth bool = false
	var usrName, idUsr string

	//SESSION
	if sess.Get("login") != nil {
		auth = true
		idUsr = sess.Get("login").(string)
		usrName = fmt.Sprintf("%s...%s", idUsr[:6], idUsr[len(idUsr)-4:len(idUsr)])
	}

	//allCoins := []s.CoinMarketCapData{}
	allCoins := srchCoin(dbSQL)

	// 1bit=0.07$ (usd)
	var _1bit_usd float64 = 0.07 // TODO: брать надо будет из биржи Binance

	allCoinsView := []CoinView{}
	id := 1
	for _, coin1 := range allCoins {
		var coinOne CoinView

		srchCoinInfoRds(dbSys, &coin1) // заполняем спец.инфой

		coinOne.ID = id
		coinOne.Name = coin1.Name
		coinOne.Ticker = coin1.CoinSymbol
		coinOne.CoinLogoImg = coin1.CoinLogoImg

		// Получение данных по паре Текущая-монета к Базовой-монете(MNT/BIP)
		// для расчета объема и цены последней
		//p2Coin := []s.PairCoins{}
		p2Coin := srchCoin2_2_24(dbSys, coin1.CoinSymbol, CoinMinter, true)

		if len(p2Coin) == 0 {
			//ничего не нашли (тогда что есть и переводим мнт?)
			// FIXME: придумать что-нибудь
			coinOne.PriceBuy = strconv.FormatFloat(float64(0.0), 'f', 2, 32)
			coinOne.PriceBuyUSD = fmt.Sprintf("$%s", strconv.FormatFloat(float64(0.0)*_1bit_usd, 'f', 2, 32))
			coinOne.Volume24 = strconv.FormatFloat(float64(0.0), 'f', 2, 32)
			coinOne.Change24 = strconv.FormatFloat(float64(0.0), 'f', 2, 32)
			coinOne.Change24f32 = 0.0
		} else {
			coinOne.PriceBuy = strconv.FormatFloat(float64(p2Coin[0].PriceBuy), 'f', 2, 32)
			coinOne.PriceBuyUSD = fmt.Sprintf("$%s", strconv.FormatFloat(float64(p2Coin[0].PriceBuy)*_1bit_usd, 'f', 2, 32))
			coinOne.Volume24 = strconv.FormatFloat(float64(p2Coin[0].Volume24), 'f', 2, 32)
			coinOne.Change24 = strconv.FormatFloat(float64(p2Coin[0].Change24), 'f', 2, 32)
			coinOne.Change24f32 = p2Coin[0].Change24
		}

		allCoinsView = append(allCoinsView, coinOne)
		id++
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
	ctx.Data["Title"] = "Coins"

	ctx.Data["AllCoins"] = allCoinsView

	// Пользователь:
	ctx.Data["UsrAuth"] = auth
	ctx.Data["UsrName"] = usrName
	ctx.Data["UsrAddress"] = idUsr

	// Инф.сообщения от системы:
	ctx.Data["AlertAct"] = alertAct
	ctx.Data["AlertMsg"] = alertMsg
	ctx.Data["AlertType"] = alertType

	// Вывод страницы:
	ctx.HTML(200, "coins")
}

// СТРАНИЦА: Информация об одной паре-монет (к базовой{по умолч.} или кастомной)
func hndCoinInfo(ctx *macaron.Context, sess session.Store) {
	var alertMsg, alertType, alertAct string
	var auth bool = false
	var usrName, idUsr string

	// SESSION
	if sess.Get("login") != nil {
		auth = true
		idUsr = sess.Get("login").(string)
		usrName = fmt.Sprintf("%s...%s", idUsr[:6], idUsr[len(idUsr)-4:len(idUsr)])
	}

	ticker1 := ctx.Params(":ticker")
	ticker2 := CoinMinter //"MNT"
	if ctx.Params(":ticker2") != "" {
		ticker2 = ctx.Params(":ticker2")
	}

	// Моя монета?!
	myCoin := false

	if ticker1 == CoinMinter && ticker2 == CoinMinter {
		// TODO: Если ticker1 есть и он базовая монета, а ticker2 нету, то нужно просто инфу о БАЗОВОЙ монете: добыто, в обороте и т.п.
		page404(ctx)
		fmt.Println("WRNG: нет страницы о базовой монете")
		return
	}

	// если покупка или продажа, то движения нужно
	ctx.Req.ParseForm()
	ctx.Resp.WriteHeader(http.StatusOK)
	typeAct := ctx.Req.PostFormValue("typeAct")

	// Покупка или продажа монеты [DEX]	или изменение создателем монеты её инфы
	if typeAct != "" && auth == true {
		if typeAct == "BUY" {
			inputBuy1 := ctx.Req.PostFormValue("inputBuy1")
			inputBuy2 := ctx.Req.PostFormValue("inputBuy2")
			inputBuy1_f32, err := strconv.ParseFloat(inputBuy1, 32)
			if err != nil {
				alertAct = "Buy"
				alertMsg = fmt.Sprintf("%s", err)
				alertType = "danger" //danger warning
			} else {
				fmt.Println("BUY", inputBuy1, inputBuy2)
				hashTr, err := BuyCoin(sess, ticker2, ticker1, float32(inputBuy1_f32))
				if err != nil {
					alertAct = "Buy"
					alertMsg = fmt.Sprintf("%s", err)
					alertType = "danger" //danger warning
				} else {
					alertAct = "Buy"
					alertMsg = hashTr
					alertType = "success" //danger warning
				}
			}
		}
		if typeAct == "SELL" {
			inputSell1 := ctx.Req.PostFormValue("inputSell1")
			inputSell2 := ctx.Req.PostFormValue("inputSell2")
			inputSell1_f32, err := strconv.ParseFloat(inputSell1, 32)
			if err != nil {
				alertAct = "Sell"
				alertMsg = fmt.Sprintf("%s", err)
				alertType = "danger" //danger warning
			} else {
				fmt.Println("SELL", inputSell1, inputSell2)
				hashTr, err := SellCoin(sess, ticker1, ticker2, float32(inputSell1_f32))
				if err != nil {
					alertAct = "Sell"
					alertMsg = fmt.Sprintf("%s", err)
					alertType = "danger" //danger warning
				} else {
					alertAct = "Sell"
					alertMsg = hashTr
					alertType = "success" //danger warning
				}
			}
		}
		// Изменение визуальной информации о монете (логотип, сайт, описание)
		if typeAct == "EDIT" {
			iWWW := ctx.Req.PostFormValue("inputWWW")
			iDesc := ctx.Req.PostFormValue("inputDescription")
			iIcon := ctx.Req.PostFormValue("inputIcon")

			// проверяем что адрес точно является создателем
			coinEdt := s.CoinMarketCapData{}
			coinEdt = srchCoinCreator(dbSQL, ticker1, idUsr) // TODO: потом будет возврат true/false а полные данные брать из Redis и там же их и менять!

			if coinEdt.CoinSymbol != "" {
				myCoin = true
				// новые данные
				coinEdt.CoinURL = iWWW
				coinEdt.CoinLogoImg = iIcon
				coinEdt.CoinDesc = iDesc

				if !updCoinInfoRds_3v(dbSys, &coinEdt) {
					alertAct = "Edit"
					alertMsg = fmt.Sprintf("Update %s coin", ticker1)
					alertType = "danger" //danger warning
				} else {
					alertAct = "Edit"
					alertMsg = ticker1
					alertType = "success" //danger warning
				}
			}
		}
	}

	_2Coin := []string{}
	one1Coins := s.CoinMarketCapData{}
	allTrans := []OneCoinTrans{}

	one1Coins = srchCoin1(dbSQL, ticker1) // получение данных монеты из SQL
	srchCoinInfoRds(dbSys, &one1Coins)    // получение данных монеты из Redis

	actionCoin := true
	directDirection := true

	p2Coin := []s.PairCoins{}
	p2Coin = srchCoin2_1_24(dbSys, ticker1, ticker2) // берём пару в одном направление

	if len(p2Coin) == 0 {
		// ничего не нашли
		//FIXME: зачем так??? может через srchCoin2_2_24()
		p2Coin = srchCoin2_1_24(dbSys, ticker2, ticker1) // теперь эту-же пару, но в другом направление
		if len(p2Coin) == 0 {
			// нету данных!!! о покупке и продажи!
			actionCoin = false
		} else {
			// надо менять местами цену!!!
			directDirection = false
		}
	}

	if one1Coins.CoinSymbol == "" {
		// нету данных!!!
		page404(ctx)
		fmt.Println("WRNG: нет страницы 2")
		return
	}
	for _, tr1 := range one1Coins.Transactions {
		new2Coin := true
		for _, pc := range _2Coin {
			if pc == fmt.Sprintf("%s/%s", tr1.CoinToBuy, tr1.CoinToSell) || pc == fmt.Sprintf("%s/%s", tr1.CoinToSell, tr1.CoinToBuy) {
				new2Coin = false
			}
		}
		if new2Coin {
			_2Coin = append(_2Coin, fmt.Sprintf("%s/%s", tr1.CoinToBuy, tr1.CoinToSell))
		}

		if (tr1.CoinToBuy == ticker1 && tr1.CoinToSell == ticker2) || (tr1.CoinToBuy == ticker2 && tr1.CoinToSell == ticker1) {
			// всё хорошо, идем дальше по алгоритму
		} else {
			// ИНАЧЕ выбрать другой элемент
			continue // переходим к следующей итерации
		}

		allTrans = append(allTrans, OneCoinTrans{
			Date:   tr1.Time.Format("2006-01-02 15:04:05"),
			Open:   tr1.Price,
			High:   tr1.Price,
			Low:    tr1.Price,
			Close:  tr1.Price,
			Volume: tr1.Volume,
		})
	}

	// проверка на моя ли монета, если еще не проводилась выше при EDIT
	if auth && !myCoin {
		coinEdt := srchCoinCreator(dbSQL, ticker1, idUsr) // TODO: будет true/false!

		if coinEdt.CoinSymbol != "" {
			myCoin = true
		}
	}

	creatorTXT := ""
	if one1Coins.Creator != "" {
		creatorTXT = fmt.Sprintf("%s...%s", one1Coins.Creator[:7], one1Coins.Creator[len(one1Coins.Creator)-7:len(one1Coins.Creator)])
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
	ctx.Data["Title"] = "Coin" //one1Coins.Name // иначе не покажется в меню выбор монет <-->

	ctx.Data["UpdateData"] = timeFormat0(one1Coins.TimeUpdate)
	ctx.Data["Time"] = timeFormat0(one1Coins.Time)
	ctx.Data["TitleCoin"] = one1Coins.Name
	ctx.Data["Transactions"] = allTrans
	ctx.Data["Ticker1"] = ticker1
	ctx.Data["Ticker2"] = ticker2
	ctx.Data["Other2Coins"] = _2Coin
	if actionCoin {
		if directDirection {
			ctx.Data["PriceNowBuy"] = strconv.FormatFloat(float64(p2Coin[0].PriceBuy), 'f', 4, 32)
			ctx.Data["PriceNowSell"] = strconv.FormatFloat(float64(p2Coin[0].PriceSell), 'f', 4, 32)
		} else {
			ctx.Data["PriceNowBuy"] = strconv.FormatFloat(float64(p2Coin[0].PriceSell), 'f', 4, 32)
			ctx.Data["PriceNowSell"] = strconv.FormatFloat(float64(p2Coin[0].PriceBuy), 'f', 4, 32)
		}
		ctx.Data["Change24"] = p2Coin[0].Change24
		ctx.Data["Volume24"] = p2Coin[0].Volume24
	} else {
		ctx.Data["PriceNowBuy"] = strconv.FormatFloat(float64(0), 'f', 4, 32)
		ctx.Data["PriceNowSell"] = strconv.FormatFloat(float64(0), 'f', 4, 32)
		ctx.Data["Change24"] = 0
		ctx.Data["Volume24"] = 0
	}
	ctx.Data["MyCoin"] = myCoin
	ctx.Data["Creator"] = one1Coins.Creator // с ссылкой на эксплорер
	ctx.Data["CreatorTXT"] = creatorTXT
	ctx.Data["CRR_prc"] = one1Coins.ConstantReserveRatio
	ctx.Data["InitialAmount"] = strconv.FormatFloat(float64(one1Coins.InitialAmount), 'f', 2, 32)   //Amount of coins to issue. Issued coins will be available to sender account.
	ctx.Data["InitialReserve"] = strconv.FormatFloat(float64(one1Coins.InitialReserve), 'f', 2, 32) // Initial reserve in base coin.
	ctx.Data["Volume"] = one1Coins.VolumeNow
	ctx.Data["ReserveBalance"] = strconv.FormatFloat(float64(one1Coins.ReserveBalanceNow), 'f', 2, 32)
	//ctx.Data["AmntTrans24x7"] = one1Coins.AmntTrans24x7 //<p class="lead">Количество транзакций (за 7д):</p>
	ctx.Data["CoinInf"] = one1Coins

	// Пользователь:
	ctx.Data["UsrAuth"] = auth
	ctx.Data["UsrName"] = usrName
	ctx.Data["UsrAddress"] = idUsr

	// Инф.сообщения от системы:
	ctx.Data["AlertAct"] = alertAct
	ctx.Data["AlertMsg"] = alertMsg
	ctx.Data["AlertType"] = alertType

	// Вывод страницы:
	ctx.HTML(200, "coin_info")
}
