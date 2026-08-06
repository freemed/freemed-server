package model

import (
	"database/sql"
	"context"
	"time"

)

const (
	TABLE_BILLKEY = "billkey"
)

type BillkeyModel struct {
	ID        int64          `db:"id" json:"id"`
	CreatedAt time.Time      `db:"created_at" json:"created_at"`
	UpdatedAt time.Time      `db:"updated_at" json:"updated_at"`
	DeletedAt sql.NullTime   `db:"deleted_at" json:"deleted_at"`
	Date       time.Time `db:"billkeydate" json:"date"`
	Data       []byte    `db:"billkey" json:"key"`
	Procedures string    `db:"bkprocs" json:"procedures"`
}

func (BillkeyModel) TableName() string {
	return TABLE_BILLKEY
}

func init() {
}

// GetBillkeyPayload retrieves a payload from a specified billkey
func GetBillkeyPayload(billkey int64) (string, error) {
	bk, err := Queries.GetBillkeyById(context.Background(), billkey)
	if err != nil {
		return "", err
	}
	return bk.Billkey.String, nil
}
