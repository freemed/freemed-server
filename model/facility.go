package model

import "gorm.io/gorm"

const (
	TABLE_FACILITY = "facility"
)

type FacilityModel struct {
	gorm.Model
	Name           string     `db:"psrname" json:"name"`
	Addr1          NullString `db:"psraddr1" json:"addr1"`
	Addr2          NullString `db:"psraddr2" json:"addr2"`
	City           NullString `db:"psrcity" json:"city"`
	State          NullString `db:"psrstate" json:"state"`
	Zip            NullString `db:"psrzip" json:"zip"`
	Country        NullString `db:"psrcountry" json:"country"`
	Note           NullString `db:"psrnote" json:"note"`
	Phone          NullString `db:"psrphone" json:"phone"`
	Fax            NullString `db:"psrfax" json:"fax"`
	Email          NullString `db:"psremail" json:"email"`
	Ein            NullString `db:"psrein" json:"ein"`
	NpiID          NullString `db:"psrnpi" json:"npi_identifier"`
	Taxonomy       NullString `db:"psrtaxonomy" json:"taxonomy"`
	Internal       NullString `db:"psrintext" json:"internal"`
	PlaceOfService int64      `db:"psrpos" json:"pos_id"`
	X12Id          NullString `db:"psrx12id" json:"x12_identifier"`
	X12IdType      NullString `db:"psrx12idtype" json:"x12_identifier_tupe"`
	Id             int64      `db:"id" json:"id"`
}

func (FacilityModel) TableName() string {
	return TABLE_FACILITY
}

func init() {
	DbTables = append(DbTables, DbTable{TableName: TABLE_FACILITY, Obj: FacilityModel{}, Key: "Id"})
	DbSupportPicklists = append(DbSupportPicklists, DbSupportPicklist{ModuleName: "facility", Query: "SELECT CONCAT(psrname, ', ', psrcity, ', ', psrstate) AS v, id AS k FROM " + TABLE_FACILITY + " WHERE CONCAT(psrname, ', ', psrcity, ', ', psrstate) LIKE CONCAT('%', :query, '%') ORDER BY psrname, psrcity, psrstate"})
}
