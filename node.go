package main

import (
	"fmt"
	"math"
	"sort"

	"net/http"
	"strconv"
	"time"

	"github.com/go-macaron/session"
	"gopkg.in/macaron.v1"

	s "github.com/ValidatorCenter/prs3r/strc"
)

type NodeExt2 struct {
	s.NodeExt
	CommissionNow int `json:"commission_now"`
}

// СТРАНИЦА: со списком валидаторов (и кандидатов)
func hndValidatorsInfo(ctx *macaron.Context, sess session.Store) {
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

	pageVldt := 0
	if ctx.Params(":pgn") != "" {
		pageVldt, err = strconv.Atoi(ctx.Params(":pgn"))
		if err != nil {
			pageVldt = 0
		}
	}
	if pageVldt < 0 {
		pageVldt = 0
	}

	// FIXME: УДАЛИТЬ
	//skipVldt := pageVldt * 50
	//fmt.Println(skipVldt)

	// Количество нод в базе
	totalBlock := srchNodeAmnt(dbSQL)

	nodeM := []NodeExt2{}
	// Получаем список нод
	nodeM2 := srchNodeList(dbSQL) //, skipVldt)
	for iNm2, _ := range nodeM2 {
		nNm1 := NodeExt2{}
		nNm1.NodeExt = nodeM2[iNm2]
		nodeM = append(nodeM, nNm1)
	}

	// Теперь бежим по списку и дозополняем данныеми из Redis
	for iNs, _ := range nodeM {
		if !srchNodeInfoRds(dbSys, &nodeM[iNs].NodeExt) { // получаем динамические данные по ноде
			log("ERR", fmt.Sprintf("[node.go] hndValidatorsInfo(srchNodeInfoRds) Redis load - %s", nodeM[iNs].PubKey), "")
		}
		nodeM[iNs].CommissionNow = nodeM[iNs].Commission
	}
	// сортируем по ORDER BY status DESC, total_stake_f32 DESC
	sort.Slice(nodeM, func(i, j int) bool {
		//FIXME: хреново отрабатывается, сортируеся почему-то не совсем по уровню
		return (nodeM[i].StatusInt >= nodeM[j].StatusInt) && (nodeM[i].TotalStake > nodeM[j].TotalStake)
	})

	// Изменение % ноды, если есть уникальное условие
	nodesNewPrc := srchNodeUserXAddress(dbSQL, "Mx----------------------------------------")
	dateNow := time.Now()
	for iN, _ := range nodeM {
		for _, nux1 := range nodesNewPrc {
			if nodeM[iN].PubKey == nux1.PubKey {
				if nux1.Start.Unix() <= dateNow.Unix() && nux1.Finish.Unix() >= dateNow.Unix() {
					if nodeM[iN].CommissionNow > nux1.Commission {
						nodeM[iN].CommissionNow = nux1.Commission
					}
				}
			}
		}
	}

	stkcA := srchNodeStakesAmntAll(dbSQL)
	bstrT := srchNodeBlockstoryTypeAll(dbSQL, "AbsentBlock")

	BtnL := 0
	BtnLL := 0
	if pageVldt != 0 {
		BtnL = pageVldt - 1
	}
	BtnR := 0
	BtnRR := 0
	BtnR = pageVldt + 1
	BtnRR = int(math.Ceil(float64(totalBlock)/50) - 1)
	if pageVldt == BtnRR {
		BtnR = 0
		BtnRR = 0
	}

	dt := time.Now()
	rtng := int(1)
	for iM, _ := range nodeM {
		nodeM[iM].RatingID = rtng
		nodeM[iM].Age = diffTimeStr(dt, nodeM[iM].Created)
		for _, dNB := range bstrT {
			if dNB.PubKey == nodeM[iM].PubKey {
				nodeM[iM].AmnNoBlocks = dNB.Amnt
			}
		}
		for _, dNB := range stkcA {
			if dNB.PubKey == nodeM[iM].PubKey {
				nodeM[iM].AmntSlots = dNB.Amnt
			}
		}
		rtng++
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
	ctx.Data["Title"] = "Nodes"

	ctx.Data["AllCandidates"] = nodeM
	ctx.Data["CoinMinter"] = CoinMinter // Базовая монета систм.

	// Кнопки навигации по страницам:
	ctx.Data["BtnNow"] = pageVldt
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
	ctx.HTML(200, "nodes_info")
}

// Возвращает полные данные по ноде - информация, награда, делегаты...
func GetNodeOneInfo(nmbrNode string) (s.NodeExt, []ResReward) {
	v := s.NodeExt{}
	rE := []ResReward{}

	v = srchNodePubkey(dbSQL, nmbrNode) // получаем общие данные по ноде
	if !srchNodeInfoRds(dbSys, &v) {    // получаем динамические данные по ноде
		log("ERR", fmt.Sprintf("[node.go] GetNodeOneInfo(srchNodeInfoRds) Redis load - %s", v.PubKey), "")
	}
	v.Stakes = srchNodeStakeSql(dbSQL, nmbrNode)   // стэк ноды
	v.Blocks = srchNodeBlockstory(dbSQL, nmbrNode) // важные исторические блоки для ноды

	rE = srchRewardPubkeySql(dbSQL, nmbrNode)

	return v, rE
}

// СТРАНИЦА: с информацией об одной ноде (валидаторе/кандидате)
func hndValidatorOneInfo(ctx *macaron.Context, sess session.Store) {
	var alertMsg, alertType, alertAct string
	var auth bool = false
	var usrName, idUsr string

	// SESSION
	if sess.Get("login") != nil {
		auth = true
		idUsr = sess.Get("login").(string)
		usrName = fmt.Sprintf("%s...%s", idUsr[:6], idUsr[len(idUsr)-4:len(idUsr)])
	}

	nmbrNode := ctx.Params(":number")

	//--------------------------------------------------------------------------
	// если был POST! - редактирование свойств по доп.комиссии
	ctx.Req.ParseForm()
	ctx.Resp.WriteHeader(http.StatusOK)

	xPubKey := ctx.Req.PostFormValue("pubKey")
	xAddress := ctx.Req.PostFormValue("addressX")
	xCommission := ctx.Req.PostFormValue("commission")
	xStart := ctx.Req.PostFormValue("start")
	xFinish := ctx.Req.PostFormValue("finish")
	xTypeAct := ctx.Req.PostFormValue("typeAct")

	if auth == true && xTypeAct == "ADD" { //Добавляем в базу
		actBad := false
		layOut := "2006-01-02"
		dateStampStart, err := time.Parse(layOut, xStart)
		if err != nil {
			alertAct = "New address X"
			alertMsg = fmt.Sprintf("%s", err)
			alertType = "danger" //danger warning
			actBad = true
		}
		dateStampFinish, err := time.Parse(layOut, xFinish)
		if err != nil {
			alertAct = "New address X"
			alertMsg = fmt.Sprintf("%s", err)
			alertType = "danger" //danger warning
			actBad = true
		}
		xComsInt, err := strconv.Atoi(xCommission)
		if err != nil {
			alertAct = "New address X"
			alertMsg = fmt.Sprintf("%s", err)
			alertType = "danger" //danger warning
			actBad = true
		}

		if actBad != true {
			// проверить что условий для пользователя такого нет, в таком периоде!
			vN := []s.NodeUserX{}
			// Обратный запрос дат
			vN = srchNodeUserXPubkeyAddress(dbSQL, xPubKey, xAddress, dateStampFinish, dateStampStart, -1)

			if len(vN) > 0 {
				alertAct = "New address X"
				alertMsg = "уже есть!"
				alertType = "warning" //danger success
			} else {
				// добавляем
				newNUX := s.NodeUserX{}
				newNUX.Address = xAddress
				newNUX.PubKey = xPubKey
				newNUX.Commission = xComsInt
				newNUX.Start = dateStampStart
				newNUX.Finish = dateStampFinish

				if !addNodeUserX(dbSQL, &newNUX) {
					alertAct = "New address X"
					alertMsg = fmt.Sprintf("Add %s in %s", xAddress, xPubKey)
					alertType = "danger" //danger warning
				} else {
					alertAct = "Address X"
					alertMsg = "adding!"
					alertType = "success" //danger warning
				}
			}
		}
	} else if auth == true && xTypeAct == "DEL" {
		// TODO: удаление с базы
	} else if auth == true && xTypeAct == "EDIT" {
		iTitle := ctx.Req.PostFormValue("inputTitle")
		iWWW := ctx.Req.PostFormValue("inputWWW")
		iDesc := ctx.Req.PostFormValue("inputDescription")
		iIcon := ctx.Req.PostFormValue("inputIcon")

		vNd := s.NodeExt{}
		vNd = srchNodeSql_oa(dbSQL, nmbrNode, idUsr)

		if vNd.PubKey != "" {
			// новые данные
			vNd.ValidatorName = iTitle
			vNd.ValidatorURL = iWWW
			vNd.ValidatorLogoImg = iIcon
			vNd.ValidatorDesc = iDesc

			if !updNodeInfoRds_ext(dbSys, &vNd) {
				alertAct = "Edit"
				alertMsg = fmt.Sprintf("PubKey %s", vNd.PubKey)
				alertType = "danger" //danger warning
			} else {
				alertAct = "Edit"
				alertMsg = nmbrNode
				alertType = "success" //danger warning
			}
		}
	}

	// Нужно показывать как в листе нод так и при открытой странице
	newCommissionUser := int(100)
	if auth {
		vN := []s.NodeUserX{}
		nowT := time.Now()
		vN = srchNodeUserXPubkeyAddress(dbSQL, nmbrNode, idUsr, nowT, nowT, 0)

		if vN[0].Address != "" {
			newCommissionUser = vN[0].Commission
		}
	}

	v := NodeExt2{}
	rE := []ResReward{}
	// получаем данные по ноде:
	v.NodeExt, rE = GetNodeOneInfo(nmbrNode)

	// проверка что нода найдена!!! иначе на страницу 404
	if v.PubKey == "" {
		page404(ctx)
		return
	}

	v.CommissionNow = v.Commission
	// Изменение % ноды, если есть уникальное условие
	nodesNewPrc := srchNodeUserXAddress(dbSQL, "Mx----------------------------------------")
	dateNow := time.Now()
	for _, nux1 := range nodesNewPrc {
		if v.PubKey == nux1.PubKey {
			if nux1.Start.Unix() <= dateNow.Unix() && nux1.Finish.Unix() >= dateNow.Unix() {
				if v.CommissionNow > nux1.Commission {
					v.CommissionNow = nux1.Commission
				}
			}
		}
	}

	// Если авторизован пользователь и Если моя мастернода, То:
	// список пользователей с другими процентами!!! +/- пользователя
	listUserX := []s.NodeUserX{}
	myNode := false
	if auth {
		if v.OwnerAddress == idUsr {
			myNode = true
			listUserX = srchNodeUserXPubkey(dbSQL, nmbrNode)
		}
	}
	//--------------------------------------------------------------------------

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
	ctx.Data["Title"] = "Node"

	ctx.Data["OneNode"] = v
	ctx.Data["AllReward"] = rE
	ctx.Data["ListAddressX"] = listUserX
	ctx.Data["MyNode"] = myNode
	ctx.Data["MyCommission"] = newCommissionUser

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
	ctx.HTML(200, "node_info")
}

// Структура возврата блока в виде JSON
type RetJSONNode struct {
	Status bool `json:"status"`
	//Node    s.NodeExt   `json:"node"`
	Node    NodeExt2    `json:"node"`
	Rewards []ResReward `json:"reawrds"`
	ErrMsg  string      `json:"err_msg"`
}

// API: один блок
func hndAPIValidatorOneInfo(ctx *macaron.Context, sess session.Store) {
	retDt := RetJSONNode{}

	retDt.Status = true // исполнен без ошибок
	retDt.ErrMsg = ""   // нет ошибок

	nmbrNode := ctx.Params(":number")

	b0, r0 := GetNodeOneInfo(nmbrNode)

	// проверка что нода найдена!!!
	if b0.PubKey == "" {
		retDt.Status = false            // исполнен с ошибкой
		retDt.ErrMsg = "No search node" // текст ошибки
	}
	retDt.Node.NodeExt = b0
	retDt.Rewards = r0

	retDt.Node.CommissionNow = retDt.Node.Commission
	// Изменение % ноды, если есть уникальное условие
	nodesNewPrc := srchNodeUserXAddress(dbSQL, "Mx----------------------------------------")
	dateNow := time.Now()
	for _, nux1 := range nodesNewPrc {
		if retDt.Node.PubKey == nux1.PubKey {
			if nux1.Start.Unix() <= dateNow.Unix() && nux1.Finish.Unix() >= dateNow.Unix() {
				if retDt.Node.CommissionNow > nux1.Commission {
					retDt.Node.CommissionNow = nux1.Commission
				}
			}
		}
	}

	// возврат JSON данных
	ctx.JSON(200, &retDt)
}
