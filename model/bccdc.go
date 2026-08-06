package model

import (
	"database/sql"
	"time"

)

const (
	TABLE_BCCDC = "bccdc"
)

type BccdcModel struct {
	ID        int64          `db:"id" json:"id"`
	CreatedAt time.Time      `db:"created_at" json:"created_at"`
	UpdatedAt time.Time      `db:"updated_at" json:"updated_at"`
	DeletedAt sql.NullTime   `db:"deleted_at" json:"deleted_at"`
	Code        string `db:"agent_code" json:"code"`
	Description string `db:"description" json:"description"`
}

func (BccdcModel) TableName() string {
	return TABLE_BCCDC
}

func init() {
	DbSupportPicklists = append(DbSupportPicklists, DbSupportPicklist{ModuleName: "bccdc", Query: "SELECT CONCAT(code, ' - ', description) AS v, id AS k FROM " + TABLE_BCCDC + " WHERE description LIKE CONCAT('%', :query, '%') ORDER BY code, description"})
}
