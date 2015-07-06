package model

import (
	"database/sql"
	"strconv"
)

// NullString is an expansion of sql.NullString from the database/sql package
// which properly marshals values to JSON.
type NullString struct {
	// Import NullString from database/sql package
	sql.NullString
}

func (s NullString) MarshalJSON() ([]byte, error) {
	if !s.Valid {
		return []byte("null"), nil
	}
	val, _ := s.Value()
	return []byte(strconv.QuoteToASCII(val.(string))), nil
}
