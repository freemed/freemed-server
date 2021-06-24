package model

import (
	"time"

	"github.com/freemed/freemed-server/common"
)

const (
	TABLE_AUTHORIZATIONS  = "authorizations"
	MODULE_AUTHORIZATIONS = "authorizations"
)

type AuthorizationModel struct {
	Added               time.Time `db:"authdtadd" json:"added"`
	Modified            time.Time `db:"authdtmod" json:"modified"`
	Patient             int64     `db:"authpatient" json:"patient_id"`
	BeginPeriod         time.Time `db:"authdtbegin" json:"begin"`
	EndPeriod           time.Time `db:"authdtend" json:"end"`
	AuthorizationNumber string    `db:"authnum" json:"authorization_number"`
	Type                int64     `db:"authtype" json:"type"`
	Provider            int64     `db:"authprov" json:"provider_id"`
	ProviderIdentifier  string    `db:"authprovid" json:"provider_identifier"`
	Payer               int64     `db:"authinsco" json:"payer_id"`
	VisitsTotal         int64     `db:"authvisits" json:"visits_total"`
	VisitsUsed          int64     `db:"authvisitsused" json:"visits_used"`
	VisitsRemaining     int64     `db:"authvisitsremain" json:"visits_remaining"`
	User                int64     `db:"user" json:"user"`
	Active              string    `db:"active" json:"active"`
	Id                  int64     `db:"id" json:"id"`
}

func init() {
	DbTables = append(DbTables,
		DbTable{
			TableName: TABLE_AUTHORIZATIONS,
			Obj:       AuthorizationModel{},
			Key:       "Id",
		},
	)
	common.EmrModuleMap[MODULE_AUTHORIZATIONS] = common.EmrModuleType{
		Name:         MODULE_AUTHORIZATIONS,
		PatientField: "Patient",
		Type:         AuthorizationModel{},
	}
}
