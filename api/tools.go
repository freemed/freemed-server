package api

import (
	"log"
	"net/http"

	"github.com/freemed/freemed-server/common"
	"github.com/freemed/freemed-server/model"
	"github.com/gin-gonic/gin"
)

func init() {
	common.ApiMap["tools"] = common.ApiMapping{
		Authenticated: true,
		RouterFunction: func(r *gin.RouterGroup) {
			r.GET("/", common.RequireRole("admin"), listTools)
			r.GET("/:id", common.RequireRole("admin"), getTool)
			r.POST("/:id/execute", common.RequireRole("admin"), executeTool)
		},
	}
}

// listTools handles GET /api/tools/
func listTools(r *gin.Context) {
	tools, err := model.Queries.ListTools(r.Request.Context())
	if err != nil {
		log.Print(err.Error())
		r.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	r.JSON(http.StatusOK, tools)
}

// getTool handles GET /api/tools/:id
func getTool(r *gin.Context) {
	id := r.Param("id")
	if id == "" {
		r.AbortWithStatus(http.StatusBadRequest)
		return
	}

	toolID := common.ParseInt(id)
	tool, err := model.Queries.GetTool(r.Request.Context(), toolID)
	if err != nil {
		log.Print(err.Error())
		r.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	r.JSON(http.StatusOK, tool)
}

// executeTool handles POST /api/tools/:id/execute
func executeTool(r *gin.Context) {
	id := r.Param("id")
	if id == "" {
		r.AbortWithStatus(http.StatusBadRequest)
		return
	}

	toolID := common.ParseInt(id)
	tool, err := model.Queries.GetTool(r.Request.Context(), toolID)
	if err != nil {
		log.Print(err.Error())
		r.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	// Parse any passed-in parameters from the JSON body
	var params map[string]interface{}
	if err := r.BindJSON(&params); err != nil {
		// Empty body is acceptable — use empty params
		params = map[string]interface{}{}
	}

	r.JSON(http.StatusOK, gin.H{
		"tool":       tool.ToolName,
		"parameters": params,
		"status":     "executed",
		"message":    "Tool execution stub — implement tool-specific logic",
	})
}
