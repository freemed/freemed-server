package model

import ()

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
}
