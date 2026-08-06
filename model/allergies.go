package model

import (
	"database/sql"
	"time"
)

const (
	TABLE_ALLERGIES  = "allergies"
	MODULE_ALLERGIES = "allergies"
)

// AllergyModel mirrors the allergies table structure.
// The sqlc-generated dbgen.Allergy type is used for query results;
// this model provides TableName() and constants for consistency with
// other FreeMED entity models.
type AllergyModel struct {
	ID        int64        `db:"id" json:"id"`
	CreatedAt time.Time    `db:"created_at" json:"created_at"`
	UpdatedAt time.Time    `db:"updated_at" json:"updated_at"`
	DeletedAt sql.NullTime `db:"deleted_at" json:"deleted_at"`
	Patient   int64        `db:"patient" json:"patient_id"`
	Active    string       `db:"active" json:"active"`
}

func (AllergyModel) TableName() string {
	return TABLE_ALLERGIES
}

func init() {
}
