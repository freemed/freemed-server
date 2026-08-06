package model

import (
	"database/sql"
	"time"

)

const (
	TABLE_INTERNALSERVICETYPE = "intservtype"
)

type InternalServiceTypeModel struct {
	ID        int64          `db:"id" json:"id"`
	CreatedAt time.Time      `db:"created_at" json:"created_at"`
	UpdatedAt time.Time      `db:"updated_at" json:"updated_at"`
	DeletedAt sql.NullTime   `db:"deleted_at" json:"deleted_at"`
	Name string `db:"intservtype" json:"name"`
}

func (InternalServiceTypeModel) TableName() string {
	return TABLE_INTERNALSERVICETYPE
}

func init() {
	DbSupportPicklists = append(DbSupportPicklists, DbSupportPicklist{ModuleName: "internalservicetype", Query: "SELECT intservtype AS v, id AS k FROM " + TABLE_INTERNALSERVICETYPE + " WHERE intservtype LIKE CONCAT('%', :query, '%') ORDER BY intservtype"})
}
