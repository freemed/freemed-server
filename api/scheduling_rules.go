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
	common.ApiMap["scheduling-rules"] = common.ApiMapping{
		Authenticated: true,
		RouterFunction: func(r *gin.RouterGroup) {
			r.GET("/", listSchedulingRules)
			r.GET("/:id", getSchedulingRule)
			r.POST("/", common.RequireRole("admin"), createSchedulingRule)
			r.PUT("/:id", common.RequireRole("admin"), updateSchedulingRule)
			r.DELETE("/:id", common.RequireRole("admin"), deleteSchedulingRule)
		},
	}
}

// listSchedulingRules handles GET /api/scheduling-rules/
func listSchedulingRules(r *gin.Context) {
	rows, err := model.Queries.ListSchedulingRules(r.Request.Context())
	if err != nil {
		log.Print(err.Error())
		common.ErrorResponseFromError(r, http.StatusInternalServerError, err)
		return
	}
	r.JSON(http.StatusOK, rows)
}

// getSchedulingRule handles GET /api/scheduling-rules/:id
func getSchedulingRule(r *gin.Context) {
	id := common.ParseInt(r.Param("id"))
	if id == 0 {
		common.ErrorResponse(r, http.StatusBadRequest, "invalid id")
		return
	}

	row, err := model.Queries.GetSchedulingRule(r.Request.Context(), id)
	if err != nil {
		if err == sql.ErrNoRows {
			common.ErrorResponse(r, http.StatusNotFound, "scheduling rule not found")
			return
		}
		log.Print(err.Error())
		common.ErrorResponseFromError(r, http.StatusInternalServerError, err)
		return
	}
	r.JSON(http.StatusOK, row)
}

type schedulingRuleInput struct {
	Provider   *string `json:"provider"`
	Reason     *string `json:"reason"`
	Dowbegin   *int32  `json:"dowbegin"`
	Dowend     *int32  `json:"dowend"`
	Datebegin  *string `json:"datebegin"`
	Dateend    *string `json:"dateend"`
	Timebegin  *string `json:"timebegin"`
	Timeend    *string `json:"timeend"`
	Newpatient *bool   `json:"newpatient"`
}

// createSchedulingRule handles POST /api/scheduling-rules
func createSchedulingRule(r *gin.Context) {
	var in schedulingRuleInput
	if err := r.BindJSON(&in); err != nil {
		common.ErrorResponseFromError(r, http.StatusBadRequest, err)
		return
	}

	sess, err := common.GetSession(r)
	if err != nil {
		common.ErrorResponseFromError(r, http.StatusUnauthorized, err)
		return
	}

	params := dbgen.CreateSchedulingRuleParams{
		UserID:    sess.UserId,
		Provider:  toNullString(in.Provider),
		Reason:    toNullString(in.Reason),
		Dowbegin:  toNullInt32(in.Dowbegin),
		Dowend:    toNullInt32(in.Dowend),
		Timebegin: toNullString(in.Timebegin),
		Timeend:   toNullString(in.Timeend),
	}
	if in.Datebegin != nil {
		params.Datebegin = parseOptionalDate(*in.Datebegin)
	}
	if in.Dateend != nil {
		params.Dateend = parseOptionalDate(*in.Dateend)
	}
	if in.Newpatient != nil {
		params.Newpatient = sql.NullBool{Bool: *in.Newpatient, Valid: true}
	}

	result, err := model.Queries.CreateSchedulingRule(r.Request.Context(), params)
	if err != nil {
		log.Print(err.Error())
		common.ErrorResponseFromError(r, http.StatusInternalServerError, err)
		return
	}

	newID, _ := result.LastInsertId()
	r.JSON(http.StatusCreated, gin.H{"id": newID})
}

// updateSchedulingRule handles PUT /api/scheduling-rules/:id
func updateSchedulingRule(r *gin.Context) {
	id := common.ParseInt(r.Param("id"))
	if id == 0 {
		common.ErrorResponse(r, http.StatusBadRequest, "invalid id")
		return
	}

	var in schedulingRuleInput
	if err := r.BindJSON(&in); err != nil {
		common.ErrorResponseFromError(r, http.StatusBadRequest, err)
		return
	}

	params := dbgen.UpdateSchedulingRuleParams{
		ID:        id,
		Provider:  toNullString(in.Provider),
		Reason:    toNullString(in.Reason),
		Dowbegin:  toNullInt32(in.Dowbegin),
		Dowend:    toNullInt32(in.Dowend),
		Timebegin: toNullString(in.Timebegin),
		Timeend:   toNullString(in.Timeend),
	}
	if in.Datebegin != nil {
		params.Datebegin = parseOptionalDate(*in.Datebegin)
	}
	if in.Dateend != nil {
		params.Dateend = parseOptionalDate(*in.Dateend)
	}
	if in.Newpatient != nil {
		params.Newpatient = sql.NullBool{Bool: *in.Newpatient, Valid: true}
	}

	err := model.Queries.UpdateSchedulingRule(r.Request.Context(), params)
	if err != nil {
		log.Print(err.Error())
		common.ErrorResponseFromError(r, http.StatusInternalServerError, err)
		return
	}

	r.JSON(http.StatusOK, true)
}

// deleteSchedulingRule handles DELETE /api/scheduling-rules/:id
func deleteSchedulingRule(r *gin.Context) {
	id := common.ParseInt(r.Param("id"))
	if id == 0 {
		common.ErrorResponse(r, http.StatusBadRequest, "invalid id")
		return
	}

	err := model.Queries.DeleteSchedulingRule(r.Request.Context(), id)
	if err != nil {
		log.Print(err.Error())
		common.ErrorResponseFromError(r, http.StatusInternalServerError, err)
		return
	}

	r.JSON(http.StatusOK, gin.H{"status": "deleted"})
}
