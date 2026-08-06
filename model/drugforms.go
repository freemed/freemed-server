package model

import (
	"database/sql"
	"time"

)

const (
	TABLE_DRUGFORM = "drugforms"
)

type DrugFormModel struct {
	ID        int64          `db:"id" json:"id"`
	CreatedAt time.Time      `db:"created_at" json:"created_at"`
	UpdatedAt time.Time      `db:"updated_at" json:"updated_at"`
	DeletedAt sql.NullTime   `db:"deleted_at" json:"deleted_at"`
	Code        string `db:"code" json:"code"`
	Description string `db:"description" json:"description"`
}

func (DrugFormModel) TableName() string {
	return TABLE_DRUGFORM
}

func init() {
	DbSupportPicklists = append(DbSupportPicklists, DbSupportPicklist{ModuleName: "drugform", Query: "SELECT CONCAT(code, ' - ', description) AS v, id AS k FROM " + TABLE_DRUGFORM + " WHERE CONCAT(code, ' - ', description) LIKE CONCAT('%', :query, '%') ORDER BY code, description"})
}
