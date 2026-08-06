package api

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/freemed/freemed-server/common"
	"github.com/freemed/freemed-server/dbgen"
	"github.com/freemed/freemed-server/model"
	"github.com/gin-gonic/gin"
)

func init() {
	common.ApiMap["patients"] = common.ApiMapping{
		Authenticated: true,
		RouterFunction: func(r *gin.RouterGroup) {
			r.POST("/", patientCreate)
			r.POST("/searchDuplicates", patientSearchForDuplicates)
			r.POST("/search", patientSearch)
			r.GET("/picklist/:param", patientPicklist)
			r.GET("/total", patientTotalInSystem)
		},
	}
}

type picklistItem struct {
	Value string `db:"value" json:"value"`
	ID    int64  `db:"id" json:"id"`
}

func patientPicklist(r *gin.Context) {
	param := r.Param("param")
	if param == "" {
		r.AbortWithStatus(http.StatusBadRequest)
		return
	}

	var first, last, either string
	if strings.Index(param, ",") > -1 {
		parts := strings.SplitN(param, ",", 2)
		last = strings.TrimSpace(parts[0])
		first = strings.TrimSpace(parts[1])
	} else if strings.Index(param, " ") > -1 {
		parts := strings.SplitN(param, " ", 2)
		first = strings.TrimSpace(parts[0])
		last = strings.TrimSpace(parts[1])
	} else {
		either = strings.TrimSpace(param)
	}

	var rows interface{}
	var err error

	if first != "" && last != "" {
		rows, err = model.Queries.PatientPicklistByName(r.Request.Context(), dbgen.PatientPicklistByNameParams{
			LastName:  last,
			FirstName: first,
		})
	} else if first != "" {
		rows, err = model.Queries.PatientPicklistByFirstNameOrId(r.Request.Context(), dbgen.PatientPicklistByFirstNameOrIdParams{
			Query: first,
		})
	} else if last != "" {
		rows, err = model.Queries.PatientPicklistByLastNameOrId(r.Request.Context(), dbgen.PatientPicklistByLastNameOrIdParams{
			Query: last,
		})
	} else if either != "" {
		rows, err = model.Queries.PatientPicklistByEither(r.Request.Context(), dbgen.PatientPicklistByEitherParams{
			Query: either,
		})
	} else {
		r.AbortWithStatus(http.StatusBadRequest)
		return
	}

	if err != nil {
		log.Print(err.Error())
		r.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	// Convert dbgen row types to picklistItem
	switch v := rows.(type) {
	case []dbgen.PatientPicklistByNameRow:
		o := make([]picklistItem, len(v))
		for i, row := range v {
			o[i] = picklistItem{Value: row.Value, ID: row.ID}
		}
		r.JSON(http.StatusOK, o)
	case []dbgen.PatientPicklistByFirstNameOrIdRow:
		o := make([]picklistItem, len(v))
		for i, row := range v {
			o[i] = picklistItem{Value: row.Value, ID: row.ID}
		}
		r.JSON(http.StatusOK, o)
	case []dbgen.PatientPicklistByLastNameOrIdRow:
		o := make([]picklistItem, len(v))
		for i, row := range v {
			o[i] = picklistItem{Value: row.Value, ID: row.ID}
		}
		r.JSON(http.StatusOK, o)
	case []dbgen.PatientPicklistByEitherRow:
		o := make([]picklistItem, len(v))
		for i, row := range v {
			o[i] = picklistItem{Value: row.Value, ID: row.ID}
		}
		r.JSON(http.StatusOK, o)
	}
}

func patientSearch(r *gin.Context) {
	var params gin.H
	if err := r.BindJSON(&params); err != nil {
		log.Print(err.Error())
		r.AbortWithError(http.StatusBadRequest, err)
		return
	}
	if len(params) < 1 {
		r.AbortWithError(http.StatusBadRequest, fmt.Errorf("no usable search parameters found"))
		return
	}

	// Try sqlc PatientSearch for simple params
	hasSimple := false
	searchParams := dbgen.PatientSearchParams{}
	for paramName, paramValue := range params {
		switch paramName {
		case "first_name":
			if sv, ok := paramValue.(string); ok && sv != "" {
				searchParams.FirstName = sv
				hasSimple = true
			}
		case "last_name":
			if sv, ok := paramValue.(string); ok && sv != "" {
				searchParams.LastName = sv
				hasSimple = true
			}
		case "patient_id":
			if sv, ok := paramValue.(string); ok && sv != "" {
				searchParams.PatientID = sv
				hasSimple = true
			}
		}
	}

	if hasSimple {
		rows, err := model.Queries.PatientSearch(r.Request.Context(), searchParams)
		if err != nil {
			log.Print(err.Error())
			r.AbortWithError(http.StatusInternalServerError, err)
			return
		}
		o := make([]picklistItem, len(rows))
		for i, row := range rows {
			o[i] = picklistItem{
				Value: fmt.Sprintf("%s, %s (%s)", row.LastName, row.FirstName, row.PatientID),
				ID:    row.ID,
			}
		}
		r.JSON(http.StatusOK, o)
		return
	}

	// Complex dynamic search with additional fields
	k := make([]string, 0)
	v := make([]interface{}, 0)
	archive := " AND p.ptarchive = 0 "
	for paramName, paramValue := range params {
		switch paramName {
		case "age":
			if iv, found := paramValue.(float64); found && iv != 0 {
				k = append(k, fmt.Sprintf("FLOOR( ( TO_DAYS(NOW()) - TO_DAYS(p.ptdob) ) / 365 ) = %d", int64(iv)))
			}
		case "archive":
			if bv, found := paramValue.(bool); found && bv {
				archive = ""
			}
		case "city":
			if sv, found := paramValue.(string); found && sv != "" {
				k = append(k, "pa.city LIKE CONCAT('%', ?, '%')")
				v = append(v, sv)
			}
		case "dmv":
			if sv, found := paramValue.(string); found && sv != "" {
				k = append(k, "p.dmv LIKE CONCAT('%', ?, '%')")
				v = append(v, sv)
			}
		case "email":
			if sv, found := paramValue.(string); found && sv != "" {
				k = append(k, "p.pemail LIKE CONCAT('%', ?, '%')")
				v = append(v, sv)
			}
		case "ssn":
			if sv, found := paramValue.(string); found && sv != "" {
				k = append(k, "p.ssn LIKE CONCAT('%', ?, '%')")
				v = append(v, sv)
			}
		case "zip":
			if sv, found := paramValue.(string); found && sv != "" {
				k = append(k, "pa.zip LIKE CONCAT('%', ?, '%')")
				v = append(v, sv)
			}
		}
	}

	if len(v) < 1 {
		r.AbortWithError(http.StatusBadRequest, fmt.Errorf("no valid parameters presented"))
		return
	}

	query := fmt.Sprintf(
		"SELECT p.ptlname AS last_name"+
			", p.ptfname AS first_name"+
			", p.ptmname AS middle_name"+
			", p.ptid AS patient_id"+
			", FLOOR( ( TO_DAYS(NOW()) - TO_DAYS(p.ptdob) ) / 365 ) AS age"+
			", p.ptdob AS date_of_birth"+
			", p.id AS id"+
			" FROM "+model.TABLE_PATIENT+" p"+
			" LEFT OUTER JOIN "+model.TABLE_PATIENT_ADDRESS+" pa ON p.id = pa.patient"+
			" WHERE "+strings.Join(k, " AND ")+" AND pa.active = 1 "+archive+
			" ORDER BY p.ptlname, p.ptfname, p.ptmname LIMIT 20")

	rows, err := model.SqlDb.QueryContext(r.Request.Context(), query, v...)
	if err != nil {
		log.Print(err.Error())
		r.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	defer rows.Close()

	var out []picklistItem
	for rows.Next() {
		var item picklistItem
		var lastName, firstName, middleName, patientID, dateOfBirth string
		var age, id int64
		if err := rows.Scan(&lastName, &firstName, &middleName, &patientID, &age, &dateOfBirth, &id); err != nil {
			log.Print(err.Error())
			continue
		}
		item.Value = fmt.Sprintf("%s, %s (%s)", lastName, firstName, patientID)
		item.ID = id
		out = append(out, item)
	}
	r.JSON(http.StatusOK, out)
}

func patientTotalInSystem(r *gin.Context) {
	count, err := model.Queries.PatientTotalInSystem(r.Request.Context())
	if err != nil {
		log.Print(err.Error())
		r.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	r.JSON(http.StatusOK, count)
}

func patientSearchForDuplicates(r *gin.Context) {
	var params gin.H
	if err := r.BindJSON(&params); err != nil {
		log.Print(err.Error())
		r.AbortWithError(http.StatusBadRequest, err)
		return
	}

	dupParams := dbgen.PatientSearchDuplicatesParams{}
	if v, ok := params["ptlname"]; ok {
		dupParams.Ptlname = v.(string)
	}
	if v, ok := params["ptfname"]; ok {
		dupParams.Ptfname = v.(string)
	}
	if v, ok := params["ptmname"]; ok {
		dupParams.Ptmname = sql.NullString{String: v.(string), Valid: true}
	}
	if v, ok := params["ptsuffix"]; ok {
		dupParams.Ptsuffix = sql.NullString{String: v.(string), Valid: true}
	}
	if _, ok := params["ptdob"]; ok {
		dupParams.Ptdob = sql.NullTime{Valid: true}
	}

	results, err := model.Queries.PatientSearchDuplicates(r.Request.Context(), dupParams)
	if err != nil {
		log.Print(err.Error())
		r.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	r.JSON(http.StatusOK, results)
}

// patientCreate handles POST /patients — creates a new patient with address.
func patientCreate(r *gin.Context) {
	var input struct {
		FirstName     string `json:"first_name"`
		LastName      string `json:"last_name"`
		MiddleName    string `json:"middle_name"`
		NameSuffix    string `json:"name_suffix"`
		DateOfBirth   string `json:"date_of_birth"`
		Gender        string `json:"gender"`
		PatientID     string `json:"patient_id"`
		AddressLine1  string `json:"address_line_1"`
		AddressLine2  string `json:"address_line_2"`
		City          string `json:"city"`
		State         string `json:"state"`
		Zip           string `json:"zip"`
		Phone         string `json:"phone"`
	}

	if err := r.BindJSON(&input); err != nil {
		log.Print(err.Error())
		r.AbortWithError(http.StatusBadRequest, err)
		return
	}

	// Get the current user from the JWT session.
	session, err := common.GetSession(r)
	if err != nil {
		log.Printf("patientCreate: failed to get session: %v", err)
		r.AbortWithError(http.StatusUnauthorized, err)
		return
	}

	// Parse date of birth if provided.
	var dob interface{}
	if input.DateOfBirth != "" {
		parsed, err := common.ParseDate(input.DateOfBirth)
		if err != nil {
			r.AbortWithError(http.StatusBadRequest, fmt.Errorf("invalid date_of_birth: %v", err))
			return
		}
		dob = parsed
	}

	// Use a transaction for atomicity.
	tx, err := model.SqlDb.BeginTx(r.Request.Context(), nil)
	if err != nil {
		log.Printf("patientCreate: begin tx: %v", err)
		r.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	defer tx.Rollback()

	// Insert into patient table.
	result, err := tx.ExecContext(r.Request.Context(),
		`INSERT INTO patient
			(ptlname, ptfname, ptmname, ptsuffix, ptsex, ptid, ptdob, ptarchive, ptbilltype, user, stamp)
		 VALUES (?, ?, ?, ?, ?, ?, ?, 0, '', ?, NOW())`,
		input.LastName, input.FirstName, input.MiddleName, input.NameSuffix,
		input.Gender, input.PatientID, dob, session.UserId,
	)
	if err != nil {
		log.Printf("patientCreate: insert patient: %v", err)
		r.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	patientID, err := result.LastInsertId()
	if err != nil {
		log.Printf("patientCreate: get last insert id: %v", err)
		r.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	// Insert into patient_address table.
	_, err = tx.ExecContext(r.Request.Context(),
		`INSERT INTO patient_address
			(patient, line1, line2, city, stpr, postal, active)
		 VALUES (?, ?, ?, ?, ?, ?, 1)`,
		patientID, input.AddressLine1, input.AddressLine2,
		input.City, input.State, input.Zip,
	)
	if err != nil {
		log.Printf("patientCreate: insert address: %v", err)
		r.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	if err := tx.Commit(); err != nil {
		log.Printf("patientCreate: commit: %v", err)
		r.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	r.JSON(http.StatusCreated, gin.H{"id": patientID})
}
