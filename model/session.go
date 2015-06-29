package model

import (
	"fmt"
	"github.com/freemed/freemed-server/util"
	"log"
	"os"
	"time"
)

const (
	TABLE_SESSION = "sessions"
)

type SessionModel struct {
	Id          int64  `db:"id"`
	SessionId   string `db:"session_id"`
	UserId      int64  `db:"user_id"`
	Expires     int64  `db:"expiry_time"`
	SessionData []byte `db:"session_data"`
}

func init() {
	DbTables = append(DbTables, DbTable{TableName: TABLE_SESSION, Obj: SessionModel{}, Key: "Id"})
}

func (u *SessionModel) GetBySessionId(id interface{}) error {
	err := DbMap.SelectOne(u, "SELECT * FROM "+TABLE_SESSION+" WHERE session_id = ?", id)
	if err != nil {
		return err
	}

	return nil
}

func CreateSession(uid int64) (SessionModel, error) {
	hn, _ := os.Hostname()
	sid := fmt.Sprintf("%d-%s", time.Now().Unix(), util.Md5hash(fmt.Sprintf("%d.%s.%d", time.Now().Unix(), hn, time.Now().UnixNano())))
	s := SessionModel{
		SessionId: sid,
		UserId:    uid,
		Expires:   time.Now().Unix() + int64(SessionLength),
	}
	err := DbMap.Insert(&s)
	if err != nil {
		log.Print(err.Error())
		return SessionModel{}, err
	}
	return s, nil
}

func FreshenSession(sid string) error {
	_, err := DbMap.Exec("UPDATE "+TABLE_SESSION+" SET expiry_time = ? WHERE session_id = ?", time.Now().Unix()+int64(SessionLength), sid)
	return err
}

func ExpireSession(sid string) error {
	_, err := DbMap.Exec("DELETE FROM "+TABLE_SESSION+" WHERE session_id = ? AND expiry_time < NOW()", sid)
	return err
}

func getSessionById(sid string) (SessionModel, error) {
	var s SessionModel
	err := DbMap.SelectOne(&s, "SELECT * FROM "+TABLE_SESSION+" WHERE session_id = ?", sid)
	if err != nil {
		log.Print("getSessionById: " + err.Error())
		return s, err
	}
	return s, nil
}

func TokenAuthFunc(sid string) (bool, SessionModel) {
	if sid == "" {
		return false, SessionModel{}
	}
	s, err := getSessionById(sid)
	log.Printf("TokenAuthFunc(): %s returned %v", sid, s)
	if err != nil {
		return false, SessionModel{}
	}
	if s.SessionId != "" {
		return true, s
	}
	return false, SessionModel{}
}

func SessionExpiryThread() {
	log.Print("SessionExpiryThread: spinning up")
	for {
		if !util.IsRunning {
			log.Print("SessionExpiryThread: !IsRunning")
			return
		}
		_, err := DbMap.Exec("DELETE FROM " + TABLE_SESSION + " WHERE expiry_time < NOW()")
		if err != nil {
			log.Print("SessionExpiryThread: " + err.Error())
		}
		util.SleepFor(30)
	}
}
