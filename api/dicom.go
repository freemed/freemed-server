package api

import (
	"crypto/md5"
	"database/sql"
	"encoding/base64"
	"encoding/hex"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/freemed/freemed-server/common"
	"github.com/freemed/freemed-server/dbgen"
	"github.com/freemed/freemed-server/model"
	"github.com/freemed/freemed-server/pkg/dicom"
	"github.com/gin-gonic/gin"
)

// DICOMweb endpoints (QIDO-RS search + WADO-RS retrieve). The patient
// sub-resource routes (list/upload/get/remove) are registered in
// api/patient.go.
func init() {
	common.ApiMap["dicom"] = common.ApiMapping{
		Authenticated: true,
		RouterFunction: func(r *gin.RouterGroup) {
			r.GET("/studies", dicomQidoStudies)
			r.GET("/studies/:studyUID/series/:seriesUID/instances/:sopUID", dicomWadoRetrieve)
		},
	}
}

// dicomListItem is the metadata-only shape returned to the frontend. The raw
// DICOM blob (d_data) is deliberately excluded to keep list responses light.
type dicomListItem struct {
	ID                 int64  `json:"id"`
	CreatedAt          string `json:"created_at"`
	StudyDescription   string `json:"study_description"`
	Filename           string `json:"filename"`
	StudyDate          string `json:"study_date"`
	InstitutionName    string `json:"institution_name"`
	InstitutionAddress string `json:"institution_address"`
	StudyUID           string `json:"study_uid"`
	SeriesUID          string `json:"series_uid"`
	SopUID             string `json:"sop_uid"`
	ReferringProvider  string `json:"referring_provider"`
	Modality           string `json:"modality"`
	StorageStatus      string `json:"storage_status"`
}

// dicomStudy is the QIDO-RS study-level response subset.
type dicomStudy struct {
	StudyInstanceUID string `json:"StudyInstanceUID"`
	StudyDate        string `json:"StudyDate"`
	StudyDescription string `json:"StudyDescription"`
	PatientID        string `json:"PatientID"`
	Modality         string `json:"Modality"`
}

func dicomNullTime(nt sql.NullTime) string {
	if !nt.Valid {
		return ""
	}
	return nt.Time.Format("2006-01-02")
}

// dicomUploadInput is the JSON fallback body for base64 DICOM uploads.
type dicomUploadInput struct {
	DataBase64 string `json:"data_base64"`
	Filename   string `json:"filename"`
}

// dicomList handles GET /api/patient/:id/dicom
func dicomList(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		common.ErrorResponse(c, http.StatusBadRequest, "bad request")
		return
	}

	patientID := common.ParseInt(id)
	rows, err := model.Queries.ListDicomByPatient(c.Request.Context(), patientID)
	if err != nil {
		log.Print(err.Error())
		common.ErrorResponseFromError(c, http.StatusInternalServerError, err)
		return
	}

	items := make([]dicomListItem, 0, len(rows))
	for _, row := range rows {
		items = append(items, dicomListItem{
			ID:                 row.ID,
			CreatedAt:          row.CreatedAt.Format(time.RFC3339),
			StudyDescription:   row.DStudyDescription.String,
			Filename:           row.DFilename.String,
			StudyDate:          dicomNullTime(row.DStudyDate),
			InstitutionName:    row.DInstitutionName.String,
			InstitutionAddress: row.DInstitutionAddress.String,
			StudyUID:           row.DStudyUid.String,
			SeriesUID:          row.DSeriesUid.String,
			SopUID:             row.DSopUid.String,
			ReferringProvider:  row.DReferringProvider.String,
			Modality:           row.DModality.String,
			StorageStatus:      row.StorageStatus,
		})
	}
	c.JSON(http.StatusOK, items)
}

// dicomUpload handles POST /api/patient/:id/dicom — multipart form field
// "file" or JSON {data_base64, filename}. Deduplicates by md5.
func dicomUpload(c *gin.Context) {
	patientID := common.ParseInt(c.Param("id"))
	if patientID == 0 {
		common.ErrorResponse(c, http.StatusBadRequest, "bad request")
		return
	}

	sess, err := common.GetSession(c)
	if err != nil {
		common.ErrorResponseFromError(c, http.StatusUnauthorized, err)
		return
	}

	var data []byte
	var filename string

	if file, header, ferr := c.Request.FormFile("file"); ferr == nil {
		defer file.Close()
		data, err = io.ReadAll(file)
		if err != nil {
			common.ErrorResponseFromError(c, http.StatusBadRequest, err)
			return
		}
		filename = header.Filename
	} else {
		var in dicomUploadInput
		if err := c.BindJSON(&in); err != nil {
			common.ErrorResponseFromError(c, http.StatusBadRequest, err)
			return
		}
		decoded, derr := base64.StdEncoding.DecodeString(in.DataBase64)
		if derr != nil {
			common.ErrorResponse(c, http.StatusBadRequest, "invalid base64 data")
			return
		}
		data = decoded
		filename = in.Filename
	}

	if len(data) == 0 {
		common.ErrorResponse(c, http.StatusBadRequest, "empty DICOM data")
		return
	}

	sum := md5.Sum(data)
	md5hex := hex.EncodeToString(sum[:])

	count, err := model.Queries.CountDicomByMd5(c.Request.Context(), md5hex)
	if err != nil {
		log.Print(err.Error())
		common.ErrorResponseFromError(c, http.StatusInternalServerError, err)
		return
	}
	if count > 0 {
		common.ErrorResponse(c, http.StatusConflict, "duplicate DICOM object already stored")
		return
	}

	// Metadata extraction is best-effort: a parse failure does not prevent
	// the object from being stored, it just yields empty metadata columns.
	meta, perr := dicom.Parse(data)
	if perr != nil {
		log.Printf("dicomUpload: metadata parse warning: %v", perr)
	}

	var studyDate sql.NullTime
	if meta.StudyDate != "" {
		if t, terr := time.Parse("20060102", meta.StudyDate); terr == nil {
			studyDate = sql.NullTime{Time: t, Valid: true}
		}
	}

	result, err := model.Queries.CreateDicom(c.Request.Context(), dbgen.CreateDicomParams{
		Md5:                md5hex,
		PatientID:          patientID,
		StudyDescription:   strToNullString(meta.StudyDescription),
		Filename:           strToNullString(filename),
		StudyDate:          studyDate,
		InstitutionName:    strToNullString(meta.InstitutionName),
		InstitutionAddress: sql.NullString{},
		StudyUid:           strToNullString(meta.StudyInstanceUID),
		SeriesUid:          strToNullString(meta.SeriesInstanceUID),
		SopUid:             strToNullString(meta.SOPInstanceUID),
		ReferringProvider:  strToNullString(meta.ReferringPhysicianName),
		Modality:           strToNullString(meta.Modality),
		DicomPatientID:     strToNullString(meta.PatientID),
		XmlData:            sql.NullString{},
		StorageStatus:      "stored",
		UserID:             sess.UserId,
		Data:               sql.NullString{String: string(data), Valid: true},
	})
	if err != nil {
		log.Print(err.Error())
		common.ErrorResponseFromError(c, http.StatusInternalServerError, err)
		return
	}

	newID, _ := result.LastInsertId()
	c.JSON(http.StatusCreated, gin.H{"id": newID})
}

// dicomGet handles GET /api/patient/:id/dicom/:itemId — raw DICOM bytes.
func dicomGet(c *gin.Context) {
	id := c.Param("itemId")
	if id == "" {
		common.ErrorResponse(c, http.StatusBadRequest, "bad request")
		return
	}

	itemID := common.ParseInt(id)
	row, err := model.Queries.GetDicom(c.Request.Context(), itemID)
	if err != nil {
		if err == sql.ErrNoRows {
			common.ErrorResponse(c, http.StatusNotFound, "DICOM object not found")
			return
		}
		log.Print(err.Error())
		common.ErrorResponseFromError(c, http.StatusInternalServerError, err)
		return
	}

	c.Data(http.StatusOK, "application/dicom", []byte(row.DData.String))
}

// dicomRemove handles DELETE /api/patient/:id/dicom/:itemId — soft delete.
func dicomRemove(c *gin.Context) {
	itemID := common.ParseInt(c.Param("itemId"))
	if itemID == 0 {
		common.ErrorResponse(c, http.StatusBadRequest, "bad request")
		return
	}

	if err := model.Queries.DeleteDicom(c.Request.Context(), itemID); err != nil {
		log.Print(err.Error())
		common.ErrorResponseFromError(c, http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "deleted"})
}

// dicomQidoStudies handles GET /api/dicom/studies — QIDO-RS study search.
// Supports ?PatientID= and ?StudyInstanceUID= filters.
func dicomQidoStudies(c *gin.Context) {
	var patientID sql.NullString
	if v := c.Query("PatientID"); v != "" {
		patientID = sql.NullString{String: v, Valid: true}
	}
	var studyUID sql.NullString
	if v := c.Query("StudyInstanceUID"); v != "" {
		studyUID = sql.NullString{String: v, Valid: true}
	}

	rows, err := model.Queries.ListDicomStudies(c.Request.Context(), dbgen.ListDicomStudiesParams{
		PatientID: patientID,
		StudyUid:  studyUID,
	})
	if err != nil {
		log.Print(err.Error())
		common.ErrorResponseFromError(c, http.StatusInternalServerError, err)
		return
	}

	out := make([]dicomStudy, 0, len(rows))
	for _, row := range rows {
		out = append(out, dicomStudy{
			StudyInstanceUID: row.DStudyUid.String,
			StudyDate:        dicomNullTime(row.DStudyDate),
			StudyDescription: row.DStudyDescription.String,
			PatientID:        row.DPatientID.String,
			Modality:         row.DModality.String,
		})
	}
	c.JSON(http.StatusOK, out)
}

// dicomWadoRetrieve handles GET
// /api/dicom/studies/:studyUID/series/:seriesUID/instances/:sopUID — WADO-RS
// instance retrieval returning raw DICOM bytes.
func dicomWadoRetrieve(c *gin.Context) {
	studyUID := c.Param("studyUID")
	seriesUID := c.Param("seriesUID")
	sopUID := c.Param("sopUID")
	if studyUID == "" || seriesUID == "" || sopUID == "" {
		common.ErrorResponse(c, http.StatusBadRequest, "bad request")
		return
	}

	row, err := model.Queries.GetDicomBySop(c.Request.Context(), dbgen.GetDicomBySopParams{
		StudyUid:  sql.NullString{String: studyUID, Valid: true},
		SeriesUid: sql.NullString{String: seriesUID, Valid: true},
		SopUid:    sql.NullString{String: sopUID, Valid: true},
	})
	if err != nil {
		if err == sql.ErrNoRows {
			common.ErrorResponse(c, http.StatusNotFound, "SOP instance not found")
			return
		}
		log.Print(err.Error())
		common.ErrorResponseFromError(c, http.StatusInternalServerError, err)
		return
	}

	c.Data(http.StatusOK, "application/dicom", []byte(row.DData.String))
}
