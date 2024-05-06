package model

import (
	"time"

	"gorm.io/gorm"
)

const (
	TABLE_CLEARINGHOUSE = "clearinghouse"
)

type ClearinghouseModel struct {
	gorm.Model
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
	Id            int64     `db:"id" json:"id"`
}

func init() {
	DbTables = append(DbTables, DbTable{TableName: TABLE_CLEARINGHOUSE, Obj: ClearinghouseModel{}, Key: "Id"})
	DbSupportPicklists = append(DbSupportPicklists, DbSupportPicklist{ModuleName: "clearinghouse", Query: "SELECT CONCAT(chname, ' (', chcity, ', ', chstate, ')') AS v, id AS k FROM " + TABLE_CLEARINGHOUSE + " WHERE chname LIKE CONCAT('%', :query, '%') ORDER BY chname, chstate, chzip"})
}
