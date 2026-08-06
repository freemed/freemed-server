package api

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/freemed/freemed-server/common"
	"github.com/freemed/freemed-server/dbgen"
	"github.com/freemed/freemed-server/model"
	"github.com/gin-gonic/gin"
)

func init() {
	common.ApiMap["messages"] = common.ApiMapping{
		Authenticated: true,
		RouterFunction: func(r *gin.RouterGroup) {
			r.GET("/list_users", messagesListUsers)
			r.GET("/view", messagesView)
			r.GET("/view/:id", messageGet)
			r.POST("/send", messageSend)
		},
	}
}

type messagesUserObj struct {
	Username string `json:"username" binding:"required"`
	ID       int64  `json:"id" binding:"required"`
}

func messagesListUsers(r *gin.Context) {
	rows, err := model.Queries.MessagesListUsers(r.Request.Context())
	if err != nil {
		log.Print(err.Error())
		r.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	// Convert dbgen rows to messagesUserObj for API response
	o := make([]messagesUserObj, len(rows))
	for i, row := range rows {
		o[i] = messagesUserObj{Username: row.Username, ID: row.ID}
	}
	r.JSON(http.StatusOK, o)
}

func messagesView(r *gin.Context) {
	session, err := common.GetSession(r)
	if err != nil {
		log.Print(err.Error())
		r.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	unreadOnly, err := strconv.ParseBool(r.Query("unread_only"))
	if err != nil {
		unreadOnly = false
	}

	patient, err := strconv.ParseInt(r.Query("patient"), 10, 64)
	if err != nil {
		patient = 0
	}

	if patient != 0 {
		if unreadOnly {
			o, err := model.Queries.MessagesViewUnreadForPatient(r.Request.Context(), dbgen.MessagesViewUnreadForPatientParams{
				PatientID: patient,
				UserID:    session.UserId,
			})
			if err != nil {
				log.Print(err.Error())
				r.AbortWithError(http.StatusInternalServerError, err)
				return
			}
			r.JSON(http.StatusOK, o)
			return
		}
		o, err := model.Queries.MessagesViewForPatient(r.Request.Context(), dbgen.MessagesViewForPatientParams{
			PatientID: patient,
			UserID:    session.UserId,
		})
		if err != nil {
			log.Print(err.Error())
			r.AbortWithError(http.StatusInternalServerError, err)
			return
		}
		r.JSON(http.StatusOK, o)
		return
	}

	if unreadOnly {
		o, err := model.Queries.MessagesViewUnreadForUser(r.Request.Context(), session.UserId)
		if err != nil {
			log.Print(err.Error())
			r.AbortWithError(http.StatusInternalServerError, err)
			return
		}
		r.JSON(http.StatusOK, o)
		return
	}
	o, err := model.Queries.MessagesViewForUser(r.Request.Context(), session.UserId)
	if err != nil {
		log.Print(err.Error())
		r.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	r.JSON(http.StatusOK, o)
}

func messageGet(r *gin.Context) {
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
	if msg.Msgfor != session.UserId {
		log.Print("MessageGet(): not allowed")
		r.AbortWithError(http.StatusBadRequest, fmt.Errorf("not allowed"))
		return
	}

	r.JSON(http.StatusOK, msg)
}

func messageSend(r *gin.Context) {
	session, err := common.GetSession(r)
	if err != nil {
		log.Print(err.Error())
		r.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	log.Printf("MessageSend(): user=%d", session.UserId)

	var msg model.MessagesModel
	if err = r.BindJSON(&msg); err != nil {
		r.AbortWithError(http.StatusBadRequest, err)
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
}
