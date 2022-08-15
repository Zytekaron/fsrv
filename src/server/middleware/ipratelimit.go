package middleware

import (
	"fsrv/src/types/response"
	"fsrv/utils"
	"github.com/gin-gonic/gin"
	"github.com/zytekaron/gotil/v2/rl"
	"time"
)

const ipRateLimitPurgeInterval = 10 * time.Minute

// IPRateLimit applies ip based rate limiting
func IPRateLimit(limit int, duration time.Duration) gin.HandlerFunc {
	bm := rl.NewSync(limit, duration)
	utils.Executor(ipRateLimitPurgeInterval, bm.Purge)

	return func(ctx *gin.Context) {
		sb := bm.Get(ctx.ClientIP())
		if sb.Draw(1) {
			ctx.Next()
		} else {
			ctx.AbortWithStatusJSON(429, response.TooManyRequests)
		}
	}
}
