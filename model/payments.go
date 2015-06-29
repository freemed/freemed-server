package model

import (
	"database/sql"
	"github.com/freemed/freemed-server/db"
)

const (
	TABLE_PAYMENTS = "payrec"
)

type PaymentModel struct {
	Added       NullTime       `db:"payrecdtadd" json:"added"`
	Modified    NullTime       `db:"payrecdtmod" json:"modified"`
	Patient     int64          `db:"payrecpatient" json:"patient_id"`
	Category    int64          `db:"payreccat" json:"category"`
	Procedure   int64          `db:"payrecproc" json:"procedure_id"`
	Source      int64          `db:"payrecsource" json:"source"`
	Link        int64          `db:"payreclink" json:"link"`
	Type        int64          `db:"payrectype" json:"type"`
	Number      sql.NullString `db:"payrecnum" json:"number"`
	Amount      float64        `db:"payrecamt" json:"amount"`
	Description sql.NullString `db:"payreclock" json:"locked"`
	Locked      string         `db:"payrecdescrip" json:"description"`
	User        int64          `db:"user" json:"user_id"`
	Active      string         `db:"active" json:"active"`
	Id          int64          `db:"id" json:"id"`
}

func init() {
	db.DbTables = append(db.DbTables, db.DbTable{TableName: TABLE_PAYMENTS, Obj: PaymentModel{}, Key: "Id"})
}
