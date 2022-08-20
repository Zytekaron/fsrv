package middleware

import (
	"database/sql"
	"fmt"
	"fsrv/src/config"
	"fsrv/src/database/dberr"
	"fsrv/src/database/dbutil"
	"fsrv/src/types/response"
	"fsrv/utils"
	"fsrv/utils/syncrl"
	"github.com/gin-gonic/gin"
	"github.com/zytekaron/gotil/v2/rl"
	"time"
)

const ipRateLimitPurgeInterval = 10 * time.Minute
const keyAttemptRateLimitPurgeInterval = 10 * time.Minute
const defaultKeyRateLimitPurgeInterval = 10 * time.Minute
const validKeyRateLimitPurgeInterval = 10 * time.Minute

func UnifiedRateLimit(db dbutil.DBInterface, serverConfig *config.Server) gin.HandlerFunc {
	ipRLMgr := rl.NewSync(serverConfig.IPAnonymousRL.Limit, time.Duration(serverConfig.IPAnonymousRL.Reset))
	keyAttemptRLMgr := rl.NewSync(serverConfig.KeyAuthAttemptRL.Limit, time.Duration(serverConfig.KeyAuthAttemptRL.Reset))
	defaultKeyRLMgr := rl.NewSync(serverConfig.KeyAuthDefaultRL.Limit, time.Duration(serverConfig.KeyAuthDefaultRL.Reset))
	keyRLSuite := syncrl.New()
	utils.Executor(ipRateLimitPurgeInterval, ipRLMgr.Purge)
	utils.Executor(keyAttemptRateLimitPurgeInterval, keyAttemptRLMgr.Purge)
	utils.Executor(defaultKeyRateLimitPurgeInterval, defaultKeyRLMgr.Purge)
	utils.Executor(validKeyRateLimitPurgeInterval, keyRLSuite.Purge)

	return func(ctx *gin.Context) {
		keyID, keyProvided := ctx.GetQuery("key")
		if keyProvided {
			//if attempting key authentication
			if keyAttemptRLMgr.Draw(keyID, 1) {
				//check if key and rate limit exists
				rtlimID, err := db.GetKeyRateLimitID(keyID)
				if err != nil {
					if err == dberr.ErrKeyMissing {
						//if key is invalid
						ctx.AbortWithStatusJSON(403, response.Forbidden)
						return
					} else {
						//if key is valid (but no viable rate limit exists)
						keyAttemptRLMgr.Draw(keyID, -1) //undraw sucessful auth attempt
						if defaultKeyRLMgr.Draw(keyID, 1) {
							ctx.Next()
							return
						}
					}
				}

				keyBm, ok := keyRLSuite.Get(rtlimID)
				if !ok {
					rateLimit, err := db.GetRateLimitData(rtlimID)
					if err != nil {
						if err == sql.ErrNoRows {
							//if key is valid (but no viable rate limit exists)
							keyAttemptRLMgr.Draw(keyID, -1) //undraw sucessful auth attempt
							if defaultKeyRLMgr.Draw(keyID, 1) {
								ctx.Next()
								return
							}
							//if key is valid, no ratelimit exists, and is ratelimited
							ctx.AbortWithStatusJSON(403, response.Forbidden)
							return
						}
						//if key is invalid
						ctx.AbortWithStatusJSON(500, response.InternalServerError)
						return
					}

					//create and add bucket manager for rate limit level
					keyBm = rl.NewSync(rateLimit.Limit, time.Duration(rateLimit.Reset))
					keyRLSuite.Put(rtlimID, keyBm)
				}

				//draw from key rate limit
				keyAttemptRLMgr.Draw(keyID, -1) //undraw sucessful auth attempt
				bucket := keyBm.Get(keyID)
				if bucket.Draw(1) {
					ctx.Next()
					return
				} else {
					ctx.AbortWithStatusJSON(429, response.TooManyRequests)
					return
				}

			}
			//if ip is rate limited from trying bad keys
			ctx.AbortWithStatusJSON(429, response.TooManyRequests)
			return
		} else {
			//if not attempting key authentication
			sb := ipRLMgr.Get(ctx.ClientIP())
			if sb.Draw(1) {
				fmt.Printf("uses remaining: %d", sb.RemainingUses())
				ctx.Next()
				return
			} else {
				ctx.AbortWithStatusJSON(429, response.TooManyRequests)
				return
			}
		}
	}
}
