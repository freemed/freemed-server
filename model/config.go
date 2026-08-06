package model

import (
	"database/sql"
	"time"
	"context"
	"fmt"
	"sync"

	"github.com/freemed/freemed-server/dbgen"
)

const (
	TABLE_CONFIG = "config"
)

var (
	configCache     map[int64]dbgen.Config
	configCacheLock *sync.RWMutex
)

type ConfigModel struct {
	ID        int64          `db:"id" json:"id"`
	CreatedAt time.Time      `db:"created_at" json:"created_at"`
	UpdatedAt time.Time      `db:"updated_at" json:"updated_at"`
	DeletedAt sql.NullTime   `db:"deleted_at" json:"deleted_at"`
	Key     string     `db:"c_option" json:"key"`
	Value   NullString `db:"c_value" json:"value"`
	Title   NullString `db:"c_title" json:"title"`
	Section NullString `db:"c_section" json:"section"`
	Type    string     `db:"c_type" json:"type"`
	Options NullString `db:"c_options" json:"options"`
}

func (ConfigModel) TableName() string {
	return TABLE_CONFIG
}

func init() {
}

// cacheConfigValues is an internal caching mechanism for reading all config values
func cacheConfigValues(force bool) error {
	configCacheLock.Lock()
	defer configCacheLock.Unlock()
	if configCache == nil || force {
		configCache = map[int64]dbgen.Config{}
	}
	if len(configCache) < 1 || force {
		cm, err := Queries.ListAllConfig(context.Background())
		if err != nil {
			return err
		}
		for _, v := range cm {
			configCache[v.ID] = v
		}
	}
	return nil
}

// ConfigGetBySectionKey returns a config entry based on the specified key and section
func ConfigGetBySectionKey(section, key string) (ConfigModel, error) {
	err := cacheConfigValues(false)
	if err != nil {
		return ConfigModel{}, err
	}
	configCacheLock.RLock()
	defer configCacheLock.RUnlock()
	for _, v := range configCache {
		if v.CSection.String == section && v.COption == key {
			return dbgenConfigToModel(v), nil
		}
	}
	return ConfigModel{}, fmt.Errorf("config value with section %s and key %s not found", section, key)
}

// ConfigGetByKey returns a config entry based on the specified key
func ConfigGetByKey(key string) (ConfigModel, error) {
	err := cacheConfigValues(false)
	if err != nil {
		return ConfigModel{}, err
	}
	configCacheLock.RLock()
	defer configCacheLock.RUnlock()
	for _, v := range configCache {
		if v.COption == key {
			return dbgenConfigToModel(v), nil
		}
	}
	return ConfigModel{}, fmt.Errorf("config value with key %s not found", key)
}

// ConfigGetByID returns a config entry based on the specified PK
func ConfigGetByID(id int64) (ConfigModel, error) {
	err := cacheConfigValues(false)
	if err != nil {
		return ConfigModel{}, err
	}
	configCacheLock.RLock()
	defer configCacheLock.RUnlock()
	v, found := configCache[id]
	if !found {
		return ConfigModel{}, fmt.Errorf("config value with key %d not found", id)
	}
	return dbgenConfigToModel(v), nil
}

func dbgenConfigToModel(c dbgen.Config) ConfigModel {
	return ConfigModel{
		ID:      c.ID,
		Value:   NullString{NullString: c.CValue},
		Title:   NullString{NullString: c.CTitle},
		Section: NullString{NullString: c.CSection},
		Type:    c.CType,
		Options: NullString{NullString: c.COptions},
	}
}
