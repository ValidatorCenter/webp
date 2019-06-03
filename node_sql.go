package main

import (
	"fmt"

	s "github.com/ValidatorCenter/prs3r/strc"

	// база данных на SQL
	"github.com/jmoiron/sqlx"
	_ "github.com/kshvakov/clickhouse"
)

// Количество нод в базе
func srchNodeAmnt(db *sqlx.DB) int {
	iAmnt := 0
	err := db.Get(&iAmnt, "SELECT count() FROM nodes")
	if err != nil {
		log("ERR", fmt.Sprint("[node_sql.go] srchNodeAmnt(Get) - ERR:", err), "")
		panic(err) //dbg
	}
	return iAmnt
}

// Поиск всех нод пользователя
func srchNodeAddress(db *sqlx.DB, addrs string) []s.NodeExt {
	v := []s.NodeExt{}
	// Некоторая информация переехала в Redis! --> Sort("-status", "-total_stake_f32")
	err := db.Select(&v, fmt.Sprintf(`
		SELECT * 
		FROM nodes
		WHERE owner_address = '%s' OR reward_address = '%s'
	`, addrs, addrs))
	if err != nil {
		log("ERR", fmt.Sprint("[node_sql.go] srchNodeAddress(Select) - [addrs ", addrs, "] ERR:", err), "")
		panic(err) //dbg
		return v
	}
	return v
}

//Поиск ноды по паблику
func srchNodePubkey(db *sqlx.DB, pub_key string) s.NodeExt {
	v := s.NodeExt{}
	// Некоторая информация переехала в Redis!
	err := db.Get(&v, fmt.Sprintf(`
		SELECT * 
		FROM nodes
		WHERE pub_key = '%s'
	`, pub_key))
	if err != nil {
		if err.Error() == "sql: no rows in result set" {
			log("WRN", fmt.Sprint("[node_sql.go] srchNodePubkey(Get) - [pub_key ", pub_key, "] ERR:", err), "")
		} else {
			log("ERR", fmt.Sprint("[node_sql.go] srchNodePubkey(Get) - [pub_key ", pub_key, "] ERR:", err), "")
			panic(err) //dbg
		}
		return v
	}
	return v
}

// Поиск ноды по публичному ключу и основному адресу кошелька в SQL
func srchNodeSql_oa(db *sqlx.DB, pub_key string, owner_address string) s.NodeExt {
	p := s.NodeExt{}
	err := db.Get(&p, fmt.Sprintf(`
		SELECT * 
		FROM nodes FINAL 
		WHERE pub_key = '%s' AND owner_address = '%s'
		LIMIT 1
		`, pub_key, owner_address))
	if err != nil {
		if err.Error() == "sql: no rows in result set" {
			log("WRN", fmt.Sprint("[node_sql.go] srchNodeSql_oa(Select) - ", err), "")
		} else {
			log("ERR", fmt.Sprint("[node_sql.go] srchNodeSql_oa(Select) - ", err), "")
		}
		return s.NodeExt{}
	}
	return p
}

//Поиск стэк-ноды по паблику
func srchNodeStakeSql(db *sqlx.DB, pub_key string) []s.StakesInfo {
	v := []s.StakesInfo{}
	err := db.Select(&v, fmt.Sprintf(`
		SELECT owner_address as owner, coin, value_f32, bip_value_f32
		FROM node_stakes
		FINAL
		WHERE pub_key = '%s'
	`, pub_key))
	if err != nil {
		log("ERR", fmt.Sprint("[node_sql.go] srchNodeStake(Select) - [pub_key ", pub_key, "] ERR:", err), "")
		panic(err) //dbg
		return v
	}
	return v
}

//Поиск блоков-истории ноды по паблику
func srchNodeBlockstory(db *sqlx.DB, pub_key string) []s.BlocksStory {
	v := []s.BlocksStory{}
	err := db.Select(&v, fmt.Sprintf(`
		SELECT block_id, block_type
		FROM node_blockstory
		WHERE pub_key = '%s'
		ORDER BY block_id
	`, pub_key))
	if err != nil {
		log("ERR", fmt.Sprint("[node_sql.go] srchNodeBlockstory(Select) - [pub_key ", pub_key, "] ERR:", err), "")
		panic(err) //dbg
		return v
	}
	return v
}

// Список всех нод
func srchNodeList(db *sqlx.DB) []s.NodeExt { //, page int) []s.NodeExt {
	nodeM := []s.NodeExt{}

	// Некоторая информация переехала в Redis! status и total_stake_f32 в REDIS
	// поэтому берем все данные, а не только по страницам
	/*
		SELECT *
		FROM nodes b
		FINAL
		ORDER BY status, total_stake_f32 DESC
		LIMIT %d, 50
	*/
	//skipVldt := page * 50
	/*
		SELECT n.*, count() AS amnt_slots
			FROM nodes n FINAL
			LEFT JOIN node_stakes ns FINAL
			ON n.pub_key = ns.pub_key
			GROUP BY n.*
	*/
	err := db.Select(&nodeM, `
		SELECT *
		FROM nodes FINAL
	`)
	if err != nil {
		log("ERR", fmt.Sprint("[node_sql.go] srchNodeList(Select) - ERR:", err), "")
		panic(err) //dbg
	}

	return nodeM
}

type Delegate struct {
	Validator    string  `json:"pub_key" db:"pub_key"`
	ValidatorMin string  `json:"pub_key_min" db:"pub_key_min"`
	Coin         string  `json:"coin" db:"coin"`
	Value        float32 `json:"value_f32" db:"value_f32"`
	ValueBip     float32 `json:"bip_value_f32" db:"bip_value_f32"`
}

//...!тут STAKES.OWNER
// Поиск всех делегирований пользователя в ноды!
func srchNodeAddrsSql(db *sqlx.DB, addrs string) []Delegate {
	//..nodesCollection := sessMDB.DB("mvc_db").C("tabl_node")
	//nodesCollection.Find(bson.M{"stakes.owner": nmbrAddrs}).Sort("-$natural").All(&nodesDeleg)
	nodesDeleg := []Delegate{}
	err := db.Select(&nodesDeleg, fmt.Sprintf(`
		SELECT ns.pub_key,ns.coin,ns.value_f32,ns.bip_value_f32
		FROM node_stakes ns FINAL
		WHERE ns.owner_address = '%s'
	`, addrs))
	if err != nil {
		log("ERR", fmt.Sprint("[node_sql.go] srchNodeList(Select) - ERR:", err), "")
		panic(err) //dbg
	}

	return nodesDeleg
}

type NodeAmntInf struct {
	PubKey string `json:"pub_key" db:"pub_key"`
	Amnt   int    `json:"amnt" db:"amnt"`
}

//Поиск блоков-истории всех нод по типу
func srchNodeBlockstoryTypeAll(db *sqlx.DB, typeBlock string) []NodeAmntInf {
	v := []NodeAmntInf{}
	err := db.Select(&v, fmt.Sprintf(`
		SELECT pub_key, count() AS amnt
		FROM node_blockstory
		WHERE block_type='%s'
		GROUP BY pub_key
	`, typeBlock)) //AbsentBlock
	if err != nil {
		log("ERR", fmt.Sprint("[node_sql.go] srchNodeBlockstoryTypeAll(Select) - [typeBlock ", typeBlock, "] ERR:", err), "")
		panic(err) //dbg
		return v
	}
	return v
}

//Поиск занятых слотов всех нод
func srchNodeStakesAmntAll(db *sqlx.DB) []NodeAmntInf {
	v := []NodeAmntInf{}
	err := db.Select(&v, `
		SELECT pub_key, count() AS amnt
		FROM node_stakes FINAL
		GROUP BY pub_key
	`)
	if err != nil {
		log("ERR", fmt.Sprint("[node_sql.go] srchNodeStakesAmntAll(Select) - ERR:", err), "")
		panic(err) //dbg
		return v
	}
	return v
}
