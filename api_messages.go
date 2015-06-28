package main

import (
	"github.com/martini-contrib/encoder"
	"github.com/martini-contrib/render"
	"log"
	"net/http"
	"strconv"
)

type messagesUserObj struct {
	Username string `json:"username" binding:"required"`
	Id       string `json:"id" binding:"required"`
}

func MessagesListUsers(enc encoder.Encoder, r render.Render) {
	var o []messagesUserObj
	_, err := dbmap.Select(&o, "SELECT username, id FROM "+TABLE_USER)
	if err != nil {
		log.Print(err.Error())
		r.JSON(http.StatusInternalServerError, false)
		return
	}
	r.JSON(http.StatusOK, o)
	return
}

func MessagesView(req *http.Request, enc encoder.Encoder, r render.Render) {
	var o []MessagesModel
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
			_, err = dbmap.Select(&o, "SELECT * FROM "+TABLE_MESSAGES+" WHERE LENGTH(msgtag) < 1 AND msgpatient = ? AND msgread=0 AND msgtag=''", patient)
		} else {
			_, err = dbmap.Select(&o, "SELECT * FROM "+TABLE_MESSAGES+" WHERE LENGTH(msgtag) < 1 AND msgpatient = ?", patient)
		}
	} else {
		if unread_only {
			_, err = dbmap.Select(&o, "SELECT * FROM "+TABLE_MESSAGES+" WHERE LENGTH(msgtag) < 1 AND msgfor = ? AND msgread=0 AND msgtag=''", patient)
		} else {
			_, err = dbmap.Select(&o, "SELECT * FROM "+TABLE_MESSAGES+" WHERE LENGTH(msgtag) < 1 AND msgfor = ?", patient)
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
