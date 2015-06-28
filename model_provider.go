package main

import (
	"database/sql"
)

const (
	TABLE_PROVIDER = "physician"
)

type ProviderModel struct {
	LastName             string        `db:"phylname" json:"last_name"`
	FirstName            string        `db:"phyfname" json:"first_name"`
	MiddleName           string        `db:"phymname" json:"middle_name"`
	Title                string        `db:"phytitle" json:"title"`
	Practice             int64         `db:"phypractice" json:"practice_id"`
	PracticeEin          string        `db:"phypracein" json:"-"`
	Addr1A               string        `db:"phyaddr1a" json:"-"`
	Addr2A               string        `db:"phyaddr2a" json:"-"`
	CityA                string        `db:"phycitya" json:"-"`
	StateA               string        `db:"phystatea" json:"-"`
	ZipA                 string        `db:"phyzipa" json:"-"`
	PhoneA               string        `db:"phyphonea" json:"-"`
	FaxA                 string        `db:"phyfaxa" json:"-"`
	Addr1B               string        `db:"phyaddr1b" json:"-"`
	Addr2B               string        `db:"phyaddr2b" json:"-"`
	CityB                string        `db:"phycityb" json:"-"`
	StateB               string        `db:"phystateb" json:"-"`
	ZipB                 string        `db:"phyzipb" json:"-"`
	PhoneB               string        `db:"phyphoneb" json:"-"`
	FaxB                 string        `db:"phyfaxb" json:"-"`
	Email                string        `db:"phyemail" json:"email"`
	Pager                string        `db:"phypager" json:"pager"`
	Upin                 string        `db:"phyupin" json:"upin"`
	SocialSecurityNumber string        `db:"physsn" json:"ssn"`
	Degrees              string        `db:"phydegrees" json:"degrees"`
	Specialties          string        `db:"physpecialties" json:"specialties"`
	Id1                  string        `db:"phyid1" json:"-"`
	Status               int64         `db:"phystatus" json:"status"`
	Referring            string        `db:"phyref" json:"referring"`
	ReferCount           int           `db:"phyrefcount" json:"-"`
	ReferAmount          float64       `db:"phyrefamt" json:"-"`
	ReferCollected       float64       `db:"phyrefcoll" json:"-"`
	ChargeMap            string        `db:"phychargemap" json:"charge_map"`
	IdMap                string        `db:"phyidmap" json:"id_map"`
	GroupPractice        sql.NullInt64 `db:"phygrpprac" json:"group_practice_id"`
	Anesthesiologist     int64         `db:"phyanesth" json:"anesth_id"`
	HL7Id                string        `db:"phyhl7id" json:"hl7_identifier"`
	DeaId                string        `db:"phydea" json:"dea_identifier"`
	CliaId               string        `db:"phyclia" json:"clia_identifier"`
	NpiId                string        `db:"phynpi" json:"npi_identifier"`
	Id                   int64         `db:"id" json:"id"`
}

func init() {
	dbTables = append(dbTables, DbTable{TableName: TABLE_PROVIDER, Obj: ProviderModel{}, Key: "Id"})
}
