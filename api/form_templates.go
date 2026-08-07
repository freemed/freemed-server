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
	common.ApiMap["form-templates"] = common.ApiMapping{
		Authenticated: true,
		RouterFunction: func(r *gin.RouterGroup) {
			r.GET("/", common.RequireRole("admin"), listFormTemplates)
			r.GET("/:id", common.RequireRole("admin"), getFormTemplate)
			r.POST("/", common.RequireRole("admin"), createFormTemplate)
			r.PUT("/:id", common.RequireRole("admin"), updateFormTemplate)
			r.DELETE("/:id", common.RequireRole("admin"), deleteFormTemplate)
		},
	}
}

// listFormTemplates handles GET /api/form-templates/
func listFormTemplates(r *gin.Context) {
	rows, err := model.Queries.ListFormTemplates(r.Request.Context())
	if err != nil {
		log.Print(err.Error())
		common.ErrorResponseFromError(r, http.StatusInternalServerError, err)
		return
	}
	r.JSON(http.StatusOK, rows)
}

// getFormTemplate handles GET /api/form-templates/:id
func getFormTemplate(r *gin.Context) {
	id := common.ParseInt(r.Param("id"))
	if id == 0 {
		common.ErrorResponse(r, http.StatusBadRequest, "invalid id")
		return
	}

	row, err := model.Queries.GetFormTemplate(r.Request.Context(), id)
	if err != nil {
		if err == sql.ErrNoRows {
			common.ErrorResponse(r, http.StatusNotFound, "form template not found")
			return
		}
		log.Print(err.Error())
		common.ErrorResponseFromError(r, http.StatusInternalServerError, err)
		return
	}
	r.JSON(http.StatusOK, row)
}

type formTemplateInput struct {
	Name         string  `json:"name" binding:"required"`
	Description  *string `json:"description"`
	FormType     *string `json:"form_type"`
	TemplateData *string `json:"template_data"`
	IsDefault    *bool   `json:"is_default"`
}

// createFormTemplate handles POST /api/form-templates
func createFormTemplate(r *gin.Context) {
	var in formTemplateInput
	if err := r.BindJSON(&in); err != nil {
		common.ErrorResponseFromError(r, http.StatusBadRequest, err)
		return
	}

	sess, err := common.GetSession(r)
	if err != nil {
		common.ErrorResponseFromError(r, http.StatusUnauthorized, err)
		return
	}

	formType := "encounter"
	if in.FormType != nil && *in.FormType != "" {
		formType = *in.FormType
	}

	isDefault := false
	if in.IsDefault != nil {
		isDefault = *in.IsDefault
	}

	var description sql.NullString
	if in.Description != nil {
		description = sql.NullString{String: *in.Description, Valid: true}
	}

	var templateData sql.NullString
	if in.TemplateData != nil {
		templateData = sql.NullString{String: *in.TemplateData, Valid: true}
	}

	result, err := model.Queries.CreateFormTemplate(r.Request.Context(), dbgen.CreateFormTemplateParams{
		Name:         in.Name,
		Description:  description,
		FormType:     formType,
		TemplateData: templateData,
		IsDefault:    isDefault,
		User:         sess.UserId,
	})
	if err != nil {
		log.Print(err.Error())
		common.ErrorResponseFromError(r, http.StatusInternalServerError, err)
		return
	}

	newID, _ := result.LastInsertId()
	r.JSON(http.StatusCreated, gin.H{"id": newID})
}

// updateFormTemplate handles PUT /api/form-templates/:id
func updateFormTemplate(r *gin.Context) {
	id := common.ParseInt(r.Param("id"))
	if id == 0 {
		common.ErrorResponse(r, http.StatusBadRequest, "invalid id")
		return
	}

	var in formTemplateInput
	if err := r.BindJSON(&in); err != nil {
		common.ErrorResponseFromError(r, http.StatusBadRequest, err)
		return
	}

	formType := "encounter"
	if in.FormType != nil && *in.FormType != "" {
		formType = *in.FormType
	}

	isDefault := false
	if in.IsDefault != nil {
		isDefault = *in.IsDefault
	}

	var description sql.NullString
	if in.Description != nil {
		description = sql.NullString{String: *in.Description, Valid: true}
	}

	var templateData sql.NullString
	if in.TemplateData != nil {
		templateData = sql.NullString{String: *in.TemplateData, Valid: true}
	}

	err := model.Queries.UpdateFormTemplate(r.Request.Context(), dbgen.UpdateFormTemplateParams{
		ID:           id,
		Name:         in.Name,
		Description:  description,
		FormType:     formType,
		TemplateData: templateData,
		IsDefault:    isDefault,
	})
	if err != nil {
		log.Print(err.Error())
		common.ErrorResponseFromError(r, http.StatusInternalServerError, err)
		return
	}

	r.JSON(http.StatusOK, true)
}

// deleteFormTemplate handles DELETE /api/form-templates/:id
func deleteFormTemplate(r *gin.Context) {
	id := common.ParseInt(r.Param("id"))
	if id == 0 {
		common.ErrorResponse(r, http.StatusBadRequest, "invalid id")
		return
	}

	err := model.Queries.DeleteFormTemplate(r.Request.Context(), id)
	if err != nil {
		log.Print(err.Error())
		common.ErrorResponseFromError(r, http.StatusInternalServerError, err)
		return
	}

	r.JSON(http.StatusOK, gin.H{"status": "deleted"})
}
