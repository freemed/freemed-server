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
	common.ApiMap["scanned-documents"] = common.ApiMapping{
		Authenticated: true,
		RouterFunction: func(r *gin.RouterGroup) {
			r.GET("/:id", scannedDocGet)
		},
	}
}

// patientScannedDocsList handles GET /api/patient/:id/scanned-documents
//
// The array shape is kept (api/portal.go portalDocuments returns the same
// ListScannedDocs rows as a bare array and is consumed by the portal clients),
// so the bound is applied server-side: ?offset=/?limit= clamped, default 50,
// maximum 200. The query is now `LIMIT ? OFFSET ?` instead of returning every
// scanned document the patient has ever had.
func patientScannedDocsList(r *gin.Context) {
	id := r.Param("id")
	if id == "" {
		r.AbortWithStatus(http.StatusBadRequest)
		return
	}

	offset, limit := pageParams(r)

	patientID := common.ParseInt(id)
	docs, err := model.Queries.ListScannedDocs(r.Request.Context(), dbgen.ListScannedDocsParams{
		PatientID: patientID,
		Limit:     limit,
		Offset:    offset,
	})
	if err != nil {
		log.Print(err.Error())
		r.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	r.JSON(http.StatusOK, docs)
}

// scannedDocGet handles GET /api/scanned-documents/:id
func scannedDocGet(r *gin.Context) {
	id := r.Param("id")
	if id == "" {
		r.AbortWithStatus(http.StatusBadRequest)
		return
	}

	docID := common.ParseInt(id)
	doc, err := model.Queries.GetScannedDoc(r.Request.Context(), docID)
	if err != nil {
		log.Print(err.Error())
		r.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	r.JSON(http.StatusOK, doc)
}
