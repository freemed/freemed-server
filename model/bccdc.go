package model

import "gorm.io/gorm"

const (
	TABLE_BCCDC = "bccdc"
)

type BccdcModel struct {
	gorm.Model
	Code        string `db:"agent_code" json:"code"`
	Description string `db:"description" json:"description"`
	Id          int64  `db:"id" json:"id"`
}

func init() {
	DbTables = append(DbTables, DbTable{TableName: TABLE_BCCDC, Obj: BccdcModel{}, Key: "Id"})
	DbSupportPicklists = append(DbSupportPicklists, DbSupportPicklist{ModuleName: "bccdc", Query: "SELECT CONCAT(code, ' - ', description) AS v, id AS k FROM " + TABLE_BCCDC + " WHERE description LIKE CONCAT('%', :query, '%') ORDER BY code, description"})
}
