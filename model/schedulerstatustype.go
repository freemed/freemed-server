package model

import (
	"database/sql"
	"time"

)

const (
	TABLE_SCHEDULERSTATUSTYPE = "schedulerstatustype"
)

type SchedulerStatusTypeModel struct {
	ID        int64          `db:"id" json:"id"`
	CreatedAt time.Time      `db:"created_at" json:"created_at"`
	UpdatedAt time.Time      `db:"updated_at" json:"updated_at"`
	DeletedAt sql.NullTime   `db:"deleted_at" json:"deleted_at"`
	Name        string `db:"sname" json:"name"`
	Description string `db:"sdescrip" json:"description"`
	Color       string `db:"scolor" json:"color"`
	Age         int    `db:"sage" json:"age"`
}

func (SchedulerStatusTypeModel) TableName() string {
	return TABLE_SCHEDULERSTATUSTYPE
}

func init() {
	DbSupportPicklists = append(DbSupportPicklists, DbSupportPicklist{ModuleName: "inscogroup", Query: "SELECT CONCAT(name, ' ', description) AS v, id AS k FROM " + TABLE_SCHEDULERSTATUSTYPE + " WHERE name LIKE :query OR description LIKE CONCAT('%', :query, '%') ORDER BY name, description"})
}
