package api

import (
	"fmt"
	"github.com/freemed/freemed-server/common"
	"github.com/freemed/freemed-server/model"
	"github.com/go-martini/martini"
	"github.com/martini-contrib/binding"
	"github.com/martini-contrib/encoder"
	"github.com/martini-contrib/render"
	"log"
	"net/http"
	"strconv"
	"time"
)

func init() {
	common.ApiMap["messages"] = common.ApiMapping{
		Authenticated: true,
		JsonArmored:   true,
		RouterFunction: func(r martini.Router) {
			r.Get("/list_users", MessagesListUsers)
			r.Get("/view", MessagesView)
			r.Get("/view/:id", MessageGet)
			r.Post("/send", binding.Json(model.MessagesModel{}, MessageSend))
		},
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

func MessagesView(req *http.Request, enc encoder.Encoder, r render.Render, session common.SessionModel) {
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
			_, err = model.DbMap.Select(&o, "SELECT m.*, u.userdescrip AS 'sender' FROM "+model.TABLE_MESSAGES+" m LEFT OUTER JOIN "+model.TABLE_USER+" u ON u.id = m.msgby WHERE (ISNULL(m.msgtag) OR LENGTH(m.msgtag) < 1) AND m.msgpatient = ? AND m.msgread=0 AND m.msgby = ?", patient, session.UserId)
		} else {
			_, err = model.DbMap.Select(&o, "SELECT m.*, u.userdescrip AS 'sender' FROM "+model.TABLE_MESSAGES+" m LEFT OUTER JOIN "+model.TABLE_USER+" u ON u.id = m.msgby WHERE (ISNULL(m.msgtag) OR LENGTH(m.msgtag) < 1) AND m.msgpatient = ? AND m.msgfor = ?", patient, session.UserId)
		}
	} else {
		if unread_only {
			_, err = model.DbMap.Select(&o, "SELECT m.*, u.userdescrip AS 'sender' FROM "+model.TABLE_MESSAGES+" m LEFT OUTER JOIN "+model.TABLE_USER+" u ON u.id = m.msgby WHERE (ISNULL(m.msgtag) OR LENGTH(m.msgtag) < 1) AND m.msgfor = ? AND m.msgread = 0", session.UserId)
		} else {
			_, err = model.DbMap.Select(&o, "SELECT m.*, u.userdescrip AS 'sender' FROM "+model.TABLE_MESSAGES+" m LEFT OUTER JOIN "+model.TABLE_USER+" u ON u.id = m.msgby WHERE (ISNULL(m.msgtag) OR LENGTH(m.msgtag) < 1) AND m.msgfor = ?", session.UserId)
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

func MessageGet(session common.SessionModel, params martini.Params, enc encoder.Encoder, r render.Render) {
	var idString string
	var ok bool
	if idString, ok = params["id"]; !ok {
		log.Print("MessageGet(): No id provided")
		r.JSON(http.StatusInternalServerError, false)
		return
	}

	id, err := strconv.ParseInt(idString, 10, 64)
	if err != nil {
		log.Print(err.Error())
		r.JSON(http.StatusInternalServerError, false)
		return
	}

	msg, err := model.MessageById(id)
	if err != nil {
		log.Print(err.Error())
		r.JSON(http.StatusInternalServerError, false)
		return
	}

	// Access control: do not allow access from other user
	if msg.For != session.UserId {
		log.Print("MessageGet(): not allowed")
		r.JSON(http.StatusInternalServerError, false)
		return
	}

	r.JSON(http.StatusOK, msg)
	return
}

func MessageSend(msg model.MessagesModel, session common.SessionModel, enc encoder.Encoder, r render.Render) {
	log.Printf("MessageSend(): user=%d", session.UserId)

	// Ensure that we can't send as any other user
	msg.Sender = session.UserId

	// Set time to be now
	msg.Sent = time.Now()

	// Set unique key
	msg.Unique = model.NewNullStringValue(fmt.Sprintf("%d", time.Now().Unix()))

	err := model.MessageSend(msg)
	if err != nil {
		log.Print(err)
		r.JSON(http.StatusInternalServerError, false)
		return
	}

	r.JSON(http.StatusOK, true)
	return
}
