package model

import (
	"database/sql"
	"time"

)

const (
	TABLE_CLAIMLOG = "claimlog"
)

type ClaimLogModel struct {
	ID        int64          `db:"id" json:"id"`
	CreatedAt time.Time      `db:"created_at" json:"created_at"`
	UpdatedAt time.Time      `db:"updated_at" json:"updated_at"`
	DeletedAt sql.NullTime   `db:"deleted_at" json:"deleted_at"`
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
}

func (ClaimLogModel) TableName() string {
	return TABLE_CLAIMLOG
}

func init() {
}
