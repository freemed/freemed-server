package model

import (
	"database/sql"
	"github.com/freemed/freemed-server/db"
)

const (
	TABLE_CPT = "cpt"
)

type CptCodeModel struct {
	Code                 string         `db:"abbrev" json:"abbrev"`
	NameInternal         sql.NullString `db:"cptnameint" json:"name_internal"`
	NameExternal         sql.NullString `db:"cptnameext" json:"name_external"`
	Gender               string         `db:"cptgender" json:"gender"`
	Taxed                string         `db:"cpttaxed" json:"taxed"`
	Type                 int64          `db:"cpttype" json:"type"`
	RequiredCptCodes     sql.NullString `db:"cptreqcpt" json:"required_cpt"`
	ExcludedCptCodes     sql.NullString `db:"cptexccpt" json:"excluded_cpt"`
	RequiredIcdCodes     sql.NullString `db:"cptreqicd" json:"required_icd"`
	ExcludedIcdCodes     sql.NullString `db:"cptrexcicd" json:"excluded_icd"`
	RelativeValue        float64        `db:"cptrelval" json:"relative_value"`
	DefaultTypeOfService int64          `db:"cptdeftos" json:"default_type"`
	DefaultStandardFee   float64        `db:"cptdefstdfee" json:"default_standard_fee"`
	StandardFees         string         `db:"cptstdfee" json:"standard_fee"`
	TypesOfService       string         `db:"cpttos" json:"types_of_service"`
	TypesOfServicePrefix string         `db:"cpttosprfx" json:"types_of_service_prefix"`
	Id                   int64          `db:"id" json:"id"`
}

func init() {
	db.DbTables = append(db.DbTables, db.DbTable{TableName: TABLE_CPT, Obj: CptCodeModel{}, Key: "Id"})
}
