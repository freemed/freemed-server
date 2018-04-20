package api

import (
	"fmt"
	"log"
	"net/http"

	"github.com/freemed/freemed-server/common"
	"github.com/freemed/freemed-server/model"
	"github.com/gin-gonic/gin"
)

func init() {
	common.ApiMap["support"] = common.ApiMapping{
		Authenticated: true,
		RouterFunction: func(r *gin.RouterGroup) {
			r.GET("/:module/picklist/:query", moduleSupportPicklist)
		},
	}
}

type iface map[string]interface{}

func resolveSupportModule(module string) (model.DbSupportPicklist, error) {
	for _, iter := range model.DbSupportPicklists {
		if iter.ModuleName == module {
			return iter, nil
		}
	}
	return model.DbSupportPicklist{}, fmt.Errorf("resolveSupportModule: no module named '%s'", module)
}

func moduleSupportPicklist(r *gin.Context) {
	module := r.Param("module")
	query := r.Param("query")
	if module == "" || query == "" {
		r.AbortWithStatus(http.StatusBadRequest)
		return
	}

	// Check for module
	mod, err := resolveSupportModule(module)
	if err != nil {
		log.Printf("moduleSupportPicklist(): %s", err.Error())
		r.AbortWithStatus(http.StatusBadRequest)
		return
	}

	var o []iface

	_, err = model.DbMap.Select(&o, mod.Query, map[string]interface{}{
		"query": query,
	})

	if err != nil {
		log.Print(err.Error())
		r.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	r.JSON(http.StatusOK, o)
	return
}
