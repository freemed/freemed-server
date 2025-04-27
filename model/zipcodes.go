package model

import "gorm.io/gorm"

const (
	TABLE_ZIPCODES = "zipcodes"
)

type ZipcodesModel struct {
	gorm.Model
	Id        int64   `db:"id" json:"id"`
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
	DbTables = append(DbTables, DbTable{TableName: TABLE_ZIPCODES, Obj: ZipcodesModel{}, Key: "Id"})
}
