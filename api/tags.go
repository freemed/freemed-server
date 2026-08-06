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
	common.ApiMap["tags"] = common.ApiMapping{
		Authenticated: true,
		RouterFunction: func(r *gin.RouterGroup) {
			r.GET("/search", tagsSearch)
		},
	}
}

// TagsInput is the JSON payload for creating a patient tag.
type TagsInput struct {
	Tag string `json:"tag"`
}

// patientTagsList handles GET /api/patient/:id/tags
func patientTagsList(r *gin.Context) {
	id := r.Param("id")
	if id == "" {
		r.AbortWithStatus(http.StatusBadRequest)
		return
	}

	patientID := common.ParseInt(id)
	tags, err := model.Queries.ListTags(r.Request.Context(), patientID)
	if err != nil {
		log.Print(err.Error())
		r.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	r.JSON(http.StatusOK, tags)
}

// patientTagsCreate handles POST /api/patient/:id/tags
func patientTagsCreate(r *gin.Context) {
	id := r.Param("id")
	if id == "" {
		r.AbortWithStatus(http.StatusBadRequest)
		return
	}

	session, err := common.GetSession(r)
	if err != nil {
		log.Printf("patientTagsCreate: failed to get session: %v", err)
		r.AbortWithError(http.StatusUnauthorized, err)
		return
	}

	var input TagsInput
	if err := r.BindJSON(&input); err != nil {
		log.Print(err.Error())
		r.AbortWithError(http.StatusBadRequest, err)
		return
	}

	params := dbgen.CreateTagParams{
		Tag:       input.Tag,
		PatientID: common.ParseInt(id),
		UserID:    session.UserId,
	}

	result, err := model.Queries.CreateTag(r.Request.Context(), params)
	if err != nil {
		log.Print(err.Error())
		r.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	newID, _ := result.LastInsertId()
	r.JSON(http.StatusCreated, gin.H{"id": newID})
}

// patientTagsExpire handles DELETE /api/patient/:id/tags/:tagId
func patientTagsExpire(r *gin.Context) {
	tagIDStr := r.Param("tagId")
	if tagIDStr == "" {
		r.AbortWithStatus(http.StatusBadRequest)
		return
	}

	tagID := common.ParseInt(tagIDStr)
	if err := model.Queries.ExpireTag(r.Request.Context(), tagID); err != nil {
		log.Print(err.Error())
		r.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	r.JSON(http.StatusOK, gin.H{"status": "expired"})
}

// tagsSearch handles GET /api/tags/search?q=
func tagsSearch(r *gin.Context) {
	q := r.Query("q")
	if q == "" {
		r.AbortWithStatus(http.StatusBadRequest)
		return
	}

	results, err := model.Queries.SearchByTag(r.Request.Context(), q)
	if err != nil {
		log.Print(err.Error())
		r.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	r.JSON(http.StatusOK, results)
}
