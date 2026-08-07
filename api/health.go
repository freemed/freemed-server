package api

import (
	"log"
	"net/http"

	"github.com/freemed/freemed-server/common"
	"github.com/freemed/freemed-server/model"
	"github.com/gin-gonic/gin"
)

func init() {
	common.ApiMap["health"] = common.ApiMapping{
		Authenticated: false,
		RouterFunction: func(r *gin.RouterGroup) {
			r.GET("/", healthCheck)
		},
	}
}

func healthCheck(c *gin.Context) {
	dbOK := true
	redisOK := true

	// Check database connectivity
	if model.SqlDb != nil {
		if err := model.SqlDb.Ping(); err != nil {
			log.Printf("health: database ping failed: %v", err)
			dbOK = false
		}
	} else {
		log.Print("health: SqlDb is nil")
		dbOK = false
	}

	// Check Redis connectivity
	if common.ActiveSession != nil {
		if err := common.ActiveSession.Ping(); err != nil {
			log.Printf("health: redis ping failed: %v", err)
			redisOK = false
		}
	} else {
		log.Print("health: ActiveSession is nil")
		redisOK = false
	}

	status := "ok"
	httpStatus := http.StatusOK
	if !dbOK || !redisOK {
		status = "degraded"
		httpStatus = http.StatusServiceUnavailable
	}

	c.JSON(httpStatus, gin.H{
		"status":   status,
		"database": dbOK,
		"redis":    redisOK,
	})
}
