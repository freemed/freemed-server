package api

import (
	"log"
	"net/http"

	"github.com/freemed/freemed-server/common"
	"github.com/gin-gonic/gin"

	"github.com/freemed/freemed-server/model"
)

func init() {
	common.ApiMap["procedures"] = common.ApiMapping{
		Authenticated: true,
		RouterFunction: func(r *gin.RouterGroup) {
			r.GET("/:id", getProcedure)
		},
	}
}

func getProcedure(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	procedureID := common.ParseInt(id)
	procedure, err := model.Queries.GetProcedure(c.Request.Context(), procedureID)
	if err != nil {
		log.Print(err.Error())
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, procedure)
}

func patientProcedures(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	patientID := common.ParseInt(id)
	procedures, err := model.Queries.PatientProcedures(c.Request.Context(), patientID)
	if err != nil {
		log.Print(err.Error())
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, procedures)
}
