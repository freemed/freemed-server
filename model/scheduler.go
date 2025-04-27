package model

import (
	"time"

	"gorm.io/gorm"
)

const (
	TABLE_SCHEDULER = "scheduler"
)

type SchedulerModel struct {
	gorm.Model
	Date          time.Time  `db:"caldateof" json:"date"`
	Created       time.Time  `db:"calcreated" json:"created"`
	Modified      NullTime   `db:"calmodified" json:"modified"`
	Type          string     `db:"caltype" json:"type"`
	Hour          int        `db:"calhour" json:"hour"`
	Minute        int        `db:"calminute" json:"minute"`
	Duration      int        `db:"calduration" json:"duration"`
	Facility      NullInt64  `db:"calfacility" json:"facility_id"`
	Room          NullInt64  `db:"calroom" json:"room"`
	Provider      int64      `db:"calphysician" json:"provider_id"`
	Patient       int64      `db:"calpatient" json:"patient_id"`
	CptCode       NullInt64  `db:"calcptcode" json:"cptcode_id"`
	Status        string     `db:"calstatus" json:"status"`
	PreNote       string     `db:"calprenote" json:"prenote"`
	PostNote      NullString `db:"calpostnote" json:"postnote"`
	Mark          int64      `db:"calmark" json:"mark"`
	GroupID       int64      `db:"calgroupid" json:"group_id"`
	GroupMembers  NullString `db:"calgroupmembers" json:"groupmemnbers"`
	RecurringNote NullString `db:"calrecurnote" json:"recurring_note"`
	RecurringID   int64      `db:"calrecurid" json:"recurring_id"`
	Template      int64      `db:"calappttemplate" json:"template"`
	Attendees     NullString `db:"calattendees" json:"attendees"`
	User          int64      `db:"user" json:"user"`
	Id            int64      `db:"id" json:"id"`
}

func (SchedulerModel) TableName() string {
	return TABLE_SCHEDULER
}

func init() {
	DbTables = append(DbTables, DbTable{TableName: TABLE_SCHEDULER, Obj: SchedulerModel{}, Key: "Id"})
}
