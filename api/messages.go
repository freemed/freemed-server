package api

import (
	"github.com/freemed/freemed-server/model"
	"github.com/go-martini/martini"
	"github.com/martini-contrib/encoder"
	"github.com/martini-contrib/render"
	"log"
	"net/http"
	"strconv"
)

func init() {
	model.ApiMap["messages"] = func(r martini.Router) {
		r.Get("/list_users", MessagesListUsers)
		r.Get("/view", MessagesView)
	}
}

type messagesUserObj struct {
	Username string `json:"username" binding:"required"`
	Id       string `json:"id" binding:"required"`
}

func MessagesListUsers(enc encoder.Encoder, r render.Render) {
	var o []messagesUserObj
	_, err := model.DbMap.Select(&o, "SELECT username, id FROM "+model.TABLE_USER)
	if err != nil {
		log.Print(err.Error())
		r.JSON(http.StatusInternalServerError, false)
		return
	}
	r.JSON(http.StatusOK, o)
	return
}

func MessagesView(req *http.Request, enc encoder.Encoder, r render.Render) {
	var o []model.MessagesModel
	q := req.URL.Query()

	unread_only, err := strconv.ParseBool(q.Get("unread_only"))
	if err != nil {
		unread_only = false
	}

	patient, err := strconv.ParseInt(q.Get("patient"), 10, 64)
	if err != nil {
		patient = 0
	}

	if patient != 0 {
		if unread_only {
			_, err = model.DbMap.Select(&o, "SELECT * FROM "+model.TABLE_MESSAGES+" WHERE LENGTH(msgtag) < 1 AND msgpatient = ? AND msgread=0 AND msgtag=''", patient)
		} else {
			_, err = model.DbMap.Select(&o, "SELECT * FROM "+model.TABLE_MESSAGES+" WHERE LENGTH(msgtag) < 1 AND msgpatient = ?", patient)
		}
	} else {
		if unread_only {
			_, err = model.DbMap.Select(&o, "SELECT * FROM "+model.TABLE_MESSAGES+" WHERE LENGTH(msgtag) < 1 AND msgfor = ? AND msgread=0 AND msgtag=''", patient)
		} else {
			_, err = model.DbMap.Select(&o, "SELECT * FROM "+model.TABLE_MESSAGES+" WHERE LENGTH(msgtag) < 1 AND msgfor = ?", patient)
		}
	}

	if err != nil {
		log.Print(err.Error())
		r.JSON(http.StatusInternalServerError, false)
		return
	}
	r.JSON(http.StatusOK, o)
	return
}
