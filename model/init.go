package model

import (
	"database/sql"

	"github.com/freemed/freemed-server/dbgen"
)

var (
	SqlDb         *sql.DB // database/sql connection pool for sqlc and direct queries
	Queries       *dbgen.Queries
	SessionLength int
)