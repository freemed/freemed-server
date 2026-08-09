package api

import (
	"bytes"
	"fmt"
	"log"
	"net/http"

	"github.com/freemed/freemed-server/common"
	"github.com/freemed/freemed-server/model"
	"github.com/freemed/freemed-server/pkg/billing"
	"github.com/gin-gonic/gin"
	"github.com/jung-kurt/gofpdf"
)

func init() {
	common.ApiMap["cms1500"] = common.ApiMapping{
		Authenticated: true,
		RouterFunction: func(r *gin.RouterGroup) {
			r.GET("/:voucher", common.RequireRole("admin"), cms1500ByVoucher)
		},
	}
}

// cms1500ByVoucher handles GET /api/cms1500/:voucher
// Generates a CMS-1500 (02/12) PDF for all procedures sharing a claim voucher.
func cms1500ByVoucher(c *gin.Context) {
	voucher := c.Param("voucher")
	if voucher == "" {
		common.ErrorResponse(c, http.StatusBadRequest, "voucher parameter is required")
		return
	}

	// Query procrec for all procedures with this voucher
	rows, err := model.SqlDb.Query(
		`SELECT id FROM procrec WHERE procvoucher = ?`, voucher)
	if err != nil {
		log.Printf("cms1500ByVoucher query: %v", err)
		common.ErrorResponseFromError(c, http.StatusInternalServerError, err)
		return
	}
	defer rows.Close()

	var procedureIDs []int64
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			log.Printf("cms1500ByVoucher scan: %v", err)
			common.ErrorResponseFromError(c, http.StatusInternalServerError, err)
			return
		}
		procedureIDs = append(procedureIDs, id)
	}
	if err := rows.Err(); err != nil {
		log.Printf("cms1500ByVoucher rows: %v", err)
		common.ErrorResponseFromError(c, http.StatusInternalServerError, err)
		return
	}

	if len(procedureIDs) == 0 {
		common.ErrorResponse(c, http.StatusNotFound, "no procedures found for voucher")
		return
	}

	// Get facility ID from patient's primary facility
	facilityID := int64(1)
	err = model.SqlDb.QueryRow(
		`SELECT ptprimaryfacility FROM patient p
		 JOIN procrec pr ON pr.procpatient = p.id
		 WHERE pr.procvoucher = ? LIMIT 1`, voucher).Scan(&facilityID)
	if err != nil {
		log.Printf("cms1500ByVoucher facility lookup: %v (using default)", err)
	}

	// Assemble CMS-1500 data
	data, err := billing.AssembleCMS1500(model.SqlDb, procedureIDs, facilityID)
	if err != nil {
		log.Printf("cms1500ByVoucher assemble: %v", err)
		common.ErrorResponseFromError(c, http.StatusInternalServerError, err)
		return
	}

	// Generate PDF
	pdfBytes, err := renderCMS1500PDF(data)
	if err != nil {
		log.Printf("cms1500ByVoucher pdf: %v", err)
		common.ErrorResponseFromError(c, http.StatusInternalServerError, err)
		return
	}

	c.Header("Content-Type", "application/pdf")
	c.Header("Content-Disposition", "attachment; filename=\""+voucher+"-cms1500.pdf\"")
	c.Data(http.StatusOK, "application/pdf", pdfBytes)
}

// renderCMS1500PDF generates a CMS-1500 (02/12) form PDF using gofpdf.
func renderCMS1500PDF(data *billing.CMS1500Data) ([]byte, error) {
	pdf := gofpdf.New("P", "mm", "Letter", "")
	pdf.SetMargins(10, 10, 10)
	pdf.SetAutoPageBreak(false, 0)
	pdf.AddPage()

	// Use a monospace font for precise positioning
	pdf.SetFont("Courier", "", 8)

	// ---- Header: Insurance Type Checkboxes ----
	// Box 1 row at ~22mm from top
	y := 22.0
	pdf.SetXY(10, y)
	pdf.CellFormat(0, 4, "1. INSURANCE: "+trunc(data.InsuranceType, 20), "", 0, "L", false, 0, "")

	// Box 1a: Insured's ID
	y = 28.0
	pdf.SetXY(55, y)
	pdf.CellFormat(0, 4, "1a. INSURED'S ID: "+trunc(data.InsuredID, 20), "", 0, "L", false, 0, "")

	// Box 2: Patient Name
	y = 32.0
	pdf.SetXY(10, y)
	pdf.CellFormat(0, 4, fmt.Sprintf("2. PATIENT: %s, %s %s", data.PatientLastName, data.PatientFirstName, data.PatientMiddleName), "", 0, "L", false, 0, "")

	// Box 3: Patient DOB / Sex
	y = 36.0
	pdf.SetXY(10, y)
	pdf.CellFormat(0, 4, fmt.Sprintf("3. DOB: %s  SEX: %s", data.PatientDOB, data.PatientSex), "", 0, "L", false, 0, "")

	// Box 4: Insured's Name
	y = 40.0
	pdf.SetXY(10, y)
	pdf.CellFormat(0, 4, fmt.Sprintf("4. INSURED: %s, %s", data.InsuredLastName, data.InsuredFirstName), "", 0, "L", false, 0, "")

	// Box 5: Patient Address
	y = 44.0
	pdf.SetXY(10, y)
	pdf.CellFormat(0, 4, fmt.Sprintf("5. ADDRESS: %s, %s %s %s", data.PatientAddress, data.PatientCity, data.PatientState, data.PatientZip), "", 0, "L", false, 0, "")

	// Box 6: Relationship
	y = 48.0
	pdf.SetXY(10, y)
	relLabel := "Self"
	switch data.Relationship {
	case "18":
		relLabel = "Self"
	case "01":
		relLabel = "Spouse"
	case "19":
		relLabel = "Child"
	case "20":
		relLabel = "Employee"
	case "39":
		relLabel = "Organ Donor"
	case "40":
		relLabel = "Cadaver Donor"
	case "53":
		relLabel = "Life Partner"
	case "G8":
		relLabel = "Other Relationship"
	}
	pdf.CellFormat(0, 4, "6. RELATIONSHIP: "+relLabel, "", 0, "L", false, 0, "")

	// Box 7: Insured's Address
	y = 52.0
	pdf.SetXY(10, y)
	pdf.CellFormat(0, 4, fmt.Sprintf("7. INSURED ADDR: %s, %s %s %s", data.InsuredAddress, data.InsuredCity, data.InsuredState, data.InsuredZip), "", 0, "L", false, 0, "")

	// Box 11: Insured's Policy/Group
	y = 56.0
	pdf.SetXY(10, y)
	pdf.CellFormat(0, 4, "11. POLICY/GROUP: "+trunc(data.PolicyGroup, 30), "", 0, "L", false, 0, "")

	// Box 11a: Insured's DOB
	pdf.SetXY(110, y)
	pdf.CellFormat(0, 4, "11a. DOB: "+data.InsuredDOB, "", 0, "L", false, 0, "")

	// Box 17: Referring Provider
	y = 60.0
	pdf.SetXY(10, y)
	pdf.CellFormat(0, 4, "17. REFERRING: "+trunc(data.ReferringProvider, 25)+"  NPI: "+data.ReferringProviderNPI, "", 0, "L", false, 0, "")

	// Box 21: Diagnosis Codes
	y = 64.0
	pdf.SetXY(10, y)
	diagStr := ""
	for i, code := range data.DiagnosisCodes {
		if i > 0 {
			diagStr += ", "
		}
		diagStr += fmt.Sprintf("%d.%s", i+1, code)
	}
	pdf.CellFormat(0, 4, "21. DIAGNOSIS: "+trunc(diagStr, 80), "", 0, "L", false, 0, "")

	// Box 23: Prior Auth
	y = 68.0
	pdf.SetXY(10, y)
	pdf.CellFormat(0, 4, "23. PRIOR AUTH: "+trunc(data.PriorAuthNumber, 25), "", 0, "L", false, 0, "")

	// ---- Box 24: Service Lines ----
	y = 74.0
	pdf.SetFont("Courier", "B", 7)
	pdf.SetXY(10, y)
	pdf.CellFormat(190, 4, "24. SERVICE LINES (Date From|To|POS|CPT|Mods|DiagPtr|Charge|Units)", "", 0, "L", false, 0, "")
	y += 5

	pdf.SetFont("Courier", "", 7)
	maxLines := 6
	for i, sl := range data.ServiceLines {
		if i >= maxLines {
			break
		}
		line := fmt.Sprintf("  %s-%s | %s | %s | %s%s%s | %s%s%s%s | $%.2f | %d",
			sl.DateFrom, sl.DateTo, sl.PlaceOfService, sl.ProcedureCode,
			sl.Modifier1, sl.Modifier2, sl.Modifier3,
			sl.DiagnosisPointer1, sl.DiagnosisPointer2, sl.DiagnosisPointer3, sl.DiagnosisPointer4,
			sl.Charge, sl.Units)
		pdf.SetXY(10, y)
		pdf.CellFormat(190, 4, trunc(line, 120), "", 0, "L", false, 0, "")
		y += 4
	}

	// ---- Footer information ----
	y += 4

	// Box 25: Federal Tax ID
	pdf.SetXY(10, y)
	pdf.SetFont("Courier", "", 8)
	pdf.CellFormat(0, 4, "25. TAX ID: "+data.FederalTaxID, "", 0, "L", false, 0, "")

	// Box 26: Patient Account
	y += 5
	pdf.SetXY(10, y)
	pdf.CellFormat(0, 4, "26. ACCOUNT#: "+data.PatientAccountNumber, "", 0, "L", false, 0, "")

	// Box 28: Total Charge
	y += 5
	pdf.SetXY(10, y)
	pdf.CellFormat(0, 4, fmt.Sprintf("28. TOTAL CHARGE: $%.2f", data.TotalCharge), "", 0, "L", false, 0, "")

	// Box 31: Rendering Provider
	y += 5
	pdf.SetXY(10, y)
	pdf.CellFormat(0, 4, "31. RENDERING: "+trunc(data.RenderingProvider, 35), "", 0, "L", false, 0, "")

	// Box 32: Service Facility
	y += 5
	pdf.SetXY(10, y)
	pdf.CellFormat(0, 4, fmt.Sprintf("32. FACILITY: %s | %s, %s %s", data.FacilityName, data.FacilityCity, data.FacilityState, data.FacilityZip), "", 0, "L", false, 0, "")

	// Box 33: Billing Provider
	y += 5
	pdf.SetXY(10, y)
	pdf.CellFormat(0, 4, fmt.Sprintf("33. BILLING: %s | %s, %s %s | NPI: %s | Ph: %s",
		data.BillingProviderName, data.BillingProviderCity, data.BillingProviderState,
		data.BillingProviderZip, data.BillingProviderNPI, data.BillingProviderPhone), "", 0, "L", false, 0, "")

	// Footer: Generation timestamp
	y += 10
	pdf.SetFont("Courier", "I", 6)
	pdf.SetXY(10, y)
	pdf.CellFormat(0, 4, "Generated by FreeMED EMR - CMS-1500 (02/12)", "", 0, "C", false, 0, "")

	var buf bytes.Buffer
	if err := pdf.Output(&buf); err != nil {
		return nil, fmt.Errorf("pdf output: %w", err)
	}
	return buf.Bytes(), nil
}

func trunc(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}
