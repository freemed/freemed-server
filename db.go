package main

import (
	"database/sql"
	"github.com/coopernurse/gorp"
	//"github.com/go-martini/martini"
	_ "github.com/go-sql-driver/mysql"
	//"github.com/martini-contrib/binding"
	//"github.com/martini-contrib/render"
	//"github.com/martini-contrib/sessionauth"
	//"github.com/martini-contrib/sessions"
	"log"
	"os"
)

var (
	dbmap *gorp.DbMap

	dbTables = make([]DbTable, 0)
)

type DbTable struct {
	TableName string
	Obj       interface{}
	Key       string
}

func initDb() *gorp.DbMap {
	db, err := sql.Open("mysql", *DB_USER+":"+*DB_PASS+"@/"+*DB_NAME)
	if err != nil {
		log.Fatalln("initDb: Fail to create database", err)
	}

	dbmap := &gorp.DbMap{
		Db:      db,
		Dialect: gorp.MySQLDialect{"InnoDB", "UTF8"},
	}

	//dbmap.AddTableWithName(MyUserModel{}, "users").SetKeys(true, "Id")
	for _, v := range dbTables {
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
