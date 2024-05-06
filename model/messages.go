package model

import (
	"time"

	"gorm.io/gorm"
)

const (
	TABLE_MESSAGES = "messages"
)

type MessagesModel struct {
	gorm.Model
	Id         int64      `db:"id" json:"id"`
	Sender     int64      `db:"msgby" json:"msgby"`
	SenderName string     `db:"sender" json:"sender"`
	Sent       time.Time  `db:"msgtime" json:"msgtime"`
	For        int64      `db:"msgfor" json:"msgfor"`
	Recipients string     `db:"msgrecip" json:"msgrecip"`
	Patient    int64      `db:"msgpatient" json:"msgpatient"`
	Person     string     `db:"msgperson" json:"msgperson"`
	Urgency    int        `db:"msgurgency" json:"msgurgency"`
	Subject    string     `db:"msgsubject" json:"msgsubject"`
	Text       string     `db:"msgtext" json:"msgtext"`
	Read       int        `db:"msgread" json:"msgread"`
	Unique     NullString `db:"msgunique" json:"msgunique"`
	Tag        NullString `db:"msgtag" json:"msgtag"`
	Active     string     `db:"active" json:"active"`
}

func init() {
	DbTables = append(DbTables, DbTable{TableName: TABLE_MESSAGES, Obj: MessagesModel{}, Key: "Id"})
}

func MessageById(id int64) (*MessagesModel, error) {
	var msg MessagesModel
	tx := Db.First(&msg, id)
	if tx.Error != nil {
		return &msg, tx.Error
	}
	return &msg, nil
}

func (msg MessagesModel) Send() error {
	tx := Db.Create(msg)
	if tx.Error != nil {
		return tx.Error
	}
	return nil
}

func MessageSend(msg MessagesModel) error {
	return msg.Send()
}
