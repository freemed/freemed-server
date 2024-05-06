package model

import (
	"time"

	"gorm.io/gorm"
)

const (
	TABLE_PATIENT         = "patient"
	TABLE_PATIENT_ADDRESS = "patient_address"
)

type PatientModel struct {
	gorm.Model
	DateAdd      time.Time `db:"ptdtadd" json:"date_added"`
	DateModified NullTime  `db:"ptdtmod" json:"date_modified"`
	// ptdoc
	// ptrefdoc
	// ptpcp
	// ptphy1
	// ptphy2
	// ptphy3
	// ptphy4
	Salutaion  string     `db:"ptsalut" json:"salutation"`
	LastName   string     `db:"ptlname" json:"last_name"`
	MaidenName NullString `db:"ptmaidenname" json:"maiden_name"`
	FirstName  string     `db:"ptfname" json:"first_name"`
	MiddleName NullString `db:"ptmname" json:"middle_name"`
	NameSuffix string     `db:"ptsuffix" json:"name_suffix"`
	// ptaddr1
	// ptaddr2
	// ptcity
	// ptstate
	// ptzip
	// ptcountry
	// pthphone
	// ptwphone
	// ptmphone
	// ptfax
	// ptemail
	// ptsex
	Gender string `db:"ptsex" json:"gender"`
	// ptdob
	// ptssn
	// ptdmv
	// ptstatus
	PatientId    string    `db:"ptid" json:"patient_id"`
	Diagnosis1   NullInt64 `db:"ptdiag1" json:"diagnosis_1"`
	Diagnosis2   NullInt64 `db:"ptdiag2" json:"diagnosis_2"`
	Diagnosis3   NullInt64 `db:"ptdiag3" json:"diagnosis_3"`
	Diagnosis4   NullInt64 `db:"ptdiag4" json:"diagnosis_4"`
	DiagnosisSet string    `db:"ptdiagset" json:"diagnosis_set"`
	// ptmarital
	// ptempl
	// ptnextofkin
	// ptpharmacy
	// ptrace
	// ptreligion
	Archive         int64      `db:"ptarchive" json:"archive"`
	Iso             string     `db:"iso" json:"iso"`
	BloodType       string     `db:"ptblood" json:"blood_type"`
	Dead            int64      `db:"ptdead" json:"dead"`
	DeathDate       NullTime   `db:"ptdeaddt" json:"death_date"`
	Budget          float64    `db:"ptbudg" json:"ptbudg"`
	BillingType     string     `db:"ptbilltype" json:"billing_type"`
	PrimaryFacility int64      `db:"ptprimaryfacility" json:"primary_facility"`
	PrimaryLanguage string     `db:"ptprimarylanguage" json:"primary_language"`
	Patient         int64      `db:"patient" json:"patient"`
	Module          string     `db:"module" json:"module"`
	RecordID        int64      `db:"oid" json:"oid"`
	Stamp           time.Time  `db:"stamp" json:"stamp"`
	Summary         string     `db:"summary" json:"summary"`
	Locked          bool       `db:"locked" json:"locked"`
	Annotation      NullString `db:"annotation" json:"annotation"`
	User            int64      `db:"user" json:"user_id"`
	Provider        int64      `db:"provider" json:"provider_id"`
	Status          string     `db:"status" json:"status"`
	Id              int64      `db:"id" json:"id"`
}

func init() {
	DbTables = append(DbTables, DbTable{TableName: TABLE_PATIENT, Obj: PatientModel{}, Key: "Id"})
}
