package model

const (
	TABLE_ENCLOSURETYPE = "enctype"
)

type EnclosureTypeModel struct {
	EnclosureType string `db:"enclosure" json:"enclosure"`
	Id            int64  `db:"id" json:"id"`
}

func init() {
	DbTables = append(DbTables, DbTable{TableName: TABLE_ENCLOSURETYPE, Obj: EnclosureTypeModel{}, Key: "Id"})
	DbSupportPicklists = append(DbSupportPicklists, DbSupportPicklist{ModuleName: "enclosuretype", Query: "SELECT enclosure AS v, id AS k FROM " + TABLE_ENCLOSURETYPE + " WHERE enclosure LIKE CONCAT('%', :query, '%') ORDER BY enclosure"})
}
