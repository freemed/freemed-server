package main

const (
	TABLE_ZIPCODES = "zipcodes"
)

type ZipcodesModel struct {
	Id        int64   `db:"id" json:"id"`
	Zip       string  `db:"zip" json:"zip"`
	City      string  `db:"city" json:"city"`
	State     string  `db:"state" json:"state"`
	Latitude  float64 `db:"latitude" json:"latitude"`
	Longitude float64 `db:"longitude" json:"longitude"`
	Timezone  int64   `db:"timezone" json:"timezone"`
	DST       int64   `db:"dst" json:"dst"`
	Country   string  `db:"country" json:"country"`
}

func init() {
	dbTables = append(dbTables, DbTable{TableName: TABLE_ZIPCODES, Obj: ZipcodesModel{}, Key: "Id"})
}
