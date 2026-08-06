package api

import (
	"fmt"
	"log"
	"net/http"
	"strings"

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

	mod, err := resolveSupportModule(module)
	if err != nil {
		log.Printf("moduleSupportPicklist(): %s", err.Error())
		r.AbortWithStatus(http.StatusBadRequest)
		return
	}

	// Convert GORM-style :query named params to standard ?
	sqlQuery := strings.ReplaceAll(mod.Query, ":query", "?")

	rows, err := model.SqlDb.QueryContext(r.Request.Context(), sqlQuery, query)
	if err != nil {
		log.Print(err.Error())
		r.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	defer rows.Close()

	cols, err := rows.Columns()
	if err != nil {
		log.Print(err.Error())
		r.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	var out []iface
	for rows.Next() {
		vals := make([]interface{}, len(cols))
		valPtrs := make([]interface{}, len(cols))
		for i := range vals {
			valPtrs[i] = &vals[i]
		}
		if err := rows.Scan(valPtrs...); err != nil {
			log.Print(err.Error())
			continue
		}
		row := make(iface, len(cols))
		for i, col := range cols {
			row[col] = vals[i]
		}
		out = append(out, row)
	}

	r.JSON(http.StatusOK, out)
}
