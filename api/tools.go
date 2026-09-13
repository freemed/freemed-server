package api

import (
	"context"
	"database/sql"
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
			r.GET("/verify-backup", common.RequireRole("admin"), verifyBackup)
		},
	}
}

// toolHandler is the Go implementation of a registered tool. The `tools` table
// is a legacy registry carried over from the PHP application, where tool_class
// named a PHP class implementing the tool. Those classes do not exist in Go, so
// a tool is only executable if a Go handler is registered for its tool_name.
type toolHandler func(ctx context.Context, params map[string]interface{}) (map[string]interface{}, error)

// toolHandlers maps tools.tool_name to its Go implementation. Empty until
// individual tools are ported; unknown names are reported as not implemented
// rather than silently succeeding.
var toolHandlers = map[string]toolHandler{}

// listTools handles GET /api/tools/
func listTools(r *gin.Context) {
	tools, err := model.Queries.ListTools(r.Request.Context())
	if err != nil {
		log.Print(err.Error())
		common.ErrorResponseFromError(r, http.StatusInternalServerError, err)
		return
	}

	// Report executability so callers do not have to guess from tool_class.
	type toolView struct {
		ID          int64  `json:"id"`
		ToolName    string `json:"tool_name"`
		Description string `json:"tool_description"`
		ToolClass   string `json:"tool_class"`
		Parameters  string `json:"tool_parameters"`
		Executable  bool   `json:"executable"`
	}
	out := make([]toolView, 0, len(tools))
	for _, t := range tools {
		desc, params := "", ""
		if t.ToolDescription.Valid {
			desc = t.ToolDescription.String
		}
		if t.ToolParameters.Valid {
			params = t.ToolParameters.String
		}
		_, executable := toolHandlers[t.ToolName]
		out = append(out, toolView{
			ID:          t.ID,
			ToolName:    t.ToolName,
			Description: desc,
			ToolClass:   t.ToolClass,
			Parameters:  params,
			Executable:  executable,
		})
	}

	r.JSON(http.StatusOK, out)
}

// getTool handles GET /api/tools/:id
func getTool(r *gin.Context) {
	id := r.Param("id")
	if id == "" {
		common.ErrorResponse(r, http.StatusBadRequest, "bad request")
		return
	}

	toolID := common.ParseInt(id)
	tool, err := model.Queries.GetTool(r.Request.Context(), toolID)
	if err != nil {
		if err == sql.ErrNoRows {
			common.ErrorResponse(r, http.StatusNotFound, "tool not found")
			return
		}
		log.Print(err.Error())
		common.ErrorResponseFromError(r, http.StatusInternalServerError, err)
		return
	}

	r.JSON(http.StatusOK, tool)
}

// executeTool handles POST /api/tools/:id/execute
//
// The tools registry is metadata-only in this codebase: tool_class referenced
// PHP classes from the 0.9.x application that have no Go equivalent. A tool is
// executed only when a Go handler is registered for its tool_name. Anything
// else returns 501 so callers are never told a side effect happened when none
// did — the previous implementation always answered "executed".
func executeTool(r *gin.Context) {
	id := r.Param("id")
	if id == "" {
		common.ErrorResponse(r, http.StatusBadRequest, "bad request")
		return
	}

	toolID := common.ParseInt(id)
	if toolID == 0 {
		common.ErrorResponse(r, http.StatusBadRequest, "invalid tool id")
		return
	}

	tool, err := model.Queries.GetTool(r.Request.Context(), toolID)
	if err != nil {
		if err == sql.ErrNoRows {
			common.ErrorResponse(r, http.StatusNotFound, "tool not found")
			return
		}
		log.Print(err.Error())
		common.ErrorResponseFromError(r, http.StatusInternalServerError, err)
		return
	}

	handler, ok := toolHandlers[tool.ToolName]
	if !ok {
		// tool_class names a PHP class that does not exist in this codebase;
		// echo the declared parameter template so an operator can see what the
		// tool expects while it remains unimplemented.
		params := ""
		if tool.ToolParameters.Valid {
			params = tool.ToolParameters.String
		}
		r.JSON(http.StatusNotImplemented, gin.H{
			"code":            http.StatusNotImplemented,
			"message":         "tool is registered but has no Go implementation; nothing was executed",
			"tool":            tool.ToolName,
			"tool_class":      tool.ToolClass,
			"tool_parameters": params,
		})
		return
	}

	var params map[string]interface{}
	if err := r.BindJSON(&params); err != nil {
		// Empty/absent body is acceptable — run with no parameters.
		params = map[string]interface{}{}
	}

	result, err := handler(r.Request.Context(), params)
	if err != nil {
		log.Printf("executeTool: %s failed: %v", tool.ToolName, err)
		common.ErrorResponseFromError(r, http.StatusInternalServerError, err)
		return
	}

	r.JSON(http.StatusOK, gin.H{
		"tool":       tool.ToolName,
		"parameters": params,
		"status":     "executed",
		"result":     result,
	})
}
