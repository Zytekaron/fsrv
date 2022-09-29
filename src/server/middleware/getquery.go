package middleware

import (
	"fsrv/src/types/response"
	"github.com/gin-gonic/gin"
)

type QueryParseFunc[V any] func(value string, present bool) (result V, err error)

// GetQuery gets a query from the URL and parses
// it based on the provided function. If an error
// is returned, the request fails with code 400
// and responds to the client with the error.
//
//		Added Context Fields:
//		 Variable; name and type provided by caller.
//	  If a property by the given name is already
//	  present, this middleware will panic.
func GetQuery[V any](param, name string, parse QueryParseFunc[V]) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		if _, ok := ctx.Get(name); ok {
			panic("duplicate context parameter name: " + name)
		}

		val, ok := ctx.GetQuery(param)
		res, err := parse(val, ok)
		if err != nil {
			ctx.AbortWithStatusJSON(400, response.NewErrorMessage("error parsing query: "+err.Error()))
			return
		}

		ctx.Set(name, res)
		ctx.Next()
	}
}
