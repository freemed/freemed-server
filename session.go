package main

import (
	"fmt"
	"log"
	"os"
	"time"
)

type SessionModel struct {
	Id          int64  `db:"id"`
	SessionId   string `db:"session_id"`
	UserId      int64  `db:"user_id"`
	Expires     int64  `db:"expiry_time"`
	SessionData []byte `db:"session_data"`
}

func init() {
	dbTables = append(dbTables, DbTable{TableName: "session", Obj: SessionModel{}})
}

// GetById will populate a user object from a database model with
// a matching id.
func (u *SessionModel) GetBySessionId(id interface{}) error {
	err := dbmap.SelectOne(u, "SELECT * FROM user WHERE session_id = ?", id)
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

func expireSession(sid string) error {
	_, err := dbmap.Exec("DELETE FROM session WHERE session_id=? AND expiry_time < NOW()", sid)
	return err
}

func getSessionById(sid string) (SessionModel, error) {
	s, err := dbmap.Get(SessionModel{}, sid)
	if err != nil {
		return s.(SessionModel), err
	}
	return s.(SessionModel), nil
}

func sessionExpiryThread() {
	log.Print("sessionExpiryThread spinning up")
	for {
		_, err := dbmap.Exec("DELETE FROM session WHERE expiry_time < NOW()")
		if err != nil {
			log.Print(err.Error())
		}
		sleepFor(30)
	}
}
