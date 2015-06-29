package model

import (
	"database/sql"
	"github.com/go-martini/martini"
	_ "github.com/go-sql-driver/mysql"
	"gopkg.in/gorp.v1"
	"log"
	"os"
)

var (
	ApiMap   = map[string]func(martini.Router){}
	DbTables = make([]DbTable, 0)
	DbUser   string
	DbPass   string
	DbName   string
	DbHost   string
)

type DbTable struct {
	TableName string
	Obj       interface{}
	Key       string
}

func InitDb() *gorp.DbMap {
	dbobj, err := sql.Open("mysql", DbUser+":"+DbPass+"@/"+DbName)
	if err != nil {
		log.Fatalln("initDb: Fail to create database", err)
	}

	dbmap := &gorp.DbMap{
		Db:      dbobj,
		Dialect: gorp.MySQLDialect{"InnoDB", "UTF8"},
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
