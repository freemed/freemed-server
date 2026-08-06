package model

import (
	"database/sql"
	"time"

)

const (
	TABLE_PLACEOFSERVICE = "pos"
)

type PlaceOfServiceModel struct {
	ID        int64          `db:"id" json:"id"`
	CreatedAt time.Time      `db:"created_at" json:"created_at"`
	UpdatedAt time.Time      `db:"updated_at" json:"updated_at"`
	DeletedAt sql.NullTime   `db:"deleted_at" json:"deleted_at"`
	Name        string   `db:"posname" json:"name"`
	Description string   `db:"posdescrip" json:"description"`
	Added       NullTime `db:"posdtadd" json:"added"`
	Modified    NullTime `db:"posdtmod" json:"modified"`
}

func (PlaceOfServiceModel) TableName() string {
	return TABLE_PLACEOFSERVICE
}

func init() {
	DbSupportPicklists = append(DbSupportPicklists, DbSupportPicklist{ModuleName: "placeofservice", Query: "SELECT CONCAT(posname, ' - ', posdescrip) AS v, id AS k FROM " + TABLE_PLACEOFSERVICE + " WHERE posname LIKE CONCAT('%', :query, '%') OR posdescrip LIKE CONCAT('%', :query, '%') ORDER BY posname, posdescrip"})
}
