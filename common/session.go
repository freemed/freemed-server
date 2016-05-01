package common

import (
	"encoding/json"
	"fmt"
	"gopkg.in/redis.v3"
	"log"
	"os"
	"time"
)

var (
	ActiveSession *SessionConnector
)

type SessionConnector struct {
	Address       string
	Password      string
	DatabaseId    int64
	SessionLength int64

	client *redis.Client
}

type SessionModel struct {
	SessionId   string `json:"session_id"`
	UserId      int64  `json:"user_id"`
	Expires     int64  `json:"expiry_time"`
	SessionData []byte `json:"session_data"`
}

// SessionFromMap is a horror-show hack, designed to get around an annoying
// limitation in Gin which encodes the SessionModel object as a
// map[string]interface{} within another one.
func SessionFromMap(m map[string]interface{}) SessionModel {
	return SessionModel{
		SessionId: m["session_id"].(string),
		UserId:    int64(m["user_id"].(float64)),
		Expires:   int64(m["expiry_time"].(float64)),
		//SessionData: m["session_data"].([]byte),
	}
}

func (s *SessionConnector) Connect() error {
	s.client = redis.NewClient(&redis.Options{
		Addr:     s.Address,
		Password: s.Password,
		DB:       s.DatabaseId,
	})
	_, err := s.client.Ping().Result()
	return err
}

func (s *SessionConnector) CreateSession(uid int64) (SessionModel, error) {
	hn, _ := os.Hostname()
	sid := fmt.Sprintf("%d-%s", time.Now().Unix(), Md5hash(fmt.Sprintf("%d.%s.%d", time.Now().Unix(), hn, time.Now().UnixNano())))
	sm := SessionModel{
		SessionId: sid,
		UserId:    uid,
	}
	log.Printf("CreateSession(%d): key %s = %v", uid, sid, sm)
	err := s.StoreFreshenSession(sm)
	if err != nil {
		log.Print(err.Error())
		return SessionModel{}, err
	}
	return sm, nil
}

func (s *SessionConnector) GetSession(sid string) (SessionModel, error) {
	log.Printf("GetSession(%s)", sid)
	val, err := s.client.Get(sid).Result()
	if err == redis.Nil {
		log.Printf("GetSession(): %s does not exist", sid)
		return SessionModel{}, err
	}
	if err != nil {
		return SessionModel{}, err
	}
	var m SessionModel
	err = json.Unmarshal([]byte(val), &m)
	if err != nil {
		return SessionModel{}, err
	}
	return m, nil
}

func (s *SessionConnector) ExpireSession(sid string) error {
	log.Printf("ExpireSession(%s)", sid)
	return s.client.Del(sid).Err()
}

func (s *SessionConnector) StoreFreshenSession(sm SessionModel) error {
	log.Printf("StoreFreshenSession(): %v", sm)
	sm.Expires = s.SessionLength
	b, err := json.Marshal(sm)
	if err != nil {
		return err
	}
	return s.client.Set(sm.SessionId, string(b), time.Duration(sm.Expires)*time.Second).Err()
}

func TokenAuthFunc(sid string) (bool, SessionModel) {
	if sid == "" {
		return false, SessionModel{}
	}
	s, err := ActiveSession.GetSession(sid)
	log.Printf("TokenAuthFunc(): %s returned %v", sid, s)
	if err != nil {
		return false, SessionModel{}
	}
	if s.SessionId != "" {
		return true, s
	}
	return false, SessionModel{}
}
