package main

import (
	"fmt"
	"strconv"
	"strings"

	s "github.com/ValidatorCenter/prs3r/strc"
	"github.com/go-redis/redis"
)

// Получение данных Системной таблицы из SQL
func srchSysSql(db *redis.Client) s.StatusDB {
	p := s.StatusDB{}

	strc := "latest_block_save,latest_block_cmc,latest_block_vld,latest_block_rwd"

	arr := strings.Split(strc, ",")
	len_arr := len(arr)
	if len_arr == 0 {
		log("ERR", fmt.Sprint("[mem_sys.go] srchSysSql(len(arr)) - ", "=0!"), "")
		return s.StatusDB{}
	}

	for i := 0; i < len_arr; i++ {
		switch arr[i] {
		case "latest_block_save":
			_lbRes, err := db.Get(arr[i]).Result()
			if err == redis.Nil {
				p.LatestBlockSave = 0
				log("WRN", fmt.Sprint("[mem_sys.go] srchSysSql(Get...", arr[i], ") - ", "=0!"), "")
			} else if err != nil {
				log("ERR", fmt.Sprint("[mem_sys.go] srchSysSql(Get...", arr[i], ") - ", err), "")
				return s.StatusDB{}
			}
			if _lbRes != "" {
				p.LatestBlockSave, err = strconv.Atoi(_lbRes)
				if err != nil {
					log("ERR", fmt.Sprint("[mem_sys.go] strconv.Atoi(", arr[i], ") - ", err), "")
					return s.StatusDB{}
				}
			} else {
				log("WRN", fmt.Sprint("[mem_sys.go] Нет в MEM - ", arr[i], " => будет значит 0 "), "")
				p.LatestBlockSave = 0
			}
		case "latest_block_cmc":
			_lbRes, err := db.Get(arr[i]).Result()
			if err == redis.Nil {
				p.LatestBlockCMC = 0
				log("WRN", fmt.Sprint("[mem_sys.go] srchSysSql(Get...", arr[i], ") - ", "=0!"), "")
			} else if err != nil {
				log("ERR", fmt.Sprint("[mem_sys.go] srchSysSql(Get...", arr[i], ") - ", err), "")
				return s.StatusDB{}
			}
			if _lbRes != "" {
				p.LatestBlockCMC, err = strconv.Atoi(_lbRes)
				if err != nil {
					log("ERR", fmt.Sprint("[mem_sys.go] strconv.Atoi(", arr[i], ") - ", err), "")
					return s.StatusDB{}
				}
			} else {
				log("WRN", fmt.Sprint("[mem_sys.go] Нет в MEM - ", arr[i], " => будет значит 0 "), "")
				p.LatestBlockCMC = 0
			}
		case "latest_block_vld":
			_lbRes, err := db.Get(arr[i]).Result()
			if err == redis.Nil {
				p.LatestBlockVld = 0
				log("WRN", fmt.Sprint("[mem_sys.go] srchSysSql(Get...", arr[i], ") - ", "=0!"), "")
			} else if err != nil {
				log("ERR", fmt.Sprint("[mem_sys.go] srchSysSql(Get...", arr[i], ") - ", err), "")
				return s.StatusDB{}
			}
			if _lbRes != "" {
				p.LatestBlockVld, err = strconv.Atoi(_lbRes)
				if err != nil {
					log("ERR", fmt.Sprint("[mem_sys.go] strconv.Atoi(", arr[i], ") - ", err), "")
					return s.StatusDB{}
				}
			} else {
				log("WRN", fmt.Sprint("[mem_sys.go] Нет в MEM - ", arr[i], " => будет значит 0 "), "")
				p.LatestBlockVld = 0
			}
		case "latest_block_rwd":
			_lbRes, err := db.Get(arr[i]).Result()
			if err == redis.Nil {
				p.LatestBlockRwd = 0
				log("WRN", fmt.Sprint("[mem_sys.go] srchSysSql(Get...", arr[i], ") - ", "=0!"), "")
			} else if err != nil {
				log("ERR", fmt.Sprint("[mem_sys.go] srchSysSql(Get...", arr[i], ") - ", err), "")
				return s.StatusDB{}
			}
			if _lbRes != "" {
				p.LatestBlockRwd, err = strconv.Atoi(_lbRes)
				if err != nil {
					log("ERR", fmt.Sprint("[mem_sys.go] strconv.Atoi(", arr[i], ") - ", err), "")
					return s.StatusDB{}
				}
			} else {
				log("WRN", fmt.Sprint("[mem_sys.go] Нет в MEM - ", arr[i], " => будет значит 0 "), "")
				p.LatestBlockRwd = 0
			}
		}
	}
	return p
}
