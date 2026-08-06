package model

import (
	"context"
	"database/sql"
	"time"

	"github.com/freemed/freemed-server/dbgen"
)

const (
	TABLE_MESSAGES = "messages"
)

type MessagesModel struct {
	ID        int64          `db:"id" json:"id"`
	CreatedAt time.Time      `db:"created_at" json:"created_at"`
	UpdatedAt time.Time      `db:"updated_at" json:"updated_at"`
	DeletedAt sql.NullTime   `db:"deleted_at" json:"deleted_at"`
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

func (MessagesModel) TableName() string {
	return TABLE_MESSAGES
}

func init() {
}

func MessageById(id int64) (*dbgen.Message, error) {
	msg, err := Queries.GetMessageById(context.Background(), id)
	if err != nil {
		return nil, err
	}
	return &msg, nil
}

func (msg MessagesModel) Send() error {
	_, err := Queries.CreateMessage(context.Background(), dbgen.CreateMessageParams{
		Msgby:      msg.Sender,
		Msgtime:    msg.Sent,
		Msgfor:     msg.For,
		Msgrecip:   msg.Recipients,
		Msgpatient: msg.Patient,
		Msgperson:  msg.Person,
		Msgurgency: int64(msg.Urgency),
		Msgsubject: msg.Subject,
		Msgtext:    msg.Text,
		Msgread:    int64(msg.Read),
		Msgunique:  toNullString(msg.Unique),
		Msgtag:     toNullString(msg.Tag),
		Active:     msg.Active,
	})
	return err
}

func MessageSend(msg MessagesModel) error {
	return msg.Send()
}

func toNullString(ns NullString) sql.NullString {
	return sql.NullString{String: ns.String, Valid: ns.Valid}
}
