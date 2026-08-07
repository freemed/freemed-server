package common

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	jwt "github.com/appleboy/gin-jwt/v2"
	"github.com/gin-gonic/gin"
)

var (
	IsRunning = true
)

// SleepFor waits for sec seconds
func SleepFor(sec int64) {
	for i := 0; i < int(sec); i++ {
		if !IsRunning {
			return
		}
		time.Sleep(time.Second)
	}
}

// JSONEncode creates a json-encoded version of an object
func JSONEncode(o interface{}) []byte {
	b, err := json.Marshal(o)
	if err != nil {
		log.Print(err.Error())
		return []byte("false")
	}
	return b
}

// GetSession returns the SessionModel associated with the current session from JWT_PAYLOAD
func GetSession(c *gin.Context) (SessionModel, error) {
	claims := jwt.ExtractClaims(c)
	if len(claims) < 1 {
		return SessionModel{}, errors.New("JWT_PAYLOAD not found")
	}
	userid, ok := claims["id"]
	if !ok {
		return SessionModel{}, errors.New("claim not found")
	}
	sm := SessionModel{}
	sm.UserId = int64(userid.(float64))
	sm.SessionId = jwt.GetToken(c)
	return sm, nil
}

// GetClaim returns a string claim from the JWT, or empty string if not found.
func GetClaim(c *gin.Context, key string) string {
	claims := jwt.ExtractClaims(c)
	if v, ok := claims[key]; ok {
		return fmt.Sprintf("%v", v)
	}
	return ""
}

// RequireRole returns a Gin handler that checks the JWT user_type claim
// against a list of allowed roles. If the claim doesn't match, it returns 403.
func RequireRole(allowed ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userType := GetClaim(c, "user_type")
		for _, role := range allowed {
			if userType == role {
				c.Next()
				return
			}
		}
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
			"code":    http.StatusForbidden,
			"message": "Insufficient permissions",
		})
	}
}

// ParseInt forces an integer to be parsed and returns 0 if unparseable
func ParseInt(s string) int64 {
	i, _ := strconv.ParseInt(s, 10, 64)
	return i
}

// ParseDate parses a string into a date
func ParseDate(s string) (t time.Time, e error) {
	formats := []string{
		"2006-01-02",
		"01/02/2006",
		// TODO: FIXME: IMPLEMENT: More commmon formats
	}
	if s == "" {
		return time.Now(), fmt.Errorf("Unable to parse null date")
	}
	for _, f := range formats {
		t, e = time.Parse(f, s)
		if e == nil {
			return
		}
	}
	return time.Now(), fmt.Errorf("Unable to parse date")
}
