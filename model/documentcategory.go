package model

const (
	TABLE_DOCUMENTCATEGORY = "documents_tc"
)

type DocumentCategoryModel struct {
	Type        string `db:"type" json:"type"`
	Category    string `db:"category" json:"category"`
	Description string `db:"description" json:"description"`
	Id          int64  `db:"id" json:"id"`
}

func init() {
	DbTables = append(DbTables, DbTable{TableName: TABLE_DOCUMENTCATEGORY, Obj: DocumentCategoryModel{}, Key: "Id"})
	DbSupportPicklists = append(DbSupportPicklists, DbSupportPicklist{ModuleName: "documentcategory", Query: "SELECT CONCAT(type, '/', category, ' - ', description) AS v, id AS k FROM " + TABLE_DOCUMENTCATEGORY + " WHERE CONCAT(type, '/', category, ' - ', description) LIKE CONCAT('%', :query, '%') ORDER BY type, category, description"})
}
