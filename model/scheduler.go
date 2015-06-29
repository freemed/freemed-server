package model

import (
	"time"
)

const (
	TABLE_SCHEDULER = "scheduler"
)

type SchedulerModel struct {
	Date          time.Time `db:"caldateof" json:"date"`
	Created       time.Time `db:"calcreated" json:"created"`
	Modified      time.Time `db:"calmodified" json:"modified"`
	Type          string    `db:"caltype" json:"type"`
	Hour          int       `db:"calhour" json:"hour"`
	Minute        int       `db:"calminute" json:"minute"`
	Duration      int       `db:"calduration" json:"duration"`
	Facility      int64     `db:"calfacility" json:"facility_id"`
	Room          int64     `db:"calroom" json:"room"`
	Provider      int64     `db:"calphysician" json:"provider_id"`
	Patient       int64     `db:"calpatient" json:"patient_id"`
	CptCode       int64     `db:"calcptcode" json:"cptcode_id"`
	Status        string    `db:"calstatus" json:"status"`
	PreNote       string    `db:"calprenote" json:"prenote"`
	PostNote      string    `db:"calpostnote" json:"postnote"`
	Mark          int64     `db:"calmark" json:"mark"`
	GroupId       int64     `db:"calgroupid" json:"group_id"`
	GroupMembers  string    `db:"calgroupmembers" json:"groupmemnbers"`
	RecurringNote string    `db:"calrecurnote" json:"recurring_note"`
	RecurringId   int64     `db:"calrecurid" json:"recurring_id"`
	Template      int64     `db:"calappttemplate" json:"template"`
	Attendees     string    `db:"calattendees" json:"attendees"`
	User          int64     `db:"user" json:"user"`
	Id            int64     `db:"id" json:"id"`
}

func init() {
	DbTables = append(DbTables, DbTable{TableName: TABLE_SCHEDULER, Obj: SchedulerModel{}, Key: "Id"})
}
