package common

import (
	"github.com/go-martini/martini"
)

var (
	ApiMap = map[string]ApiMapping{}
)

type ApiMapping struct {
	Authenticated  bool
	JsonArmored    bool
	RouterFunction func(martini.Router)
}
