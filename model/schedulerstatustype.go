package model

import "gorm.io/gorm"

const (
	TABLE_SCHEDULERSTATUSTYPE = "schedulerstatustype"
)

type SchedulerStatusTypeModel struct {
	gorm.Model
	Name        string `db:"sname" json:"name"`
	Description string `db:"sdescrip" json:"description"`
	Color       string `db:"scolor" json:"color"`
	Age         int    `db:"sage" json:"age"`
	Id          int64  `db:"id" json:"id"`
}

func (SchedulerStatusTypeModel) TableName() string {
	return TABLE_SCHEDULERSTATUSTYPE
}

func init() {
	DbTables = append(DbTables, DbTable{TableName: TABLE_SCHEDULERSTATUSTYPE, Obj: SchedulerStatusTypeModel{}, Key: "Id"})
	DbSupportPicklists = append(DbSupportPicklists, DbSupportPicklist{ModuleName: "inscogroup", Query: "SELECT CONCAT(name, ' ', description) AS v, id AS k FROM " + TABLE_SCHEDULERSTATUSTYPE + " WHERE name LIKE :query OR description LIKE CONCAT('%', :query, '%') ORDER BY name, description"})
}
