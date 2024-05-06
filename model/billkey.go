package model

import (
	"time"

	"gorm.io/gorm"
)

const (
	TABLE_BILLKEY = "billkey"
)

type BillkeyModel struct {
	gorm.Model
	Date       time.Time `db:"billkeydate" json:"date"`
	Data       []byte    `db:"billkey" json:"key"`
	Procedures string    `db:"bkprocs" json:"procedures"`
	Id         int64     `db:"id" json:"id"`
}

func init() {
	DbTables = append(DbTables, DbTable{TableName: TABLE_BILLKEY, Obj: BillkeyModel{}, Key: "Id"})
}

// GetBillkeyPayload retrieves a payload from a specified billkey
func GetBillkeyPayload(billkey int64) (string, error) {
	var bk BillkeyModel
	tx := Db.First(&bk, billkey)
	if tx.Error != nil {
		return "", tx.Error
	}
	return string(bk.Data), nil
}
