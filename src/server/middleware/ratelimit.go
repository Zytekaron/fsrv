package middleware

import (
	"fsrv/src/database"
	"fsrv/src/types"
	"fsrv/src/types/response"
	"fsrv/utils/syncrl"
	"github.com/gin-gonic/gin"
	"log"
	"time"
)

const rateLimitPurgeInterval = 10 * time.Minute

// RateLimit applies key-based rate limiting.
func RateLimit(rateLimits database.RateLimitController) gin.HandlerFunc {
	rl := syncrl.New()
	rl.Purger(rateLimitPurgeInterval)

	return func(ctx *gin.Context) {
		key := ctx.MustGet("key").(*types.Key)
		level := key.RateLimit
		if level == "" {
			ctx.Next()
			return
		}

		manager, ok := rl.GetManager(level)
		if !ok {
			data, err := rateLimits.Get(level)
			if err != nil {
				log.Printf("error loading rate limit information for level '%s': %s\n", level, err)
				ctx.AbortWithStatusJSON(500, response.InternalServerError)
				return
			}

			manager = syncrl.NewManager(data.Limit, time.Duration(data.Reset))
			go manager.Purge()
			rl.AddManager(level, manager)
		}

		bucket := manager.GetBucket(key.ID)
		if !bucket.Draw() {
			ctx.AbortWithStatusJSON(429, response.TooManyRequests)
			return
		}

		ctx.Next()
	}
}
