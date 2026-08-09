package api

import (
	"log"
	"net/http"

	"github.com/freemed/freemed-server/common"
	"github.com/freemed/freemed-server/dbgen"
	"github.com/freemed/freemed-server/model"
	"github.com/gin-gonic/gin"
)

func init() {
	common.ApiMap["lock"] = common.ApiMapping{
		Authenticated: true,
		RouterFunction: func(r *gin.RouterGroup) {
			r.POST("/acquire", acquireLock)
			r.DELETE("/release", releaseLock)
		},
	}
}

type lockInput struct {
	TableName string `json:"table_name" binding:"required"`
	RecordID  int64  `json:"record_id" binding:"required"`
}

func acquireLock(c *gin.Context) {
	session, err := common.GetSession(c)
	if err != nil {
		common.ErrorResponseFromError(c, http.StatusUnauthorized, err)
		return
	}

	var in lockInput
	if err := c.BindJSON(&in); err != nil {
		common.ErrorResponseFromError(c, http.StatusBadRequest, err)
		return
	}

	_, err = model.Queries.AcquireRecordLock(c.Request.Context(), dbgen.AcquireRecordLockParams{
		TableName: in.TableName,
		RecordID:  in.RecordID,
		UserID:    session.UserId,
	})
	if err != nil {
		log.Printf("acquireLock: %s", err.Error())
		common.ErrorResponseFromError(c, http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"locked": true})
}

type releaseInput struct {
	TableName string `json:"table_name" binding:"required"`
	RecordID  int64  `json:"record_id" binding:"required"`
}

func releaseLock(c *gin.Context) {
	session, err := common.GetSession(c)
	if err != nil {
		common.ErrorResponseFromError(c, http.StatusUnauthorized, err)
		return
	}

	var in releaseInput
	if err := c.BindJSON(&in); err != nil {
		common.ErrorResponseFromError(c, http.StatusBadRequest, err)
		return
	}

	err = model.Queries.ReleaseRecordLock(c.Request.Context(), dbgen.ReleaseRecordLockParams{
		TableName: in.TableName,
		RecordID:  in.RecordID,
		UserID:    session.UserId,
	})
	if err != nil {
		log.Printf("releaseLock: %s", err.Error())
		common.ErrorResponseFromError(c, http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"released": true})
}
