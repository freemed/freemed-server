package api

import (
	"github.com/freemed/freemed-server/common"
	"github.com/freemed/freemed-server/model"
	"github.com/go-martini/martini"
	"github.com/martini-contrib/binding"
	"github.com/martini-contrib/encoder"
	"github.com/martini-contrib/render"
	"log"
	"net/http"
)

func init() {
	common.ApiMap["auth"] = common.ApiMapping{
		Authenticated: false,
		JsonArmored:   true,
		RouterFunction: func(r martini.Router) {
			r.Post("/login", binding.Json(AuthLoginObj{}), AuthLogin)
			r.Delete("/logout", AuthLogout)
		},
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
		s, err := common.ActiveSession.CreateSession(uid)
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
		log.Printf("AuthLogout(): Invalid Authorization header : %s", authHeader)
		r.JSON(http.StatusInternalServerError, false)
		return
	}
	sid := authHeader[7:]
	log.Printf("AuthLogout(): Expire session %s", sid)
	common.ActiveSession.ExpireSession(sid)
	r.JSON(http.StatusOK, true)
}
