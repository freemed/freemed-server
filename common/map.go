package common

import (
	"github.com/gin-gonic/gin"
)

var (
	ApiMap = map[string]ApiMapping{}
)

type ApiMapping struct {
	Authenticated  bool
	RouterFunction func(*gin.RouterGroup)
}
