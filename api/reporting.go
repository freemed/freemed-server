package api

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/freemed/freemed-server/common"
	"github.com/freemed/freemed-server/dbgen"
	"github.com/freemed/freemed-server/model"
	"github.com/gin-gonic/gin"
	"github.com/xuri/excelize/v2"
)

func init() {
	common.ApiMap["reports"] = common.ApiMapping{
		Authenticated: true,
		RouterFunction: func(r *gin.RouterGroup) {
			r.GET("/", reportsList)
			r.GET("/:id", reportsGet)
			r.POST("/export", exportReports)
		},
	}
}

func reportsList(r *gin.Context) {
	o, err := model.Queries.ListReports(r.Request.Context())
	if err != nil {
		log.Print(err.Error())
		r.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	r.JSON(http.StatusOK, o)
}

func reportsGet(r *gin.Context) {
	id := common.ParseInt(r.Param("id"))
	if id == 0 {
		r.AbortWithStatus(http.StatusBadRequest)
		return
	}
	o, err := model.Queries.GetReportById(r.Request.Context(), id)
	if err != nil {
		log.Print(err.Error())
		r.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	r.JSON(http.StatusOK, o)
}

type exportInput struct {
	ReportType string                 `json:"report_type" binding:"required"`
	Params     map[string]interface{} `json:"params"`
}

func exportReports(c *gin.Context) {
	_, err := common.GetSession(c)
	if err != nil {
		common.ErrorResponseFromError(c, http.StatusUnauthorized, err)
		return
	}

	var in exportInput
	if err := c.BindJSON(&in); err != nil {
		common.ErrorResponseFromError(c, http.StatusBadRequest, err)
		return
	}

	switch in.ReportType {
	case "aging":
		exportAging(c)
	case "patient_list":
		exportPatientList(c)
	case "medications":
		exportMedications(c, in.Params)
	case "problems":
		exportProblems(c, in.Params)
	case "encounters":
		exportEncounters(c, in.Params)
	default:
		common.ErrorResponse(c, http.StatusBadRequest, "unsupported report type: "+in.ReportType)
	}
}

func exportAging(c *gin.Context) {
	rows, err := model.Queries.AgingSummary(c.Request.Context())
	if err != nil {
		log.Printf("exportAging: %s", err.Error())
		common.ErrorResponseFromError(c, http.StatusInternalServerError, err)
		return
	}

	f := excelize.NewFile()
	defer f.Close()

	sheet := "Sheet1"
	f.SetCellValue(sheet, "A1", "Age Bucket")
	f.SetCellValue(sheet, "B1", "Balance")

	for i, row := range rows {
		rowNum := i + 2
		f.SetCellValue(sheet, cellName("A", rowNum), row.AgeBucket)
		f.SetCellValue(sheet, cellName("B", rowNum), row.Balance)
	}

	c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Header("Content-Disposition", "attachment; filename=aging_report.xlsx")
	if err := f.Write(c.Writer); err != nil {
		log.Printf("exportAging: write error: %s", err.Error())
	}
}

func exportPatientList(c *gin.Context) {
	rows, err := model.Queries.ListPatients(c.Request.Context(), dbgen.ListPatientsParams{
		Limit:  10000,
		Offset: 0,
	})
	if err != nil {
		log.Printf("exportPatientList: %s", err.Error())
		common.ErrorResponseFromError(c, http.StatusInternalServerError, err)
		return
	}

	f := excelize.NewFile()
	defer f.Close()

	sheet := "Sheet1"
	f.SetCellValue(sheet, "A1", "ID")
	f.SetCellValue(sheet, "B1", "Last Name")
	f.SetCellValue(sheet, "C1", "First Name")
	f.SetCellValue(sheet, "D1", "Date of Birth")
	f.SetCellValue(sheet, "E1", "Gender")
	f.SetCellValue(sheet, "F1", "Patient ID")

	for i, row := range rows {
		rowNum := i + 2
		f.SetCellValue(sheet, cellName("A", rowNum), row.ID)
		f.SetCellValue(sheet, cellName("B", rowNum), row.LastName)
		f.SetCellValue(sheet, cellName("C", rowNum), row.FirstName)
		if row.DateOfBirth.Valid {
			f.SetCellValue(sheet, cellName("D", rowNum), row.DateOfBirth.Time.Format("2006-01-02"))
		}
		f.SetCellValue(sheet, cellName("E", rowNum), row.Gender)
		f.SetCellValue(sheet, cellName("F", rowNum), row.PatientID)
	}

	c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Header("Content-Disposition", "attachment; filename=patient_list.xlsx")
	if err := f.Write(c.Writer); err != nil {
		log.Printf("exportPatientList: write error: %s", err.Error())
	}
}

func exportMedications(c *gin.Context, params map[string]interface{}) {
	patientID := getPatientIDFromParams(c, params)
	if patientID == 0 {
		common.ErrorResponse(c, http.StatusBadRequest, "patient_id is required")
		return
	}

	rows, err := model.Queries.ListMedications(c.Request.Context(), patientID)
	if err != nil {
		log.Printf("exportMedications: %s", err.Error())
		common.ErrorResponseFromError(c, http.StatusInternalServerError, err)
		return
	}

	f := excelize.NewFile()
	defer f.Close()

	sheet := "Sheet1"
	f.SetCellValue(sheet, "A1", "ID")
	f.SetCellValue(sheet, "B1", "Drug Name")
	f.SetCellValue(sheet, "C1", "Dosage")
	f.SetCellValue(sheet, "D1", "Frequency")
	f.SetCellValue(sheet, "E1", "Start Date")
	f.SetCellValue(sheet, "F1", "End Date")
	f.SetCellValue(sheet, "G1", "Active")

	for i, row := range rows {
		rowNum := i + 2
		f.SetCellValue(sheet, cellName("A", rowNum), row.ID)
		f.SetCellValue(sheet, cellName("B", rowNum), row.DrugName)
		f.SetCellValue(sheet, cellName("C", rowNum), row.Dosage)
		f.SetCellValue(sheet, cellName("D", rowNum), row.Frequency)
		if row.StartDate.Valid {
			f.SetCellValue(sheet, cellName("E", rowNum), row.StartDate.Time.Format("2006-01-02"))
		}
		if row.EndDate.Valid {
			f.SetCellValue(sheet, cellName("F", rowNum), row.EndDate.Time.Format("2006-01-02"))
		}
		f.SetCellValue(sheet, cellName("G", rowNum), row.Active)
	}

	c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Header("Content-Disposition", "attachment; filename=medications.xlsx")
	if err := f.Write(c.Writer); err != nil {
		log.Printf("exportMedications: write error: %s", err.Error())
	}
}

func exportProblems(c *gin.Context, params map[string]interface{}) {
	patientID := getPatientIDFromParams(c, params)
	if patientID == 0 {
		common.ErrorResponse(c, http.StatusBadRequest, "patient_id is required")
		return
	}

	rows, err := model.Queries.ListCurrentProblemsByPatient(c.Request.Context(), patientID)
	if err != nil {
		log.Printf("exportProblems: %s", err.Error())
		common.ErrorResponseFromError(c, http.StatusInternalServerError, err)
		return
	}

	f := excelize.NewFile()
	defer f.Close()

	sheet := "Sheet1"
	f.SetCellValue(sheet, "A1", "ID")
	f.SetCellValue(sheet, "B1", "Date")
	f.SetCellValue(sheet, "C1", "Problem")
	f.SetCellValue(sheet, "D1", "Active")

	for i, row := range rows {
		rowNum := i + 2
		f.SetCellValue(sheet, cellName("A", rowNum), row.ID)
		f.SetCellValue(sheet, cellName("B", rowNum), row.Date.Format("2006-01-02"))
		f.SetCellValue(sheet, cellName("C", rowNum), row.Problem)
		f.SetCellValue(sheet, cellName("D", rowNum), row.Active)
	}

	c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Header("Content-Disposition", "attachment; filename=problems.xlsx")
	if err := f.Write(c.Writer); err != nil {
		log.Printf("exportProblems: write error: %s", err.Error())
	}
}

func exportEncounters(c *gin.Context, params map[string]interface{}) {
	patientID := getPatientIDFromParams(c, params)
	if patientID == 0 {
		common.ErrorResponse(c, http.StatusBadRequest, "patient_id is required")
		return
	}

	rows, err := model.Queries.ListEncounters(c.Request.Context(), patientID)
	if err != nil {
		log.Printf("exportEncounters: %s", err.Error())
		common.ErrorResponseFromError(c, http.StatusInternalServerError, err)
		return
	}

	f := excelize.NewFile()
	defer f.Close()

	sheet := "Sheet1"
	f.SetCellValue(sheet, "A1", "ID")
	f.SetCellValue(sheet, "B1", "Date")
	f.SetCellValue(sheet, "C1", "Summary")
	f.SetCellValue(sheet, "D1", "Status")
	f.SetCellValue(sheet, "E1", "Provider")
	f.SetCellValue(sheet, "F1", "User")

	for i, row := range rows {
		rowNum := i + 2
		f.SetCellValue(sheet, cellName("A", rowNum), row.ID)
		f.SetCellValue(sheet, cellName("B", rowNum), row.Stamp.Format("2006-01-02"))
		f.SetCellValue(sheet, cellName("C", rowNum), row.Summary)
		f.SetCellValue(sheet, cellName("D", rowNum), row.Status)
		f.SetCellValue(sheet, cellName("E", rowNum), row.Provider)
		f.SetCellValue(sheet, cellName("F", rowNum), row.User)
	}

	c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Header("Content-Disposition", "attachment; filename=encounters.xlsx")
	if err := f.Write(c.Writer); err != nil {
		log.Printf("exportEncounters: write error: %s", err.Error())
	}
}

// getPatientIDFromParams extracts a patient_id from the params map.
func getPatientIDFromParams(c *gin.Context, params map[string]interface{}) int64 {
	if params == nil {
		return 0
	}
	val, ok := params["patient_id"]
	if !ok {
		return 0
	}
	switch v := val.(type) {
	case float64:
		return int64(v)
	case int64:
		return v
	case json.Number:
		n, _ := v.Int64()
		return n
	default:
		return 0
	}
}

func cellName(col string, row int) string {
	return col + itoa(row)
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	digits := ""
	for n > 0 {
		digits = string(rune('0'+n%10)) + digits
		n /= 10
	}
	return digits
}
