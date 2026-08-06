package model

import (
	"database/sql"
	"time"

)

const (
	TABLE_PATIENT_EMR = "patient_emr"
)

type PatientEmrModel struct {
	ID        int64          `db:"id" json:"id"`
	CreatedAt time.Time      `db:"created_at" json:"created_at"`
	UpdatedAt time.Time      `db:"updated_at" json:"updated_at"`
	DeletedAt sql.NullTime   `db:"deleted_at" json:"deleted_at"`
	Patient    int64      `db:"patient" json:"patient_id"`
	Module     string     `db:"module" json:"module"`
	RecordId   int64      `db:"oid" json:"oid"`
	Stamp      time.Time  `db:"stamp" json:"stamp"`
	Summary    string     `db:"summary" json:"summary"`
	Locked     bool       `db:"locked" json:"locked"`
	Annotation NullString `db:"annotation" json:"annotation"`
	User       int64      `db:"user" json:"user_id"`
	Provider   int64      `db:"provider" json:"provider_id"`
	Language   string     `db:"language" json:"language"`
	Status     string     `db:"status" json:"status"`
}

func (PatientEmrModel) TableName() string {
	return TABLE_PATIENT_EMR
}

func init() {
}
