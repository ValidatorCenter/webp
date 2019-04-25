package main

import (
	"fmt"
	"time"

	s "github.com/ValidatorCenter/prs3r/strc"

	// база данных на SQL
	"github.com/jmoiron/sqlx"
	_ "github.com/kshvakov/clickhouse"
)

// Поиск всех задач по ноде
func srchNodeTask(db *sqlx.DB, addrs string, nodes []s.NodeExt) []s.NodeTodo {
	allList := []s.NodeTodo{}
	strWhere := ""
	for _, sM := range nodes {
		strWhere = fmt.Sprintf("%spub_key = '%s' OR ", strWhere, sM.PubKey)
	}
	if strWhere != "" {
		strWhere = fmt.Sprintf("%saddress = '%s'", strWhere, addrs)
	} else {
		strWhere = fmt.Sprintf("address = '%s'", addrs)
	}

	err := db.Select(&allList, fmt.Sprintf(`
		SELECT * 
		FROM node_tasks FINAL
		WHERE %s
	`, strWhere))
	if err != nil {
		log("ERR", fmt.Sprint("[node_task_sql.go] srchNodeTask(Select) - [addrs ", addrs, "] ERR:", err), "")
		panic(err) //dbg
		return allList
	}

	return allList
}

// Поиск задач по адресу
func srchNodeTaskAddress(db *sqlx.DB, addrs string) []s.NodeTodo {
	allList := []s.NodeTodo{}

	err := db.Select(&allList, fmt.Sprintf(`
		SELECT * 
		FROM node_tasks FINAL
		WHERE address = '%s'
	`, addrs))
	if err != nil {
		log("ERR", fmt.Sprint("[node_task_sql.go] srchNodeTaskAddress(Select) - [addrs ", addrs, "] ERR:", err), "")
		panic(err) //dbg
		return allList
	}

	return allList
}

// Поиск не исполненых задач по ноде
func srchNodeTaskNoDone(db *sqlx.DB, addrs string, nodes []s.NodeExt) []s.NodeTodo {
	//maxTodo := 100
	list100 := []s.NodeTodo{}

	strWhere := ""
	for _, sM := range nodes {
		if strWhere != "" {
			strWhere = fmt.Sprintf("%s OR pub_key = '%s'", strWhere, sM.PubKey)
		} else {
			strWhere = fmt.Sprintf("%spub_key = '%s'", strWhere, sM.PubKey)
		}
	}
	// надо разделить, если это владелец ноды, то все АДРЕСА, если не владелец, то только адрес
	if strWhere != "" {
		// ЭТО владелец ноды, значит не ограничиваемся по адресу!

		// TODO: а если владелец ноды, только для теста, а потом не использует а делегирует в другие ноды?!
		// FIXME: strWhere = fmt.Sprintf("%saddress = '%s'", strWhere, addrs)
	} else {
		strWhere = fmt.Sprintf("address = '%s'", addrs)
	}

	// порядок: дата, очередность, не исполненные -> Sort("priority", "created")
	/*err := db.Select(&list100, fmt.Sprintf(`
		SELECT *
		FROM node_tasks FINAL
		WHERE done=0 AND (%s)
		ORDER BY priority, created
		LIMIT %d
	`, strWhere, maxTodo))*/ // будем брать всё, а потом разбирать в цикле
	err := db.Select(&list100, fmt.Sprintf(`
		SELECT * 
		FROM node_tasks FINAL
		WHERE done=0 AND (%s)
		ORDER BY priority, created
	`, strWhere))
	if err != nil {
		log("ERR", fmt.Sprint("[node_task_sql.go] srchNodeTaskNoDone(Select) - [addrs ", addrs, "] ERR:", err), "")
		panic(err) //dbg
		return list100
	}

	return list100
}

// Обновление задачи
func updNodeTask(db *sqlx.DB, newData s.NodeTodo) error {
	var err error

	err = addNodeTaskSql(db, &newData)

	return err
}

// Добавить задачу для ноды в SQL
func addNodeTaskSql(db *sqlx.DB, dt *s.NodeTodo) error {
	var err error
	tx := db.MustBegin()

	dt.UpdYCH = time.Now()

	qPg := `
		INSERT INTO node_tasks (
			priority,
			done,
			created,
			donet,
			type,
			height_i32,
			pub_key,
			address,
			amount_f32,
			comment,
			tx_hash,
			updated_date
		) VALUES (
			:priority,
			:done,
			:created,
			:donet,
			:type,
			:height_i32,
			:pub_key,
			:address,
			:amount_f32,
			:comment,
			:tx_hash,
			:updated_date
		)`

	_, err = tx.NamedExec(qPg, &dt)
	if err != nil {
		log("ERR", fmt.Sprint("[node_task_sql.go] addNodeTaskSql(NamedExec) -", err), "")
		return err
	}
	log("INF", "INSERT", fmt.Sprint("node-task ", dt.Address, " ", dt.PubKey))

	err = tx.Commit()
	if err != nil {
		log("ERR", fmt.Sprint("[node_task_sql.go] addNodeTaskSql(Commit() -", err), "")
		return err
	}
	return err
}
