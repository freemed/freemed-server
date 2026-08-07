package api

import (
	"database/sql"
	"log"
	"net/http"

	"github.com/freemed/freemed-server/common"
	"github.com/freemed/freemed-server/dbgen"
	"github.com/freemed/freemed-server/model"
	"github.com/gin-gonic/gin"
)

func init() {
	common.ApiMap["scheduler/templates"] = common.ApiMapping{
		Authenticated: true,
		RouterFunction: func(r *gin.RouterGroup) {
			r.GET("/", templateList)
			r.GET("/:id", templateGet)
			r.POST("/", common.RequireRole("admin"), templateCreate)
			r.PUT("/:id", common.RequireRole("admin"), templateUpdate)
			r.DELETE("/:id", common.RequireRole("admin"), templateDelete)
		},
	}
}

func templateList(c *gin.Context) {
	rows, err := model.Queries.ListTemplates(c.Request.Context())
	if err != nil {
		log.Print(err.Error())
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, rows)
}

func templateGet(c *gin.Context) {
	id := common.ParseInt(c.Param("id"))
	if id == 0 {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	row, err := model.Queries.GetTemplate(c.Request.Context(), id)
	if err != nil {
		log.Print(err.Error())
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, row)
}

type templateInput struct {
	Name      string  `json:"name" binding:"required"`
	Duration  int64   `json:"duration" binding:"required"`
	Equipment *string `json:"equipment"`
	Color     string  `json:"color"`
}

func templateCreate(c *gin.Context) {
	var in templateInput
	if err := c.BindJSON(&in); err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	var equip sql.NullString
	if in.Equipment != nil && *in.Equipment != "" {
		equip = sql.NullString{String: *in.Equipment, Valid: true}
	}

	result, err := model.Queries.CreateTemplate(c.Request.Context(), dbgen.CreateTemplateParams{
		Atname:      in.Name,
		Atduration:  in.Duration,
		Atequipment: equip,
		Atcolor:     in.Color,
	})
	if err != nil {
		log.Print(err.Error())
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	newID, _ := result.LastInsertId()
	c.JSON(http.StatusCreated, gin.H{"id": newID})
}

func templateUpdate(c *gin.Context) {
	id := common.ParseInt(c.Param("id"))
	if id == 0 {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	var in templateInput
	if err := c.BindJSON(&in); err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	var equip sql.NullString
	if in.Equipment != nil && *in.Equipment != "" {
		equip = sql.NullString{String: *in.Equipment, Valid: true}
	}

	err := model.Queries.UpdateTemplate(c.Request.Context(), dbgen.UpdateTemplateParams{
		ID:          id,
		Atname:      in.Name,
		Atduration:  in.Duration,
		Atequipment: equip,
		Atcolor:     in.Color,
	})
	if err != nil {
		log.Print(err.Error())
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "updated"})
}

func templateDelete(c *gin.Context) {
	id := common.ParseInt(c.Param("id"))
	if id == 0 {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	err := model.Queries.DeleteTemplate(c.Request.Context(), id)
	if err != nil {
		log.Print(err.Error())
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "deleted"})
}
