package middleware

import (
	"context"
	"fsrv/src/database/dberr"
	"fsrv/src/database/dbutil"
	"fsrv/src/types/response"
	"fsrv/utils"
	"fsrv/utils/syncrl"
	"github.com/gin-gonic/gin"
	"github.com/zytekaron/gotil/v2/rl"
	"log"
	"time"
)

const keyRateLimitPurgeInterval = 10 * time.Minute

// KeyRateLimit applies key based rate limiting
func KeyRateLimit(db dbutil.DBInterface, ipBM *rl.SyncBucketManager) gin.HandlerFunc {
	suite := syncrl.New()
	utils.Executor(keyRateLimitPurgeInterval, suite.Purge)
	badKeyRLBM := rl.NewSync(5, keyRateLimitPurgeInterval)
	badKeyRateLimitHandler := IPRateLimit(badKeyRLBM)

	return func(ctx *gin.Context) {
		keyID := ctx.GetString("KeyID")
		level, err := db.GetKeyRateLimitID(keyID)
		if err != nil {
			if err == dberr.ErrKeyMissing {
				badKeyRateLimitHandler(ctx)
				if ctx.IsAborted() {
					ctx.AbortWithStatusJSON(403, response.Forbidden)
				}
				return
			}
		}
		ipBM.Draw(ctx.ClientIP(), -1) //reverse ip rate limit decrement //todo: replace with something better, this fails if default untrusted limit is 0

		if level == "" {
			ctx.Next()
			return
		}

		bm, ok := suite.Get(level)
		if !ok {
			_, cancel := context.WithTimeout(ctx, 10*time.Second)
			defer cancel()

			rtLimID, err := db.GetKeyRateLimitID(keyID)
			if err != nil {
				log.Printf("error loading rate limit for key '%s': %s\n", ctx.GetString("keyID"), err)
				ctx.AbortWithStatusJSON(500, response.InternalServerError)
				return
			}

			rtLim, err := db.GetRateLimitData(rtLimID)
			if err != nil {
				log.Printf("error loading rate limit information for level '%s': %s\n", level, err)
				ctx.AbortWithStatusJSON(500, response.InternalServerError)
				return
			}

			bm = rl.NewSync(rtLim.Limit, time.Duration(rtLim.Reset))
			suite.Put(level, bm)
		}

		bucket := bm.Get(keyID)
		if !bucket.Draw(1) {
			ctx.AbortWithStatusJSON(429, response.TooManyRequests)
			return
		}

		ctx.Next()
	}
}
