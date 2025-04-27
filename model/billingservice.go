package model

import (
	"time"

	"gorm.io/gorm"
)

const (
	TABLE_BILLINGSERVICE = "bservice"
)

type BillingServiceModel struct {
	gorm.Model
	Name    string    `db:"bsname" json:"name"`
	Address string    `db:"bsaddr" json:"address"`
	City    string    `db:"bscity" json:"city"`
	State   string    `db:"bsstate" json:"state"`
	Zip     string    `db:"bszip" json:"zip"`
	Phone   string    `db:"bsphone" json:"phone"`
	Etin    string    `db:"bsetin" json:"etin"`
	Tin     string    `db:"bstin" json:"tin"`
	Stamp   time.Time `db:"stamp" json:"stamp"`
	User    int64     `db:"user" json:"user"`
	Id      int64     `db:"id" json:"id"`
}

func (BillingServiceModel) TableName() string {
	return TABLE_BILLINGSERVICE
}

func init() {
	DbTables = append(DbTables, DbTable{TableName: TABLE_BILLINGSERVICE, Obj: BillingServiceModel{}, Key: "Id"})
	DbSupportPicklists = append(DbSupportPicklists, DbSupportPicklist{ModuleName: "billingservice", Query: "SELECT CONCAT(bsname, ' ', bscity, ',  ', bsstate, ')') AS v, id AS k FROM " + TABLE_BILLINGSERVICE + " WHERE bsname LIKE CONCAT('%', :query, '%') ORDER BY name, bsstate, bscity"})
}
