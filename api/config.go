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
	common.ApiMap["config"] = common.ApiMapping{
		Authenticated: true,
		RouterFunction: func(r *gin.RouterGroup) {
			r.GET("/all", configGetAll)
			r.GET("/sections", configGetSections)
			r.PUT("/:id", common.RequireRole("admin"), configUpdate)
		},
	}
	common.ApiMap["preferences"] = common.ApiMapping{
		Authenticated: true,
		RouterFunction: func(r *gin.RouterGroup) {
			r.PUT("/", preferencesBatchUpdate)
		},
	}
}

func configGetAll(r *gin.Context) {
	o, err := model.Queries.ListAllConfig(r.Request.Context())
	if err != nil {
		log.Print(err.Error())
		r.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	r.JSON(http.StatusOK, o)
}

type configValueInput struct {
	CValue *string `json:"c_value" binding:"required"`
}

func configUpdate(r *gin.Context) {
	id := common.ParseInt(r.Param("id"))
	if id == 0 {
		r.AbortWithStatus(http.StatusBadRequest)
		return
	}

	var in configValueInput
	if err := r.BindJSON(&in); err != nil {
		r.AbortWithError(http.StatusBadRequest, err)
		return
	}

	var val sql.NullString
	if in.CValue != nil {
		val = sql.NullString{String: *in.CValue, Valid: true}
	}

	err := model.Queries.UpdateConfig(r.Request.Context(), dbgen.UpdateConfigParams{
		ID:     id,
		CValue: val,
	})
	if err != nil {
		log.Print(err.Error())
		r.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	r.JSON(http.StatusOK, true)
}

func configGetSections(r *gin.Context) {
	sections, err := model.Queries.GetConfigSections(r.Request.Context())
	if err != nil {
		log.Print(err.Error())
		r.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	r.JSON(http.StatusOK, sections)
}

// preferencesBatchUpdate handles PUT /api/preferences
// Accepts a JSON object of {key: value, ...} and batch-updates config rows
func preferencesBatchUpdate(r *gin.Context) {
	var prefs map[string]string
	if err := r.BindJSON(&prefs); err != nil {
		r.AbortWithError(http.StatusBadRequest, err)
		return
	}

	for key, value := range prefs {
		var val sql.NullString
		if value != "" {
			val = sql.NullString{String: value, Valid: true}
		}
		err := model.Queries.UpdateConfigByOption(r.Request.Context(), dbgen.UpdateConfigByOptionParams{
			COption: key,
			CValue:  val,
		})
		if err != nil {
			log.Printf("preferencesBatchUpdate: error updating %s: %v", key, err)
			r.AbortWithError(http.StatusInternalServerError, err)
			return
		}
	}

	r.JSON(http.StatusOK, true)
}
