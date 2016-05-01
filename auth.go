package main

import (
	"github.com/appleboy/gin-jwt"
	"github.com/freemed/freemed-server/common"
	"github.com/freemed/freemed-server/model"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
	"time"
)

var (
	authMiddleware            *jwt.GinJWTMiddleware
	authMiddlewareInitialized bool
)

func getAuthMiddleware() *jwt.GinJWTMiddleware {
	if !authMiddlewareInitialized {
		authMiddleware = &jwt.GinJWTMiddleware{
			Realm:   "FreeMED",
			Key:     []byte(*SESSION_KEY),
			Timeout: time.Minute * time.Duration(*SESSION_LENGTH),
			Authenticator: func(userId string, password string, c *gin.Context) (string, bool) {
				_, res := model.CheckUserPassword(userId, password)
				return userId, res
			},
			Authorizator: func(userId string, c *gin.Context) bool {
				// TODO: FIXME: XXX
				return true
			},
			PayloadFunc: func(userId string) map[string]interface{} {
				user, err := model.GetUserByName(userId)
				if err != nil {
					log.Printf("PayloadFunc((): " + err.Error())
				}
				s, _ := common.ActiveSession.CreateSession(user.Id)
				//log.Printf("PayloadFunc(): user = %s, session = %v", user, s)
				return map[string]interface{}{
					"uid":        user.Id,
					"session":    s,
					"session_id": s.SessionId,
					"expires":    s.Expires,
				}
			},
			Unauthorized: func(c *gin.Context, code int, message string) {
				c.JSON(code, gin.H{
					"code":    code,
					"message": message,
				})
			},
		}
		authMiddlewareInitialized = true
	}
	return authMiddleware
}

func authMiddlewareLogout(c *gin.Context) {
	// As this exists outside of the normal middleware, we have to load it first,
	// *manually*. This is awful, but is the easiest way to keep it in the /auth
	// namespace.
	getAuthMiddleware().MiddlewareFunc()(c)

	session, err := common.GetSession(c)
	if err != nil {
		log.Printf("AuthLogout(): Expire session: %v", err)
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	log.Printf("AuthLogout(): Expire session %s", session.SessionId)
	common.ActiveSession.ExpireSession(session.SessionId)
	c.JSON(http.StatusOK, true)
}
