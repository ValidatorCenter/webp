package main

import (
	"fmt"

	"github.com/go-redis/redis"
)

// Поместить задачу в MEM
func setATasksMem(db *redis.Client, hashID string, dataJSON string) bool {
	err := db.Set(hashID, dataJSON, 0).Err()
	if err != nil {
		log("ERR", fmt.Sprint("[node_task_redis.go] setATasksMem(set...", hashID, ") - ", err), "")
		return false
	}
	return true
}

// Получить задачу из MEM
func getATasksMem(db *redis.Client, hashID string) string {
	var _lbRes string
	_lbRes, err := db.Get(hashID).Result()
	if err == redis.Nil {
		log("WRN", fmt.Sprint("[node_task_redis.go] getATasksMem(Get...", hashID, ") - ", "=0!"), "")
	} else if err != nil {
		log("ERR", fmt.Sprint("[node_task_redis.go] getATasksMem(Get...", hashID, ") - ", err), "")
	}
	return _lbRes
}

// Удалить задачу из MEM
func delATasksMem(db *redis.Client, hashID string) bool {
	return setATasksMem(db, hashID, "")
}
