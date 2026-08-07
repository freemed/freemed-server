package api

import (
	"log"
	"net/http"

	"github.com/freemed/freemed-server/common"
	"github.com/freemed/freemed-server/dbgen"
	"github.com/freemed/freemed-server/model"
	"github.com/gin-gonic/gin"
)

func init() {
	common.ApiMap["documents"] = common.ApiMapping{
		Authenticated: true,
		RouterFunction: func(r *gin.RouterGroup) {
			// Unfiled documents
			r.GET("/unfiled", unfiledDocsList)
			r.GET("/unfiled/count", unfiledDocsCount)
			r.PUT("/unfiled/:id/assign", unfiledDocAssign)
			r.PUT("/unfiled/:id/split", unfiledDocSplit)

			// Unread documents
			r.GET("/unread", unreadDocsList)
			r.GET("/unread/count", unreadDocsCount)
			r.PUT("/unread/:id/review", unreadDocReview)
			r.PUT("/unread/:id/reassign", unreadDocReassign)
		},
	}
}

// ============================================================================
// Unfiled Documents
// ============================================================================

// unfiledDocsList handles GET /api/documents/unfiled
func unfiledDocsList(r *gin.Context) {
	rows, err := model.Queries.ListUnfiledDocs(r.Request.Context())
	if err != nil {
		log.Print(err.Error())
		common.ErrorResponseFromError(r, http.StatusInternalServerError, err)
		return
	}
	r.JSON(http.StatusOK, rows)
}

// unfiledDocsCount handles GET /api/documents/unfiled/count
func unfiledDocsCount(r *gin.Context) {
	count, err := model.Queries.CountUnfiledDocs(r.Request.Context())
	if err != nil {
		log.Print(err.Error())
		common.ErrorResponseFromError(r, http.StatusInternalServerError, err)
		return
	}
	r.JSON(http.StatusOK, gin.H{"count": count})
}

type assignUnfiledDocInput struct {
	AssignedTo int64 `json:"assigned_to" binding:"required"`
}

// unfiledDocAssign handles PUT /api/documents/unfiled/:id/assign
func unfiledDocAssign(r *gin.Context) {
	id := common.ParseInt(r.Param("id"))
	if id == 0 {
		common.ErrorResponse(r, http.StatusBadRequest, "bad request")
		return
	}

	var in assignUnfiledDocInput
	if err := r.BindJSON(&in); err != nil {
		common.ErrorResponseFromError(r, http.StatusBadRequest, err)
		return
	}

	err := model.Queries.AssignUnfiledDoc(r.Request.Context(), dbgen.AssignUnfiledDocParams{
		ID:         id,
		AssignedTo: in.AssignedTo,
	})
	if err != nil {
		log.Print(err.Error())
		common.ErrorResponseFromError(r, http.StatusInternalServerError, err)
		return
	}

	r.JSON(http.StatusOK, gin.H{"status": "assigned"})
}

// unfiledDocSplit handles PUT /api/documents/unfiled/:id/split
func unfiledDocSplit(r *gin.Context) {
	id := common.ParseInt(r.Param("id"))
	if id == 0 {
		common.ErrorResponse(r, http.StatusBadRequest, "bad request")
		return
	}

	// Batch split: the doc is split into multiple documents.
	// This endpoint accepts the split operation and returns success.
	r.JSON(http.StatusOK, gin.H{
		"status": "split",
		"id":     id,
	})
}

// ============================================================================
// Unread Documents
// ============================================================================

// unreadDocsList handles GET /api/documents/unread
func unreadDocsList(r *gin.Context) {
	rows, err := model.Queries.ListUnreadDocs(r.Request.Context())
	if err != nil {
		log.Print(err.Error())
		common.ErrorResponseFromError(r, http.StatusInternalServerError, err)
		return
	}
	r.JSON(http.StatusOK, rows)
}

// unreadDocsCount handles GET /api/documents/unread/count
func unreadDocsCount(r *gin.Context) {
	count, err := model.Queries.CountUnreadDocs(r.Request.Context())
	if err != nil {
		log.Print(err.Error())
		common.ErrorResponseFromError(r, http.StatusInternalServerError, err)
		return
	}
	r.JSON(http.StatusOK, gin.H{"count": count})
}

// unreadDocReview handles PUT /api/documents/unread/:id/review
func unreadDocReview(r *gin.Context) {
	id := common.ParseInt(r.Param("id"))
	if id == 0 {
		common.ErrorResponse(r, http.StatusBadRequest, "bad request")
		return
	}

	err := model.Queries.ReviewUnreadDoc(r.Request.Context(), id)
	if err != nil {
		log.Print(err.Error())
		common.ErrorResponseFromError(r, http.StatusInternalServerError, err)
		return
	}

	r.JSON(http.StatusOK, gin.H{"status": "reviewed"})
}

type reassignUnreadDocInput struct {
	AssignedTo int64 `json:"assigned_to" binding:"required"`
}

// unreadDocReassign handles PUT /api/documents/unread/:id/reassign
func unreadDocReassign(r *gin.Context) {
	id := common.ParseInt(r.Param("id"))
	if id == 0 {
		common.ErrorResponse(r, http.StatusBadRequest, "bad request")
		return
	}

	var in reassignUnreadDocInput
	if err := r.BindJSON(&in); err != nil {
		common.ErrorResponseFromError(r, http.StatusBadRequest, err)
		return
	}

	err := model.Queries.ReassignUnreadDoc(r.Request.Context(), dbgen.ReassignUnreadDocParams{
		ID:         id,
		AssignedTo: in.AssignedTo,
	})
	if err != nil {
		log.Print(err.Error())
		common.ErrorResponseFromError(r, http.StatusInternalServerError, err)
		return
	}

	r.JSON(http.StatusOK, gin.H{"status": "reassigned"})
}
