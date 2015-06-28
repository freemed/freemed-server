package main

import (
	"flag"
	"fmt"
	"github.com/go-martini/martini"
	"github.com/martini-contrib/render"
	"log"
	"net/http"
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
	SESSION_LENGTH = flag.Int("session-length", 600, "Session expiry in seconds")

	IsRunning = true
	apimap = map[string]func(martini.Router){}
)

func main() {
	flag.Parse()

	log.Print("Initializing database backend")
	dbmap = initDb()

	log.Print("Initializing background services")
	go sessionExpiryThread()

	log.Print("Initializing web services")
	m := martini.Classic()

	m.Use(render.Renderer())

	static := Static("ui", StaticOptions{
		Exclude: "/api",
	})

	for k, v := range apimap {
		if k == "auth" {
			m.Group("/api/"+k, v, contentMiddleware)
		} else {
			m.Group("/api/"+k, v, contentMiddleware, TokenFunc(tokenAuthFunc))
		}
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
	//m.Run()
	//go func() {
	log.Printf("Launching http on port :%d", *HTTP_PORT)
	if err := http.ListenAndServe(fmt.Sprintf(":%d", *HTTP_PORT), m); err != nil {
		log.Fatal(err)
	}
	//}()

}
