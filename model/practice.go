package model

import (
	"database/sql"
	"github.com/freemed/freemed-server/db"
)

const (
	TABLE_PRACTICE = "practice"
)

type PracticeModel struct {
	Name        string         `db:"pracname" json:"name"`
	PracticeEin sql.NullString `db:"pracein" json:"ein"`
	Addr1A      sql.NullString `db:"addr1a" json:"addr1_1"`
	Addr2A      sql.NullString `db:"addr2a" json:"addr2_1"`
	CityA       sql.NullString `db:"citya" json:"city_1"`
	StateA      sql.NullString `db:"statea" json:"state_1"`
	ZipA        sql.NullString `db:"zipa" json:"zip_1"`
	PhoneA      sql.NullString `db:"phonea" json:"phone_1"`
	FaxA        sql.NullString `db:"faxa" json:"fax_1"`
	Addr1B      sql.NullString `db:"addr1b" json:"addr1_2"`
	Addr2B      sql.NullString `db:"addr2b" json:"addr2_2"`
	CityB       sql.NullString `db:"cityb" json:"city_2"`
	StateB      sql.NullString `db:"stateb" json:"state_2"`
	ZipB        sql.NullString `db:"zipb" json:"zip_2"`
	PhoneB      sql.NullString `db:"phoneb" json:"phone_2"`
	FaxB        sql.NullString `db:"faxb" json:"fac_2"`
	Email       sql.NullString `db:"email" json:"email"`
	Cellular    sql.NullString `db:"cellular" json:"cellular"`
	NpiId       string         `db:"pracnpi" json:"npi_identifier"`
	Id          int64          `db:"id" json:"id"`
}

func init() {
	db.DbTables = append(db.DbTables, db.DbTable{TableName: TABLE_PRACTICE, Obj: PracticeModel{}, Key: "Id"})
}
