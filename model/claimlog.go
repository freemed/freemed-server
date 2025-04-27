package model

import (
	"time"

	"gorm.io/gorm"
)

const (
	TABLE_CLAIMLOG = "claimlog"
)

type ClaimLogModel struct {
	gorm.Model
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

func (ClaimLogModel) TableName() string {
	return TABLE_CLAIMLOG
}

func init() {
	DbTables = append(DbTables, DbTable{TableName: TABLE_CLAIMLOG, Obj: ClaimLogModel{}, Key: "Id"})
}
