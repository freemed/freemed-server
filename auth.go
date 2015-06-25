package main

import (
	//"github.com/go-martini/martini"
	"encoding/json"
	"github.com/martini-contrib/encoder"
	"github.com/martini-contrib/render"
	"log"
	"net/http"
)

type AuthLoginObj struct {
	Username string `json:"user" binding:"required"`
	Password string `json:"pass" binding:"required"`
}

func AuthLogin(data AuthLoginObj, enc encoder.Encoder, r render.Render) {
	log.Printf("user=%s", data.Username)
	uid, res := checkUserPassword(data.Username, data.Password)
	if res {
		s, err := createSession(uid)
		if err != nil {
			r.JSON(http.StatusInternalServerError, false)
			return
		}
		r.JSON(http.StatusOK, map[string]interface{}{
			"session_id": s.SessionId,
			"expires":    s.Expires,
		})
	}
	r.JSON(http.StatusOK, res)
}

func AuthLogout(req *http.Request, enc encoder.Encoder, r render.Render) {
	authHeader := req.Header.Get("Authorization")
	if authHeader == "" || len(authHeader) < 20 || authHeader[:7] != "Bearer " {
		log.Printf("Found Authorization header : %s", authHeader)
		r.JSON(http.StatusInternalServerError, false)
		return
	}
	sid := authHeader[7:]
	log.Printf("Expire session %s", sid)
	expireSession(sid)
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
