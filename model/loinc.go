package model

const (
	TABLE_LOINC = "loinc"
)

type LoincModel struct {
	LoincNum          string `db:"loinc_num" json:"loinc_num"`
	Component         string `db:"component" json:"component"`
	Property          string `db:"property" json:"property"`
	TypeAspect        string `db:"type_aspct" json:"type_aspect"`
	System            string `db:"system" json:"system"`
	ScaleType         string `db:"scale_typ" json:"scale_type"`
	MethodType        string `db:"method_typ" json:"method_type"`
	AnswerList        string `db:"answerlist" json:"answer_list"`
	Status            string `db:"status" json:"status"`
	ShortName         string `db:"shortname" json:"short_name"`
	ExternalCopyright string `db:"external_copyright_notice" json:"external_copyright_notice"`
	Id                int64  `db:"id" json:"id"`
}

func init() {
	DbTables = append(DbTables, DbTable{TableName: TABLE_LOINC, Obj: LoincModel{}})
	DbSupportPicklists = append(DbSupportPicklists, DbSupportPicklist{ModuleName: "loinc", Query: "SELECT CONCAT(loinc_num, ' ', shortname) AS v, abbrev AS k FROM " + TABLE_LOINC + " WHERE CONCAT(loinc_num, ' ', shortname) LIKE CONCAT('%', :query, '%') ORDER BY shortname, loinc_num"})
}
