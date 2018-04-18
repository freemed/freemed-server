package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"net/http"
	"runtime/pprof"
	"strings"

	"github.com/braintree/manners"
	_ "github.com/freemed/freemed-server/api"
	"github.com/freemed/freemed-server/common"
	"github.com/freemed/freemed-server/config"
	"github.com/freemed/freemed-server/model"
	"github.com/gin-gonic/contrib/gzip"
	"github.com/gin-gonic/gin"
	"gopkg.in/natefinch/lumberjack.v2"
)

var (
	ConfigFile  = flag.String("config", "config.yml", "Configuration file")
	Debug       = flag.Bool("debug", false, "Enable debugging")
	LogToStdout = flag.Bool("log-stdout", false, "Enable redirecting all log output to stdout")
	CpuProfile  = flag.String("cpu-profile", "", "Write cpu profile to file")

	Version string
)

func main() {
	flag.Parse()

	if *Debug {
		log.SetFlags(log.Lshortfile | log.LstdFlags | log.Lmicroseconds)
	} else {
		log.SetFlags(log.LstdFlags | log.Lmicroseconds)
	}

	if *CpuProfile != "" {
		f, err := os.Create(*CpuProfile)
		if err != nil {
			log.Fatal(err)
		}
		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	}

	c, err := config.LoadYamlConfigWithDefaults(*ConfigFile)
	if err != nil {
		log.Printf("FreeMED version %s\n\n", Version)
		panic(err)
	}
	if c == nil {
		log.Printf("FreeMED version %s\n\n", Version)
		panic("UNABLE TO LOAD CONFIG")
	}
	config.Config = *c

	if !*LogToStdout {
		log.SetOutput(&lumberjack.Logger{
			Filename:   fmt.Sprintf("%s/%s/server.log", config.Config.Paths.BasePath, config.Config.Paths.Logs),
			MaxSize:    500, // megabytes
			MaxBackups: 20,
			MaxAge:     28,   // days
			LocalTime:  true, // don't use UTC
		})
	}

	if *Debug {
		log.Print("Overriding existing debug configuration")
		config.Config.Debug = true
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	log.Print("Initializing database backend")
	model.DbMap = model.InitDb()

	log.Print("Initializing session backend")
	common.ActiveSession = &common.SessionConnector{
		Address:    config.Config.Redis.Host,
		Password:   config.Config.Redis.Pass,
		DatabaseId: int64(config.Config.Redis.DatabaseId),
	}
	err = common.ActiveSession.Connect()
	if err != nil {
		panic(err)
	}

	log.Print("Initializing web services")
	m := gin.New()
	m.Use(gin.Logger())
	m.Use(gin.Recovery())

	// Enable gzip compression
	m.Use(gzip.Gzip(gzip.DefaultCompression))

	// Serve up the static UI...
	m.Static("/ui", "./ui")
	m.StaticFile("/favicon.ico", "./ui/favicon.ico")

	// ... with a redirection for the root page
	m.GET("/", func(c *gin.Context) {
		c.Redirect(http.StatusMovedPermanently, "./ui/index.html")
	})

	// All authorized pieces live in /api
	a := m.Group("/api")

	// JWT pieces
	auth := m.Group("/auth")
	auth.POST("/login", getAuthMiddleware().LoginHandler)
	auth.GET("/refresh_token", getAuthMiddleware().RefreshHandler)
	auth.DELETE("/logout", authMiddlewareLogout)
	auth.GET("/logout", authMiddlewareLogout) // for compatibility -- really shouldn't use this

	// Iterate through initializing API maps
	for k, v := range common.ApiMap {
		f := make([]string, 0)
		if v.Authenticated {
			f = append(f, "AUTH")
		}

		log.Printf("Adding handler /api/%s [%s]", k, strings.Join(f, ","))
		g := a.Group("/" + k)
		if v.Authenticated {
			g.Use(getAuthMiddleware().MiddlewareFunc())
		}
		v.RouterFunction(g)
	}

	if config.Config.Web.Keys.Key != "" && config.Config.Web.Keys.Cert != "" {
		log.Printf("Launching https on port :%d", config.Config.Web.TlsPort)
		go func() {
			log.Fatal(manners.ListenAndServeTLS(fmt.Sprintf(":%d", config.Config.Web.Port), config.Config.Web.Keys.Cert, config.Config.Web.Keys.Key, m))
		}()
	}

	// HTTP
	log.Printf("Launching http on port :%d", config.Config.Web.Port)
	log.Fatal(manners.ListenAndServe(fmt.Sprintf(":%d", config.Config.Web.Port), m))
}
