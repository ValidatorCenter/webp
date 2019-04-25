package main

import (
	"fmt"
	"os"
	"strings"

	// web-framework
	"github.com/go-macaron/session"
	"gopkg.in/ini.v1"
	"gopkg.in/macaron.v1"

	// база данных на SQL
	"github.com/jmoiron/sqlx"
	_ "github.com/kshvakov/clickhouse"

	// база данных на Redis
	"github.com/go-redis/redis"

	ms "github.com/ValidatorCenter/minter-go-sdk"
	//s "github.com/ValidatorCenter/prs3r/strc"
)

// Глобально установленные объекты
var (
	CoinMinter string // Основная монета Minter
	srvAuth    string // сервер авторизации
	sdk        ms.SDK

	dbSQL *sqlx.DB
	dbSys *redis.Client
)

// СТРАНИЦА: Лендиг приветствия
func myHello(ctx *macaron.Context, sess session.Store) {
	var alertMsg, alertType, alertAct string
	var auth bool = false
	var usrName, idUsr string

	//SESSION
	if sess.Get("login") != nil {
		auth = true
		idUsr = sess.Get("login").(string)
		if len(idUsr) > 0 {
			usrName = fmt.Sprintf("%s...%s", idUsr[:6], idUsr[len(idUsr)-4:len(idUsr)])
		}
	}

	//GET
	alertType = ctx.QueryEscape("alert")
	alertAct = ctx.QueryEscape("act")
	alertMsg = ctx.QueryEscape("msg")

	// Заголовк страницы:
	ctx.Data["Title"] = "Hello"

	// Инф.сообщения от системы:
	ctx.Data["AlertAct"] = alertAct
	ctx.Data["AlertMsg"] = alertMsg
	ctx.Data["AlertType"] = alertType

	// Пользователь:
	ctx.Data["UsrAuth"] = auth
	ctx.Data["UsrName"] = usrName
	ctx.Data["UsrAddress"] = idUsr

	// Вывод страницы:
	ctx.HTML(200, "hello")
}

// СТРАНИЦА: Нет такой страницы - 404
func page404(ctx *macaron.Context) {
	// Последний синхронизированный блок
	ResultNetwork, _ := sdk.GetStatus()
	// получаем системуную коллекцию из Redis
	statusMDB := srchSysSql(dbSys)

	// Инф. о синхронизации БД с БлокЧейном:
	ctx.Data["LastSync"] = statusMDB.LatestBlockSave
	ctx.Data["Current"] = ResultNetwork.LatestBlockHeight
	if sdk.ChainMainnet {
		ctx.Data["ChainNet"] = "mainnet"
	} else {
		ctx.Data["ChainNet"] = "testnet"
	}

	//ctx.Error(404, "Нет такой страницы!")
	ctx.Data["Title"] = "404"
	ctx.HTML(200, "_404")
}

// Главая функция! Вход в программу
func main() {
	var err error
	ConfFileName := "webd.ini"
	srvAuth = "http://localhost:3999" // default

	if len(os.Args) == 2 {
		ConfFileName = os.Args[1]
	}

	//////////////////////////////////////////////////////
	// INI
	cfg, err := ini.LoadSources(ini.LoadOptions{IgnoreInlineComment: true}, ConfFileName)
	if err != nil {
		log("ERR", fmt.Sprintf("Загрузка INI '%s' - %s", ConfFileName, err.Error()), "")
		log("INF", "webd /path/file_config.ini", "")
		log("INF", "ИЛИ берётся webd.ini в директории программы", "")
		return
	} else {
		log("OK", fmt.Sprintf("Данные с INI '%s' файла - загружены!", ConfFileName), "")
	}
	secMN := cfg.Section("masternode")
	sdk.MnAddress = secMN.Key("ADDRESS").String()

	_strChain := secMN.Key("CHAIN").String()
	if strings.ToLower(_strChain) == "main" {
		sdk.ChainMainnet = true
	} else {
		sdk.ChainMainnet = false
	}

	CoinMinter = ms.GetBaseCoin()
	secSite := cfg.Section("site")
	SiteTemplatesDir := secSite.Key("TemplateDir").String()
	if SiteTemplatesDir == "" {
		// берем по умолчанию если нет в конфиге
		SiteTemplatesDir = "templates"
	}
	SitePublicDir := secSite.Key("PublicDir").String()
	if SitePublicDir == "" {
		SitePublicDir = "public"
	}

	secAuth := cfg.Section("auth")
	srvAuth = secAuth.Key("ADDRESS").String()

	secDB := cfg.Section("database")
	//////////////////////////////////////////////////////
	// DB:: Redis
	r_db, err := secDB.Key("REDIS_DB").Int()
	if err != nil {
		r_db = 0
	}
	dbSys = redis.NewClient(&redis.Options{
		Addr:     secDB.Key("REDIS_ADDRESS").String(),
		Password: secDB.Key("REDIS_PSWRD").String(), // no password set
		DB:       r_db,                              // use default DB
	})
	defer dbSys.Close()
	log("OK", "Подключились к БД - Redis", "")

	//////////////////////////////////////////////////////
	// DB:: ClickHouse
	dbSQL, err = sqlx.Open("clickhouse", secDB.Key("CLICKHOUSE_ADDRESS").String())
	if err != nil {
		log("ERR", fmt.Sprint("Подключение к БД - ClickHouse ", err), "")
	}
	defer dbSQL.Close()
	log("OK", "Подключились к БД - ClickHouse", "")

	//////////////////////////////////////////////////////
	// MACARON web-framework
	m := macaron.Classic()
	m.Use(macaron.Renderer(macaron.RenderOptions{
		Directory:  SiteTemplatesDir,
		Extensions: []string{".html"},
	}))
	m.Use(macaron.Static(SitePublicDir))
	m.Use(session.Sessioner(session.Options{
		Provider: "memory", // FIXME: переделать на Redis!
		// Cookie name to save session ID. Default is "MacaronSession".
		CookieName: "MVCSession",
	}))

	m.Get("/", myHello)    // Лендинг страница
	m.Get("/404", page404) // ТЕСТ: нет такой страницы
	// API v1
	m.Get("/api/v1/newMnemonic", hndAPINewMnemonic)                              // новая seed-фраза, регистрация нового аккаунта в сети Minter
	m.Get("/api/v1/autoDeleg/:number", hndAPIAutoDelegAddress)                   // получить JSON конфигурацию автоделегирования
	m.Get("/api/v1/autoTaskOut/:tokenauth/:pubkey", hndAPIAutoTodo)              // получить JSON задачи на исполнение (max=100 за раз)
	m.Get("/api/v1/autoTaskIn/:tokenauth/:returndataJSON", hndAPIAutoTodoReturn) // результат выполнения автоделегатор

	// о сессиях тут -> https://go-macaron.com/docs/middlewares/session
	//m.Route("/auth", "GET,POST", hndAuth)
	m.Post("/auth", hndAuthUser)    // регистрация и авторизация
	m.Get("/logout", hndLogoutUser) // выход с сеанса

	// Эксплорер::Блоки
	m.Get("/blocks", hndBlocksInfo)                    // лист блоков
	m.Get("/blocks/:pgn", hndBlocksInfo)               // лист блоков
	m.Get("/block/:number", hndBlockOneInfo)           // 1 блок
	m.Get("/api/v1/block/:number", hndAPIBlockOneInfo) // API v1: 1 блок JSON
	//TODO: получить лист блоков по API в виде JSON

	// Эксплорер::Транизакции
	m.Get("/transactions", hndTransactionsInfo)                    // Страница транзакций
	m.Get("/transactions/:pgn", hndTransactionsInfo)               // Страница транзакций
	m.Get("/transaction/:number", hndTransactionOneInfo)           // 1 транзакция
	m.Get("/api/v1/transaction/:number", hndAPITransactionOneInfo) // API v1: 1 транзакция JSON
	//TODO: получить список транзакций по API в виде JSON

	// Валидаторы/Ноды
	m.Get("/nodes", hndValidatorsInfo)                        // Страница валидаторов
	m.Get("/nodes/:pgn", hndValidatorsInfo)                   // Страница валидаторов
	m.Route("/node/:number", "GET,POST", hndValidatorOneInfo) // 1 валидатор
	m.Get("/api/v1/node/:number", hndAPIValidatorOneInfo)     // API v1: 1 валидатор JSON
	//TODO: получить список нод/валидаторов по API в виде JSON

	// Адрес кошелька
	m.Get("/address/:number", hndAddressOneInfo)      // 1 адрес
	m.Get("/address/:number/:pgn", hndAddressOneInfo) // 1 адрес
	//TODO: получить данные адреса по API в виде JSON

	// Монеты
	m.Get("/coins", hndCoins)                                  // Страница монет
	m.Route("/coin/:ticker", "GET,POST", hndCoinInfo)          // 1 монета
	m.Route("/coin/:ticker/:ticker2", "GET,POST", hndCoinInfo) // 1 пара монет
	//TODO: получить лист Монет по API в виде JSON
	//TODO: получить данные монеты по API в виде JSON

	// ЛК::Кошелёк
	m.Route("/sendcoin", "GET,POST", hndWalletTxSend)         // TX: Отправить монету
	m.Route("/delegation", "GET,POST", hndWalletTxDelegation) // TX: Делегирование монет
	m.Route("/masternode", "GET,POST", hndWalletTxMasternode) // TX: Управление нодой
	m.Route("/coiner", "GET,POST", hndWalletTxCoiner)         // TX: Создание монет
	m.Route("/checks", "GET,POST", hndWalletTxChecks)         // TX: Управление чеками
	m.Get("/tasklist", hndWalletListTask)                     // Страница листа задач
	m.Get("/tasklist/:number", hndWalletListTaskAddrs)        // Страница листа задач по адресу

	//TODO: Выполнение транзакции SEND, DELEG, DECLARE/START/STOP-NODE, COINER, CREATE/ACT-CHECK по API в виде JSON
	//TODO: получить список задач Task по API в виде JSON

	//[+] convert конвертация - Обмен на другие монеты (coins - уже есть)
	//[?] checks чеки - Создание и обналичивание
	//[+] delegation делегирование - Делегирование и отзыв
	//[+] masternode мастернода - Декларирование, включение и отключение
	//[+] coiner создание монет - Создание монет)
	//[-] профиль, - адрес, seed фраза и приваткей

	m.Run()
}
