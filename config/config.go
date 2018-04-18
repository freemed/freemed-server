package config

import (
	"encoding/xml"
	"io/ioutil"

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
		XMLName              xml.Name `yaml:"-" xml:"urls"`
	} `yaml:"urls"`
	Session struct {
		XMLName xml.Name `yaml:"-" xml:"session"`
		Expiry  int64    `yaml:"expiry" xml:"expiry"`
		Key     string   `yaml:"key" xml:"key"`
	} `yaml:"session"`
}

func (c *AppConfig) SetDefaults() {
	c.Debug = false
	c.Web.Port = 3000
	c.Web.TlsPort = 4000
	c.Database.Name = "freemed"
	c.Database.User = "freemed"
	c.Database.Pass = "freemed"
	c.Database.Host = ""
	c.Database.Migrations = true
	c.Redis.Host = "localhost:6379"
	c.Redis.Pass = ""
	c.Redis.DatabaseId = 0
	c.Paths.BasePath = "."
	c.Paths.DbMigrationsPath = "db/migrations"
	c.Paths.Logs = "logs"
	c.Session.Expiry = 10
	c.Session.Key = "freemed"
}

func LoadYamlConfigWithDefaults(configPath string) (*AppConfig, error) {
	c := &AppConfig{}
	c.SetDefaults()
	data, err := ioutil.ReadFile(configPath)
	if err != nil {
		return c, err
	}
	err = yaml.Unmarshal([]byte(data), c)
	return c, err
}

func LoadXmlConfigWithDefaults(configPath string) (*AppConfig, error) {
	c := &AppConfig{}
	c.SetDefaults()
	data, err := ioutil.ReadFile(configPath)
	if err != nil {
		return c, err
	}
	err = xml.Unmarshal([]byte(data), c)
	return c, err
}
