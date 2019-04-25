package main

import (
	"fmt"
	"math"
	"strconv"
	"time"

	"github.com/go-macaron/session"
	"gopkg.in/macaron.v1"

	//ms "github.com/ValidatorCenter/minter-go-sdk"
	s "github.com/ValidatorCenter/prs3r/strc"
)

// Структура одного блока для вывода на страницу
type Block1 struct {
	Hash             string              `json:"hash"`
	HashMin          string              `json:"hash_min"`
	Height           int                 `json:"height_i32"`
	Time             time.Time           `json:"time"`
	Age              string              `json:"age"`
	NumTxs           int                 `json:"num_txs_i32"`
	TotalTxs         int                 `json:"total_txs_i32"`
	Transactions     []Block1Transaction `json:"transactions"`
	Events           []Block1Events      `json:"events"`
	Validators       []Block1Validator   `json:"validators"`
	Proposer         string              `json:"proposer"`
	ProposerName     string              `json:"proposer_name"`
	ProposerLogo     string              `json:"proposer_logo"`
	BlockReward      float32             `json:"block_reward_f32"`
	Size             int                 `json:"size_i32"`
	TransactionsAmnt int                 `json:"transactions_amnt_i32"`
	EventsAmnt       int                 `json:"events_amnt_i32"`
	PrecommitsAmnt   int                 `json:"validators_amnt_i32"`
	PrecommitsBlock  int                 `json:"-"`
}

// Структура одной транзакции в блоке для вывода на страницу
type Block1Transaction struct {
	Hash     string      `json:"hash"`
	HashMin  string      `json:"hash_min"`
	From     string      `json:"from"`
	FromMin  string      `json:"from_min"`
	Nonce    int         `json:"nonce_i32"`
	GasPrice int         `json:"gas_price_i32"`
	Type     int         `json:"type"`
	TypeTxt  string      `json:"type_txt"`
	Amount   float32     `json:"amount_f32"`
	Data     interface{} `json:"-"`
	Payload  string      `json:"payload"` // комментарий зашифрован Base64
	Gas      int         `json:"gas_i32"`
	GasCoin  string      `json:"gas_coin"`
	GasUsed  int         `json:"gas_used_i32"`
}

// Структура одного события в блоке для вывода на страницу
type Block1Events struct {
	Type    string `json:"type"`
	TypeTxt string `json:"type_txt"`
	//--------------------------
	Role               string  `json:"role"` //DAO,Developers,Validator,Delegator
	Address            string  `json:"address"`
	AddressMin         string  `json:"address_min"`
	Amount             float32 `json:"amount_f32"`
	Coin               string  `json:"coin"`
	ValidatorPubKey    string  `json:"validator_pub_key"`
	ValidatorPubKeyMin string  `json:"validator_pub_key_min"`
}

// Структура одного валидатор в блоке для вывода на страницу
type Block1Validator struct {
	PubKey string `json:"pub_key"`
	Signed bool   `json:"signed"`
	Name   string `json:"name"`
	Logo   string `json:"logo"`
}

func getTxOperationTxt(txI int, lng string) string {
	switch txI {
	case 1:
		return "Send"
	case 2:
		return "SellCoin"
	case 3:
		return "SellAllCoin"
	case 4:
		return "BuyCoin"
	case 5:
		return "CreateCoin"
	case 6:
		return "DeclareCandidacy"
	case 7:
		return "Delegate"
	case 8:
		return "Unbond"
	case 9:
		return "RedeemCheck"
	case 10:
		return "SetCandidateOn"
	case 11:
		return "SetCandidateOff"
	case 12:
		return "CreateMultisig"
	}
	return "?"
}

// СТРАНИЦА: список блоков по убыванию
func hndBlocksInfo(ctx *macaron.Context, sess session.Store) {
	var alertMsg, alertType, alertAct string
	var auth bool = false
	var usrName, idUsr string
	var err error

	//SESSION
	if sess.Get("login") != nil {
		auth = true
		idUsr = sess.Get("login").(string)
		usrName = fmt.Sprintf("%s...%s", idUsr[:6], idUsr[len(idUsr)-4:len(idUsr)])
	}

	// Пагинация
	pageNmbrNow := 1
	if ctx.Params(":pgn") != "" {
		pageNmbrNowTxt := ctx.Params(":pgn")
		pageNmbrNow, err = strconv.Atoi(pageNmbrNowTxt)
		if err != nil {
			pageNmbrNow = 1
		}
	}
	if pageNmbrNow < 1 {
		pageNmbrNow = 1
	}

	totalBlock := srchBlockAmnt(dbSQL) // количество блоков уже загружено в БД

	// для навигации по страницам списка блоков из БД
	minBlock := totalBlock - 50*(pageNmbrNow-1)   // от последнего блока - Начало (!важно, minBlock больше maxBlock)
	maxBlock := totalBlock - 50*(pageNmbrNow) + 1 // от последнего блока - Конец (!важно, minBlock больше maxBlock)
	if maxBlock < 1 {
		maxBlock = 1
	}

	blsM := srchBlockN(dbSQL, uint32(minBlock), uint32(maxBlock)) // получаем из БД список блоков по заданному промежутку

	dt := time.Now()

	// Перерабатываем данные полученные из БД по блокам
	blsM1 := []Block1{}
	for _, v := range blsM {
		b0 := Block1{}

		// TODO: Идеи по отображение страницы Лист-блоков (взято с https://stargazer.certus.one/blocks)
		// 1 - показывать пропозера блока, и показывать его по имени! которое он себе установил
		// 2 - ...или можно просто его логотип показывать, меньше занимаего места и очень круто

		b0.Hash = v.Hash
		b0.HashMin = getMinString(v.Hash)
		b0.Height = v.Height
		b0.Time = v.Time
		b0.Age = diffTimeStr(dt, v.Time)
		b0.NumTxs = v.NumTxs // Количество транзакций в блоке
		b0.TotalTxs = v.TotalTxs
		b0.BlockReward = v.BlockReward
		b0.Size = v.Size
		b0.Proposer = v.Proposer
		b0.PrecommitsAmnt = v.ValidatorAmnt

		// FIXME: содержимое блока - должно получаться потом через JSON, а не как сейчас: получает всё! и модальное окно просто показывает

		//b0.TransactionsAmnt = len(v.Transactions)
		//b0.EventsAmnt = len(v.Events)
		//b0.Validators = v.Validators

		//.......................... for _, vT := range v.Transactions {....
		//.......................... for _, vE := range v.Events {....

		blsM1 = append(blsM1, b0)
	}

	// Расчет кнопок навигации
	BtnLL := 0
	BtnL := 0
	if pageNmbrNow != 1 {
		BtnLL = 1
		BtnL = pageNmbrNow - 1
	}
	BtnR := 0
	BtnRR := 0
	BtnR = pageNmbrNow + 1
	BtnRR = int(math.Ceil(float64(totalBlock) / 50))
	if pageNmbrNow == BtnRR {
		BtnR = 0
		BtnRR = 0
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
	ctx.Data["Title"] = "Blocks"

	ctx.Data["AllBlocks"] = blsM1
	ctx.Data["TotalBlock"] = totalBlock
	ctx.Data["MinBlock"] = minBlock
	ctx.Data["MaxBlock"] = maxBlock
	ctx.Data["CoinMinter"] = CoinMinter // Базовая монета систм.

	// Кнопки навигации по страницам:
	ctx.Data["BtnLL"] = BtnLL
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
	ctx.HTML(200, "blocks_info")
}

// Получаем данные по блоку для визуализации, с Event,Valid и Trx
func GetBlockOneInfo(nmbrBlockInt uint32) Block1 {
	b0 := Block1{}

	//v := ms.BlockResponse{} // FIXME: устаревший способ, т.к. доработал и внес Events обратно в блок!
	v := s.BlockResponse2{}
	v = srchBlockMin(dbSQL, nmbrBlockInt)

	// проверка что блок найден!!! иначе на страницу 404
	if v.Hash == "" {
		return b0
	}

	b0.Hash = v.Hash
	b0.HashMin = getMinString(v.Hash)
	b0.Height = v.Height
	b0.Time = v.Time
	b0.NumTxs = v.NumTxs
	b0.TotalTxs = v.TotalTxs
	b0.BlockReward = v.BlockReward
	b0.Size = v.Size
	b0.Proposer = v.Proposer
	b0.TransactionsAmnt = len(v.Transactions)
	b0.EventsAmnt = len(v.Events)
	b0.PrecommitsAmnt = len(v.Validators)

	vPro := s.NodeExt{}
	vPro.PubKey = v.Proposer
	if !srchNodeInfoRds(dbSys, &vPro) { // получаем динамические данные по ноде
		log("ERR", fmt.Sprintf("[block.go] GetBlockOneInfo(srchNodeInfoRds) Redis load - %s", vPro.PubKey), "")
	} else {
		if vPro.ValidatorName != "" {
			b0.ProposerName = vPro.ValidatorName
		}
		if vPro.ValidatorLogoImg != "" {
			b0.ProposerLogo = vPro.ValidatorLogoImg
		}
	}

	// Подстраница: Валидаторы блока
	for _, vV := range v.Validators {
		v0 := Block1Validator{}
		v0.PubKey = vV.PubKey
		v0.Signed = vV.Signed

		v0Ext := s.NodeExt{}
		v0Ext.PubKey = vV.PubKey
		if srchNodeInfoRds(dbSys, &v0Ext) {
			if v0Ext.ValidatorName != "" {
				v0.Name = v0Ext.ValidatorName
			}
			if v0Ext.ValidatorLogoImg != "" {
				v0.Logo = v0Ext.ValidatorLogoImg
			}
		}

		b0.Validators = append(b0.Validators, v0)
	}

	// Подстраница: Транзакции блока
	for _, vT := range v.Transactions {
		t0 := Block1Transaction{}

		t0.Hash = vT.Hash
		t0.HashMin = getMinString(vT.Hash)
		t0.From = vT.From
		t0.FromMin = getMinString(vT.From)
		t0.Type = vT.Type
		t0.TypeTxt = getTxOperationTxt(vT.Type, "en") // внизу отработка
		t0.Amount = vT.Amount

		b0.Transactions = append(b0.Transactions, t0)
	}

	// Подстраница: События блока
	for _, vE := range v.Events {
		e0 := Block1Events{}

		e0.Type = vE.Type
		e0.TypeTxt = ""         // TODO: расшифровка от e0.Type
		e0.Role = vE.Value.Role //DAO,Developers,Validator,Delegator
		e0.Address = vE.Value.Address
		e0.AddressMin = getMinString(vE.Value.Address)
		e0.Amount = vE.Value.Amount
		e0.Coin = vE.Value.Coin
		e0.ValidatorPubKey = vE.Value.ValidatorPubKey
		e0.ValidatorPubKeyMin = getMinString(vE.Value.ValidatorPubKey)

		b0.Events = append(b0.Events, e0)
	}

	return b0
}

// СТРАНИЦА: одного блока
func hndBlockOneInfo(ctx *macaron.Context, sess session.Store) {
	var alertMsg, alertType, alertAct string
	var auth bool = false
	var usrName, idUsr string

	//SESSION
	if sess.Get("login") != nil {
		auth = true
		idUsr = sess.Get("login").(string)
		usrName = fmt.Sprintf("%s...%s", idUsr[:6], idUsr[len(idUsr)-4:len(idUsr)])
	}

	nmbrBlock := ctx.Params(":number")
	nmbrBlockInt, _ := strconv.Atoi(nmbrBlock)

	b0 := GetBlockOneInfo(uint32(nmbrBlockInt))

	// проверка что блок найден!!! иначе на страницу 404
	if b0.Hash == "" {
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
	ctx.Data["Title"] = "Block"

	ctx.Data["OneBlocks"] = b0
	ctx.Data["CoinMinter"] = CoinMinter

	// Пользователь:
	ctx.Data["UsrAuth"] = auth
	ctx.Data["UsrName"] = usrName
	ctx.Data["UsrAddress"] = idUsr

	// Инф.сообщения от системы:
	ctx.Data["AlertAct"] = alertAct
	ctx.Data["AlertMsg"] = alertMsg
	ctx.Data["AlertType"] = alertType

	// Вывод страницы:
	ctx.HTML(200, "block_info")
}

// Структура возврата блока в виде JSON
type RetJSONBlock struct {
	Status bool   `json:"status"`
	Block  Block1 `json:"block"`
	ErrMsg string `json:"err_msg"`
}

// API: один блок
func hndAPIBlockOneInfo(ctx *macaron.Context, sess session.Store) {
	retDt := RetJSONBlock{}

	retDt.Status = true // исполнен без ошибок
	retDt.ErrMsg = ""   // нет ошибок

	nmbrBlock := ctx.Params(":number")
	nmbrBlockInt, _ := strconv.Atoi(nmbrBlock)

	b0 := GetBlockOneInfo(uint32(nmbrBlockInt))

	// проверка что блок найден!!!
	if b0.Hash == "" {
		retDt.Status = false             // исполнен с ошибкой
		retDt.ErrMsg = "No search block" // текст ошибки
	}
	retDt.Block = b0

	// возврат JSON данных
	ctx.JSON(200, &retDt)
}
