package common

import (
	"crypto/md5"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"time"

	jwt "github.com/appleboy/gin-jwt/v2"
	"github.com/gin-gonic/gin"
)

var (
	IsRunning = true
)

// Md5hash produces an MD5 sum for a string
func Md5hash(orig string) string {
	return fmt.Sprintf("%x", md5.Sum([]byte(orig)))
}

// SleepFor waits for sec seconds
func SleepFor(sec int64) {
	for i := 0; i < int(sec); i++ {
		if !IsRunning {
			return
		}
		time.Sleep(time.Second)
	}
}

// JsonEncode creates a json-encoded version of an object
func JsonEncode(o interface{}) []byte {
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
