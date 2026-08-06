package model

import (
	"database/sql"
	"time"

)

const (
	TABLE_PRACTICE = "practice"
)

type PracticeModel struct {
	ID        int64          `db:"id" json:"id"`
	CreatedAt time.Time      `db:"created_at" json:"created_at"`
	UpdatedAt time.Time      `db:"updated_at" json:"updated_at"`
	DeletedAt sql.NullTime   `db:"deleted_at" json:"deleted_at"`
	Name        string     `db:"pracname" json:"name"`
	PracticeEin NullString `db:"pracein" json:"ein"`
	Addr1A      NullString `db:"addr1a" json:"addr1_1"`
	Addr2A      NullString `db:"addr2a" json:"addr2_1"`
	CityA       NullString `db:"citya" json:"city_1"`
	StateA      NullString `db:"statea" json:"state_1"`
	ZipA        NullString `db:"zipa" json:"zip_1"`
	PhoneA      NullString `db:"phonea" json:"phone_1"`
	FaxA        NullString `db:"faxa" json:"fax_1"`
	Addr1B      NullString `db:"addr1b" json:"addr1_2"`
	Addr2B      NullString `db:"addr2b" json:"addr2_2"`
	CityB       NullString `db:"cityb" json:"city_2"`
	StateB      NullString `db:"stateb" json:"state_2"`
	ZipB        NullString `db:"zipb" json:"zip_2"`
	PhoneB      NullString `db:"phoneb" json:"phone_2"`
	FaxB        NullString `db:"faxb" json:"fac_2"`
	Email       NullString `db:"email" json:"email"`
	Cellular    NullString `db:"cellular" json:"cellular"`
	NpiId       string     `db:"pracnpi" json:"npi_identifier"`
}

func (PracticeModel) TableName() string {
	return TABLE_PRACTICE
}

func init() {
	DbSupportPicklists = append(DbSupportPicklists, DbSupportPicklist{ModuleName: "practice", Query: "SELECT CONCAT(pracname, ', ', citya, ', ', statea) AS v, id AS k FROM " + TABLE_PRACTICE + " WHERE pracname LIKE CONCAT('%', :query, '%') ORDER BY pracname, citya, statea"})
}
