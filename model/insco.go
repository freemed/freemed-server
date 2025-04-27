package model

import (
	"time"

	"gorm.io/gorm"
)

const (
	TABLE_INSCO = "insco"
)

type InscoModel struct {
	gorm.Model
	DateAdded                      time.Time `db:"inscodtadd" json:"date_added"`
	DateModified                   time.Time `db:"inscodtmod" json:"date_modified"`
	Name                           string    `db:"insconame" json:"name"`
	Alias                          string    `db:"inscoalias" json:"alias"`
	AddressLine1                   string    `db:"inscoaddr1" json:"address_1"`
	AddressLine2                   string    `db:"inscoaddr2" json:"address_2"`
	City                           string    `db:"inscocity" json:"city"`
	State                          string    `db:"inscostate" json:"state"`
	Zip                            string    `db:"inscozip" json:"zip"`
	PhoneNumber                    string    `db:"inscophone" json:"phone_number"`
	FaxNumber                      string    `db:"inscofax" json:"fax_number"`
	GroupId                        int64     `db:"inscogroup" json:"group_id"`
	TypeId                         int64     `db:"inscotype" json:"type_id"`
	Assigned                       int64     `db:"inscoassign" json:"assigned"`
	Modifiers                      string    `db:"inscomod" json:"modifiers"`
	IdMap                          string    `db:"inscoidmap" json:"id_map"`
	X12Id                          string    `db:"inscox12id" json:"x12_id"`
	DefaultPaperFormat             string    `db:"inscodefformat" json:"default_paper_format"`
	DefaultPaperTarget             string    `db:"inscodeftarget" json:"default_paper_target"`
	DefaultPaperTargetOptions      string    `db:"inscodeftargetopt" json:"default_paper_target_options"`
	DefaultElectronicFormat        string    `db:"inscodefformate" json:"default_electronic_format"`
	DefaultElectronicTarget        string    `db:"inscodeftargete" json:"default_electronic_target"`
	DefaultElectronicTargetOptions string    `db:"inscodeftargetopte" json:"default_electronic_target_options"`
	Archived                       int64     `db:"inscoarchive" json:"archived"`
	Id                             int64     `db:"id" json:"id"`
}

func (InscoModel) TableName() string {
	return TABLE_INSCO
}

func init() {
	DbTables = append(DbTables, DbTable{TableName: TABLE_INSCO, Obj: InscoModel{}, Key: "Id"})
	DbSupportPicklists = append(DbSupportPicklists, DbSupportPicklist{ModuleName: "insco", Query: "SELECT name AS v, id AS k FROM " + TABLE_INSCO + " WHERE name LIKE CONCAT('%', :query, '%') ORDER BY name"})
}
