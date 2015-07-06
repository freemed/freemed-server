package model

import (
	"database/sql"
	"time"
)

const (
	TABLE_PROCEDURE = "procrec"
)

type ProcedureModel struct {
	Patient             int64           `db:"procpatient" json:"patient_id"`
	EpisodeOfCare       NullString      `db:"proceoc" json:"episode_of_care"`
	CptCode             int64           `db:"proccpt" json:"cpt"`
	CptCodeModifier1    int64           `db:"proccptmod" json:"cpt_modifier_1"`
	CptCodeModifier2    int64           `db:"proccptmod2" json:"cpt_modifier_2"`
	CptCodeModifier3    int64           `db:"proccptmod3" json:"cpt_modifier_3"`
	Diagnosis1          int64           `db:"procdiag1" json:"dx_1"`
	Diagnosis2          int64           `db:"procdiag2" json:"dx_2"`
	Diagnosis3          int64           `db:"procdiag3" json:"dx_3"`
	Diagnosis4          int64           `db:"procdiag4" json:"dx_4"`
	DiagnosisSet        string          `db:"procdiagset" json:"dx_set"`
	Charges             float64         `db:"proccharges" json:"charges"`
	Units               float64         `db:"procunits" json:"units"`
	Voucher             NullString      `db:"procvoucher" json:"voucher"`
	Provider            int64           `db:"procphysician" json:"provider_id"`
	Date                time.Time       `db:"procdt" json:"date"`
	DateEnd             NullTime        `db:"procdtend" json:"date_end"`
	PlaceOfService      int64           `db:"procpos" json:"pos_id"`
	Comment             NullString      `db:"proccomment" json:"comment"`
	OriginalBalance     float64         `db:"procbalorig" json:"balance_original"`
	CurrentBalance      float64         `db:"procbalcurrent" json:"balance_current"`
	AmountPaid          float64         `db:"procamtpain" json:"amount_paid"`
	Billed              int64           `db:"procbilled" json:"billed"`
	Billable            int64           `db:"procbillable" json:"billable"`
	Authorization       int64           `db:"procauth" json:"authorization_id"`
	ReferringProvider   int64           `db:"procrefdoc" json:"referring_provider_id"`
	ReferralDate        NullTime        `db:"procrefdt" json:"referral_date"`
	AmountAllowed       sql.NullFloat64 `db:"procamtallowed" json:"amount_allowed`
	BilledDate          NullTime        `db:"procdtbilled" json:"billed_date"`
	CoverageCurrent     int64           `db:"proccurcovid" json:"current_coverage_id"`
	CoverageCurrentType int64           `db:"proccurcovtp" json:"current_coverage_type_id"`
	Coverage1           int64           `db:"proccov1" json:"coverage_id_1"`
	Coverage2           int64           `db:"proccov2" json:"coverage_id_2"`
	Coverage3           int64           `db:"proccov3" json:"coverage_id_3"`
	Coverage4           int64           `db:"proccov4" json:"coverage_id_4"`

	MedicaidReference     NullString `db:"procmedicaidref" json:"medicaid_ref"`
	MedicaidResubmission  NullString `db:"procmedicaidresub" json:"medicaid_resubmission"`
	LabCharges            float64    `db:"proclabcharges" json:"lab_charges"`
	Status                NullString `db:"procstatus" json:"status"`
	SlidingScale          NullString `db:"procslidingscale" json:"sliding_scale"`
	TypeOfServiceOverride int64      `db:"proctosoverride" json:"type_of_service_override"`
	Order                 int64      `db:"orderid" json:"order_id"`
	User                  int64      `db:"user" json:"user_id"`
	Active                string     `db:"active" json:"active"`
	Id                    int64      `db:"id" json:"id"`
}

func init() {
	DbTables = append(DbTables, DbTable{TableName: TABLE_PROCEDURE, Obj: ProcedureModel{}, Key: "Id"})
}
