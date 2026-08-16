package model

const TABLE_PHARMACY = "pharmacy"

type PharmacyModel struct {
	Id       int64  `json:"id"`
	Pharmacy string `json:"pharmacy"` // DB column is phname
}

func (PharmacyModel) TableName() string {
	return TABLE_PHARMACY
}

func init() {
	DbSupportPicklists = append(DbSupportPicklists, DbSupportPicklist{
		ModuleName: "pharmacy",
		Query:      "SELECT phname AS v, id AS k FROM " + TABLE_PHARMACY + " WHERE phname LIKE CONCAT('%', :query, '%') ORDER BY phname",
	})
}
