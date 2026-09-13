package config

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/xml"
	"fmt"
	"os"
	"strconv"

	"gopkg.in/yaml.v2"
)

var (
	Config AppConfig
)

// JWT signing-key domains. Every credential type this server issues is signed
// with its OWN key, derived from the single configured master secret, so a token
// minted for one domain can never be replayed against another. Before this split
// all four domains shared one key, which meant a patient's portal token verified
// as a staff FHIR session.
const (
	KeyDomainStaff  = "staff"  // provider/staff session JWT (cookie "jwt")
	KeyDomainPortal = "portal" // patient portal JWT (cookie "portal_jwt")
	KeyDomainFHIR   = "fhir"   // SMART on FHIR access tokens
	// KeyDomainOAuth2 is reserved for signed OAuth2 authorization codes. Codes are
	// currently opaque random DB rows (InsertFhirAuthCode), so nothing signs with
	// this domain yet; it is defined here so the domain set stays complete and the
	// code-binding change has a key waiting for it.
	KeyDomainOAuth2 = "oauth2"
)

// minSessionKeyLen is the shortest master secret accepted at startup. 32 bytes
// of entropy is the floor for an HS256 signing key.
const minSessionKeyLen = 32

// DomainKey derives the signing key for one credential domain from the master
// session secret using HMAC-SHA256 with a domain-separated label. The result is
// deterministic (so a restart does not invalidate live sessions) and distinct per
// domain (so token classes cannot be confused).
//
// Operators manage ONE secret — session.key / FREEMED_SESSION_KEY — and the four
// domain keys follow from it. Rotating the master secret rotates all of them.
func (c *AppConfig) DomainKey(domain string) []byte {
	mac := hmac.New(sha256.New, []byte(c.Session.Key))
	mac.Write([]byte("freemed/jwt/v1/" + domain))
	return mac.Sum(nil)
}

// ValidateStartup performs fail-fast configuration checks that must abort the
// process rather than log a warning. A weak or default JWT signing key is a
// total authentication bypass: anyone can mint an admin token offline, so this
// returns an error instead of the advisory produced by ValidateProduction.
func (c *AppConfig) ValidateStartup() error {
	switch {
	case c.Session.Key == "":
		return fmt.Errorf("session.key is not configured: set session.key in config.yml or export FREEMED_SESSION_KEY (e.g. openssl rand -base64 48)")
	case c.Session.Key == defaultSessionKey:
		return fmt.Errorf("session.key is the built-in default %q, which is public: set a strong random value via session.key or FREEMED_SESSION_KEY", defaultSessionKey)
	case len(c.Session.Key) < minSessionKeyLen:
		return fmt.Errorf("session.key is only %d bytes; at least %d bytes of entropy are required for an HS256 signing key", len(c.Session.Key), minSessionKeyLen)
	}
	return nil
}

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
	Sms struct {
		XMLName  xml.Name          `yaml:"-" xml:"sms"`
		Provider string            `yaml:"provider" xml:"provider"`
		Settings map[string]string `yaml:"settings" xml:"-"`
	} `yaml:"sms"`
	Tickler struct {
		XMLName  xml.Name `yaml:"-" xml:"tickler"`
		Interval int      `yaml:"interval" xml:"interval"`
	} `yaml:"tickler"`
	Ncpdp struct {
		XMLName  xml.Name `yaml:"-" xml:"ncpdp"`
		SenderID string   `yaml:"sender-id" xml:"sender-id"`
	} `yaml:"ncpdp"`
	Mwl struct {
		XMLName         xml.Name `yaml:"-" xml:"mwl"`
		Port            int      `yaml:"port" xml:"port"`
		AETitle         string   `yaml:"ae-title" xml:"ae-title"`
		MaxAssociations int      `yaml:"max-associations" xml:"max-associations"`
	} `yaml:"mwl"`
	LogFormat string `yaml:"log-format" xml:"log-format"`
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
	c.Sms.Provider = "noop"
	c.Tickler.Interval = 15
	// NCPDP sender ID is assigned by NCPDP (or the Surescripts gateway) per
	// organization — there is no safe default, so leave it empty and require
	// configuration before e-prescribing XML can be generated.
	c.Ncpdp.SenderID = ""
	// DICOM Modality Worklist C-FIND SCP. Port 0 disables the listener.
	c.Mwl.Port = 0
	c.Mwl.AETitle = "FREEMED"
	c.Mwl.MaxAssociations = 16
	c.LogFormat = "text"
}

// ValidateProduction returns a list of security warnings when default values
// are still in use. Call at startup and log findings; non-blocking.
//
// A default or weak session key is NOT listed here: it is fatal and handled by
// ValidateStartup, because it allows an attacker to mint an admin token offline.
func (c *AppConfig) ValidateProduction() []string {
	var warnings []string
	if c.Database.Pass == defaultDatabasePass {
		warnings = append(warnings, "database password is using the default value — change it in config.yml or set FREEMED_DB_PASS")
	}
	if c.Debug {
		warnings = append(warnings, "debug mode is enabled — turn off in production")
	}
	if c.Ncpdp.SenderID == "" {
		warnings = append(warnings, "NCPDP sender ID is not configured — e-prescribing (NCPDP SCRIPT) will be unavailable; set ncpdp.sender-id or FREEMED_NCPDP_SENDER_ID to the organization's NCPDP ID")
	}
	if c.Mwl.Port != 0 && c.Mwl.AETitle == "" {
		warnings = append(warnings, "DICOM Modality Worklist is enabled but no AE title is configured — the worklist SCP will refuse every association; set mwl.ae-title or FREEMED_MWL_AE_TITLE (max 16 characters)")
	}
	if c.Mwl.Port != 0 && len(c.Mwl.AETitle) > 16 {
		warnings = append(warnings, "DICOM Modality Worklist AE title is longer than 16 characters — no DICOM peer will be able to address it; set mwl.ae-title or FREEMED_MWL_AE_TITLE to 16 characters or fewer")
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
	if v := os.Getenv("FREEMED_LOG_FORMAT"); v != "" {
		c.LogFormat = v
	}
	if v := os.Getenv("FREEMED_SMS_PROVIDER"); v != "" {
		c.Sms.Provider = v
	}
	if v := os.Getenv("FREEMED_TICKLER_INTERVAL"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			c.Tickler.Interval = n
		}
	}
	if v := os.Getenv("FREEMED_NCPDP_SENDER_ID"); v != "" {
		c.Ncpdp.SenderID = v
	}
	if v := os.Getenv("FREEMED_MWL_PORT"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			c.Mwl.Port = n
		}
	}
	if v := os.Getenv("FREEMED_MWL_AE_TITLE"); v != "" {
		c.Mwl.AETitle = v
	}
	if v := os.Getenv("FREEMED_MWL_MAX_ASSOCIATIONS"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			c.Mwl.MaxAssociations = n
		}
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
