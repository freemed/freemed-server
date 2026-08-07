package api

import (
	"database/sql"
	"log"
	"net/http"

	"github.com/freemed/freemed-server/common"
	"github.com/freemed/freemed-server/dbgen"
	"github.com/freemed/freemed-server/model"
	"github.com/gin-gonic/gin"
)

func init() {
	common.ApiMap["user-preferences"] = common.ApiMapping{
		Authenticated: true,
		RouterFunction: func(r *gin.RouterGroup) {
			r.GET("/", listUserPreferences)
			r.GET("/:key", getUserPreferenceByKey)
			r.PUT("/:key", common.RequireRole("admin"), upsertUserPreference)
		},
	}
}

func listUserPreferences(r *gin.Context) {
	rows, err := model.Queries.ListUserPreferences(r.Request.Context())
	if err != nil {
		log.Print(err.Error())
		r.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	r.JSON(http.StatusOK, rows)
}

func getUserPreferenceByKey(r *gin.Context) {
	key := r.Param("key")
	if key == "" {
		r.AbortWithStatus(http.StatusBadRequest)
		return
	}
	row, err := model.Queries.GetUserPreferenceByKey(r.Request.Context(), key)
	if err != nil {
		if err == sql.ErrNoRows {
			r.AbortWithStatus(http.StatusNotFound)
			return
		}
		log.Print(err.Error())
		r.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	r.JSON(http.StatusOK, row)
}

type userPreferenceInput struct {
	DefaultValue *string `json:"default_value"`
	Title        *string `json:"title"`
	Section      *string `json:"section"`
	OptionType   *string `json:"option_type"`
	Options      *string `json:"options"`
}

func upsertUserPreference(r *gin.Context) {
	key := r.Param("key")
	if key == "" {
		r.AbortWithStatus(http.StatusBadRequest)
		return
	}

	var in userPreferenceInput
	if err := r.BindJSON(&in); err != nil {
		r.AbortWithError(http.StatusBadRequest, err)
		return
	}

	var options sql.NullString
	if in.Options != nil {
		options = sql.NullString{String: *in.Options, Valid: true}
	}

	err := model.Queries.UpsertUserPreference(r.Request.Context(), dbgen.UpsertUserPreferenceParams{
		OptionKey:    key,
		DefaultValue: derefString(in.DefaultValue),
		Title:        derefString(in.Title),
		Section:      derefString(in.Section),
		OptionType:   derefString(in.OptionType),
		Options:      options,
	})
	if err != nil {
		log.Print(err.Error())
		r.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	r.JSON(http.StatusOK, true)
}
