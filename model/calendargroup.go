package model

import (
	"github.com/freemed/freemed-server/db"
)

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
	db.DbTables = append(db.DbTables, db.DbTable{TableName: TABLE_CALENDARGROUP, Obj: CalendarGroupModel{}, Key: "Id"})
}
