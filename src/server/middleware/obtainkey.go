package middleware

import (
	"fsrv/src/database/dbutil"
	"fsrv/src/types/response"
	"github.com/gin-gonic/gin"
)

// ObtainKey retrieves data for a specific key and adds it to the context
func ObtainKey(db dbutil.DBInterface) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		key, err := db.GetKeyData(ctx.GetString("KeyID"))
		if err != nil {
			ctx.AbortWithStatusJSON(401, response.Forbidden)
			return
		}

		ctx.Set("key", &key)
		ctx.Next()
	}
}
