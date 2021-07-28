package model

const (
	TABLE_PLACEOFSERVICE = "pos"
)

type PlaceOfServiceModel struct {
	Name        string   `db:"posname" json:"name"`
	Description string   `db:"posdescrip" json:"description"`
	Added       NullTime `db:"posdtadd" json:"added"`
	Modified    NullTime `db:"posdtmod" json:"modified"`
	Id          int64    `db:"id" json:"id"`
}

func init() {
	DbTables = append(DbTables, DbTable{TableName: TABLE_PLACEOFSERVICE, Obj: PlaceOfServiceModel{}, Key: "Id"})
	DbSupportPicklists = append(DbSupportPicklists, DbSupportPicklist{ModuleName: "placeofservice", Query: "SELECT CONCAT(posname, ' - ', posdescrip) AS v, id AS k FROM " + TABLE_PLACEOFSERVICE + " WHERE posname LIKE CONCAT('%', :query, '%') OR posdescrip LIKE CONCAT('%', :query, '%') ORDER BY posname, posdescrip"})
}
