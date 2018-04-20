package model

const (
	TABLE_DRUGQUANTITYQUALIFIER = "drugquantityqual"
)

type DrugQuantityQualifierModel struct {
	Code        string `db:"code" json:"code"`
	Description string `db:"description" json:"description"`
	Id          int64  `db:"id" json:"id"`
}

func init() {
	DbTables = append(DbTables, DbTable{TableName: TABLE_DRUGQUANTITYQUALIFIER, Obj: DrugQuantityQualifierModel{}, Key: "Id"})
	DbSupportPicklists = append(DbSupportPicklists, DbSupportPicklist{ModuleName: "drugquantityqualifier", Query: "SELECT CONCAT(code, ' - ', description) AS v, id AS k FROM " + TABLE_DRUGQUANTITYQUALIFIER + " WHERE CONCAT(code, ' - ', description) LIKE CONCAT('%', :query, '%') ORDER BY code, description"})
}
