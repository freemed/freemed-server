package model

import (
	"database/sql"
	"time"

)

const (
	TABLE_CLAIMTYPE = "claimtype"
)

type ClaimTypeModel struct {
	ID        int64          `db:"id" json:"id"`
	CreatedAt time.Time      `db:"created_at" json:"created_at"`
	UpdatedAt time.Time      `db:"updated_at" json:"updated_at"`
	DeletedAt sql.NullTime   `db:"deleted_at" json:"deleted_at"`
	Name        string    `db:"clmtpname" json:"name"`
	Description string    `db:"clmtpdescrip" json:"description"`
	Added       time.Time `db:"clmtpadd" json:"added"`
	Modified    time.Time `db:"clmtpmod" json:"modified"`
}

func (ClaimTypeModel) TableName() string {
	return TABLE_CLAIMTYPE
}
func init() {
	DbSupportPicklists = append(DbSupportPicklists, DbSupportPicklist{ModuleName: "claimtype", Query: "SELECT CONCAT(clmtpname, ' - ', clmtpdescrip) AS v, id AS k FROM " + TABLE_CLAIMTYPE + " WHERE CONCAT(clmtpname, ' - ', clmtpdescrip) LIKE CONCAT('%', :query, '%') ORDER BY clmtpname, clmtpdescrip"})
}
