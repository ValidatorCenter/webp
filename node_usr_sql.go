package main

import (
	"fmt"
	"time"

	s "github.com/ValidatorCenter/prs3r/strc"

	// база данных на SQL
	"github.com/jmoiron/sqlx"
	_ "github.com/kshvakov/clickhouse"
)

// Поиск акции для Адреса-кошелька по ноде(pub_key), адресу и промежетку во времени
//если limit=-1 то без лимит, если 0, то если ничего нет, в нулевую структуру вставить пустой элемент!!! 1 и более, как обычно
func srchNodeUserXPubkeyAddress(db *sqlx.DB, pub_key string, address string, start time.Time, finish time.Time, limit int) []s.NodeUserX {

	// если limit=-1 то без лимит, если 0, то если ничего нет, в нулевую структуру вставить пустой элемент!!! 1 и более, как обычно
	// проверить что условий для пользователя такого нет, в таком периоде!
	limStr := "" // if limit == -1 { limStr="" }
	if limit == -1 || limit == 0 {
		limStr = ""
	} else if limit > 0 {
		limStr = fmt.Sprintf(" LIMIT = %d", limit)
	}

	vN := []s.NodeUserX{}
	err := db.Select(&vN, fmt.Sprintf(`
		SELECT * FROM node_userx 
		WHERE pub_key = '%s' AND 
			address = '%s' AND 
			start <= '%s' AND 
			finish >= '%s'%s
		`, pub_key, address, start.Format("2006-01-02 15:04:05"), finish.Format("2006-01-02 15:04:05"), limStr))
	if err != nil {
		log("ERR", fmt.Sprint("[node_usr_sql.go] srchNodeUserXPubkeyAddress(Select) - [pub_key №", pub_key, "] ERR:", err), "")
		panic(err) //dbg
	}

	if limit == 0 && len(vN) == 0 {
		vN = append(vN, s.NodeUserX{})
	}

	return vN
}

// Берем всех пользователей с новым % из SQL
func srchNodeUserXSql(db *sqlx.DB) []s.NodeUserX {
	pp := []s.NodeUserX{}
	err := db.Select(&pp, "SELECT * FROM node_userx")
	if err != nil {
		log("ERR", fmt.Sprint("[node_usr_sql.go] srchNodeUserXSql(Select) - ", err), "")
		return []s.NodeUserX{}
	}
	return pp
}

// Поиск акции для Адреса-кошелька
func srchNodeUserXAddress(db *sqlx.DB, address string) []s.NodeUserX {
	listUserX := []s.NodeUserX{}
	err := db.Select(&listUserX, fmt.Sprintf("SELECT * FROM node_userx WHERE address = '%s'", address))
	if err != nil {
		log("ERR", fmt.Sprint("[node_usr_sql.go] srchNodeUserXAddress(Select) - [address №", address, "] ERR:", err), "")
		panic(err) //dbg
	}
	return listUserX
}

// Поиск акции для Адреса-кошелька по ноде(pub_key)
func srchNodeUserXPubkey(db *sqlx.DB, pub_key string) []s.NodeUserX {
	listUserX := []s.NodeUserX{}
	err := db.Select(&listUserX, fmt.Sprintf("SELECT * FROM node_userx WHERE pub_key = '%s'", pub_key))
	if err != nil {
		log("ERR", fmt.Sprint("[node_usr_sql.go] srchNodeUserXPubkey(Select) - [pub_key №", pub_key, "] ERR:", err), "")
		panic(err) //dbg
	}
	return listUserX
}

// Добавляем условие акции для кошелька(пользователя)
func addNodeUserX(db *sqlx.DB, dt *s.NodeUserX) bool {
	var err error
	tx := db.MustBegin()

	dt.UpdYCH = time.Now()

	qPg := `
		INSERT INTO node_userx (
			pub_key,
			address,
			start,
			finish,
			commission,
			updated_date
		) VALUES (
			:pub_key,
			:address,
			:start,
			:finish,
			:commission,
			:updated_date
		)`

	_, err = tx.NamedExec(qPg, &dt)
	if err != nil {
		log("ERR", fmt.Sprint("[node_usr_sql.go] addNodeUserX(NamedExec) -", err), "")
		return false
	}
	log("INF", "INSERT", fmt.Sprint("node-userx ", dt.Address, " ", dt.PubKey))

	err = tx.Commit()
	if err != nil {
		log("ERR", fmt.Sprint("[node_usr_sql.go] addNodeUserX(Commit) -", err), "")
		return false
	}
	return true
}
