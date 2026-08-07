package config

import (
	"encoding/xml"
	"os"
	"strconv"

	"gopkg.in/yaml.v2"
)

var (
	Config AppConfig
)

type AppConfig struct {
	XMLName  xml.Name `yaml:"-" xml:"config"`
	Debug    bool     `yaml:"debug" xml:"debug"`
	Database struct {
		XMLName    xml.Name `yaml:"-" xml:"database"`
		Name       string   `yaml:"name" xml:"name"`
		User       string   `yaml:"user" xml:"user"`
		Pass       string   `yaml:"pass" xml:"pass"`
		Host       string   `yaml:"host" xml:"host"`
		Migrations bool     `yaml:"migrations" xml:"migrations"`
	} `yaml:"database"`
	Redis struct {
		XMLName    xml.Name `yaml:"-" xml:"redis"`
		Host       string   `yaml:"host" xml:"host"`
		Pass       string   `yaml:"pass" xml:"pass"`
		DatabaseId int      `yaml:"dbid" xml:"dbid"`
	} `yaml:"redis"`
	Web struct {
		XMLName xml.Name `yaml:"-" xml:"web"`
		Port    int      `yaml:"port" xml:"port"`
		TlsPort int      `yaml:"tls-port" xml:"tls-port"`
		Keys    struct {
			XMLName xml.Name `yaml:"-" xml:"keys"`
			RootCA  string   `yaml:"ca" xml:"ca"`
			Cert    string   `yaml:"cert" xml:"cert"`
			Key     string   `yaml:"key" xml:"key"`
		} `yaml:"keys"`
	} `yaml:"web"`
	Paths struct {
		XMLName          xml.Name `yaml:"-" xml:"paths"`
		BasePath         string   `yaml:"base-path" xml:"base-path"`
		DbMigrationsPath string   `yaml:"db-migrations" xml:"db-migrations"`
		Logs             string   `yaml:"logs" xml:"logs"`
	} `yaml:"paths"`
	Urls struct {
		XMLName xml.Name `yaml:"-" xml:"urls"`
	} `yaml:"urls"`
	Session struct {
		XMLName xml.Name `yaml:"-" xml:"session"`
		Expiry  int64    `yaml:"expiry" xml:"expiry"`
		Key     string   `yaml:"key" xml:"key"`
	} `yaml:"session"`
	Scheduler struct {
		Start    int `yaml:"start" xml:"start"`
		End      int `yaml:"end" xml:"end"`
		Interval int `yaml:"interval" xml:"interval"`
	} `yaml:"scheduler"`
}

var (
	defaultDatabasePass = "freemed"
	defaultSessionKey   = "freemed"
)

func (c *AppConfig) SetDefaults() {
	c.Debug = false
	c.Web.Port = 3000
	c.Web.TlsPort = 4000
	c.Database.Name = "freemed"
	c.Database.User = "freemed"
	c.Database.Pass = defaultDatabasePass
	c.Database.Host = ""
	c.Database.Migrations = true
	c.Redis.Host = "localhost:6379"
	c.Redis.Pass = ""
	c.Redis.DatabaseId = 0
	c.Paths.BasePath = "."
	c.Paths.DbMigrationsPath = "db/migrations"
	c.Paths.Logs = "logs"
	c.Session.Expiry = 10
	c.Session.Key = defaultSessionKey
}

// ValidateProduction returns a list of security warnings when default values
// are still in use. Call at startup and log findings; non-blocking.
func (c *AppConfig) ValidateProduction() []string {
	var warnings []string
	if c.Session.Key == defaultSessionKey {
		warnings = append(warnings, "FREEMED_SESSION_KEY is using the default value — set it via environment or config.yml for production")
	}
	if c.Database.Pass == defaultDatabasePass {
		warnings = append(warnings, "database password is using the default value — change it in config.yml or set FREEMED_DB_PASS")
	}
	if c.Debug {
		warnings = append(warnings, "debug mode is enabled — turn off in production")
	}
	return warnings
}

// applyEnvOverrides applies environment variable overrides for Docker/cloud deployment.
func (c *AppConfig) applyEnvOverrides() {
	if v := os.Getenv("FREEMED_DB_HOST"); v != "" {
		c.Database.Host = v
	}
	if v := os.Getenv("FREEMED_DB_USER"); v != "" {
		c.Database.User = v
	}
	if v := os.Getenv("FREEMED_DB_PASS"); v != "" {
		c.Database.Pass = v
	}
	if v := os.Getenv("FREEMED_DB_NAME"); v != "" {
		c.Database.Name = v
	}
	if v := os.Getenv("FREEMED_REDIS_HOST"); v != "" {
		c.Redis.Host = v
	}
	if v := os.Getenv("FREEMED_REDIS_PASS"); v != "" {
		c.Redis.Pass = v
	}
	if v := os.Getenv("FREEMED_REDIS_DB"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			c.Redis.DatabaseId = n
		}
	}
	if v := os.Getenv("FREEMED_WEB_PORT"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			c.Web.Port = n
		}
	}
	if v := os.Getenv("FREEMED_SESSION_KEY"); v != "" {
		c.Session.Key = v
	}
	if v := os.Getenv("FREEMED_DEBUG"); v != "" {
		c.Debug = v == "true" || v == "1"
	}
}

func LoadYamlConfigWithDefaults(configPath string) (*AppConfig, error) {
	c := &AppConfig{}
	c.SetDefaults()
	data, err := os.ReadFile(configPath)
	if err != nil {
		// Config file is optional when using env vars
		c.applyEnvOverrides()
		return c, nil
	}
	err = yaml.Unmarshal(data, c)
	if err != nil {
		return c, err
	}
	c.applyEnvOverrides()
	return c, nil
}

func LoadXmlConfigWithDefaults(configPath string) (*AppConfig, error) {
	c := &AppConfig{}
	c.SetDefaults()
	data, err := os.ReadFile(configPath)
	if err != nil {
		return c, err
	}
	err = xml.Unmarshal(data, c)
	if err != nil {
		return c, err
	}
	c.applyEnvOverrides()
	return c, nil
}
