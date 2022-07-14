package middleware

import (
	"context"
	"fsrv/src/database"
	"fsrv/src/database/entities"
	"fsrv/src/types/response"
	"fsrv/utils"
	"fsrv/utils/syncrl"
	"github.com/gin-gonic/gin"
	"github.com/zytekaron/gotil/v2/rl"
	"log"
	"time"
)

const rateLimitPurgeInterval = 10 * time.Minute

// RateLimit applies key-based rate limiting.
func RateLimit(rateLimits database.RateLimitController) gin.HandlerFunc {
	suite := syncrl.New()
	utils.Executor(rateLimitPurgeInterval, suite.Purge)

	return func(ctx *gin.Context) {
		key := ctx.MustGet("key").(*entities.Key)
		level := key.RateLimit
		if level == "" {
			ctx.Next()
			return
		}

		bm, ok := suite.Get(level)
		if !ok {
			c, cancel := context.WithTimeout(ctx, 10*time.Second)
			defer cancel()

			data, err := rateLimits.Get(c, level)
			if err != nil {
				log.Printf("error loading rate limit information for level '%s': %s\n", level, err)
				ctx.AbortWithStatusJSON(500, response.InternalServerError)
				return
			}

			bm = rl.NewSync(data.Limit, time.Duration(data.Reset))
			suite.Put(level, bm)
		}

		bucket := bm.Get(key.ID)
		if !bucket.Draw(1) {
			ctx.AbortWithStatusJSON(429, response.TooManyRequests)
			return
		}

		ctx.Next()
	}
}
