package api

import (
	"log"
	"net/http"
	"time"

	"github.com/freemed/freemed-server/common"
	"github.com/freemed/freemed-server/dbgen"
	"github.com/freemed/freemed-server/model"
	"github.com/gin-gonic/gin"
)

func init() {
	common.ApiMap["claimlog"] = common.ApiMapping{
		Authenticated: true,
		RouterFunction: func(r *gin.RouterGroup) {
			r.GET("/", common.RequireRole("admin"), claimLogList)
			r.POST("/", common.RequireRole("admin"), claimLogCreate)
		},
	}
}

type claimLogInput struct {
	Clprocedure int64  `json:"clprocedure" binding:"required"`
	Clpayrec    int64  `json:"clpayrec"`
	Claction    string `json:"claction"`
	Clcomment   string `json:"clcomment"`
	Clformat    string `json:"clformat"`
	Cltarget    string `json:"cltarget"`
	Cltargetopt string `json:"cltargetopt"`
	Clbillkey   int64  `json:"clbillkey"`
}

// claimLogList handles GET /api/claimlog?claim=ID
func claimLogList(r *gin.Context) {
	claimID := common.ParseInt(r.DefaultQuery("claim", "0"))
	if claimID == 0 {
		common.ErrorResponse(r, http.StatusBadRequest, "claim parameter is required")
		return
	}

	rows, err := model.Queries.ListClaimLog(r.Request.Context(), claimID)
	if err != nil {
		log.Print(err.Error())
		common.ErrorResponseFromError(r, http.StatusInternalServerError, err)
		return
	}
	r.JSON(http.StatusOK, rows)
}

// claimLogCreate handles POST /api/claimlog
func claimLogCreate(r *gin.Context) {
	var in claimLogInput
	if err := r.BindJSON(&in); err != nil {
		common.ErrorResponseFromError(r, http.StatusBadRequest, err)
		return
	}

	sess, err := common.GetSession(r)
	if err != nil {
		common.ErrorResponseFromError(r, http.StatusUnauthorized, err)
		return
	}

	result, err := model.Queries.InsertClaimLog(r.Request.Context(), dbgen.InsertClaimLogParams{
		Cltimestamp: time.Now(),
		Cluser:      sess.UserId,
		Clprocedure: in.Clprocedure,
		Clpayrec:    in.Clpayrec,
		Claction:    in.Claction,
		Clcomment:   in.Clcomment,
		Clformat:    in.Clformat,
		Cltarget:    in.Cltarget,
		Cltargetopt: in.Cltargetopt,
		Clbillkey:   in.Clbillkey,
	})
	if err != nil {
		log.Print(err.Error())
		common.ErrorResponseFromError(r, http.StatusInternalServerError, err)
		return
	}

	newID, _ := result.LastInsertId()
	r.JSON(http.StatusCreated, gin.H{"id": newID})
}
