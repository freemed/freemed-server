package db

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/golang-migrate/migrate/v4"
	migratemysql "github.com/golang-migrate/migrate/v4/database/mysql"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/go-sql-driver/mysql"
	"github.com/freemed/freemed-server/config"
)

// Pool is the standard library database/sql connection pool.
// It is used by sqlc-generated code and direct *sql.DB queries.
var Pool *sql.DB

// Open creates a new database/sql connection pool using go-sql-driver/mysql.
// If config.Config.Database.Migrations is true, pending migrations are applied.
func Open() (*sql.DB, error) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s)/%s?parseTime=true&multiStatements=true",
		config.Config.Database.User,
		config.Config.Database.Pass,
		config.Config.Database.Host,
		config.Config.Database.Name,
	)
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, fmt.Errorf("db.Open: %w", err)
	}
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("db.Ping: %w", err)
	}
	Pool = db

	if config.Config.Database.Migrations {
		if err := runMigrations(db); err != nil {
			return nil, fmt.Errorf("db.runMigrations: %w", err)
		}
	}

	return db, nil
}

func runMigrations(db *sql.DB) error {
	driver, err := migratemysql.WithInstance(db, &migratemysql.Config{})
	if err != nil {
		return fmt.Errorf("migrate driver: %w", err)
	}
	m, err := migrate.NewWithDatabaseInstance(
		"file://"+config.Config.Paths.DbMigrationsPath,
		config.Config.Database.Name,
		driver,
	)
	if err != nil {
		return fmt.Errorf("migrate new: %w", err)
	}
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("migrate up: %w", err)
	}
	log.Print("Migrations applied successfully")
	return nil
}
