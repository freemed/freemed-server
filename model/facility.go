package model

import (
	"database/sql"
)

const (
	TABLE_FACILITY = "facility"
)

type FacilityModel struct {
	Name           string         `db:"psrname" json:"name"`
	Addr1          sql.NullString `db:"psraddr1" json:"addr1"`
	Addr2          sql.NullString `db:"psraddr2" json:"addr2"`
	City           sql.NullString `db:"psrcity" json:"city"`
	State          sql.NullString `db:"psrstate" json:"state"`
	Zip            sql.NullString `db:"psrzip" json:"zip"`
	Country        sql.NullString `db:"psrcountry" json:"country"`
	Note           sql.NullString `db:"psrnote" json:"note"`
	Phone          sql.NullString `db:"psrphone" json:"phone"`
	Fax            sql.NullString `db:"psrfax" json:"fax"`
	Email          sql.NullString `db:"psremail" json:"email"`
	Ein            sql.NullString `db:"psrein" json:"ein"`
	NpiId          sql.NullString `db:"psrnpi" json:"npi_identifier"`
	Taxonomy       sql.NullString `db:"psrtaxonomy" json:"taxonomy"`
	Internal       sql.NullString `db:"psrintext" json:"internal"`
	PlaceOfService int64          `db:"psrpos" json:"pos_id"`
	X12Id          sql.NullString `db:"psrx12id" json:"x12_identifier"`
	X12IdType      sql.NullString `db:"psrx12idtype" json:"x12_identifier_tupe"`
	Id             int64          `db:"id" json:"id"`
}

func init() {
	DbTables = append(DbTables, DbTable{TableName: TABLE_FACILITY, Obj: FacilityModel{}, Key: "Id"})
}
