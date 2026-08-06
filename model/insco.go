package model

import (
	"database/sql"
	"time"

)

const (
	TABLE_INSCO = "insco"
)

type InscoModel struct {
	ID        int64          `db:"id" json:"id"`
	CreatedAt time.Time      `db:"created_at" json:"created_at"`
	UpdatedAt time.Time      `db:"updated_at" json:"updated_at"`
	DeletedAt sql.NullTime   `db:"deleted_at" json:"deleted_at"`
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
}

func (InscoModel) TableName() string {
	return TABLE_INSCO
}

func init() {
	DbSupportPicklists = append(DbSupportPicklists, DbSupportPicklist{ModuleName: "insco", Query: "SELECT name AS v, id AS k FROM " + TABLE_INSCO + " WHERE name LIKE CONCAT('%', :query, '%') ORDER BY name"})
}
