package model

import ()

const (
	TABLE_CALENDARGROUP = "calgroup"
)

type CalendarGroupModel struct {
	Name      string `db:"groupname" json:"name"`
	Facility  int64  `db:"groupfacility" json:"facility_id"`
	Frequency int    `db:"groupfrequency" json:"frequency"`
	Length    int    `db:"grouplength" json:"length"`
	Members   string `db:"groupmembers" json:"members"`
	Id        int64  `db:"id" json:"id"`
}

func init() {
	DbTables = append(DbTables, DbTable{TableName: TABLE_CALENDARGROUP, Obj: CalendarGroupModel{}, Key: "Id"})
}
