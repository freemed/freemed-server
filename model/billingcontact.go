package model

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
	DbTables = append(DbTables, DbTable{TableName: TABLE_BILLINGCONTACT, Obj: BillingContactModel{}, Key: "Id"})
	DbSupportPicklists = append(DbSupportPicklists, DbSupportPicklist{
		ModuleName: "billingcontact",
		Query: "SELECT CONCAT(bclname, ',  ', bcfname, ' ', bcmname) AS v" +
			", id AS k" +
			" FROM " + TABLE_BILLINGCONTACT +
			" WHERE CONCAT(bclname, ', ', bcfname, ' ', bcmname) LIKE CONCAT('%%', :query, '%%')" +
			" ORDER BY bclname, bcfname, bcmname"})
}
