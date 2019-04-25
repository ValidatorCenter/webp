package main

import (
	"fmt"

	s "github.com/ValidatorCenter/prs3r/strc"

	// база данных на SQL
	"github.com/jmoiron/sqlx"
	_ "github.com/kshvakov/clickhouse"
)

// FIXME: Переделать из SQL в MEM.DB

// Поиск правил для Автоделегатора
func srchAutodeleg(db *sqlx.DB, xPubKey string, xAddress string, xCoin string) []s.AutodelegCfg {
	vN := []s.AutodelegCfg{}
	err := db.Select(&vN, fmt.Sprintf(`
		SELECT *
		FROM autodeleg FINAL
		WHERE pub_key='%s' AND address='%s' AND coin='%s'
	`, xPubKey, xAddress, xCoin))
	if err != nil {
		if err.Error() == "sql: no rows in result set" {
			log("WRN", fmt.Sprint("[wallet_deleg_sql.go] srchAutodeleg(Select) - ERR:", err), "")
		} else {
			log("ERR", fmt.Sprint("[wallet_deleg_sql.go] srchAutodeleg(Select) - ERR:", err), "")
			panic(err) //dbg
		}
		return vN
	}
	return vN
}

// Поиск правил по адресу для Автоделегатора
func srchAutodelegAddress(db *sqlx.DB, idUsr string) []s.AutodelegCfg {
	vN := []s.AutodelegCfg{}
	err := db.Select(&vN, fmt.Sprintf(`
		SELECT *
		FROM autodeleg FINAL
		WHERE address='%s'
	`, idUsr))
	if err != nil {
		if err.Error() == "sql: no rows in result set" {
			log("WRN", fmt.Sprint("[wallet_deleg_sql.go] srchAutodelegAddress(Select) - ERR:", err), "")
		} else {
			log("ERR", fmt.Sprint("[wallet_deleg_sql.go] srchAutodelegAddress(Select) - ERR:", err), "")
			panic(err) //dbg
		}
		return vN
	}
	return vN
}

// Если при вставке указать version = -1, запись будет удалена.
// При значениях version = 1 запись будет оставлена в таблице ОБНОВЛЕНА.

// Добавить новое правило для Автоделегатора
func addAutodelegSql(db *sqlx.DB, newData s.AutodelegCfg) error {
	var err error

	/*newNUX.Address = xAddress
	newNUX.PubKey = xPubKey
	newNUX.Coin = xCoin
	newNUX.WalletPrc = xAmntPrc

	err = adlgColl.Insert(newNUX)*/

	return err
}

// Обновление правила для Автоделегатора
func updAutodelegSql(db *sqlx.DB, updData s.AutodelegCfg) error {
	var err error

	/*
		qUx.Address = xAddress
		qUx.PubKey = xPubKey
		qUx.Coin = xCoin
	*/
	// err = adlgColl.Update(qUx, bson.M{"$set": bson.M{"wallet_prc": xAmntPrc}})

	// 1 Поиск правила --- srchAutodeleg
	// 2 Обновления правила -- addAutodelegSql(...

	return err
}

// Удаление правила для Автоделегатора
func delAutodelegSql(db *sqlx.DB, xPubKey string, xAddress string, xCoin string) error {
	var err error

	// err = adlgColl.Remove(qUx)

	return err
}
