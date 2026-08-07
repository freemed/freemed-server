package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"runtime/pprof"
	"strings"
	"syscall"
	"time"

	"github.com/freemed/freemed-server/api"
	"github.com/freemed/freemed-server/common"
	"github.com/freemed/freemed-server/config"
	dbpkg "github.com/freemed/freemed-server/internal/db"
	"github.com/freemed/freemed-server/model"
	"github.com/freemed/freemed-server/dbgen"
	"github.com/freemed/freemed-server/internal/middleware"
	"github.com/gin-gonic/contrib/gzip"
	"github.com/gin-gonic/gin"
	"gopkg.in/natefinch/lumberjack.v2"
)

var (
	configFile  = flag.String("config", "config.yml", "Configuration file")
	debug       = flag.Bool("debug", false, "Enable debugging")
	logToStdout = flag.Bool("log-stdout", false, "Enable redirecting all log output to stdout")
	cpuProfile  = flag.String("cpu-profile", "", "Write cpu profile to file")

	Version string
)

// slogWriter adapts a *slog.Logger to io.Writer so the standard log package
// can emit structured JSON through slog. Each Write call is logged as an
// Info-level message.
type slogWriter struct {
	logger *slog.Logger
}

func (w *slogWriter) Write(p []byte) (n int, err error) {
	msg := strings.TrimSpace(string(p))
	w.logger.Info(msg)
	return len(p), nil
}

func main() {
	flag.Parse()

	if *cpuProfile != "" {
		f, err := os.Create(*cpuProfile)
		if err != nil {
			log.Fatal(err)
		}
		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	}

	c, err := config.LoadYamlConfigWithDefaults(*configFile)
	if err != nil {
		log.Printf("FreeMED version %s\n\n", Version)
		panic(err)
	}
	if c == nil {
		log.Printf("FreeMED version %s\n\n", Version)
		panic("UNABLE TO LOAD CONFIG")
	}
	config.Config = *c

	for _, w := range c.ValidateProduction() {
		log.Printf("SECURITY WARNING: %s", w)
	}

	// Configure log output: file (via lumberjack rotation) or stdout,
	// optionally wrapped with slog JSON handler for structured logging.
	var logWriter io.Writer
	if !*logToStdout {
		logWriter = &lumberjack.Logger{
			Filename:   fmt.Sprintf("%s/%s/server.log", config.Config.Paths.BasePath, config.Config.Paths.Logs),
			MaxSize:    500, // megabytes
			MaxBackups: 20,
			MaxAge:     28,   // days
			LocalTime:  true, // don't use UTC
		}
	} else {
		logWriter = os.Stderr
	}

	if config.Config.LogFormat == "json" {
		// Bridge standard log → structured JSON via slog.
		jsonHandler := slog.NewJSONHandler(logWriter, &slog.HandlerOptions{Level: slog.LevelInfo})
		slogLogger := slog.New(jsonHandler)
		log.SetOutput(&slogWriter{logger: slogLogger})
		log.SetFlags(0) // slog adds its own timestamps
	} else {
		log.SetOutput(logWriter)
		if *debug {
			log.SetFlags(log.Lshortfile | log.LstdFlags | log.Lmicroseconds)
		} else {
			log.SetFlags(log.LstdFlags | log.Lmicroseconds)
		}
	}

	if *debug {
		log.Print("Overriding existing debug configuration")
		config.Config.Debug = true
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	// Initialize database/sql pool for sqlc
	log.Print("Initializing sqlc-compatible database/sql pool")
	sqlDB, err := dbpkg.Open()
	if err != nil {
		panic(err)
	}
	model.SqlDb = sqlDB

	// Start Prometheus DB pool metrics collection
	middleware.DBConnectionsGauge(sqlDB)

	// Initialize sqlc Queries wrapper
	model.Queries = dbgen.New(sqlDB)

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
	m.Use(gin.Recovery())
	m.Use(middleware.RequestID())
	m.Use(middleware.PrometheusMetrics())
	m.Use(gin.Logger())
	m.Use(middleware.SecurityHeaders())

	// Enable gzip compression
	m.Use(gzip.Gzip(gzip.DefaultCompression))

	// Serve SvelteKit SPA frontend
	if _, err := os.Stat("./frontend/build/index.html"); err == nil {
		log.Print("Serving SvelteKit frontend from frontend/build/")
		m.Static("/_app", "./frontend/build/_app")
		m.StaticFile("/favicon.ico", "./frontend/build/favicon.ico")
		m.StaticFile("/logo.png", "./frontend/build/logo.png")
		// SPA fallback: serve index.html for all non-API routes
		m.NoRoute(func(c *gin.Context) {
			c.File("./frontend/build/index.html")
		})
	}

	mw := getAuthMiddleware()

	// Rate limiting for login endpoint (10 attempts per minute per IP)
	loginLimiter := middleware.NewLoginRateLimiter(10, time.Minute)
	go loginLimiter.Cleanup(5 * time.Minute)

	// All authorized pieces live in /api
	a := m.Group("/api")

	// JWT pieces
	auth := m.Group("/auth")
	auth.GET("/csrf", middleware.GenerateCSRF)
	auth.POST("/login", loginLimiter.Middleware(), middleware.ValidateCSRF(), mw.LoginHandler)
	auth.GET("/me", mw.MiddlewareFunc(), authMe)
	auth.GET("/refresh_token", mw.RefreshHandler)
	auth.DELETE("/logout", authMiddlewareLogout)

	// Patient portal auth
	portalMw := getPortalAuthMiddleware()
	portalLimiter := middleware.NewLoginRateLimiter(5, 15*time.Minute)
	go portalLimiter.Cleanup(5 * time.Minute)

	portal := m.Group("/portal")
	portal.POST("/auth/login", portalLimiter.Middleware(), portalMw.LoginHandler)
	portal.GET("/auth/me", portalMw.MiddlewareFunc(), portalAuthMe)
	portal.GET("/auth/refresh_token", portalMw.RefreshHandler)
	portal.DELETE("/auth/logout", portalAuthMiddlewareLogout)

	// Portal appointment request with scheduling hours validation
	portal.POST("/appointments/request", portalMw.MiddlewareFunc(), api.PortalAppointmentRequest)

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

	// Create HTTP server
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", config.Config.Web.Port),
		Handler: m,
	}

	// Start server in goroutine
	go func() {
		log.Printf("Launching http on port :%d", config.Config.Web.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("HTTP server error: %v", err)
		}
	}()

	// TLS server if configured
	var tlsSrv *http.Server
	if config.Config.Web.Keys.Key != "" && config.Config.Web.Keys.Cert != "" {
		tlsSrv = &http.Server{
			Addr:    fmt.Sprintf(":%d", config.Config.Web.TlsPort),
			Handler: m,
		}
		go func() {
			log.Printf("Launching https on port :%d", config.Config.Web.TlsPort)
			if err := tlsSrv.ListenAndServeTLS(config.Config.Web.Keys.Cert, config.Config.Web.Keys.Key); err != nil && err != http.ErrServerClosed {
				log.Fatalf("HTTPS server error: %v", err)
			}
		}()
	}

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Print("Shutting down servers...")

	// Give outstanding requests 5 seconds to complete
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("HTTP server forced to shutdown: %v", err)
	}
	if tlsSrv != nil {
		if err := tlsSrv.Shutdown(ctx); err != nil {
			log.Fatalf("HTTPS server forced to shutdown: %v", err)
		}
	}
	log.Print("Servers stopped")
}
