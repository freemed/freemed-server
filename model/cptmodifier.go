package model

import (
	"database/sql"
	"time"

)

const (
	TABLE_CPTMODIFIER = "cptmod"
)

type CptModifierModel struct {
	ID        int64          `db:"id" json:"id"`
	CreatedAt time.Time      `db:"created_at" json:"created_at"`
	UpdatedAt time.Time      `db:"updated_at" json:"updated_at"`
	DeletedAt sql.NullTime   `db:"deleted_at" json:"deleted_at"`
	Modifier    string `db:"cptmod" json:"modifier"`
	Description string `db:"cptmoddescrip" json:"description"`
}

func (CptModifierModel) TableName() string {
	return TABLE_CPTMODIFIER
}

func init() {
	DbSupportPicklists = append(DbSupportPicklists, DbSupportPicklist{ModuleName: "cptmodifier", Query: "SELECT CONCAT(cptmod, ' ', cptmoddescrip') AS v, id AS k FROM " + TABLE_CPTMODIFIER + " WHERE CONCAT(cptmod, ' ', cptmoddescrip) LIKE CONCAT('%', :query, '%') ORDER BY cptmod, cptmoddescrip"})
}
