package api

import (
	"log"

	"github.com/freemed/freemed-server/common"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func init() {
	log.Print("Registering /api/metrics endpoint (Prometheus)")
	common.ApiMap["metrics"] = common.ApiMapping{
		Authenticated: false,
		RouterFunction: func(r *gin.RouterGroup) {
			r.GET("/", prometheusHandler)
		},
	}
}

func prometheusHandler(c *gin.Context) {
	promhttp.HandlerFor(prometheus.DefaultGatherer, promhttp.HandlerOpts{}).ServeHTTP(c.Writer, c.Request)
}
