package model

const (
	TABLE_ICDCODE = "icd9"
)

type IcdModel struct {
	Abbreviation       string     `db:"abbrev" json:"abbrev"`
	Language           string     `db:"language" json:"language"`
	Icd9Code           string     `db:"icd9code" json:"icd_9_code"`
	Icd10Code          NullString `db:"icd10code" json:"icd_10_code"`
	Icd9Description    string     `db:"icd9descrip" json:"icd_9_description"`
	Icd10Description   NullString `db:"icd10descrip" json:"icd_10_description"`
	IcdMetaDescription NullString `db:"icdmetadesc" json:"meta_description"`
	Icdng              NullTime   `db:"icdng" json:"ng"`
	IcdDrg             string     `db:"icddrg" json:"drg"`
	IcdNum             int64      `db:"icdnum" json:"number"`
	IcdAmount          float32    `db:"icdamt" json:"amount"`
	IcdColl            float64    `db:"icdcoll" json:"coll"`
	Id                 int64      `db:"id" json:"id"`
}

func init() {
	DbTables = append(DbTables, DbTable{TableName: TABLE_ICDCODE, Obj: IcdModel{}})
	DbSupportPicklists = append(DbSupportPicklists, DbSupportPicklist{ModuleName: "icd", Query: "SELECT CONCAT(icd10code, ' ', icd10descrip) AS v, abbrev AS k FROM " + TABLE_ICDCODE + " WHERE CONCAT(icd10code, ' ', icd10descrip) LIKE CONCAT('%', :query, '%') ORDER BY icd10code, icd10descrip"})
}
