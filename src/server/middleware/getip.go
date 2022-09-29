package middleware

import "github.com/gin-gonic/gin"

// GetIP gets the client's ip and assigns it to the context.
//
//	Added Context Fields:
//	 ip -> string
func GetIP() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		ctx.Set("ip", ctx.ClientIP())
	}
}
