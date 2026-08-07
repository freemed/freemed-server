package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	jwt "github.com/appleboy/gin-jwt/v2"
	"github.com/freemed/freemed-server/common"
	"github.com/freemed/freemed-server/config"
	"github.com/freemed/freemed-server/dbgen"
	"github.com/freemed/freemed-server/model"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
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
				return &dbgen.User{}, jwt.ErrFailedAuthentication
			},
			Authorizator: func(data interface{}, c *gin.Context) bool {
				return authorizeRequest(c, data)
			},
			PayloadFunc: func(data interface{}) jwt.MapClaims {
				if v, ok := data.(*dbgen.User); ok {
					userType := ""
					if v.Usertype.Valid {
						userType = v.Usertype.String
					}
					return jwt.MapClaims{
						identityKey:  v.ID,
						"jti":        uuid.New().String(),
						"user_type":  userType,
						"username":   v.Username,
						"provider_id": v.Userrealphy,
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
			// Token only via Authorization header or httpOnly cookie
			TokenLookup:   "header:Authorization,cookie:jwt",
			TokenHeadName: "Bearer",
			TimeFunc:      time.Now,
			// LoginResponse sets the JWT as an httpOnly cookie alongside the JSON response.
			LoginResponse: func(c *gin.Context, code int, token string, expire time.Time) {
				maxAge := int(time.Until(expire).Seconds())
				if maxAge < 0 {
					maxAge = 0
				}
				secure := config.Config.Web.Keys.Cert != "" && config.Config.Web.Keys.Key != ""
				c.SetCookie("jwt", token, maxAge, "/", "", secure, true)
				c.JSON(http.StatusOK, gin.H{
					"code":    http.StatusOK,
					"expire":  expire.Format(time.RFC3339),
					"message": "login successful",
				})
			},
			RefreshResponse: func(c *gin.Context, code int, token string, t time.Time) {
				maxAge := int(time.Until(t).Seconds())
				if maxAge < 0 {
					maxAge = 0
				}
				secure := config.Config.Web.Keys.Cert != "" && config.Config.Web.Keys.Key != ""
				c.SetCookie("jwt", token, maxAge, "/", "", secure, true)
				c.JSON(http.StatusOK, gin.H{
					"code":    http.StatusOK,
					"token":   token,
					"expire":  t.Format(time.RFC3339),
					"message": "refresh successfully",
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

	claims := jwt.ExtractClaims(c)
	if jti, ok := claims["jti"]; ok {
		jtiStr := fmt.Sprintf("%v", jti)
		// Blacklist the JWT token ID so it cannot be reused
		err := common.ActiveSession.BlacklistToken(jtiStr, time.Minute*time.Duration(config.Config.Session.Expiry))
		if err != nil {
			log.Printf("AuthLogout(): failed to blacklist token %s: %v", jtiStr, err)
		}
	}

	session, err := common.GetSession(c)
	if err != nil {
		log.Printf("AuthLogout(): Expire session: %v", err)
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	log.Printf("AuthLogout(): Expire session %s", session.SessionId)
	common.ActiveSession.ExpireSession(session.SessionId)

	// Clear the httpOnly jwt cookie
	c.SetCookie("jwt", "", -1, "/", "", false, true)
	c.JSON(http.StatusOK, true)
}

// authMe returns the current user's session info from JWT claims.
func authMe(c *gin.Context) {
	claims := jwt.ExtractClaims(c)
	c.JSON(http.StatusOK, gin.H{
		"user_id":     claims[identityKey],
		"user_type":   claims["user_type"],
		"username":    claims["username"],
		"provider_id": claims["provider_id"],
	})
}

// authorizeRequest performs basic RBAC based on the user_type claim.
func authorizeRequest(c *gin.Context, data interface{}) bool {
	claims := jwt.ExtractClaims(c)
	userID, ok := claims[identityKey]
	if !ok {
		return false
	}
	_ = userID

	// Check if this token has been blacklisted
	if jti, ok := claims["jti"]; ok {
		jtiStr := fmt.Sprintf("%v", jti)
		blacklisted, err := common.ActiveSession.IsTokenBlacklisted(jtiStr)
		if err == nil && blacklisted {
			return false
		}
	}

	// For now, any authenticated user with a valid (non-blacklisted) token is authorized.
	// TODO: Add role-based checks per route.
	return true
}
