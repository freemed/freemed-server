package common

import (
	"github.com/gin-gonic/gin"
)

var (
	ApiMap = map[string]ApiMapping{}
)

type ApiMapping struct {
	Authenticated  bool
	JsonArmored    bool
	RouterFunction func(*gin.RouterGroup)
}
