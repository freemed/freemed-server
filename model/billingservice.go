package model

import (
	"github.com/freemed/freemed-server/db"
	"time"
)

const (
	TABLE_BILLINGSERVICE = "bservice"
)

type BillingServiceModel struct {
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

func init() {
	db.DbTables = append(db.DbTables, db.DbTable{TableName: TABLE_BILLINGSERVICE, Obj: BillingServiceModel{}, Key: "Id"})
}
