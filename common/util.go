package common

import (
	"crypto/md5"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strconv"
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
