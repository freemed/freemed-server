package model

import (
	"time"

	"gorm.io/gorm"
)

const (
	TABLE_CLAIMTYPE = "claimtype"
)

type ClaimTypeModel struct {
	gorm.Model
	Name        string    `db:"clmtpname" json:"name"`
	Description string    `db:"clmtpdescrip" json:"description"`
	Added       time.Time `db:"clmtpadd" json:"added"`
	Modified    time.Time `db:"clmtpmod" json:"modified"`
	Id          int64     `db:"id" json:"id"`
}

func init() {
	DbTables = append(DbTables, DbTable{TableName: TABLE_CLAIMTYPE, Obj: ClaimTypeModel{}, Key: "Id"})
	DbSupportPicklists = append(DbSupportPicklists, DbSupportPicklist{ModuleName: "claimtype", Query: "SELECT CONCAT(clmtpname, ' - ', clmtpdescrip) AS v, id AS k FROM " + TABLE_CLAIMTYPE + " WHERE CONCAT(clmtpname, ' - ', clmtpdescrip) LIKE CONCAT('%', :query, '%') ORDER BY clmtpname, clmtpdescrip"})
}
