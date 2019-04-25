package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	//"strings"
	"net/url"

	"github.com/go-macaron/session"
	"gopkg.in/macaron.v1"
)

type RetJSONSeed struct {
	Status   bool   `json:"status"`
	Mnemonic string `json:"mnemonic"`
	ErrMsg   string `json:"err_msg"`
}

type RetJSONPriv struct {
	RetJSONSeed
	Address string `json:"address"`
	Privkey string `json:"priv_key"`
}

// СТРАНИЦА: Регистрация и авторизация
func hndAuthUser(ctx *macaron.Context, sess session.Store) {
	var alertMsg, alertType, alertAct string

	ctx.Req.ParseForm()
	ctx.Resp.WriteHeader(http.StatusOK)
	seedPhr := ctx.Req.PostFormValue("seed_phrase")

	if seedPhr != "" {
		// возможно авторизация, но всё равно проверим авторизованы ли мы уже?!
		if sess.Get("login") == nil {
			// не авторизованы

			retDt := RetJSONPriv{}

			// GET
			//urlS := fmt.Sprintf("%s/api/v1/authSeed?sp=%s", srvAuth, strings.Replace(seedPhr, " ", "%20", -1))
			//res1, err := http.Get(urlS)

			// POST
			formData := url.Values{
				"sp": {seedPhr},
			}
			res1, err := http.PostForm(fmt.Sprintf("%s/api/v1/authSeed", srvAuth), formData)

			if err != nil {
				log("ERR", fmt.Sprint("[auth.go] hndAuthUser(PostForm) - ERR:", err), "")
				alertAct = "Auth"
				alertMsg = "Authorization server is not responding"
				alertType = "danger" //danger warning success

				// Инф.сообщения от системы:
				ctx.Data["AlertAct"] = alertAct
				ctx.Data["AlertMsg"] = alertMsg
				ctx.Data["AlertType"] = alertType

				// редирект
				ctx.HTML(200, "auth_redir") // на странице так-же поля авторизации
				return

			}
			defer res1.Body.Close()

			body, err := ioutil.ReadAll(res1.Body)
			if err != nil {
				log("ERR", fmt.Sprint("[auth.go] hndAuthUser(ReadAll) - ERR:", err), "")
				alertAct = "Auth"
				alertMsg = "Authorization Server - Unknown Response"
				alertType = "danger" //danger warning success

				// Инф.сообщения от системы:
				ctx.Data["AlertAct"] = alertAct
				ctx.Data["AlertMsg"] = alertMsg
				ctx.Data["AlertType"] = alertType

				// редирект
				ctx.HTML(200, "auth_redir") // на странице так-же поля авторизации
				return

			}

			json.Unmarshal(body, &retDt)

			if retDt.Privkey != "" {
				sess.Set("login", retDt.Address) // TODO: можно с приватного ключа брать!
				sess.Set("priv_k", retDt.Privkey)
			} else {
				alertAct = "Auth"
				alertMsg = "Not logged in"
				alertType = "danger" //danger warning success
			}

		}
	} else {
		alertAct = "Auth"
		alertMsg = "No Seed-phrases"
		alertType = "danger" //danger warning success
	}

	// Инф.сообщения от системы:
	ctx.Data["AlertAct"] = alertAct
	ctx.Data["AlertMsg"] = alertMsg
	ctx.Data["AlertType"] = alertType

	// редирект
	ctx.HTML(200, "auth_redir") // на странице так-же поля авторизации
}

// СТРАНИЦА: Выход с сеанса
func hndLogoutUser(ctx *macaron.Context, sess session.Store) {
	if sess.Get("login") != nil {
		sess.Delete("login")
		sess.Delete("priv_k")
	}
	// редирект
	ctx.HTML(200, "auth_redir") // на странице так-же поля авторизации
}

// API: Возвращает JSON новую seed-фразу
func hndAPINewMnemonic(ctx *macaron.Context) {
	retDt := RetJSONSeed{}

	urlS := fmt.Sprintf("%s/api/v1/newMnemonic", srvAuth)
	res, err := http.Get(urlS)
	if err != nil {
		retDt.Status = false
		retDt.ErrMsg = err.Error()
		ctx.JSON(200, &retDt)
		return
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		retDt.Status = false
		retDt.ErrMsg = err.Error()
		ctx.JSON(200, &retDt)
		return
	}

	json.Unmarshal(body, &retDt)

	// возврат JSON данных
	ctx.JSON(200, &retDt)
}
