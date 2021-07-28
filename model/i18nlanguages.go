package model

const (
	TABLE_I18NLANGUAGES = "i18nlanguages"
)

type I18nLanguageModel struct {
	Abbreviation string `db:"abbrev" json:"abbrev"`
	Language     string `db:"language" json:"language"`
}

func init() {
	DbTables = append(DbTables, DbTable{TableName: TABLE_I18NLANGUAGES, Obj: I18nLanguageModel{}})
	DbSupportPicklists = append(DbSupportPicklists, DbSupportPicklist{ModuleName: "i18nlanguages", Query: "SELECT language AS v, abbrev AS k FROM " + TABLE_I18NLANGUAGES + " WHERE language LIKE CONCAT('%', :query, '%') ORDER BY language"})
}
