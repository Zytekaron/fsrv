package middleware

import (
	"context"
	"fsrv/src/database"
	"fsrv/src/database/entities"
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
		key := ctx.MustGet("key").(*entities.Key)
		level := key.RateLimit
		if level == "" {
			ctx.Next()
			return
		}

		manager, ok := rl.GetManager(level)
		if !ok {
			c, cancel := context.WithTimeout(ctx, 10*time.Second)
			defer cancel()

			data, err := rateLimits.Get(c, level)
			if err != nil {
				log.Printf("error loading rate limit information for level '%s': %s\n", level, err)
				ctx.AbortWithStatusJSON(500, response.InternalServerError)
				return
			}

			manager = syncrl.NewManager(data.Limit, time.Duration(data.Reset))
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
