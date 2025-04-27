package model

import "gorm.io/gorm"

const (
	TABLE_CPT = "cpt"
)

type CptCodeModel struct {
	gorm.Model
	Code                 string     `db:"abbrev" json:"abbrev"`
	NameInternal         NullString `db:"cptnameint" json:"name_internal"`
	NameExternal         NullString `db:"cptnameext" json:"name_external"`
	Gender               string     `db:"cptgender" json:"gender"`
	Taxed                string     `db:"cpttaxed" json:"taxed"`
	Type                 int64      `db:"cpttype" json:"type"`
	RequiredCptCodes     NullString `db:"cptreqcpt" json:"required_cpt"`
	ExcludedCptCodes     NullString `db:"cptexccpt" json:"excluded_cpt"`
	RequiredIcdCodes     NullString `db:"cptreqicd" json:"required_icd"`
	ExcludedIcdCodes     NullString `db:"cptrexcicd" json:"excluded_icd"`
	RelativeValue        float64    `db:"cptrelval" json:"relative_value"`
	DefaultTypeOfService int64      `db:"cptdeftos" json:"default_type"`
	DefaultStandardFee   float64    `db:"cptdefstdfee" json:"default_standard_fee"`
	StandardFees         string     `db:"cptstdfee" json:"standard_fee"`
	TypesOfService       string     `db:"cpttos" json:"types_of_service"`
	TypesOfServicePrefix string     `db:"cpttosprfx" json:"types_of_service_prefix"`
	Id                   int64      `db:"id" json:"id"`
}

func (CptCodeModel) TableName() string {
	return TABLE_CPT
}

func init() {
	DbTables = append(DbTables, DbTable{TableName: TABLE_CPT, Obj: CptCodeModel{}, Key: "Id"})
	DbSupportPicklists = append(DbSupportPicklists, DbSupportPicklist{ModuleName: "cpt", Query: "SELECT cptnameext AS v, id AS k FROM " + TABLE_CPT + " WHERE cptnameext LIKE CONCAT('%', :query, '%') ORDER BY cptnameext"})
}
