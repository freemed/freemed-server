package model

import (
	"database/sql"
	"time"

)

const (
	TABLE_ZIPCODES = "zipcodes"
)

type ZipcodesModel struct {
	ID        int64          `db:"id" json:"id"`
	CreatedAt time.Time      `db:"created_at" json:"created_at"`
	UpdatedAt time.Time      `db:"updated_at" json:"updated_at"`
	DeletedAt sql.NullTime   `db:"deleted_at" json:"deleted_at"`
	Zip       string  `db:"zip" json:"zip"`
	City      string  `db:"city" json:"city"`
	State     string  `db:"state" json:"state"`
	Latitude  float64 `db:"latitude" json:"latitude"`
	Longitude float64 `db:"longitude" json:"longitude"`
	Timezone  int64   `db:"timezone" json:"timezone"`
	DST       int64   `db:"dst" json:"dst"`
	Country   string  `db:"country" json:"country"`
}

func (ZipcodesModel) TableName() string {
	return TABLE_ZIPCODES
}

func init() {
}
