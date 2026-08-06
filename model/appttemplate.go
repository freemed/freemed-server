package model

import (
	"database/sql"
	"time"

)

const (
	TABLE_APPTTEMPLATE = "appttemplate"
)

type AppointmentTemplateModel struct {
	ID        int64          `db:"id" json:"id"`
	CreatedAt time.Time      `db:"created_at" json:"created_at"`
	UpdatedAt time.Time      `db:"updated_at" json:"updated_at"`
	DeletedAt sql.NullTime   `db:"deleted_at" json:"deleted_at"`
	Name      string `db:"atname" json:"name"`
	Duration  int    `db:"atduration" json:"duration"`
	Equipment []byte `db:"atequipment" json:"equipment"`
	Color     string `db:"atcolor" json:"color"`
}

func (AppointmentTemplateModel) TableName() string {
	return TABLE_APPTTEMPLATE
}

func init() {
	DbSupportPicklists = append(DbSupportPicklists, DbSupportPicklist{ModuleName: "appttemplate", Query: "SELECT CONCAT(atname, ' (', atduration, 'm)') AS v, id AS k FROM " + TABLE_APPTTEMPLATE + " WHERE atname LIKE CONCAT('%', :query, '%') ORDER BY atname, atduration"})
}
