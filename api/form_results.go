package api

import (
	"context"
	"database/sql"
	"encoding/xml"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/freemed/freemed-server/common"
	"github.com/freemed/freemed-server/dbgen"
	"github.com/freemed/freemed-server/model"
	"github.com/gin-gonic/gin"
)

// controlField describes a single typed control parsed from a form template's
// XML (//controls/control element attributes).
type controlField struct {
	Variable string `json:"variable"`
	Name     string `json:"name"`
	Type     string `json:"type"`
	Default  string `json:"default"`
	Options  string `json:"options"`
	Limits   string `json:"limits"`
	UUID     string `json:"uuid"`
}

// parseTemplateControls parses form template XML and returns the list of
// controls found under //controls/control. Missing attributes yield empty
// strings. It uses only encoding/xml — no external XML libraries.
func parseTemplateControls(xmlData string) ([]controlField, error) {
	dec := xml.NewDecoder(strings.NewReader(xmlData))
	var controls []controlField
	for {
		tok, err := dec.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		se, ok := tok.(xml.StartElement)
		if !ok || se.Name.Local != "control" {
			continue
		}
		var cf controlField
		for _, a := range se.Attr {
			switch a.Name.Local {
			case "variable":
				cf.Variable = a.Value
			case "name":
				cf.Name = a.Value
			case "type":
				cf.Type = a.Value
			case "default":
				cf.Default = a.Value
			case "options":
				cf.Options = a.Value
			case "limits":
				cf.Limits = a.Value
			case "uuid":
				cf.UUID = a.Value
			}
		}
		controls = append(controls, cf)
	}
	return controls, nil
}

// resolveTemplateControls looks up a form template by name and returns its
// parsed controls. Returns an empty slice (no error) when the template name is
// blank or no matching template exists.
func resolveTemplateControls(ctx context.Context, templateName string) ([]controlField, error) {
	if templateName == "" {
		return []controlField{}, nil
	}
	rows, err := model.Queries.ListFormTemplates(ctx)
	if err != nil {
		return nil, err
	}
	for _, t := range rows {
		if t.Name == templateName {
			if !t.TemplateData.Valid {
				return []controlField{}, nil
			}
			return parseTemplateControls(t.TemplateData.String)
		}
	}
	return []controlField{}, nil
}

func controlUUID(cf controlField) string {
	if cf.UUID != "" {
		return cf.UUID
	}
	return cf.Variable
}

// formTemplateControls handles GET /api/form-templates/:id/controls
func formTemplateControls(r *gin.Context) {
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

	if !row.TemplateData.Valid {
		r.JSON(http.StatusOK, []controlField{})
		return
	}

	controls, err := parseTemplateControls(row.TemplateData.String)
	if err != nil {
		common.ErrorResponseFromError(r, http.StatusBadRequest, err)
		return
	}
	r.JSON(http.StatusOK, controls)
}

// formResultsList handles GET /api/patient/:id/forms
func formResultsList(r *gin.Context) {
	patientID := common.ParseInt(r.Param("id"))
	if patientID == 0 {
		common.ErrorResponse(r, http.StatusBadRequest, "bad request")
		return
	}

	// The array shape is kept: frontend/src/routes/patients/[id]/forms/+page.svelte
	// types this as `FormResult[]`. ?offset=/?limit= are honoured and clamped,
	// and the query carries `LIMIT ? OFFSET ?` instead of returning every form
	// the patient has ever had.
	offset, limit := pageParams(r)

	rows, err := model.Queries.ListFormResultsByPatient(r.Request.Context(), dbgen.ListFormResultsByPatientParams{
		PatientID: patientID,
		Limit:     limit,
		Offset:    offset,
	})
	if err != nil {
		log.Print(err.Error())
		common.ErrorResponseFromError(r, http.StatusInternalServerError, err)
		return
	}
	r.JSON(http.StatusOK, rows)
}

type formValuesInput struct {
	TemplateID int64             `json:"template_id"`
	Values     map[string]string `json:"values"`
}

// formResultsCreate handles POST /api/patient/:id/forms
func formResultsCreate(r *gin.Context) {
	patientID := common.ParseInt(r.Param("id"))
	if patientID == 0 {
		common.ErrorResponse(r, http.StatusBadRequest, "bad request")
		return
	}

	sess, err := common.GetSession(r)
	if err != nil {
		common.ErrorResponseFromError(r, http.StatusUnauthorized, err)
		return
	}

	var in formValuesInput
	if err := r.BindJSON(&in); err != nil {
		common.ErrorResponseFromError(r, http.StatusBadRequest, err)
		return
	}
	if in.TemplateID == 0 {
		common.ErrorResponse(r, http.StatusBadRequest, "template_id is required")
		return
	}

	template, err := model.Queries.GetFormTemplate(r.Request.Context(), in.TemplateID)
	if err != nil {
		if err == sql.ErrNoRows {
			common.ErrorResponse(r, http.StatusNotFound, "form template not found")
			return
		}
		log.Print(err.Error())
		common.ErrorResponseFromError(r, http.StatusInternalServerError, err)
		return
	}

	templateData := ""
	if template.TemplateData.Valid {
		templateData = template.TemplateData.String
	}
	controls, err := parseTemplateControls(templateData)
	if err != nil {
		common.ErrorResponseFromError(r, http.StatusBadRequest, err)
		return
	}

	result, err := model.Queries.CreateFormResult(r.Request.Context(), dbgen.CreateFormResultParams{
		PatientID:  patientID,
		FrTemplate: template.Name,
		FrFormname: template.Name,
		User:       sess.UserId,
		Active:     "active",
	})
	if err != nil {
		log.Print(err.Error())
		common.ErrorResponseFromError(r, http.StatusInternalServerError, err)
		return
	}
	formID, _ := result.LastInsertId()

	for _, cf := range controls {
		value := in.Values[cf.Variable]
		_, err := model.Queries.CreateFormRecord(r.Request.Context(), dbgen.CreateFormRecordParams{
			FrID:    formID,
			FrUuid:  controlUUID(cf),
			FrName:  cf.Name,
			FrValue: sql.NullString{String: value, Valid: true},
		})
		if err != nil {
			log.Print(err.Error())
			common.ErrorResponseFromError(r, http.StatusInternalServerError, err)
			return
		}
	}

	r.JSON(http.StatusCreated, gin.H{"id": formID})
}

// formResultsGet handles GET /api/patient/:id/forms/:itemId
func formResultsGet(r *gin.Context) {
	patientID := common.ParseInt(r.Param("id"))
	itemID := common.ParseInt(r.Param("itemId"))
	if patientID == 0 || itemID == 0 {
		common.ErrorResponse(r, http.StatusBadRequest, "bad request")
		return
	}

	row, err := model.Queries.GetFormResult(r.Request.Context(), itemID)
	if err != nil {
		if err == sql.ErrNoRows {
			common.ErrorResponse(r, http.StatusNotFound, "form not found")
			return
		}
		log.Print(err.Error())
		common.ErrorResponseFromError(r, http.StatusInternalServerError, err)
		return
	}
	if row.FrPatient != patientID {
		common.ErrorResponse(r, http.StatusNotFound, "form not found")
		return
	}

	records, err := model.Queries.ListFormRecords(r.Request.Context(), itemID)
	if err != nil {
		log.Print(err.Error())
		common.ErrorResponseFromError(r, http.StatusInternalServerError, err)
		return
	}

	values := map[string]string{}
	for _, rec := range records {
		if rec.FrValue.Valid {
			values[rec.FrName] = rec.FrValue.String
		} else {
			values[rec.FrName] = ""
		}
	}

	controls, err := resolveTemplateControls(r.Request.Context(), row.FrTemplate)
	if err != nil {
		log.Print(err.Error())
		common.ErrorResponseFromError(r, http.StatusInternalServerError, err)
		return
	}

	r.JSON(http.StatusOK, gin.H{
		"form":     row,
		"values":   values,
		"controls": controls,
	})
}

// formResultsUpdate handles PUT /api/patient/:id/forms/:itemId
func formResultsUpdate(r *gin.Context) {
	patientID := common.ParseInt(r.Param("id"))
	itemID := common.ParseInt(r.Param("itemId"))
	if patientID == 0 || itemID == 0 {
		common.ErrorResponse(r, http.StatusBadRequest, "bad request")
		return
	}

	var in formValuesInput
	if err := r.BindJSON(&in); err != nil {
		common.ErrorResponseFromError(r, http.StatusBadRequest, err)
		return
	}

	row, err := model.Queries.GetFormResult(r.Request.Context(), itemID)
	if err != nil {
		if err == sql.ErrNoRows {
			common.ErrorResponse(r, http.StatusNotFound, "form not found")
			return
		}
		log.Print(err.Error())
		common.ErrorResponseFromError(r, http.StatusInternalServerError, err)
		return
	}
	if row.FrPatient != patientID {
		common.ErrorResponse(r, http.StatusNotFound, "form not found")
		return
	}

	controls, err := resolveTemplateControls(r.Request.Context(), row.FrTemplate)
	if err != nil {
		log.Print(err.Error())
		common.ErrorResponseFromError(r, http.StatusInternalServerError, err)
		return
	}

	existing, err := model.Queries.ListFormRecords(r.Request.Context(), itemID)
	if err != nil {
		log.Print(err.Error())
		common.ErrorResponseFromError(r, http.StatusInternalServerError, err)
		return
	}
	existingUUIDs := map[string]bool{}
	for _, rec := range existing {
		existingUUIDs[rec.FrUuid] = true
	}

	for _, cf := range controls {
		value, ok := in.Values[cf.Variable]
		if !ok {
			continue
		}
		uuid := controlUUID(cf)
		nv := sql.NullString{String: value, Valid: true}
		if existingUUIDs[uuid] {
			err = model.Queries.UpdateFormRecord(r.Request.Context(), dbgen.UpdateFormRecordParams{
				FrID:    itemID,
				FrUuid:  uuid,
				FrValue: nv,
			})
		} else {
			_, err = model.Queries.CreateFormRecord(r.Request.Context(), dbgen.CreateFormRecordParams{
				FrID:    itemID,
				FrUuid:  uuid,
				FrName:  cf.Name,
				FrValue: nv,
			})
		}
		if err != nil {
			log.Print(err.Error())
			common.ErrorResponseFromError(r, http.StatusInternalServerError, err)
			return
		}
	}

	if err := model.Queries.UpdateFormResult(r.Request.Context(), dbgen.UpdateFormResultParams{
		ID:         itemID,
		FrTemplate: row.FrTemplate,
		FrFormname: row.FrFormname,
		Active:     row.Active,
	}); err != nil {
		log.Print(err.Error())
		common.ErrorResponseFromError(r, http.StatusInternalServerError, err)
		return
	}

	r.JSON(http.StatusOK, gin.H{"status": "updated"})
}

// formResultsRemove handles DELETE /api/patient/:id/forms/:itemId
func formResultsRemove(r *gin.Context) {
	patientID := common.ParseInt(r.Param("id"))
	itemID := common.ParseInt(r.Param("itemId"))
	if patientID == 0 || itemID == 0 {
		common.ErrorResponse(r, http.StatusBadRequest, "bad request")
		return
	}

	row, err := model.Queries.GetFormResult(r.Request.Context(), itemID)
	if err != nil {
		if err == sql.ErrNoRows {
			common.ErrorResponse(r, http.StatusNotFound, "form not found")
			return
		}
		log.Print(err.Error())
		common.ErrorResponseFromError(r, http.StatusInternalServerError, err)
		return
	}
	if row.FrPatient != patientID {
		common.ErrorResponse(r, http.StatusNotFound, "form not found")
		return
	}

	if err := model.Queries.DeleteFormRecords(r.Request.Context(), itemID); err != nil {
		log.Print(err.Error())
		common.ErrorResponseFromError(r, http.StatusInternalServerError, err)
		return
	}
	if err := model.Queries.DeleteFormResult(r.Request.Context(), itemID); err != nil {
		log.Print(err.Error())
		common.ErrorResponseFromError(r, http.StatusInternalServerError, err)
		return
	}

	r.JSON(http.StatusOK, gin.H{"status": "deactivated"})
}
