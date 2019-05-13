package main

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/go-macaron/session"
	"gopkg.in/macaron.v1"

	ms "github.com/ValidatorCenter/minter-go-sdk"
	s "github.com/ValidatorCenter/prs3r/strc"
)

// СТРАНИЦА: список задач на исполнение
func hndWalletListTask(ctx *macaron.Context, sess session.Store) {
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

	myNodes := false
	allList := []s.NodeTodo{}
	if auth {
		// Список нод пользователя
		nodeM := []s.NodeExt{}
		nodeM = srchNodeAddress(dbSQL, idUsr)

		if len(nodeM) > 0 {
			allList = srchNodeTask(dbSQL, idUsr, nodeM)
			myNodes = true
		} else {
			allList = srchNodeTaskAddress(dbSQL, idUsr)
		}

		for iSt, _ := range allList {
			allList[iSt].AddressMin = getMinString(allList[iSt].Address)
			allList[iSt].PubKeyMin = getMinString(allList[iSt].PubKey)
			allList[iSt].TxHashMin = getMinString(allList[iSt].TxHash)
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

	ctx.Data["Title"] = "List ToDo"

	ctx.Data["AllTodo"] = allList
	ctx.Data["MyNodes"] = myNodes

	ctx.Data["CoinMinter"] = CoinMinter
	ctx.Data["UsrAuth"] = auth
	ctx.Data["UsrName"] = usrName
	ctx.Data["UsrAddress"] = idUsr
	ctx.Data["AlertAct"] = alertAct
	ctx.Data["AlertMsg"] = alertMsg
	ctx.Data["AlertType"] = alertType
	ctx.HTML(200, "list_todo")
}

// СТРАНИЦА: список задач на исполнение по Адресу кошелька
func hndWalletListTaskAddrs(ctx *macaron.Context, sess session.Store) {
	var alertMsg, alertType, alertAct string
	var auth bool = false
	var usrName, idUsr string

	if sess.Get("login") != nil {
		auth = true
		idUsr = sess.Get("login").(string)
		usrName = fmt.Sprintf("%s...%s", idUsr[:6], idUsr[len(idUsr)-4:len(idUsr)])
	}

	nmbrAddrs := ctx.Params(":number")

	myNodes := false
	allList := []s.NodeTodo{}

	// Список нод пользователя
	nodeM := []s.NodeExt{}
	nodeM = srchNodeAddress(dbSQL, nmbrAddrs)

	if len(nodeM) > 0 {
		allList = srchNodeTask(dbSQL, nmbrAddrs, nodeM)
		myNodes = true
	} else {
		allList = srchNodeTaskAddress(dbSQL, nmbrAddrs)
	}

	for iSt, _ := range allList {
		allList[iSt].AddressMin = getMinString(allList[iSt].Address)
		allList[iSt].PubKeyMin = getMinString(allList[iSt].PubKey)
		allList[iSt].TxHashMin = getMinString(allList[iSt].TxHash)
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

	ctx.Data["Title"] = "List ToDo"

	ctx.Data["AllTodo"] = allList
	ctx.Data["MyNodes"] = myNodes

	ctx.Data["CoinMinter"] = CoinMinter
	ctx.Data["UsrAuth"] = auth
	ctx.Data["UsrName"] = usrName
	ctx.Data["UsrAddress"] = idUsr
	ctx.Data["AlertAct"] = alertAct
	ctx.Data["AlertMsg"] = alertMsg
	ctx.Data["AlertType"] = alertType
	ctx.HTML(200, "list_todo")
}

type ReturnAPITask struct {
	WalletCash float32   `json:"wallet_cash_f32"` // на сумму
	List       []TaskOne `json:"list"`
}

// Структура v.1.1
type ReturnAPITask1_1 struct {
	WalletCash float32      `json:"wallet_cash_f32"` // на сумму
	HashID     string       `json:"hash"`
	List       []TaskOne1_1 `json:"list"`
}

// Задачи для исполнения ноде
type TaskOne struct {
	Done    bool      `json:"done"`       // выполнено
	Created time.Time `json:"created"`    // создана time
	Type    string    `json:"type"`       // тип задачи: SEND-CASHBACK,...
	Height  uint32    `json:"height_i32"` // блок
	PubKey  string    `json:"pub_key"`    // мастернода
	Address string    `json:"address"`    // адрес кошелька X
	Amount  float32   `json:"amount_f32"` // сумма
}

// Задачи для исполнения ноде v.1.1
type TaskOne1_1 struct {
	Address string  `json:"address"`    // адрес кошелька X
	Amount  float32 `json:"amount_f32"` // сумма
}

// API: получить JSON задачи на исполнение (max=100 за раз)
func hndAPIAutoTodo(ctx *macaron.Context) {
	var err error
	idUsr := ""    // адрес пользователя
	idWallet := "" // адрес пользователя, с которого оплата будет, может совподать с idUsr
	//maxTodo := 100
	retAPI := ReturnAPITask{}
	list100 := []s.NodeTodo{}
	tokenAuth := ctx.Params(":tokenauth")
	pubkeyNode := ctx.Params(":pubkey")
	//TODO: пока передаем PrivateKey, а нужно базу токенов и передавать токен
	// Получаем из приватника -> публичный адрес кошелька

	idUsr, err = ms.GetAddressPrivateKey(tokenAuth)
	if err != nil {
		fmt.Println("ERROR: convert private wallet to wallet address")
		ctx.JSON(200, &retAPI)
		return
	}

	// Получаем весь список нод пользователя
	nodeM := srchNodeAddress(dbSQL, idUsr)
	for iN, _ := range nodeM {
		if pubkeyNode == nodeM[iN].PubKey {
			idWallet = nodeM[iN].RewardAddress
		}
	}

	// Кошелек для возвратов, надо собирать еще в разрезе нод... гемор, реализуем проще - если несколько нод, ТО! будем в запросе API получать
	// конкретную ноду! просто и сердито!!!
	// получаем список
	coinsAddrs, _, _ := sdk.GetAddress(idWallet)

	// Нужно получить сумму, которая имеется на кошельке - для возвратов, потом  от неё резервируем на комиссию
	retAPI.WalletCash = 0
	if _, ok := coinsAddrs[CoinMinter]; ok {
		retAPI.WalletCash = (coinsAddrs[CoinMinter] - 0.1)
		if retAPI.WalletCash < 0 {
			retAPI.WalletCash = 0
		}
	}

	// Надо получить список задач на "сумму", но что-бы адресатов было не более определенного количества
	// Получаем не исполненых задач по нодам пользователя
	listMAX := srchNodeTaskNoDone(dbSQL, idUsr, nodeM)
	amntNowCash := float32(0)
	for iL, _ := range listMAX {
		if retAPI.WalletCash >= (amntNowCash + listMAX[iL].Amount) {
			// добавляем
			list100 = append(list100, listMAX[iL])
			amntNowCash += listMAX[iL].Amount
		} else {
			//break // набрали, выходим с цикла
			continue // будем донабирать
		}
	}

	for iL1, _ := range list100 {
		retAPI.List = append(retAPI.List, TaskOne{
			Done:    list100[iL1].Done,
			Created: list100[iL1].Created,
			Type:    list100[iL1].Type,
			Height:  list100[iL1].Height,
			PubKey:  list100[iL1].PubKey,
			Address: list100[iL1].Address,
			Amount:  list100[iL1].Amount,
		})
	}

	// возврат JSON данных, если нет, то пустой массив
	ctx.JSON(200, &retAPI)
}

// API:v1.1 получить JSON задачи на исполнение (max=100 за раз)
func hndAPIAutoTodo1_1(ctx *macaron.Context) {
	var err error
	idUsr := ""    // адрес пользователя
	idWallet := "" // адрес пользователя, с которого оплата будет, может совподать с idUsr
	//maxTodo := 100
	retDt := ReturnAPITask{}
	retAPI := ReturnAPITask1_1{}
	list100 := []s.NodeTodo{}
	tokenAuth := ctx.Params(":tokenauth")
	pubkeyNode := ctx.Params(":pubkey")
	//TODO: пока передаем PrivateKey, а нужно базу токенов и передавать токен
	// Получаем из приватника -> публичный адрес кошелька

	idUsr, err = ms.GetAddressPrivateKey(tokenAuth)
	if err != nil {
		fmt.Println("ERROR: convert private wallet to wallet address")
		ctx.JSON(200, &retAPI)
		return
	}

	// Получаем весь список нод пользователя
	nodeM := srchNodeAddress(dbSQL, idUsr)
	for iN, _ := range nodeM {
		if pubkeyNode == nodeM[iN].PubKey {
			idWallet = nodeM[iN].RewardAddress
		}
	}

	// Кошелек для возвратов, надо собирать еще в разрезе нод... гемор, реализуем проще - если несколько нод, ТО! будем в запросе API получать
	// конкретную ноду! просто и сердито!!!
	// получаем список
	coinsAddrs, _, _ := sdk.GetAddress(idWallet)

	// Для MEM
	retAPI.HashID = newHash()

	// Нужно получить сумму, которая имеется на кошельке - для возвратов, потом  от неё резервируем на комиссию
	retAPI.WalletCash = 0
	if _, ok := coinsAddrs[CoinMinter]; ok {
		retAPI.WalletCash = (coinsAddrs[CoinMinter] - 0.1)
		if retAPI.WalletCash < 0 {
			retAPI.WalletCash = 0
		}
	}

	// Надо получить список задач на "сумму", но что-бы адресатов было не более определенного количества
	// Получаем не исполненых задач по нодам пользователя
	listMAX := srchNodeTaskNoDone(dbSQL, idUsr, nodeM)
	amntNowCash := float32(0)
	for iL, _ := range listMAX {
		if retAPI.WalletCash >= (amntNowCash + listMAX[iL].Amount) {
			// добавляем
			list100 = append(list100, listMAX[iL])
			amntNowCash += listMAX[iL].Amount
		} else {
			//break // набрали, выходим с цикла
			continue // будем донабирать
		}
	}

	// Сворачиваем данные по адресу
	for _, d := range list100 {
		if d.Amount == 0 {
			continue
		}

		// Лист мультиотправки
		srchInList := false
		posic := 0
		for iL, _ := range retAPI.List {
			if retAPI.List[iL].Address == d.Address {
				srchInList = true
				posic = iL
			}
		}

		if !srchInList {
			// новый адрес+монета
			retAPI.List = append(retAPI.List, TaskOne1_1{
				Address: d.Address, //Кому переводим
				Amount:  d.Amount,
			})
		} else {
			// уже есть такой адрес, суммируем
			retAPI.List[posic].Amount += d.Amount
		}
	}

	retDt.WalletCash = retAPI.WalletCash
	for iL1, _ := range list100 {
		retDt.List = append(retDt.List, TaskOne{
			Done:    list100[iL1].Done,
			Created: list100[iL1].Created,
			Type:    list100[iL1].Type,
			Height:  list100[iL1].Height,
			PubKey:  list100[iL1].PubKey,
			Address: list100[iL1].Address,
			Amount:  list100[iL1].Amount,
		})
	}

	res1B, _ := json.Marshal(retDt) // TODO: сразу list100

	// Помещаем в Redis строку(это данные JSON) по HASH сгенерированному
	// TODO: Нужно с временем жизни помещать, т.е. удалять если давно лежит не реализованное
	if !setATasksMem(dbSys, retAPI.HashID, string(res1B)) {
		// TODO: что-то произошло не так...
	}

	// возврат JSON данных, если нет, то пустой массив
	ctx.JSON(200, &retAPI)
}

// Результат принятия ответа сервера от автоделегатора, по задачам валидатора
type ResQ struct {
	Status  int    `json:"sts"` // если не 0, то код ошибки
	Message string `json:"msg"`
}

// Результат выполнения задач валидатора
type NodeTodoQ struct {
	TxHash string     `json:"tx"` // транзакция исполнения
	QList  []TodoOneQ `json:"ql"`
}

// Идентификатор одной задачи
type TodoOneQ struct {
	Type    string    `json:"type"`       // тип задачи: SEND-CASHBACK,...
	Height  uint32    `json:"height"`     // блок
	PubKey  string    `json:"pubkey"`     // мастернода
	Address string    `json:"address"`    // адрес кошелька X
	Created time.Time `json:"created"`    // создана time
	Amount  float32   `json:"amount_f32"` // сумма
}

// API: результат от автоделегатора по возвратам
func hndAPIAutoTodoReturn(ctx *macaron.Context) {
	var err error
	retOk := ResQ{}
	idUsr := "" // адрес пользователя
	tokenAuth := ctx.Params(":tokenauth")
	//TODO: пока передаем PrivateKey, а нужно базу токенов и передавать токен

	resActive_txt := ctx.Params(":returndataJSON")
	resActive := NodeTodoQ{}
	json.Unmarshal([]byte(resActive_txt), &resActive)

	idUsr, err = ms.GetAddressPrivateKey(tokenAuth)
	if err != nil {
		fmt.Println("ERROR: convert private wallet to wallet address")
		retOk.Status = 1 // error!
		retOk.Message = "ERROR: convert private wallet to wallet address"
		ctx.JSON(200, &retOk)
		return
	}
	// Проверить что данная нода принадлежит именно данному пользователю!
	nodeM := []s.NodeExt{}
	nodeM = srchNodeAddress(dbSQL, idUsr) // Получаем весь список нод пользователя

	dateNow := time.Now()
	//перебираем массив и обновляем статус в базе //
	for _, d := range resActive.QList {
		nodeUser := false
		for _, cN := range nodeM {
			if d.PubKey == cN.PubKey {
				nodeUser = true
			}
		}
		if nodeUser {
			updData := s.NodeTodo{}
			// Search
			updData.Type = d.Type
			updData.Height = d.Height
			updData.PubKey = d.PubKey
			updData.Address = d.Address
			updData.Created = d.Created
			updData.Amount = d.Amount
			// Update
			updData.Done = true
			updData.DoneT = dateNow
			updData.TxHash = resActive.TxHash

			err = updNodeTask(dbSQL, updData)

			if err != nil {
				retOk.Status = 1 // error!
				retOk.Message = err.Error()
				ctx.JSON(200, &retOk)
				return
			}
		} else {
			// пропускаем! Данная нода не принадлежит пользователю!
			// TODO: результат отдать о том что нода не пользователя (может кто то пытается взломать)
		}
	}

	retOk.Status = 0 // ok!
	// возврат JSON данных: ok(0) или bad(код ошибки)
	ctx.JSON(200, &retOk)
}

// API:v1.1 результат от автоделегатора по возвратам
func hndAPIAutoTodoReturn1_1(ctx *macaron.Context) {
	var err error
	retOk := ResQ{}
	idUsr := "" // адрес пользователя
	tokenAuth := ctx.Params(":tokenauth")
	hashID := ctx.Params(":hashid")
	txHash := ctx.Params(":hashtrx")

	resActive_txt := getATasksMem(dbSys, hashID)
	//resActive := NodeTodoQ{}

	resActive := ReturnAPITask{}
	json.Unmarshal([]byte(resActive_txt), &resActive)

	idUsr, err = ms.GetAddressPrivateKey(tokenAuth)
	if err != nil {
		fmt.Println("ERROR: convert private wallet to wallet address")
		retOk.Status = 1 // error!
		retOk.Message = "ERROR: convert private wallet to wallet address"
		ctx.JSON(200, &retOk)
		return
	}
	// Проверить что данная нода принадлежит именно данному пользователю!
	nodeM := []s.NodeExt{}
	nodeM = srchNodeAddress(dbSQL, idUsr) // Получаем весь список нод пользователя

	dateNow := time.Now()
	//перебираем массив и обновляем статус в базе //
	for _, d := range resActive.List {
		nodeUser := false
		for _, cN := range nodeM {
			if d.PubKey == cN.PubKey {
				nodeUser = true
			}
		}
		if nodeUser {
			updData := s.NodeTodo{}
			// Search
			updData.Type = d.Type
			updData.Height = d.Height
			updData.PubKey = d.PubKey
			updData.Address = d.Address
			updData.Created = d.Created
			updData.Amount = d.Amount
			// Update
			updData.Done = true
			updData.DoneT = dateNow
			updData.TxHash = txHash

			err = updNodeTask(dbSQL, updData)

			if err != nil {
				retOk.Status = 1 // error!
				retOk.Message = err.Error()
				ctx.JSON(200, &retOk)
				return
			}
		} else {
			// пропускаем! Данная нода не принадлежит пользователю!
			// TODO: результат отдать о том что нода не пользователя (может кто то пытается взломать)
		}
	}

	if !delATasksMem(dbSys, hashID) {
		log("ERR", fmt.Sprint("[wallet_task.go] hndAPIAutoTodoReturn1_1(delATasksMem) - HASH = ", hashID), "")
	}

	retOk.Status = 0 // ok!
	// возврат JSON данных: ok(0) или bad(код ошибки)
	ctx.JSON(200, &retOk)
}
