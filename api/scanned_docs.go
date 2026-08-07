package api

import (
	"log"
	"net/http"

	"github.com/freemed/freemed-server/common"
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
func patientScannedDocsList(r *gin.Context) {
	id := r.Param("id")
	if id == "" {
		r.AbortWithStatus(http.StatusBadRequest)
		return
	}

	patientID := common.ParseInt(id)
	docs, err := model.Queries.ListScannedDocs(r.Request.Context(), patientID)
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
