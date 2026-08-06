package model

import (
	"database/sql"
	"time"

)

const (
	TABLE_CALENDARGROUP = "calgroup"
)

type CalendarGroupModel struct {
	ID        int64          `db:"id" json:"id"`
	CreatedAt time.Time      `db:"created_at" json:"created_at"`
	UpdatedAt time.Time      `db:"updated_at" json:"updated_at"`
	DeletedAt sql.NullTime   `db:"deleted_at" json:"deleted_at"`
	Name      string `db:"groupname" json:"name"`
	Facility  int64  `db:"groupfacility" json:"facility_id"`
	Frequency int    `db:"groupfrequency" json:"frequency"`
	Length    int    `db:"grouplength" json:"length"`
	Members   string `db:"groupmembers" json:"members"`
}

func (CalendarGroupModel) TableName() string {
	return TABLE_CALENDARGROUP
}

func init() {
}
