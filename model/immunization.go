package model

import (
	"database/sql"
	"time"

	"github.com/freemed/freemed-server/common"
	"gorm.io/gorm"
)

const (
	TABLE_IMMUNIZATION  = "immunization"
	MODULE_IMMUNIZATION = "immunization"
)

type ImmunizationModel struct {
	gorm.Model
	Stamp                 time.Time     `db:"dateof" json:"stamp"`
	Patient               int64         `db:"patient" json:"patient_id"`
	Provider              int64         `db:"provider" json:"provider_id"`
	AdministeringProvider int64         `db:"admin_provider" json:"administering_provider_id"`
	EpisodeOfCare         sql.NullInt64 `db:"eoc" json:"episode_of_care"`
	Immunization          int64         `db:"immunization" json:"immunization_id"`
	Route                 int64         `db:"route" json:"route_id"`
	BodySite              int64         `db:"body_site" json:"body_site_id"`
	Manufacturer          NullString    `db:"manufacturer" json:"manufacturer"`
	LotNumber             NullString    `db:"lot_number" json:"lot_number"`
	PreviousDoses         int64         `db:"previous_doses" json:"previous_doses"`
	Recovered             bool          `db:"recovered" json:"recovered"`
	Notes                 NullString    `db:"notes" json:"notes"`
	OrderID               int64         `db:"orderid" json:"order_id"`
	Locked                int64         `db:"locked" json:"locked"`
	User                  int64         `db:"user" json:"user"`
	Active                string        `db:"active" json:"active"`
	Id                    int64         `db:"id" json:"id"`
}

func (ImmunizationModel) TableName() string {
	return TABLE_IMMUNIZATION
}

func init() {
	DbTables = append(DbTables,
		DbTable{
			TableName: TABLE_IMMUNIZATION,
			Obj:       ImmunizationModel{},
			Key:       "Id",
		},
	)
	common.EmrModuleMap[MODULE_IMMUNIZATION] = common.EmrModuleType{
		Name:         MODULE_IMMUNIZATION,
		PatientField: "Patient",
		Type:         ImmunizationModel{},
	}
}
