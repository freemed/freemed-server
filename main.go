package main

import (
	"flag"
	"fmt"
	"github.com/go-martini/martini"
	"github.com/martini-contrib/binding"
	"github.com/martini-contrib/encoder"
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

	static := martini.Static("ui", martini.StaticOptions{
		Exclude: "/api",
	})
	m.Use(func(c martini.Context, w http.ResponseWriter) {
		c.MapTo(encoder.JsonEncoder{}, (*encoder.Encoder)(nil))
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
	})
	m.Group("/api/auth", func(r martini.Router) {

		r.Post("/login", binding.Json(AuthLoginObj{}), AuthLogin)
		r.Delete("/logout", AuthLogout)
	})
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
