package model

import (
	"database/sql"
	"time"

)

const (
	TABLE_CLEARINGHOUSE = "clearinghouse"
)

type ClearinghouseModel struct {
	ID        int64          `db:"id" json:"id"`
	CreatedAt time.Time      `db:"created_at" json:"created_at"`
	UpdatedAt time.Time      `db:"updated_at" json:"updated_at"`
	DeletedAt sql.NullTime   `db:"deleted_at" json:"deleted_at"`
	Name          string    `db:"chname" json:"name"`
	Address       string    `db:"chaddr" json:"address"`
	City          string    `db:"chcity" json:"city"`
	State         string    `db:"chstate" json:"state"`
	Zip           string    `db:"chzip" json:"zip"`
	Phone         string    `db:"chphone" json:"phone"`
	Etin          string    `db:"chetin" json:"etin"`
	X12GsSender   string    `db:"chx12gssender" json:"x12gssender"`
	X12GsReceiver string    `db:"chx12gsreceiver" json:"x12gsreceiver"`
	Stamp         time.Time `db:"stamp" json:"stamp"`
	User          int64     `db:"user" json:"user"`
}

func (ClearinghouseModel) TableName() string {
	return TABLE_CLEARINGHOUSE
}

func init() {
	DbSupportPicklists = append(DbSupportPicklists, DbSupportPicklist{ModuleName: "clearinghouse", Query: "SELECT CONCAT(chname, ' (', chcity, ', ', chstate, ')') AS v, id AS k FROM " + TABLE_CLEARINGHOUSE + " WHERE chname LIKE CONCAT('%', :query, '%') ORDER BY chname, chstate, chzip"})
}
