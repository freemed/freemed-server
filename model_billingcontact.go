package main

import (
	"time"
)

const (
	TABLE_BILLINGCONTACT = "bcontact"
)

type BillingContactModel struct {
	FirstName  string    `db:"bcfname" json:"first_name"`
	MiddleName string    `db:"bcmname" json:"middle_name"`
	LastName   string    `db:"bclname" json:"last_name"`
	Address    string    `db:"bcaddr" json:"address"`
	City       string    `db:"bccity" json:"city"`
	State      string    `db:"bcstate" json:"state"`
	Zip        string    `db:"bczip" json:"zip"`
	Phone      string    `db:"bcphone" json:"phone"`
	Stamp      time.Time `db:"stamp" json:"stamp"`
	User       int64     `db:"user" json:"user"`
	Id         int64     `db:"id" json:"id"`
}

func init() {
	dbTables = append(dbTables, DbTable{TableName: TABLE_BILLINGCONTACT, Obj: BillingContactModel{}, Key: "Id"})
}
