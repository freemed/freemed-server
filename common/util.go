package common

import (
	"crypto/md5"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"log"
	"time"
)

var (
	IsRunning = true
)

func Md5hash(orig string) string {
	return fmt.Sprintf("%x", md5.Sum([]byte(orig)))
}

func SleepFor(sec int64) {
	for i := 0; i < int(sec); i++ {
		if !IsRunning {
			return
		}
		time.Sleep(time.Second)
	}
}

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
	jwtPayload, ok := c.Get("JWT_PAYLOAD")
	if !ok {
		return SessionModel{}, errors.New("JWT_PAYLOAD not found")
	}
	switch v := jwtPayload.(type) {
	default:
		log.Printf("GetSession(): Unexpected type %T", v)
		return SessionModel{}, errors.New(fmt.Sprintf("JWT_PAYLOAD has type %T", v))
	case map[string]interface{}:
		payloadMap := jwtPayload.(map[string]interface{})
		log.Printf("session map : %v", payloadMap["session"])
		return SessionFromMap(payloadMap["session"].(map[string]interface{})), nil
	}
}
