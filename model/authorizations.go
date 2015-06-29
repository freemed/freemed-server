package model

import (
	"github.com/freemed/freemed-server/db"
	"time"
)

const (
	TABLE_AUTHORIZATIONS = "authorizations"
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
	db.DbTables = append(db.DbTables, db.DbTable{TableName: TABLE_AUTHORIZATIONS, Obj: AuthorizationModel{}, Key: "Id"})
}