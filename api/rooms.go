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
	common.ApiMap["rooms"] = common.ApiMapping{
		Authenticated: true,
		RouterFunction: func(r *gin.RouterGroup) {
			r.GET("/", listRooms)
			r.POST("/", common.RequireRole("admin"), createRoom)
			r.PUT("/:id", common.RequireRole("admin"), updateRoom)
			r.DELETE("/:id", common.RequireRole("admin"), deleteRoom)
		},
	}
}

func listRooms(c *gin.Context) {
	rooms, err := model.Queries.ListRooms(c.Request.Context())
	if err != nil {
		log.Printf("listRooms: %s", err.Error())
		common.ErrorResponseFromError(c, http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, rooms)
}

type roomInput struct {
	Name       string `json:"name" binding:"required"`
	FacilityID int64  `json:"facility_id"`
	RoomType   string `json:"room_type"`
}

func createRoom(c *gin.Context) {
	var in roomInput
	if err := c.BindJSON(&in); err != nil {
		common.ErrorResponseFromError(c, http.StatusBadRequest, err)
		return
	}

	result, err := model.Queries.CreateRoom(c.Request.Context(), dbgen.CreateRoomParams{
		Name:       in.Name,
		FacilityID: in.FacilityID,
		RoomType:   in.RoomType,
	})
	if err != nil {
		log.Printf("createRoom: %s", err.Error())
		common.ErrorResponseFromError(c, http.StatusInternalServerError, err)
		return
	}

	newID, _ := result.LastInsertId()
	c.JSON(http.StatusCreated, gin.H{"id": newID})
}

func updateRoom(c *gin.Context) {
	id := common.ParseInt(c.Param("id"))
	if id == 0 {
		common.ErrorResponse(c, http.StatusBadRequest, "invalid room id")
		return
	}

	var in roomInput
	if err := c.BindJSON(&in); err != nil {
		common.ErrorResponseFromError(c, http.StatusBadRequest, err)
		return
	}

	err := model.Queries.UpdateRoom(c.Request.Context(), dbgen.UpdateRoomParams{
		ID:         id,
		Name:       in.Name,
		FacilityID: in.FacilityID,
		RoomType:   in.RoomType,
	})
	if err != nil {
		log.Printf("updateRoom: %s", err.Error())
		common.ErrorResponseFromError(c, http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, true)
}

func deleteRoom(c *gin.Context) {
	id := common.ParseInt(c.Param("id"))
	if id == 0 {
		common.ErrorResponse(c, http.StatusBadRequest, "invalid room id")
		return
	}

	err := model.Queries.DeactivateRoom(c.Request.Context(), id)
	if err != nil {
		log.Printf("deleteRoom: %s", err.Error())
		common.ErrorResponseFromError(c, http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, true)
}
