package api

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/freemed/freemed-server/dbgen"
	"github.com/gin-gonic/gin"
)

// stubDicomQueries replaces the DICOM query seams with in-memory stubs so the
// ownership guard can be exercised without a live database. It returns a pointer
// to the slice recording delete calls (used to prove the destructive path is
// gated) and restores the real implementation when the test finishes.
func stubDicomQueries(t *testing.T, row dbgen.Dicom, getErr error) *[]int64 {
	t.Helper()
	origGet, origDelete := dicomGetRow, dicomDeleteRow
	deleted := &[]int64{}

	dicomGetRow = func(ctx context.Context, id int64) (dbgen.Dicom, error) {
		if getErr != nil {
			return dbgen.Dicom{}, getErr
		}
		if id != row.ID {
			return dbgen.Dicom{}, sql.ErrNoRows
		}
		return row, nil
	}
	dicomDeleteRow = func(ctx context.Context, id int64) error {
		*deleted = append(*deleted, id)
		return nil
	}
	t.Cleanup(func() { dicomGetRow, dicomDeleteRow = origGet, origDelete })
	return deleted
}

func dicomTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/api/patient/:id/dicom/:itemId", dicomGet)
	r.DELETE("/api/patient/:id/dicom/:itemId", dicomRemove)
	return r
}

func dicomRow(id, owner int64, payload string) dbgen.Dicom {
	return dbgen.Dicom{
		ID:       id,
		DPatient: owner,
		DData:    sql.NullString{String: payload, Valid: true},
	}
}

// TestDicomGetRejectsCrossPatientObject is the H1 regression test for the read
// path: an object owned by patient 2 must not be readable through patient 5's
// route. Before the ownership check existed this returned 200 plus the other
// patient's DICOM bytes.
func TestDicomGetRejectsCrossPatientObject(t *testing.T) {
	const secret = "PATIENT-2-CONFIDENTIAL-DICOM-BYTES"
	row := dicomRow(1, 2, secret)
	stubDicomQueries(t, row, nil)

	r := dicomTestRouter()
	for _, path := range []string{
		"/api/patient/5/dicom/1",     // wrong (existing) patient
		"/api/patient/99999/dicom/1", // non-existent patient
	} {
		req := httptest.NewRequest(http.MethodGet, path, nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("GET %s: status = %d, want 404", path, w.Code)
		}
		if strings.Contains(w.Body.String(), secret) {
			t.Errorf("GET %s: response leaked the owning patient's DICOM payload", path)
		}
		if !strings.Contains(w.Body.String(), "DICOM object not found") {
			t.Errorf("GET %s: body = %q, want a not-found message", path, w.Body.String())
		}
	}
}

// TestDicomGetAllowsOwningPatient proves the guard does not break legitimate
// access.
func TestDicomGetAllowsOwningPatient(t *testing.T) {
	const payload = "PATIENT-2-OWN-DICOM-BYTES"
	stubDicomQueries(t, dicomRow(1, 2, payload), nil)

	req := httptest.NewRequest(http.MethodGet, "/api/patient/2/dicom/1", nil)
	w := httptest.NewRecorder()
	dicomTestRouter().ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200 (body %q)", w.Code, w.Body.String())
	}
	if got := w.Body.String(); got != payload {
		t.Errorf("body = %q, want %q", got, payload)
	}
	if ct := w.Header().Get("Content-Type"); ct != "application/dicom" {
		t.Errorf("Content-Type = %q, want application/dicom", ct)
	}
}

// TestDicomGetBadIDs covers the 400 path for zero/non-numeric path parameters.
func TestDicomGetBadIDs(t *testing.T) {
	stubDicomQueries(t, dicomRow(1, 2, "x"), nil)

	for _, path := range []string{
		"/api/patient/5/dicom/0",
		"/api/patient/0/dicom/1",
		"/api/patient/0/dicom/0",
		"/api/patient/abc/dicom/1",
	} {
		req := httptest.NewRequest(http.MethodGet, path, nil)
		w := httptest.NewRecorder()
		dicomTestRouter().ServeHTTP(w, req)
		if w.Code != http.StatusBadRequest {
			t.Errorf("GET %s: status = %d, want 400", path, w.Code)
		}
	}
}

// TestDicomGetMissingObject keeps the original sql.ErrNoRows behaviour.
func TestDicomGetMissingObject(t *testing.T) {
	stubDicomQueries(t, dbgen.Dicom{}, sql.ErrNoRows)

	req := httptest.NewRequest(http.MethodGet, "/api/patient/2/dicom/1", nil)
	w := httptest.NewRecorder()
	dicomTestRouter().ServeHTTP(w, req)
	if w.Code != http.StatusNotFound {
		t.Errorf("status = %d, want 404", w.Code)
	}
}

// TestDicomRemoveRejectsCrossPatientObject is the destructive half of H1: a
// DELETE addressed to the wrong patient must not reach DeleteDicom at all.
func TestDicomRemoveRejectsCrossPatientObject(t *testing.T) {
	deleted := stubDicomQueries(t, dicomRow(1, 2, "bytes"), nil)

	r := dicomTestRouter()
	for _, path := range []string{"/api/patient/5/dicom/1", "/api/patient/99999/dicom/1"} {
		req := httptest.NewRequest(http.MethodDelete, path, nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("DELETE %s: status = %d, want 404", path, w.Code)
		}
		if len(*deleted) != 0 {
			t.Fatalf("DELETE %s: DeleteDicom was called with %v — the cross-patient delete is not gated", path, *deleted)
		}
	}
}

// TestDicomRemoveAllowsOwningPatient proves legitimate soft-deletion still works.
func TestDicomRemoveAllowsOwningPatient(t *testing.T) {
	deleted := stubDicomQueries(t, dicomRow(1, 2, "bytes"), nil)

	req := httptest.NewRequest(http.MethodDelete, "/api/patient/2/dicom/1", nil)
	w := httptest.NewRecorder()
	dicomTestRouter().ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200 (body %q)", w.Code, w.Body.String())
	}
	if len(*deleted) != 1 || (*deleted)[0] != 1 {
		t.Errorf("DeleteDicom calls = %v, want [1]", *deleted)
	}
	if !strings.Contains(w.Body.String(), "deleted") {
		t.Errorf("body = %q, want a deleted status", w.Body.String())
	}
}

/* ------------------------------------------------------------------ */
/* DICOMweb patient scoping (audit finding M10)                        */
/* ------------------------------------------------------------------ */

const (
	wadoStudyUID  = "1.2.840.113619.2.STUDY"
	wadoSeriesUID = "1.2.840.113619.2.SERIES"
	wadoSOPUID    = "1.2.840.113619.2.SOP"
)

// dicomWebTestRouter mounts the DICOMweb routes exactly as api/dicom.go's
// ApiMap registers them, including the patient context segment. The pre-scoping
// bare routes are deliberately absent (that is what the real server does too).
func dicomWebTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/api/dicom/patient/:id/studies", dicomQidoStudies)
	r.GET("/api/dicom/patient/:id/studies/:studyUID/series/:seriesUID/instances/:sopUID", dicomWadoRetrieve)
	return r
}

func wadoPathFor(patient int64) string {
	return fmt.Sprintf("/api/dicom/patient/%d/studies/%s/series/%s/instances/%s",
		patient, wadoStudyUID, wadoSeriesUID, wadoSOPUID)
}

// stubDicomBySop answers every UID triple with the supplied row — exactly what
// the real SQL does once the UIDs match — so the patient guard is the only thing
// that can reject a cross-patient request.
func stubDicomBySop(t *testing.T, row dbgen.Dicom, err error) {
	t.Helper()
	orig := dicomGetBySop
	dicomGetBySop = func(context.Context, dbgen.GetDicomBySopParams) (dbgen.Dicom, error) {
		if err != nil {
			return dbgen.Dicom{}, err
		}
		return row, nil
	}
	t.Cleanup(func() { dicomGetBySop = orig })
}

// stubListStudies serves a fixed row set and, when params is non-nil, records
// every query the handler built.
func stubListStudies(t *testing.T, rows []dbgen.ListDicomStudiesRow, params *[]dbgen.ListDicomStudiesParams) {
	t.Helper()
	orig := dicomListStudies
	dicomListStudies = func(_ context.Context, arg dbgen.ListDicomStudiesParams) ([]dbgen.ListDicomStudiesRow, error) {
		if params != nil {
			*params = append(*params, arg)
		}
		return rows, nil
	}
	t.Cleanup(func() { dicomListStudies = orig })
}

// TestDicomWadoRejectsCrossPatientInstance is the M10 regression test for the
// retrieve path: an instance owned by patient 2 must not be retrievable through
// patient 5's route, even though the caller knows all three UIDs. Before the
// guard existed this returned 200 plus the owning patient's pixels.
func TestDicomWadoRejectsCrossPatientInstance(t *testing.T) {
	const secret = "PATIENT-2-CONFIDENTIAL-PIXELS"
	stubDicomBySop(t, dicomRow(1, 2, secret), nil)

	r := dicomWebTestRouter()
	for _, patient := range []int64{5, 99999} {
		req := httptest.NewRequest(http.MethodGet, wadoPathFor(patient), nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("GET patient %d: status = %d, want 404", patient, w.Code)
		}
		if strings.Contains(w.Body.String(), secret) {
			t.Errorf("GET patient %d: response leaked the owning patient's DICOM payload", patient)
		}
		if !strings.Contains(w.Body.String(), "SOP instance not found") {
			t.Errorf("GET patient %d: body = %q, want the same not-found message as a missing instance", patient, w.Body.String())
		}
	}
}

// TestDicomWadoCrossPatientIsIndistinguishableFromMissingInstance pins the
// 404-not-403 requirement: the two responses must be byte-identical so the
// endpoint cannot be used to confirm that another patient's instance exists.
func TestDicomWadoCrossPatientIsIndistinguishableFromMissingInstance(t *testing.T) {
	// Cross-patient: the row exists but belongs to patient 2.
	stubDicomBySop(t, dicomRow(1, 2, "pixels"), nil)
	w1 := httptest.NewRecorder()
	dicomWebTestRouter().ServeHTTP(w1, httptest.NewRequest(http.MethodGet, wadoPathFor(5), nil))

	// Missing: no row at all.
	stubDicomBySop(t, dbgen.Dicom{}, sql.ErrNoRows)
	w2 := httptest.NewRecorder()
	dicomWebTestRouter().ServeHTTP(w2, httptest.NewRequest(http.MethodGet, wadoPathFor(5), nil))

	if w1.Code != http.StatusNotFound || w2.Code != http.StatusNotFound {
		t.Fatalf("statuses = %d and %d, want 404 and 404", w1.Code, w2.Code)
	}
	if w1.Body.String() != w2.Body.String() {
		t.Errorf("cross-patient body %q differs from missing-instance body %q", w1.Body.String(), w2.Body.String())
	}
}

// TestDicomWadoAllowsOwningPatient proves the guard does not break the viewer.
func TestDicomWadoAllowsOwningPatient(t *testing.T) {
	const payload = "PATIENT-2-OWN-PIXELS"
	stubDicomBySop(t, dicomRow(1, 2, payload), nil)

	w := httptest.NewRecorder()
	dicomWebTestRouter().ServeHTTP(w, httptest.NewRequest(http.MethodGet, wadoPathFor(2), nil))

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200 (body %q)", w.Code, w.Body.String())
	}
	if got := w.Body.String(); got != payload {
		t.Errorf("body = %q, want %q", got, payload)
	}
	if ct := w.Header().Get("Content-Type"); ct != "application/dicom" {
		t.Errorf("Content-Type = %q, want application/dicom", ct)
	}
}

// TestDicomWadoRequiresPatientContext covers a zero / non-numeric patient
// segment. (A path with empty UID segments does not match the route at all and
// falls through to the 404 handler — also acceptable, never a 200.)
func TestDicomWadoRequiresPatientContext(t *testing.T) {
	stubDicomBySop(t, dicomRow(1, 2, "pixels"), nil)

	for _, path := range []string{
		"/api/dicom/patient/0/studies/" + wadoStudyUID + "/series/" + wadoSeriesUID + "/instances/" + wadoSOPUID,
		"/api/dicom/patient/abc/studies/" + wadoStudyUID + "/series/" + wadoSeriesUID + "/instances/" + wadoSOPUID,
	} {
		w := httptest.NewRecorder()
		dicomWebTestRouter().ServeHTTP(w, httptest.NewRequest(http.MethodGet, path, nil))
		if w.Code != http.StatusBadRequest {
			t.Errorf("GET %s: status = %d, want 400", path, w.Code)
		}
	}

	// Empty UID segments never reach the handler.
	w := httptest.NewRecorder()
	dicomWebTestRouter().ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/dicom/patient/2/studies//series//instances/", nil))
	if w.Code == http.StatusOK {
		t.Errorf("GET with empty UID segments: status = %d, want a non-200", w.Code)
	}
}

// TestDicomQidoStudiesScopesToPathPatient proves the study list is bound to the
// path patient: the query is built with Patient = the path id, while the DICOM
// PatientID query parameter stays a separate attribute filter.
func TestDicomQidoStudiesScopesToPathPatient(t *testing.T) {
	params := &[]dbgen.ListDicomStudiesParams{}
	stubListStudies(t, []dbgen.ListDicomStudiesRow{{
		DStudyUid: sql.NullString{String: wadoStudyUID, Valid: true},
		DPatient:  7,
	}}, params)

	req := httptest.NewRequest(http.MethodGet, "/api/dicom/patient/7/studies?PatientID=999&StudyInstanceUID="+wadoStudyUID, nil)
	w := httptest.NewRecorder()
	dicomWebTestRouter().ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200 (body %q)", w.Code, w.Body.String())
	}
	if len(*params) != 1 {
		t.Fatalf("ListDicomStudies calls = %d, want 1", len(*params))
	}
	got := (*params)[0]
	if got.Patient != 7 {
		t.Errorf("query Patient = %d, want 7 (the patient segment)", got.Patient)
	}
	if !got.PatientID.Valid || got.PatientID.String != "999" {
		t.Errorf("query PatientID = %+v, want the DICOM attribute filter 999", got.PatientID)
	}
	if !got.StudyUid.Valid || got.StudyUid.String != wadoStudyUID {
		t.Errorf("query StudyUid = %+v, want %q", got.StudyUid, wadoStudyUID)
	}
	if !strings.Contains(w.Body.String(), wadoStudyUID) {
		t.Errorf("body = %q, want the study row the query returned", w.Body.String())
	}
}

// TestDicomQidoStudiesCannotWidenScope proves no query parameter can change the
// patient scope — only the path segment reaches the query.
func TestDicomQidoStudiesCannotWidenScope(t *testing.T) {
	params := &[]dbgen.ListDicomStudiesParams{}
	stubListStudies(t, nil, params)

	req := httptest.NewRequest(http.MethodGet, "/api/dicom/patient/7/studies?patient=2&patient_id=2&id=2&PatientID=2", nil)
	w := httptest.NewRecorder()
	dicomWebTestRouter().ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200 (body %q)", w.Code, w.Body.String())
	}
	if len(*params) != 1 {
		t.Fatalf("ListDicomStudies calls = %d, want 1", len(*params))
	}
	if (*params)[0].Patient != 7 {
		t.Errorf("query Patient = %d, want 7 — a query parameter widened the patient scope", (*params)[0].Patient)
	}
}

// TestDicomQidoStudiesBadPatientID covers the zero / non-numeric segment.
func TestDicomQidoStudiesBadPatientID(t *testing.T) {
	params := &[]dbgen.ListDicomStudiesParams{}
	stubListStudies(t, nil, params)

	for _, path := range []string{"/api/dicom/patient/0/studies", "/api/dicom/patient/abc/studies"} {
		w := httptest.NewRecorder()
		dicomWebTestRouter().ServeHTTP(w, httptest.NewRequest(http.MethodGet, path, nil))
		if w.Code != http.StatusBadRequest {
			t.Errorf("GET %s: status = %d, want 400", path, w.Code)
		}
	}
	if len(*params) != 0 {
		t.Errorf("ListDicomStudies was called %d times for an invalid patient id", len(*params))
	}
}

// TestDicomQidoStudiesReturnsScopedRowsOnly proves the response carries exactly
// the rows the (already patient-filtered) query returned.
func TestDicomQidoStudiesReturnsScopedRowsOnly(t *testing.T) {
	stubListStudies(t, []dbgen.ListDicomStudiesRow{
		{DStudyUid: sql.NullString{String: "1.2.3.A", Valid: true}, DPatient: 7},
		{DStudyUid: sql.NullString{String: "1.2.3.B", Valid: true}, DPatient: 7},
	}, nil)

	w := httptest.NewRecorder()
	dicomWebTestRouter().ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/dicom/patient/7/studies", nil))

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200 (body %q)", w.Code, w.Body.String())
	}
	body := w.Body.String()
	if !strings.Contains(body, "1.2.3.A") || !strings.Contains(body, "1.2.3.B") {
		t.Errorf("body = %q, want both study rows", body)
	}
}

// TestDicomBareStudiesRoutesAreGone proves the pre-scoping routes no longer
// exist, so there is no unscoped path left to enumerate studies or pull
// instances from. (In the real server these fall through to the JSON 404
// handler; here the router simply does not register them.)
func TestDicomBareStudiesRoutesAreGone(t *testing.T) {
	stubDicomBySop(t, dicomRow(1, 2, "pixels"), nil)
	stubListStudies(t, []dbgen.ListDicomStudiesRow{{DStudyUid: sql.NullString{String: wadoStudyUID, Valid: true}}}, nil)

	for _, path := range []string{
		"/api/dicom/studies",
		"/api/dicom/studies/" + wadoStudyUID + "/series/" + wadoSeriesUID + "/instances/" + wadoSOPUID,
	} {
		w := httptest.NewRecorder()
		dicomWebTestRouter().ServeHTTP(w, httptest.NewRequest(http.MethodGet, path, nil))
		if w.Code != http.StatusNotFound {
			t.Errorf("GET %s: status = %d, want 404 (route must not exist)", path, w.Code)
		}
	}
}
