package main

import (
	"fmt"
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
	dbTables = append(dbTables, DbTable{TableName: TABLE_SESSION, Obj: SessionModel{}, Key: "Id"})
}

func (u *SessionModel) GetBySessionId(id interface{}) error {
	err := dbmap.SelectOne(u, "SELECT * FROM "+TABLE_SESSION+" WHERE session_id = ?", id)
	if err != nil {
		return err
	}

	return nil
}

func createSession(uid int64) (SessionModel, error) {
	hn, _ := os.Hostname()
	sid := fmt.Sprintf("%d-%s", time.Now().Unix(), md5hash(fmt.Sprintf("%d.%s.%d", time.Now().Unix(), hn, time.Now().UnixNano())))
	s := SessionModel{
		SessionId: sid,
		UserId:    uid,
		Expires:   time.Now().Unix() + int64(*SESSION_LENGTH),
	}
	err := dbmap.Insert(&s)
	if err != nil {
		log.Print(err.Error())
		return SessionModel{}, err
	}
	return s, nil
}

func freshenSession(sid string) error {
	_, err := dbmap.Exec("UPDATE "+TABLE_SESSION+" SET expiry_time = ? WHERE session_id = ?", time.Now().Unix()+int64(*SESSION_LENGTH), sid)
	return err
}

func expireSession(sid string) error {
	_, err := dbmap.Exec("DELETE FROM "+TABLE_SESSION+" WHERE session_id = ? AND expiry_time < NOW()", sid)
	return err
}

func getSessionById(sid string) (SessionModel, error) {
	var s SessionModel
	err := dbmap.SelectOne(&s, "SELECT * FROM "+TABLE_SESSION+" WHERE session_id = ?", sid)
	if err != nil {
		log.Print("getSessionById: " + err.Error())
		return s, err
	}
	return s, nil
}

func tokenAuthFunc(sid string) (bool, SessionModel) {
	if sid == "" {
		return false, SessionModel{}
	}
	s, err := getSessionById(sid)
	log.Printf("tokenAuthFunc(): %s returned %v", sid, s)
	if err != nil {
		return false, SessionModel{}
	}
	if s.SessionId != "" {
		return true, s
	}
	return false, SessionModel{}
}

func sessionExpiryThread() {
	log.Print("sessionExpiryThread: spinning up")
	for {
		_, err := dbmap.Exec("DELETE FROM " + TABLE_SESSION + " WHERE expiry_time < NOW()")
		if err != nil {
			log.Print("sessionExpiryThread: " + err.Error())
		}
		sleepFor(30)
	}
}
