package model

import ()

const (
	TABLE_APPTTEMPLATE = "appttemplate"
)

type AppointmentTemplateModel struct {
	Name      string `db:"atname" json:"name"`
	Duration  int    `db:"atduration" json:"duration"`
	Equipment []byte `db:"atequipment" json:"equipment"`
	Color     string `db:"atcolor" json:"color"`
	Id        int64  `db:"id" json:"id"`
}

func init() {
	DbTables = append(DbTables, DbTable{TableName: TABLE_APPTTEMPLATE, Obj: AppointmentTemplateModel{}, Key: "Id"})
}
