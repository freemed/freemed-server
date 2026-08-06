package model

import (
	"database/sql"
	"time"

)

const (
	TABLE_DRUGQUANTITYQUALIFIER = "drugquantityqual"
)

type DrugQuantityQualifierModel struct {
	ID        int64          `db:"id" json:"id"`
	CreatedAt time.Time      `db:"created_at" json:"created_at"`
	UpdatedAt time.Time      `db:"updated_at" json:"updated_at"`
	DeletedAt sql.NullTime   `db:"deleted_at" json:"deleted_at"`
	Code        string `db:"code" json:"code"`
	Description string `db:"description" json:"description"`
}

func (DrugQuantityQualifierModel) TableName() string {
	return TABLE_DRUGQUANTITYQUALIFIER
}

func init() {
	DbSupportPicklists = append(DbSupportPicklists, DbSupportPicklist{ModuleName: "drugquantityqualifier", Query: "SELECT CONCAT(code, ' - ', description) AS v, id AS k FROM " + TABLE_DRUGQUANTITYQUALIFIER + " WHERE CONCAT(code, ' - ', description) LIKE CONCAT('%', :query, '%') ORDER BY code, description"})
}
