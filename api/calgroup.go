package api

import (
	"log"
	"net/http"

	"github.com/freemed/freemed-server/common"
	"github.com/freemed/freemed-server/model"
	"github.com/gin-gonic/gin"
)

// listCalGroups handles GET /api/scheduler/groups
func listCalGroups(c *gin.Context) {
	rows, err := model.Queries.ListCalGroups(c.Request.Context())
	if err != nil {
		log.Print(err.Error())
		common.ErrorResponseFromError(c, http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, rows)
}

// getCalGroup handles GET /api/scheduler/groups/:id
func getCalGroup(c *gin.Context) {
	id := common.ParseInt(c.Param("id"))
	if id == 0 {
		common.ErrorResponse(c, http.StatusBadRequest, "bad request")
		return
	}

	row, err := model.Queries.GetCalGroup(c.Request.Context(), id)
	if err != nil {
		log.Print(err.Error())
		common.ErrorResponseFromError(c, http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, row)
}
