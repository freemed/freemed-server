package model

import (
	"database/sql"
	"time"

)

const (
	TABLE_INSURANCEMODIFIER = "insmod"
)

type InsuranceModifierModel struct {
	ID        int64          `db:"id" json:"id"`
	CreatedAt time.Time      `db:"created_at" json:"created_at"`
	UpdatedAt time.Time      `db:"updated_at" json:"updated_at"`
	DeletedAt sql.NullTime   `db:"deleted_at" json:"deleted_at"`
	Modifier    string `db:"insmod" json:"modifier"`
	Description string `db:"insmoddesc" json:"description"`
}

func (InsuranceModifierModel) TableName() string {
	return TABLE_INSURANCEMODIFIER
}

func init() {
	DbSupportPicklists = append(DbSupportPicklists, DbSupportPicklist{ModuleName: "insurancemodifier", Query: "SELECT CONCAT(insmod, ' - ', insmoddesc) AS v, id AS k FROM " + TABLE_INSURANCEMODIFIER + " WHERE CONCAT(insmod, ' - ', insmoddesc) LIKE CONCAT('%', :query, '%') ORDER BY insmod,insmoddesc"})
}
