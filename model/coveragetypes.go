package model

import (
	"database/sql"
	"time"

)

const (
	TABLE_COVERAGETYPES = "covtypes"
)

type CoverageTypeModel struct {
	ID        int64          `db:"id" json:"id"`
	CreatedAt time.Time      `db:"created_at" json:"created_at"`
	UpdatedAt time.Time      `db:"updated_at" json:"updated_at"`
	DeletedAt sql.NullTime   `db:"deleted_at" json:"deleted_at"`
	Name        string    `db:"covtpname" json:"name"`
	Description string    `db:"covtpdescrip" json:"description"`
	Added       time.Time `db:"covtpdtadd" json:"added"`
	Modified    time.Time `db:"covtpdtmod" json:"modified"`
}

func (CoverageTypeModel) TableName() string {
	return TABLE_COVERAGETYPES
}

func init() {
	DbSupportPicklists = append(DbSupportPicklists, DbSupportPicklist{ModuleName: "coveragetypes", Query: "SELECT covtpname AS v, id AS k FROM " + TABLE_COVERAGETYPES + " WHERE covtpname LIKE CONCAT('%', :query, '%') ORDER BY covtpname"})
}
