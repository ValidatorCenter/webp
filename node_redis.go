package main

import (
	"fmt"
	"strconv"
	"time"

	s "github.com/ValidatorCenter/prs3r/strc"
	"github.com/go-redis/redis"
)

// Поиск информации о Ноде
func srchNodeInfoRds(db *redis.Client, dt *s.NodeExt) bool {
	if dt.PubKey != "" {
		_lbRes, err := db.HGetAll(fmt.Sprintf("%s_info", dt.PubKey)).Result()
		if err != nil {
			log("ERR", fmt.Sprint("[node_redis.go] updNodeInfoRds(hgetall...", dt.PubKey, ") - ", err), "")
			return false
		}
		// Всё "хорошо", заносим в dt новые данные
		dt.ValidatorAddress = _lbRes["validator_address"]
		dt.ValidatorName = _lbRes["validator_name"]
		dt.ValidatorURL = _lbRes["validator_url"]
		dt.ValidatorLogoImg = _lbRes["validator_logo_img"]
		dt.ValidatorDesc = _lbRes["validator_desciption"]
		dt_Uptime, _ := strconv.ParseFloat(_lbRes["uptime"], 32)
		dt.Uptime = float32(dt_Uptime)
		dt.StatusInt, _ = strconv.Atoi(_lbRes["status"])
		dt.TimeUpdate, _ = time.Parse(time.RFC3339, _lbRes["time_update"])
		dt_AmntBlocks, _ := strconv.Atoi(_lbRes["amnt_blocks"]) // Количество подписанных блоков
		dt.AmntBlocks = uint64(dt_AmntBlocks)
		dt.AmntSlashed, _ = strconv.Atoi(_lbRes["amnt_slashed"])
		dt_TotalStake, _ := strconv.ParseFloat(_lbRes["total_stake_f32"], 32)
		dt.TotalStake = float32(dt_TotalStake)
	} else {
		log("ERR", "[node_redis.go] srchNodeInfoRds(...) PubKey = 0", "")
		return false
	}

	return true
}

// Обновить информационную запись о Ноде (сокращенная)
func updNodeInfoRds_ext(db *redis.Client, dt *s.NodeExt) bool {
	var err error

	if dt.PubKey != "" {
		m2 := map[string]interface{}{
			//"validator_address":    dt.WWW,
			"validator_name":       dt.ValidatorName,
			"validator_url":        dt.ValidatorURL,
			"validator_logo_img":   dt.ValidatorLogoImg,
			"validator_desciption": dt.ValidatorDesc,
		}
		err = db.HMSet(fmt.Sprintf("%s_info", dt.PubKey), m2).Err()
		if err != nil {
			log("ERR", fmt.Sprint("[node_redis.go] updNodeInfoRds_ext(hmset...", dt.PubKey, ") - ", err), "")
			return false
		}
	} else {
		log("ERR", "[node_redis.go] updNodeInfoRds_ext(...) PubKey = 0", "")
		return false
	}

	return true
}
