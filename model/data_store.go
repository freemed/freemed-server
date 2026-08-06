package model

import (
	"database/sql"
	"time"
)

const (
	TABLE_DATA_STORE = "pds"
)

type DataStoreModel struct {
	ID        int64          `db:"id" json:"id"`
	CreatedAt time.Time      `db:"created_at" json:"created_at"`
	UpdatedAt time.Time      `db:"updated_at" json:"updated_at"`
	DeletedAt sql.NullTime   `db:"deleted_at" json:"deleted_at"`
	Patient  int64  `db:"patient" json:"patient_id"`
	Module   string `db:"module" json:"module"`
	Contents []byte `db:"contents" json:"contents"`
}

func (DataStoreModel) TableName() string {
	return TABLE_DATA_STORE
}

