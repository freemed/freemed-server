package model

// DbSupportPicklist represents dynamically assembled maintenance module picklist targets for "maintenance" modules
type DbSupportPicklist struct {
	ModuleName string
	Query      string
}

// DbSupportPicklists is the internal representation of all picklists
var DbSupportPicklists = make([]DbSupportPicklist, 0)
