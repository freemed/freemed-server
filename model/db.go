package model

import (
	"database/sql"
	"log"
	"os"

	"github.com/freemed/freemed-server/config"
	_ "github.com/go-sql-driver/mysql"
	"gopkg.in/gorp.v1"
)

var (
	DbTables           = make([]DbTable, 0)
	DbSupportPicklists = make([]DbSupportPicklist, 0)
	DbFlags            = "parseTime=true&multiStatements=true"
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

func InitDb() *gorp.DbMap {
	dbobj, err := sql.Open("mysql", config.Config.Database.User+":"+config.Config.Database.Pass+"@"+config.Config.Database.Host+"/"+config.Config.Database.Name+"?"+DbFlags)
	if err != nil {
		log.Fatalln("initDb: Fail to create database", err)
	}

	// Remove all idle connections to stop long running failures
	dbobj.SetMaxIdleConns(0)
	dbobj.SetMaxOpenConns(50)

	dbmap := &gorp.DbMap{
		Db:      dbobj,
		Dialect: gorp.MySQLDialect{Engine: "InnoDB", Encoding: "UTF8"},
	}

	//dbmap.AddTableWithName(MyUserModel{}, "users").SetKeys(true, "Id")
	for _, v := range DbTables {
		keyName := v.Key
		log.Printf("initDb: Adding table %s", v.TableName)
		if keyName != "" {
			dbmap.AddTableWithName(v.Obj, v.TableName).SetKeys(true, keyName)
		} else {
			dbmap.AddTableWithName(v.Obj, v.TableName)
		}
	}

	err = dbmap.CreateTablesIfNotExists()
	if err != nil {
		log.Fatalln("initDb: Could not build tables", err)
	}

	dbmap.TraceOn("[gorp]", log.New(os.Stdout, "db: ", log.Lmicroseconds))

	return dbmap
}
