package main

import (
	"encoding/json"
	"github.com/freemed/freemed-server/db"
	"github.com/freemed/freemed-server/model"
	"github.com/go-martini/martini"
	"github.com/martini-contrib/binding"
	"github.com/martini-contrib/encoder"
	"github.com/martini-contrib/render"
	"log"
	"net/http"
)

func init() {
	db.ApiMap["auth"] = func(r martini.Router) {
		r.Post("/login", binding.Json(AuthLoginObj{}), AuthLogin)
		r.Delete("/logout", AuthLogout)
	}
}

type AuthLoginObj struct {
	Username string `json:"user" binding:"required"`
	Password string `json:"pass" binding:"required"`
}

func AuthLogin(data AuthLoginObj, enc encoder.Encoder, r render.Render) {
	log.Printf("AuthLogin(): user=%s", data.Username)
	uid, res := model.CheckUserPassword(data.Username, data.Password)
	if res {
		s, err := model.CreateSession(uid)
		if err != nil {
			log.Printf("AuthLogin(): " + err.Error())
			r.JSON(http.StatusInternalServerError, false)
			return
		}
		r.JSON(http.StatusOK, map[string]interface{}{
			"session_id": s.SessionId,
			"expires":    s.Expires,
		})
		return
	}
	r.JSON(http.StatusUnauthorized, false)
}

func AuthLogout(req *http.Request, enc encoder.Encoder, r render.Render) {
	authHeader := req.Header.Get("Authorization")
	if authHeader == "" || len(authHeader) < 20 || authHeader[:7] != "Bearer " {
		log.Printf("AuthLogout(): Found Authorization header : %s", authHeader)
		r.JSON(http.StatusInternalServerError, false)
		return
	}
	sid := authHeader[7:]
	log.Printf("AuthLogout(): Expire session %s", sid)
	model.ExpireSession(sid)
	r.JSON(http.StatusOK, true)
}

func jsonEncode(o interface{}) []byte {
	b, err := json.Marshal(o)
	if err != nil {
		log.Print(err.Error())
		return []byte("false")
	}
	return b
}
