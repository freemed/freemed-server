package api

import (
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

// SearchResult is the unified response type for global search.
type SearchResult struct {
	ID         int64  `json:"id"`
	Title      string `json:"title"`
	ResultType string `json:"result_type"`
	Label      string `json:"label,omitempty"`
	PatientID  string `json:"patient_id,omitempty"`
}

func search(c *gin.Context) {
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
		rows, err := model.Queries.SearchPatients(c.Request.Context(), dbgen.SearchPatientsParams{
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

	// Search messages
	wg.Add(1)
	go func() {
		defer wg.Done()
		rows, err := model.Queries.SearchMessages(c.Request.Context(), q)
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
		rows, err := model.Queries.SearchAppointments(c.Request.Context(), q)
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
