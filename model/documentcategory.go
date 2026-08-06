package model

import (
	"database/sql"
	"time"

)

const (
	TABLE_DOCUMENTCATEGORY = "documents_tc"
)

type DocumentCategoryModel struct {
	ID        int64          `db:"id" json:"id"`
	CreatedAt time.Time      `db:"created_at" json:"created_at"`
	UpdatedAt time.Time      `db:"updated_at" json:"updated_at"`
	DeletedAt sql.NullTime   `db:"deleted_at" json:"deleted_at"`
	Type        string `db:"type" json:"type"`
	Category    string `db:"category" json:"category"`
	Description string `db:"description" json:"description"`
}

func (DocumentCategoryModel) TableName() string {
	return TABLE_DOCUMENTCATEGORY
}

func init() {
	DbSupportPicklists = append(DbSupportPicklists, DbSupportPicklist{ModuleName: "documentcategory", Query: "SELECT CONCAT(type, '/', category, ' - ', description) AS v, id AS k FROM " + TABLE_DOCUMENTCATEGORY + " WHERE CONCAT(type, '/', category, ' - ', description) LIKE CONCAT('%', :query, '%') ORDER BY type, category, description"})
}
