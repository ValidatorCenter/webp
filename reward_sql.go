package main

import (
	"fmt"
	"time"

	// база данных на SQL
	"github.com/jmoiron/sqlx"
	_ "github.com/kshvakov/clickhouse"
)

// Награда
type ResReward struct {
	Role    string    `json:"role" bson:"role" db:"role"`
	Address string    `json:"address" bson:"address" db:"address"`
	PubKey  string    `json:"pub_key" bson:"pub_key" db:"pub_key"`
	Amnt    float32   `json:"amount_f32" bson:"amount_f32" db:"amount_f32"`
	UpdYCH  time.Time `json:"-" bson:"-" db:"updated_date"` // ClickHouse::UpdateDate
}

// Возвращает все награды по Паблику валидатора
func srchRewardPubkeySql(db *sqlx.DB, pub_key string) []ResReward {
	rE := []ResReward{}
	// Суммировать !!!
	err := db.Select(&rE, fmt.Sprintf(`
		SELECT role, address, validator_pub_key as pub_key, sum(amount_f32) AS amount_f32 
		FROM block_event 
		WHERE validator_pub_key = '%s' AND type = 'minter/RewardEvent'
		GROUP BY role, address ,pub_key
	`, pub_key))
	if err != nil {
		log("ERR", fmt.Sprint("[reward_sql.go] srchRewardPubkeySql(Select) - [pub_key №", pub_key, "] ERR:", err), "")
		panic(err) //dbg
	}
	return rE
}

// Возвращает все награды по Адресу кошелька
func srchRewardAddrsSql(db *sqlx.DB, addrs string) []ResReward {
	rE := []ResReward{}
	err := db.Select(&rE, fmt.Sprintf(`
		SELECT role, address, validator_pub_key as pub_key, sum(amount_f32) AS amount_f32
		FROM block_event 
		WHERE address = '%s' AND type = 'minter/RewardEvent'
		GROUP BY role, address ,pub_key
	`, addrs))
	if err != nil {
		log("ERR", fmt.Sprint("[sql_reward.go] srchRewardAddrsSql(Select) - [addrs №", addrs, "] ERR:", err), "")
		panic(err) //dbg
	}
	return rE
}
