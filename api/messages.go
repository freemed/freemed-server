package api

import (
	"fmt"
	"github.com/freemed/freemed-server/common"
	"github.com/freemed/freemed-server/model"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
	"strconv"
	"time"
)

func init() {
	common.ApiMap["messages"] = common.ApiMapping{
		Authenticated: true,
		RouterFunction: func(r *gin.RouterGroup) {
			r.GET("/list_users", MessagesListUsers)
			r.GET("/view", MessagesView)
			r.GET("/view/:id", MessageGet)
			r.POST("/send", MessageSend)
		},
	}
}

type messagesUserObj struct {
	Username string `json:"username" binding:"required"`
	Id       string `json:"id" binding:"required"`
}

func MessagesListUsers(r *gin.Context) {
	var o []messagesUserObj
	_, err := model.DbMap.Select(&o, "SELECT username, id FROM "+model.TABLE_USER)
	if err != nil {
		log.Print(err.Error())
		r.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	r.JSON(http.StatusOK, o)
	return
}

func MessagesView(r *gin.Context) {
	session, err := common.GetSession(r)
	if err != nil {
		log.Print(err.Error())
		r.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	var o []model.MessagesModel

	unread_only, err := strconv.ParseBool(r.Query("unread_only"))
	if err != nil {
		unread_only = false
	}

	patient, err := strconv.ParseInt(r.Query("patient"), 10, 64)
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
		r.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	r.JSON(http.StatusOK, o)
	return
}

func MessageGet(r *gin.Context) {
	session, err := common.GetSession(r)
	if err != nil {
		log.Print(err.Error())
		r.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	idString := r.Param("id")
	if idString == "" {
		log.Print("MessageGet(): No id provided")
		r.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	id, err := strconv.ParseInt(idString, 10, 64)
	if err != nil {
		log.Print(err.Error())
		r.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	msg, err := model.MessageById(id)
	if err != nil {
		log.Print(err.Error())
		r.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	// Access control: do not allow access from other user
	if msg.For != session.UserId {
		log.Print("MessageGet(): not allowed")
		r.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	r.JSON(http.StatusOK, msg)
	return
}

func MessageSend(r *gin.Context) {
	session, err := common.GetSession(r)
	if err != nil {
		log.Print(err.Error())
		r.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	log.Printf("MessageSend(): user=%d", session.UserId)

	var msg model.MessagesModel
	if r.BindJSON(&msg) == nil {
		r.AbortWithStatus(http.StatusBadRequest)
		return
	}

	// Ensure that we can't send as any other user
	msg.Sender = session.UserId

	// Set time to be now
	msg.Sent = time.Now()

	// Set unique key
	msg.Unique = model.NewNullStringValue(fmt.Sprintf("%d", time.Now().Unix()))

	err = model.MessageSend(msg)
	if err != nil {
		log.Print(err)
		r.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	r.JSON(http.StatusOK, true)
	return
}
