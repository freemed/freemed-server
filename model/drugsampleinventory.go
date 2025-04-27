package model

import (
	"time"

	"gorm.io/gorm"
)

const (
	TABLE_DRUGSAMPLEINVENTORY = "drugsampleinv"
)

type DrugSampleInventoryModel struct {
	gorm.Model
	DrugCode             string    `db:"drugcode" json:"drug_code"`
	NDC                  string    `db:"drugndc" json:"ndc"`
	DrugClass            string    `db:"drugclass" json:"drug_class"`
	PackageCount         int       `db:"packagecount" json:"package_count"`
	Location             string    `db:"location" json:"location"`
	DrugCompany          string    `db:"drugco" json:"drug_company"`
	DrugRepresentative   string    `db:"drugrep" json:"drug_representative"`
	Invoice              string    `db:"invoice" json:"invoice"`
	SampleCount          int       `db:"samplecount" json:"sample_count"`
	SampleCountRemaining int       `db:"samplecountremain" json:"sample_count_remaining"`
	Lot                  string    `db:"lot" json:"lot"`
	Expiration           NullTime  `db:"expiration" json:"expiration"`
	Received             NullTime  `db:"received" json:"received"`
	AssignedTo           string    `db:"assignedto" json:"assigned_to"`
	LogUser              int64     `db:"loguser" json:"log_user"`
	LogDate              time.Time `db:"logdate" json:"log_date"`
	DisposedBy           string    `db:"disposeby" json:"disposed_by"`
	DisposalDate         NullTime  `db:"disposedate" json:"disposal_date"`
	DisposalMethod       string    `db:"disposemethod" json:"disposal_method"`
	DisposalReason       string    `db:"disposereason" json:"disposal_reason"`
	Witness              string    `db:"witness" json:"witness"`
	Id                   int64     `db:"id" json:"id"`
}

func (DrugSampleInventoryModel) TableName() string {
	return TABLE_DRUGSAMPLEINVENTORY
}

func init() {
	DbTables = append(DbTables, DbTable{TableName: TABLE_DRUGSAMPLEINVENTORY, Obj: DrugSampleInventoryModel{}, Key: "Id"})
	DbSupportPicklists = append(DbSupportPicklists, DbSupportPicklist{ModuleName: "drugsampleinventory", Query: "SELECT CONCAT(logdate, ' - ', drugformal, ' ', samplecountremain, '/', samplecount, ' (', lot, ')') AS v, id AS k FROM " + TABLE_DRUGSAMPLEINVENTORY + " WHERE CONCAT(logdate, ' - ', drugformal, ' ', samplecountremain, '/', samplecount, ' (', lot, ')') LIKE CONCAT('%', :query, '%') ORDER BY logdate DESC"})
}
