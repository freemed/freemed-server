package api

import (
	"context"
	"database/sql"
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

// messagesDeleteRows is a query seam. It exists so the delete handler can be
// tested — the M5 regression is "can another user's message be deleted?", which
// must be answered without a live database. Same pattern as api/dicom.go's
// dicomGetRow seam.
var messagesDeleteRows = func(ctx context.Context, arg dbgen.DeleteMessagesForUserParams) (sql.Result, error) {
	return model.Queries.DeleteMessagesForUser(ctx, arg)
}

func init() {
	common.ApiMap["messages"] = common.ApiMapping{
		Authenticated: true,
		RouterFunction: func(r *gin.RouterGroup) {
			r.GET("/list_users", messagesListUsers)
			r.GET("/view", messagesView)
			r.GET("/view/:id", messageGet)
			r.GET("/tags", messagesListTags)
			r.GET("/tag/:tag", messagesByTag)
			r.POST("/send", messageSend)
			r.POST("/delete", messagesDelete)
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
		common.ErrorResponseFromError(r, http.StatusInternalServerError, err)
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
		common.ErrorResponseFromError(r, http.StatusInternalServerError, err)
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

	offset, limit := pageParams(r)

	ctx := r.Request.Context()

	// Every branch uses the paginated query plus its companion COUNT. The
	// non-paginated variants loaded the user's entire message history into
	// memory and then sliced it, so the row count — not the page size — was the
	// real bound (M7).
	var (
		messages interface{}
		total    int64
	)
	if patient != 0 {
		if unreadOnly {
			messages, err = model.Queries.MessagesViewUnreadForPatientPaginated(ctx, dbgen.MessagesViewUnreadForPatientPaginatedParams{
				PatientID: patient,
				UserID:    session.UserId,
				Limit:     limit,
				Offset:    offset,
			})
			if err == nil {
				total, err = model.Queries.CountMessagesUnreadForPatient(ctx, dbgen.CountMessagesUnreadForPatientParams{
					PatientID: patient,
					UserID:    session.UserId,
				})
			}
		} else {
			messages, err = model.Queries.MessagesViewForPatientPaginated(ctx, dbgen.MessagesViewForPatientPaginatedParams{
				PatientID: patient,
				UserID:    session.UserId,
				Limit:     limit,
				Offset:    offset,
			})
			if err == nil {
				total, err = model.Queries.CountMessagesForPatient(ctx, dbgen.CountMessagesForPatientParams{
					PatientID: patient,
					UserID:    session.UserId,
				})
			}
		}
	} else if unreadOnly {
		messages, err = model.Queries.MessagesViewUnreadForUserPaginated(ctx, dbgen.MessagesViewUnreadForUserPaginatedParams{
			UserID: session.UserId,
			Limit:  limit,
			Offset: offset,
		})
		if err == nil {
			total, err = model.Queries.CountMessagesUnreadForUser(ctx, session.UserId)
		}
	} else {
		messages, err = model.Queries.MessagesViewForUserPaginated(ctx, dbgen.MessagesViewForUserPaginatedParams{
			UserID: session.UserId,
			Limit:  limit,
			Offset: offset,
		})
		if err == nil {
			total, err = model.Queries.CountMessagesForUser(ctx, session.UserId)
		}
	}

	if err != nil {
		log.Print(err.Error())
		common.ErrorResponseFromError(r, http.StatusInternalServerError, err)
		return
	}

	r.JSON(http.StatusOK, gin.H{
		"data":   messagesToSlice(messages),
		"total":  total,
		"offset": offset,
		"limit":  limit,
	})
}

// messagesToSlice normalizes a sqlc message result set into a []interface{} so
// an empty page serializes as [] rather than null, and so the response has the
// same shape for all four message filters.
func messagesToSlice(messages interface{}) []interface{} {
	switch v := messages.(type) {
	case []dbgen.MessagesViewForUserPaginatedRow:
		out := make([]interface{}, len(v))
		for i, row := range v {
			out[i] = row
		}
		return out
	case []dbgen.MessagesViewUnreadForUserPaginatedRow:
		out := make([]interface{}, len(v))
		for i, row := range v {
			out[i] = row
		}
		return out
	case []dbgen.MessagesViewForPatientPaginatedRow:
		out := make([]interface{}, len(v))
		for i, row := range v {
			out[i] = row
		}
		return out
	case []dbgen.MessagesViewUnreadForPatientPaginatedRow:
		out := make([]interface{}, len(v))
		for i, row := range v {
			out[i] = row
		}
		return out
	default:
		return []interface{}{}
	}
}

func messageGet(r *gin.Context) {
	session, err := common.GetSession(r)
	if err != nil {
		log.Print(err.Error())
		common.ErrorResponseFromError(r, http.StatusInternalServerError, err)
		return
	}

	idString := r.Param("id")
	if idString == "" {
		log.Print("MessageGet(): No id provided")
		common.ErrorResponse(r, http.StatusInternalServerError, "internal server error")
		return
	}

	id, err := strconv.ParseInt(idString, 10, 64)
	if err != nil {
		log.Print(err.Error())
		common.ErrorResponseFromError(r, http.StatusInternalServerError, err)
		return
	}

	msg, err := model.MessageById(id)
	if err != nil {
		log.Print(err.Error())
		common.ErrorResponseFromError(r, http.StatusInternalServerError, err)
		return
	}

	// Access control: do not allow access from other user
	if msg.Msgfor != session.UserId {
		log.Print("MessageGet(): not allowed")
		common.ErrorResponse(r, http.StatusBadRequest, "not allowed")
		return
	}

	r.JSON(http.StatusOK, msg)
}

func messageSend(r *gin.Context) {
	session, err := common.GetSession(r)
	if err != nil {
		log.Print(err.Error())
		common.ErrorResponseFromError(r, http.StatusInternalServerError, err)
		return
	}
	log.Printf("MessageSend(): user=%d", session.UserId)

	var msg model.MessagesModel
	if err = r.BindJSON(&msg); err != nil {
		common.ErrorResponseFromError(r, http.StatusBadRequest, err)
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
		common.ErrorResponseFromError(r, http.StatusInternalServerError, err)
		return
	}

	r.JSON(http.StatusOK, true)
}

// messagesListTags returns all distinct message tags
func messagesListTags(r *gin.Context) {
	tags, err := model.Queries.ListMessageTags(r.Request.Context())
	if err != nil {
		log.Print(err.Error())
		common.ErrorResponseFromError(r, http.StatusInternalServerError, err)
		return
	}
	// Convert []sql.NullString to []string, skipping empty/nil
	out := make([]string, 0, len(tags))
	for _, t := range tags {
		if t.Valid && t.String != "" {
			out = append(out, t.String)
		}
	}
	r.JSON(http.StatusOK, out)
}

// messagesByTag returns messages matching a specific tag
func messagesByTag(r *gin.Context) {
	tag := r.Param("tag")
	if tag == "" {
		r.AbortWithStatus(http.StatusBadRequest)
		return
	}
	messages, err := model.Queries.MessagesByTag(r.Request.Context(), sql.NullString{String: tag, Valid: true})
	if err != nil {
		log.Print(err.Error())
		common.ErrorResponseFromError(r, http.StatusInternalServerError, err)
		return
	}
	r.JSON(http.StatusOK, messages)
}

type messagesDeleteRequest struct {
	IDs []int64 `json:"ids" binding:"required"`
}

// messagesDelete performs a bulk delete of the caller's own messages.
//
// The ids arrive in the request body and are NOT trusted to belong to the
// caller: the delete is scoped to the session user in SQL
// (`DELETE FROM messages WHERE msgfor = ? AND id IN (...)`), so a non-owned id
// is simply not matched. messageGet in this file already enforced the same
// `msg.Msgfor == session.UserId` rule for single-message reads.
//
// Semantics: non-owned (and non-existent) ids are silently skipped rather than
// rejected. A per-id 403 would turn the endpoint into an existence oracle for
// other users' message ids, and a mixed batch would fail wholesale, leaving the
// UI unable to delete the ids the caller does own. The response therefore
// reports the number of rows actually deleted rather than `true`, so a caller
// that asked to delete 5 and got `deleted: 2` can see that 3 were not theirs.
func messagesDelete(r *gin.Context) {
	session, err := common.GetSession(r)
	if err != nil {
		log.Print(err.Error())
		common.ErrorResponseFromError(r, http.StatusUnauthorized, err)
		return
	}

	var req messagesDeleteRequest
	if err := r.BindJSON(&req); err != nil {
		common.ErrorResponseFromError(r, http.StatusBadRequest, err)
		return
	}
	if len(req.IDs) == 0 {
		common.ErrorResponse(r, http.StatusBadRequest, "no message ids supplied")
		return
	}

	result, err := messagesDeleteRows(r.Request.Context(), dbgen.DeleteMessagesForUserParams{
		UserID: session.UserId,
		Ids:    req.IDs,
	})
	if err != nil {
		log.Print(err.Error())
		common.ErrorResponseFromError(r, http.StatusInternalServerError, err)
		return
	}

	affected, err := result.RowsAffected()
	if err != nil {
		log.Print(err.Error())
		common.ErrorResponseFromError(r, http.StatusInternalServerError, err)
		return
	}

	r.JSON(http.StatusOK, gin.H{
		"status":    "deleted",
		"deleted":   affected,
		"requested": int64(len(req.IDs)),
	})
}
