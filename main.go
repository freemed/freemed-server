package main

import (
	"flag"
	"fmt"
	_ "github.com/freemed/freemed-server/api"
	"github.com/freemed/freemed-server/common"
	"github.com/freemed/freemed-server/model"
	"github.com/go-martini/martini"
	"github.com/martini-contrib/render"
	"log"
	"net/http"
	"strings"
)

var (
	HTTP_PORT      = flag.Int("http-port", 3000, "HTTP serving port")
	HTTPS_PORT     = flag.Int("https-port", 3443, "HTTPS serving port")
	HTTPS_CERT     = flag.String("https-cert", "", "HTTPS PEM certificate")
	HTTPS_KEY      = flag.String("https-key", "", "HTTPS PEM key")
	DB_NAME        = flag.String("db-name", "freemed", "Database name")
	DB_USER        = flag.String("db-user", "freemed", "Database username")
	DB_PASS        = flag.String("db-pass", "freemed", "Database password")
	DB_HOST        = flag.String("db-host", "localhost", "Database host")
	REDIS_HOST     = flag.String("redis-host", "localhost:6379", "Redis database host")
	REDIS_PASSWORD = flag.String("redis-password", "", "Redis database password")
	REDIS_DBID     = flag.Int("redis-database-id", 0, "Redis database ID")
	SESSION_LENGTH = flag.Int("session-length", 600, "Session expiry in seconds")
)

func main() {
	flag.Parse()

	// Pass variables to packages
	model.SessionLength = *SESSION_LENGTH
	model.DbUser = *DB_USER
	model.DbPass = *DB_PASS
	model.DbName = *DB_NAME
	model.DbHost = *DB_HOST

	log.Print("Initializing database backend")
	model.DbMap = model.InitDb()

	log.Print("Initializing session backend")
	common.ActiveSession = &common.SessionConnector{
		Address:    *REDIS_HOST,
		Password:   *REDIS_PASSWORD,
		DatabaseId: int64(*REDIS_DBID),
	}
	err := common.ActiveSession.Connect()
	if err != nil {
		panic(err)
	}

	log.Print("Initializing web services")
	m := martini.Classic()

	m.Use(render.Renderer())

	static := Static("ui", StaticOptions{
		Exclude: "/api",
	})

	for k, v := range common.ApiMap {
		mw := make([]martini.Handler, 0)
		f := make([]string, 0)
		if v.Authenticated {
			mw = append(mw, TokenFunc(common.TokenAuthFunc))
			f = append(f, "AUTH")
		}
		if v.JsonArmored {
			mw = append(mw, common.ContentMiddleware)
			f = append(f, "JSON")
		}

		log.Printf("Adding handler /api/%s [%s]", k, strings.Join(f, ","))
		m.Group("/api/"+k, v.RouterFunction, mw...)
	}

	m.NotFound(static, http.NotFound)

	if *HTTPS_KEY != "" && *HTTPS_CERT != "" {
		log.Printf("Launching https on port :%d", *HTTPS_PORT)
		go func() {
			if err := http.ListenAndServeTLS(fmt.Sprintf(":%d", *HTTPS_PORT), *HTTPS_CERT, *HTTPS_KEY, m); err != nil {
				log.Fatal(err)
			}
		}()
	}

	// HTTP
	log.Printf("Launching http on port :%d", *HTTP_PORT)
	if err := http.ListenAndServe(fmt.Sprintf(":%d", *HTTP_PORT), m); err != nil {
		log.Fatal(err)
	}
}
