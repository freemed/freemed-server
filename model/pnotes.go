package model

import (
	"github.com/freemed/freemed-server/common"
	"github.com/freemed/remitt-server/model"
	"gorm.io/gorm"
)

const (
	TABLE_PNOTES  = "pnotes"
	MODULE_PNOTES = ""
)

type ProgressNotesModel struct {
	gorm.Model
	Date              model.NullTime   `db:"pnotesdt" json:"date"`
	DateAdded         model.NullTime   `db:"pnotesdtadd" json:"date_added"`
	DateModified      model.NullTime   `db:"pnotesdtmod" json:"date_modified"`
	Patient           int64            `db:"pnotespat" json:"patient_id"`
	Description       string           `db:"pnotesdescrip" json:"description"`
	Provider          int64            `db:"pnotesdoc" json:"provider_id"`
	EpisodeOfCare     model.NullInt64  `db:"pnoteseoc" json:"eoc_id"`
	NotesSubjective   string           `db:"pnotes_S" json:"notes_subjective"`
	NotesObjective    string           `db:"pnotes_O" json:"notes_objective"`
	NotesAssessment   string           `db:"pnotes_A" json:"notes_assessment"`
	NotesPlan         string           `db:"pnotes_P" json:"notes_plan"`
	NotesIntervention string           `db:"pnotes_I" json:"notes_intervention"`
	NotesEvaluation   string           `db:"pnotes_E" json:"notes_evaluation"`
	NotesRevision     string           `db:"pnotes_R" json:"notes_revision"`
	BPSystolic        uint             `db:"pnotessbp" json:"sbp"`
	BPDiastolic       uint             `db:"pnotesdbp" json:"dbp"`
	Temperature       float32          `db:"pnotestemp" json:"temperature"`
	HeartRate         uint             `db:"pnotesheartrate" json:"heart_rate"`
	RespiratoryRate   uint             `db:"pnotesresprate" json:"respiratory_rate"`
	Weight            uint             `db:"weight" json:"weight"`
	Height            uint             `db:"height" json:"height"`
	BMI               uint             `db:"bmi" json:"bmi"`
	ISO               model.NullString `db:"iso" json:"iso"`
	Locked            int64            `db:"locked" json:"locked"`
	User              int64            `db:"user" json:"user_id"`
	Active            string           `db:"active" json:"active"`
	Id                int64            `db:"id" json:"id"`
}

func init() {
	DbTables = append(DbTables, DbTable{
		TableName: TABLE_PNOTES,
		Obj:       ProgressNotesModel{},
		Key:       "Id",
	})
	common.EmrModuleMap[MODULE_PNOTES] = common.EmrModuleType{
		Name:         MODULE_PNOTES,
		PatientField: "Patient",
		Type:         ProgressNotesModel{},
	}
}
