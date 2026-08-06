package model

import (
	"database/sql"
	"time"

)

const (
	TABLE_INSCOGROUP = "inscogroup"
)

type InscoGroupModel struct {
	ID        int64          `db:"id" json:"id"`
	CreatedAt time.Time      `db:"created_at" json:"created_at"`
	UpdatedAt time.Time      `db:"updated_at" json:"updated_at"`
	DeletedAt sql.NullTime   `db:"deleted_at" json:"deleted_at"`
	Name string `db:"inscogroup" json:"name"`
}

func (InscoGroupModel) TableName() string {
	return TABLE_INSCOGROUP
}

func init() {
	DbSupportPicklists = append(DbSupportPicklists, DbSupportPicklist{ModuleName: "inscogroup", Query: "SELECT name AS v, id AS k FROM " + TABLE_INSCOGROUP + " WHERE name LIKE CONCAT('%', :query, '%') ORDER BY name"})
}
