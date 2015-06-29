package model

import (
	"github.com/freemed/freemed-server/db"
	"time"
)

const (
	TABLE_CLAIMLOG = "claimlog"
)

type ClaimLogModel struct {
	Stamp         time.Time `db:"cltimestamp" json:"stamp"`
	User          int64     `db:"cluser" json:"user_id"`
	Procedure     int64     `db:"clprocedure" json:"procedure_id"`
	PaymentRecord int64     `db:"clpayrec" json:"payment_id"`
	Action        string    `db:"claction" json:"action"`
	Comment       string    `db:"clcomment" json:"comment"`
	Format        string    `db:"clformat" json:"format"`
	Target        string    `db:"cltarget" json:"target"`
	TargetOptions string    `db:"cltargetopt" json:"target_options"`
	BillKey       int64     `db:"clbillkey" json:"billkey_id"`
	Id            int64     `db:"id" json:"id"`
}

func init() {
	db.DbTables = append(db.DbTables, db.DbTable{TableName: TABLE_CLAIMLOG, Obj: ClaimLogModel{}, Key: "Id"})
}