package model

import (
	"log"

	"github.com/freemed/freemed-server/config"
	_ "github.com/go-sql-driver/mysql"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var (
	// DbTables is the internal representation of all database tables
	DbTables = make([]DbTable, 0)
	// DbSupportPicklists is the internal representation of all picklists
	DbSupportPicklists = make([]DbSupportPicklist, 0)
	// DbFlags is the passed connection flags
	DbFlags = "parseTime=true&multiStatements=true"
)

// DbTable represents the internal metadata for a managed database table
type DbTable struct {
	TableName string
	Obj       interface{}
	Key       string
}

// DbModuleWithHooks defines local behavior for programmatic hooks regarding data for modules
type DbModuleWithHooks interface {
	// PreAdd defines code executed before an insert
	PreAdd() error
	// PreMod defines code executed before an update
	PreMod() error
	// PostAdd defines code executed after an insert
	PostAdd() error
	// PostMod defines code executed after an update
	PostMod() error
	// PreDel defines code executed before a deletion
	PreDel() error
}

// DbSupportPicklist represents dynamically assembled maintenance module picklist targets for "maintenance" modules
type DbSupportPicklist struct {
	ModuleName string
	Query      string
}

// InitDb initializes all database connections
func InitDb() *gorm.DB {
	db, err := gorm.Open(
		mysql.New(
			mysql.Config{
				DSN:                       config.Config.Database.User + ":" + config.Config.Database.Pass + "@" + config.Config.Database.Host + "/" + config.Config.Database.Name + "?" + DbFlags,
				DefaultStringSize:         256,   // default size for string fields
				DisableDatetimePrecision:  true,  // disable datetime precision, which not supported before MySQL 5.6
				DontSupportRenameIndex:    true,  // drop & create when rename index, rename index not supported before MySQL 5.7, MariaDB
				DontSupportRenameColumn:   true,  // `change` when rename column, rename column not supported before MySQL 8, MariaDB
				SkipInitializeWithVersion: false, // auto configure based on currently MySQL version
			},
		),
		&gorm.Config{},
	)

	if err != nil {
		log.Fatalln("initDb: Fail to create database", err)
	}

	for _, p := range DbTables {
		log.Printf("Migrating %s", p.TableName)
		err = db.AutoMigrate(&(p.Obj))
		if err != nil {
			log.Printf("ERR: %s", err.Error())
		}
	}

	if err != nil {
		log.Printf("initDb: Could not build tables: %s", err.Error())
	}

	return db
}
