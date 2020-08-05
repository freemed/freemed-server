package common

var (
	// EmrModuleMap tracks all of the EMR attachment database tables
	EmrModuleMap = map[string]EmrModuleType{}
)

// EmrModuleType defines the meta data tracked for all EMR attachment database tables
type EmrModuleType struct {
	Name         string
	PatientField string
	Type         interface{}
}

type EmrModule interface {
}
