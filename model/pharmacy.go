package model

const TABLE_PHARMACY = "pharmacy"

type PharmacyModel struct {
	Id       int64  `json:"id"`
	Pharmacy string `json:"pharmacy"` // FIXME: schema uses 'pharmacy' as both table and column name
}

func (PharmacyModel) TableName() string {
	return TABLE_PHARMACY
}

func init() {
	DbSupportPicklists = append(DbSupportPicklists, DbSupportPicklist{
		ModuleName: "pharmacy",
		Query:      "SELECT pharmacy AS v, id AS k FROM " + TABLE_PHARMACY + " WHERE pharmacy LIKE CONCAT('%', :query, '%') ORDER BY pharmacy",
	})
}
