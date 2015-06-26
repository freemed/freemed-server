package main

import (
	"github.com/martini-contrib/encoder"
	"github.com/martini-contrib/render"
	"log"
	"net/http"
)

type MessagesUserObj struct {
	Username string `json:"username" binding:"required"`
	Id       string `json:"id" binding:"required"`
}

func MessagesListUsers(enc encoder.Encoder, r render.Render) {
	var o []MessagesUserObj
	_, err := dbmap.Select(&o, "SELECT username, id FROM user")
	if err != nil {
		log.Print(err.Error())
		r.JSON(http.StatusInternalServerError, false)
		return
	}
	r.JSON(http.StatusOK, o)
	return
}
