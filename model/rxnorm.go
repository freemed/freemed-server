package model

import (
	"database/sql"
	"time"
)

const (
	TABLE_RXNORM = "rxnorm"
)

type RxNormModel struct {
	ID        int64        `db:"id" json:"id"`
	CreatedAt time.Time    `db:"created_at" json:"created_at"`
	UpdatedAt time.Time    `db:"updated_at" json:"updated_at"`
	DeletedAt sql.NullTime `db:"deleted_at" json:"deleted_at"`
	Rxcui     string       `db:"rxcui" json:"rxcui"`
	DrugName  string       `db:"drug_name" json:"drug_name"`
	DrugClass string       `db:"drug_class" json:"drug_class"`
}

func (RxNormModel) TableName() string {
	return TABLE_RXNORM
}

func init() {
	DbSupportPicklists = append(DbSupportPicklists, DbSupportPicklist{
		ModuleName: "rxnorm",
		Query:      "SELECT CONCAT(drug_name, ' (', rxcui, ')') AS v, id AS k FROM " + TABLE_RXNORM + " WHERE drug_name LIKE CONCAT('%', :query, '%') ORDER BY drug_name",
	})
}
