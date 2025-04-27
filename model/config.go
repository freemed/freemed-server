package model

import (
	"fmt"
	"sync"

	"gorm.io/gorm"
)

const (
	TABLE_CONFIG = "config"
)

var (
	configCache     map[int64]ConfigModel
	configCacheLock *sync.RWMutex
)

type ConfigModel struct {
	gorm.Model
	Key     string     `db:"c_option" json:"key"`
	Value   NullString `db:"c_value" json:"value"`
	Title   NullString `db:"c_title" json:"title"`
	Section NullString `db:"c_section" json:"section"`
	Type    string     `db:"c_type" json:"type"`
	Options NullString `db:"c_options" json:"options"`
	Id      int64      `db:"id" json:"id"`
}

func (ConfigModel) TableName() string {
	return TABLE_CONFIG
}

func init() {
	DbTables = append(DbTables, DbTable{TableName: TABLE_CONFIG, Obj: ConfigModel{}, Key: "Id"})
}

// cacheConfigValues is an internal caching mechanism for reading all config values
func cacheConfigValues(force bool) error {
	configCacheLock.Lock()
	defer configCacheLock.Unlock()
	if configCache == nil || force {
		configCache = map[int64]ConfigModel{}
	}
	if len(configCache) < 1 || force {
		var cm []ConfigModel

		tx := Db.Find(&cm)
		if tx.Error != nil {
			return tx.Error
		}
		for _, v := range cm {
			configCache[v.Id] = v
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
		if v.Section.String == section && v.Key == key {
			return v, nil
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
		if v.Key == key {
			return v, nil
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
	return v, nil
}
