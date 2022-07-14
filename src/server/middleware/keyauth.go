package middleware

import (
	"context"
	"fsrv/src/database"
	"fsrv/src/types/response"
	"fsrv/utils"
	"github.com/gin-gonic/gin"
	"github.com/zytekaron/gotil/v2/rl"
	"log"
	"time"
)

const authLimit = 10
const authReset = time.Minute
const authPurgeInterval = 10 * time.Minute

// KeyAuth validates that a key exists and is not expired
// and then adds the "key" property to the gin context.
func KeyAuth(keys database.KeyController) gin.HandlerFunc {
	bm := rl.NewSync(authLimit, authReset)
	utils.Executor(authPurgeInterval, bm.Purge)

	return func(ctx *gin.Context) {
		auth, ok := extractKey(ctx)
		if !ok {
			ctx.AbortWithStatusJSON(403, response.Forbidden)
			return
		}

		// implement ip-level rate limiting to prevent repeated
		// failed attempts to authenticate with bad credentials
		ip := ctx.ClientIP()
		bucket := bm.Get(ip)
		if !bucket.CanDraw(1) {
			ctx.AbortWithStatusJSON(429, response.TooManyRequests)
			return
		}

		c, cancel := context.WithTimeout(ctx, 10*time.Second)
		defer cancel()

		key, err := keys.Get(c, auth)
		if err != nil {
			if err == database.ErrNoDocuments {
				// only draw from the bucket when the authentication fails
				// due to an invalid token (expired tokens are acceptable)
				bucket.Draw(1)
				ctx.AbortWithStatusJSON(403, response.Unauthorized)
				return
			}

			log.Printf("error loading key information for key '%s': %s\n", auth, err)
			ctx.AbortWithStatusJSON(500, response.InternalServerError)
			return
		}

		if key.IsExpired() {
			ctx.AbortWithStatusJSON(401, response.UnauthorizedExpired)
			return
		}

		ctx.Set("key", key)
		ctx.Next()
	}
}

func extractKey(ctx *gin.Context) (string, bool) {
	auth := ctx.GetHeader("authorization")
	if len(auth) > 0 {
		return auth, true
	}
	query := ctx.Query("key")
	if len(query) > 0 {
		return query, true
	}
	return "", false
}
