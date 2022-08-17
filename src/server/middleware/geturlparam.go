package middleware

import (
	"fsrv/src/types/response"
	"github.com/gin-gonic/gin"
)

func GetURLParamValue(param string, valueName string, denyEmpty bool) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		val, ok := ctx.GetQuery(param)
		if !ok || (denyEmpty && val == "") {
			ctx.AbortWithStatusJSON(400, response.NewError("Missing Request Parameter", param))
			return
		}

		ctx.Set(valueName, val)
		ctx.Next()
	}
}
