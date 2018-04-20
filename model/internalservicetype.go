package model

const (
	TABLE_INTERNALSERVICETYPE = "intservtype"
)

type InternalServiceTypeModel struct {
	Name string `db:"intservtype" json:"name"`
	Id   int64  `db:"id" json:"id"`
}

func init() {
	DbTables = append(DbTables, DbTable{TableName: TABLE_INTERNALSERVICETYPE, Obj: InternalServiceTypeModel{}, Key: "Id"})
	DbSupportPicklists = append(DbSupportPicklists, DbSupportPicklist{ModuleName: "internalservicetype", Query: "SELECT intservtype AS v, id AS k FROM " + TABLE_INTERNALSERVICETYPE + " WHERE intservtype LIKE CONCAT('%', :query, '%') ORDER BY intservtype"})
}
