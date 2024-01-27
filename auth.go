package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	jwt "github.com/appleboy/gin-jwt/v2"
	"github.com/freemed/freemed-server/common"
	"github.com/freemed/freemed-server/config"
	"github.com/freemed/freemed-server/model"
	"github.com/gin-gonic/gin"
)

var (
	authMiddleware            *jwt.GinJWTMiddleware
	authMiddlewareInitialized bool
	identityKey               = "id"
)

type login struct {
	Username string `form:"username" json:"username" binding:"required"`
	Password string `form:"password" json:"password" binding:"required"`
}

func getAuthMiddleware() *jwt.GinJWTMiddleware {
	var err error
	if !authMiddlewareInitialized {
		authMiddleware, err = jwt.New(&jwt.GinJWTMiddleware{
			Realm:       "FreeMED",
			Key:         []byte(config.Config.Session.Key),
			Timeout:     time.Minute * time.Duration(config.Config.Session.Expiry),
			MaxRefresh:  time.Hour,
			IdentityKey: identityKey,
			Authenticator: func(c *gin.Context) (interface{}, error) {
				var loginVals login
				if err := c.ShouldBind(&loginVals); err != nil {
					return nil, jwt.ErrMissingLoginValues
				}
				userID := loginVals.Username
				password := loginVals.Password

				id, res := model.CheckUserPassword(userID, password)
				log.Printf("Authenticator(): id = %d, res = %#v", id, res)
				if res && id > 0 {
					mod, err := model.GetUserById(fmt.Sprintf("%d", id))
					log.Printf("Authenticator(): mod = %#v, err = %#v", mod, err)
					if err != nil {
						return nil, err
					}
					return &mod, nil
				}
				return &model.UserModel{}, jwt.ErrFailedAuthentication
			},
			Authorizator: func(data interface{}, c *gin.Context) bool {
				// TODO: FIXME: XXX
				return true
			},
			PayloadFunc: func(data interface{}) jwt.MapClaims {
				if v, ok := data.(*model.UserModel); ok {
					return jwt.MapClaims{
						identityKey: v.Id,
					}
				}
				return jwt.MapClaims{}
			},
			Unauthorized: func(c *gin.Context, code int, message string) {
				c.JSON(code, gin.H{
					"code":    code,
					"message": message,
				})
			},
			TokenLookup:   "header:Authorization,query:token,cookie:jwt",
			TokenHeadName: "Bearer",
			TimeFunc:      time.Now,
			RefreshResponse: func(c *gin.Context, code int, token string, t time.Time) {
				cookie, err := c.Cookie("jwt")
				if err != nil {
					log.Println(err)
				}

				c.JSON(http.StatusOK, gin.H{
					"code":    http.StatusOK,
					"token":   token,
					"expire":  t.Format(time.RFC3339),
					"message": "refresh successfully",
					"cookie":  cookie,
				})
			},
		})
		if err != nil {
			panic(err)
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
