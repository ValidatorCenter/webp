package main

import (
	"fmt"
	"strconv"
	"time"

	s "github.com/ValidatorCenter/prs3r/strc"
	"github.com/go-redis/redis"
)

// Поиск информации о Монете
func srchCoinInfoRds(db *redis.Client, dt *s.CoinMarketCapData) bool {
	if dt.CoinSymbol != "" {

		_lbRes, err := db.HGetAll(fmt.Sprintf("%s_info", dt.CoinSymbol)).Result()
		if err != nil {
			log("ERR", fmt.Sprint("[coin_redis.go] srchCoinInfoRds(hgetall...", dt.CoinSymbol, ") - ", err), "")
			return false
		}
		// Всё "хорошо", заносим в dt новые данные
		dt.CoinURL = _lbRes["coin_url"]
		dt.CoinLogoImg = _lbRes["coin_logo_img"]
		dt.CoinDesc = _lbRes["coin_desciption"]
		dt.TimeUpdate, _ = time.Parse(time.RFC3339, _lbRes["time_update"])
		dt_VolumeNow, _ := strconv.ParseFloat(_lbRes["volume_now_f32"], 32)
		dt.VolumeNow = float32(dt_VolumeNow)
		dt_ReserveBalanceNow, _ := strconv.ParseFloat(_lbRes["reserve_balance_now_f32"], 32)
		dt.ReserveBalanceNow = float32(dt_ReserveBalanceNow)
		dt.AmntTrans24x7, _ = strconv.Atoi(_lbRes["amnt_trans_24x7"])

	} else {
		log("ERR", "[coin_redis.go] srchCoinInfoRds(...)  = '???'", "")
		return false
	}

	return true
}

// Получить движения за 24ч пары монет (строго заданного направления) из Транзакций [нужен: объём движения, % и цена в USD]
func srchCoin2_1_24(db *redis.Client, ticker1 string, ticker2 string) []s.PairCoins {
	return srchCoin2_2_24(db, ticker1, ticker2, false)
}

// Получить движения за 24ч пары монет (не зависемо от направления) из Транзакций [нужен: объём движения, % и цена в USD]
func srchCoin2_2_24(db *redis.Client, coin1 string, coin2 string, multi bool) []s.PairCoins {
	/*
		УСЛОВИЯ ЗАПРОСА:
		Получаем: PriceBuy,PriceSell,Change24,Volume24
		Volume24 - объём за 24 часа
		Change24 - изменение цены за 24 часа относительно старой цены (нужна старая цена и новая{текущая}) считаем: ((НоваяЦена-СтараяЦена)/СтараяЦена)*100%
		Транзакции покупки и продажи: (2)SellCoin, (3)SellAllCoin, (4)BuyCoin
		Период: последние 24 часа
		Заданная пара монет
	*/

	retDt := []s.PairCoins{}
	if coin1 != "" && coin2 != "" {
		_lbRes1, err := db.HGetAll(fmt.Sprintf("%s%s_c2c", coin1, coin2)).Result()
		if err != nil {
			log("ERR", fmt.Sprint("[coin_redis.go] srchCoin2_2_24(coin1,coin2>", coin1, "-", coin2, ") - ", err), "")
			return retDt
		}
		_lbRes2, err := db.HGetAll(fmt.Sprintf("%s%s_c2c", coin2, coin1)).Result()
		if err != nil && multi {
			log("ERR", fmt.Sprint("[coin_redis.go] srchCoin2_2_24(coin2,coin1>", coin2, "-", coin1, ") - ", err), "")
			return retDt
		}
		//.....
		if _lbRes1["coin_to_buy"] != "" && _lbRes1["coin_to_sell"] != "" {
			rDt1 := s.PairCoins{}

			rDt1.CoinToBuy = _lbRes1["coin_to_buy"]
			rDt1.CoinToSell = _lbRes1["coin_to_sell"]

			rDt1_PriceBuy, _ := strconv.ParseFloat(_lbRes1["price_buy_f32"], 32)
			rDt1.PriceBuy = float32(rDt1_PriceBuy)

			rDt1_PriceSell, _ := strconv.ParseFloat(_lbRes1["price_sell_f32"], 32)
			rDt1.PriceSell = float32(rDt1_PriceSell)

			rDt1_Volume24, _ := strconv.ParseFloat(_lbRes1["volume_24_f32"], 32)
			rDt1.Volume24 = float32(rDt1_Volume24)

			rDt1_Change24, _ := strconv.ParseFloat(_lbRes1["change_24_f32"], 32)
			rDt1.Change24 = float32(rDt1_Change24)

			rDt1.TimeUpdate, _ = time.Parse(time.RFC3339, _lbRes1["time_update"])

			retDt = append(retDt, rDt1)
		}

		if _lbRes2["coin_to_buy"] != "" && _lbRes2["coin_to_sell"] != "" && multi {
			rDt1 := s.PairCoins{}

			rDt1.CoinToBuy = _lbRes2["coin_to_buy"]
			rDt1.CoinToSell = _lbRes2["coin_to_sell"]

			rDt1_PriceBuy, _ := strconv.ParseFloat(_lbRes2["price_buy_f32"], 32)
			rDt1.PriceBuy = float32(rDt1_PriceBuy)

			rDt1_PriceSell, _ := strconv.ParseFloat(_lbRes2["price_sell_f32"], 32)
			rDt1.PriceSell = float32(rDt1_PriceSell)

			rDt1_Volume24, _ := strconv.ParseFloat(_lbRes2["volume_24_f32"], 32)
			rDt1.Volume24 = float32(rDt1_Volume24)

			rDt1_Change24, _ := strconv.ParseFloat(_lbRes2["change_24_f32"], 32)
			rDt1.Change24 = float32(rDt1_Change24)

			rDt1.TimeUpdate, _ = time.Parse(time.RFC3339, _lbRes2["time_update"])

			retDt = append(retDt, rDt1)
		}

	} else {
		log("ERR", "[coin_redis.go] srchCoin2_2_24(...) Coin_1 = '???' & Coin_2 = '???'", "")
	}
	return retDt
}

// Обновить информацию о 3-х View записях в Монете (coin_url, coin_logo_img, coin_desciption)
func updCoinInfoRds_3v(db *redis.Client, dt *s.CoinMarketCapData) bool {
	if dt.CoinSymbol != "" {
		m2 := map[string]interface{}{
			"coin_url":        dt.CoinURL,
			"coin_logo_img":   dt.CoinLogoImg,
			"coin_desciption": dt.CoinDesc,
		}

		err := db.HMSet(fmt.Sprintf("%s_info", dt.CoinSymbol), m2).Err()
		if err != nil {
			log("ERR", fmt.Sprint("[coin_redis.go] updCoinInfoRds_3v(hmset...", dt.CoinSymbol, ") - ", err), "")
			return false
		}
	} else {
		log("ERR", "[coin_redis.go] updCoinInfoRds_3v(...) Coin = '???'", "")
		return false
	}
	log("INF", "UPDATE", fmt.Sprintf("%s_info(3v)", dt.CoinSymbol))
	return true
}
