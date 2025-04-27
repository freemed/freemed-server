package model

import "gorm.io/gorm"

const (
	TABLE_DOCUMENTCATEGORY = "documents_tc"
)

type DocumentCategoryModel struct {
	gorm.Model
	Type        string `db:"type" json:"type"`
	Category    string `db:"category" json:"category"`
	Description string `db:"description" json:"description"`
	Id          int64  `db:"id" json:"id"`
}

func (DocumentCategoryModel) TableName() string {
	return TABLE_DOCUMENTCATEGORY
}

func init() {
	DbTables = append(DbTables, DbTable{TableName: TABLE_DOCUMENTCATEGORY, Obj: DocumentCategoryModel{}, Key: "Id"})
	DbSupportPicklists = append(DbSupportPicklists, DbSupportPicklist{ModuleName: "documentcategory", Query: "SELECT CONCAT(type, '/', category, ' - ', description) AS v, id AS k FROM " + TABLE_DOCUMENTCATEGORY + " WHERE CONCAT(type, '/', category, ' - ', description) LIKE CONCAT('%', :query, '%') ORDER BY type, category, description"})
}
