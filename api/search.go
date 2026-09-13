package api

import (
	"context"
	"log"
	"net/http"
	"sync"

	"github.com/freemed/freemed-server/common"
	"github.com/freemed/freemed-server/dbgen"
	"github.com/freemed/freemed-server/model"
	"github.com/gin-gonic/gin"
)

func init() {
	common.ApiMap["search"] = common.ApiMapping{
		Authenticated: true,
		RouterFunction: func(r *gin.RouterGroup) {
			r.GET("", search)
		},
	}
}

// Search query seams. The three searches run concurrently, so the handler test
// needs each behind a var to run without a database — the same pattern
// api/dicom.go uses for dicomGetRow.
var (
	searchPatientsQuery = func(ctx context.Context, arg dbgen.SearchPatientsParams) ([]dbgen.SearchPatientsRow, error) {
		return model.Queries.SearchPatients(ctx, arg)
	}
	searchMessagesQuery = func(ctx context.Context, arg dbgen.SearchMessagesParams) ([]dbgen.SearchMessagesRow, error) {
		return model.Queries.SearchMessages(ctx, arg)
	}
	searchAppointmentsQuery = func(ctx context.Context, query interface{}) ([]dbgen.SearchAppointmentsRow, error) {
		return model.Queries.SearchAppointments(ctx, query)
	}
)

// SearchResult is the unified response type for global search.
type SearchResult struct {
	ID         int64  `json:"id"`
	Title      string `json:"title"`
	ResultType string `json:"result_type"`
	Label      string `json:"label,omitempty"`
	PatientID  string `json:"patient_id,omitempty"`
}

// search handles GET /api/search?q=...
//
// The message leg is scoped to the session user: SearchMessages used to match on
// msgsubject alone, so every message subject in the system was readable by any
// authenticated caller. The other two legs are patient/appointment metadata the
// staff session is already entitled to.
func search(c *gin.Context) {
	session, err := common.GetSession(c)
	if err != nil {
		log.Print(err.Error())
		common.ErrorResponseFromError(c, http.StatusUnauthorized, err)
		return
	}

	q := c.Query("q")
	if q == "" {
		c.JSON(http.StatusOK, []SearchResult{})
		return
	}

	var (
		wg       sync.WaitGroup
		mu       sync.Mutex
		results  []SearchResult
		errCount int
	)

	// Search patients
	wg.Add(1)
	go func() {
		defer wg.Done()
		rows, err := searchPatientsQuery(c.Request.Context(), dbgen.SearchPatientsParams{
			Query: q,
		})
		if err != nil {
			log.Printf("SearchPatients error: %s", err.Error())
			mu.Lock()
			errCount++
			mu.Unlock()
			return
		}
		mu.Lock()
		for _, row := range rows {
			results = append(results, SearchResult{
				ID:         row.ID,
				Title:      row.Ptlname + ", " + row.Ptfname,
				ResultType: row.ResultType,
				Label:      row.Ptid,
			})
		}
		mu.Unlock()
	}()

	// Search messages (the caller's own, as sender or recipient)
	wg.Add(1)
	go func() {
		defer wg.Done()
		rows, err := searchMessagesQuery(c.Request.Context(), dbgen.SearchMessagesParams{
			Query:  q,
			UserID: session.UserId,
		})
		if err != nil {
			log.Printf("SearchMessages error: %s", err.Error())
			mu.Lock()
			errCount++
			mu.Unlock()
			return
		}
		mu.Lock()
		for _, row := range rows {
			results = append(results, SearchResult{
				ID:         row.ID,
				Title:      row.Title,
				ResultType: row.ResultType,
			})
		}
		mu.Unlock()
	}()

	// Search appointments
	wg.Add(1)
	go func() {
		defer wg.Done()
		rows, err := searchAppointmentsQuery(c.Request.Context(), q)
		if err != nil {
			log.Printf("SearchAppointments error: %s", err.Error())
			mu.Lock()
			errCount++
			mu.Unlock()
			return
		}
		mu.Lock()
		for _, row := range rows {
			results = append(results, SearchResult{
				ID:         row.ID,
				Title:      row.Title,
				ResultType: row.ResultType,
			})
		}
		mu.Unlock()
	}()

	wg.Wait()

	c.JSON(http.StatusOK, results)
}
