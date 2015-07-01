package common

var (
	EmrModuleMap = map[string]EmrModuleType{}
)

type EmrModuleType struct {
	Name string
	Type interface{}
}

type EmrModule interface {
}
