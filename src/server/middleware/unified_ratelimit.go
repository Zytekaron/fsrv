package middleware

import (
	"bytes"
	"crypto/sha512"
	"database/sql"
	"encoding/base64"
	"fsrv/src/config"
	"fsrv/src/database/dberr"
	"fsrv/src/database/dbutil"
	"fsrv/src/types/response"
	"fsrv/utils"
	"fsrv/utils/syncrl"
	"github.com/gin-gonic/gin"
	"github.com/zytekaron/gotil/v2/rl"
	"log"
	"math"
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

	validateKey := KeySourceValidator(serverConfig.KeyRandomBytes, serverConfig.KeyCheckBytes, []byte(serverConfig.KeyValidationSecret))

	return func(ctx *gin.Context) {
		keyID, keyProvided := extractKey(ctx)
		if keyProvided {
			//if attempting key authentication
			if keyAttemptRLMgr.Draw(keyID, 1) {
				//validate that key was issued by the server
				if !validateKey(keyID) {
					ctx.AbortWithStatusJSON(403, response.Forbidden)
					return
				}
				//check if key and rate limit exists
				rtLimID, err := db.GetKeyRateLimitID(keyID)
				if err != nil {
					if err == dberr.ErrKeyMissing {
						//if key is invalid
						ctx.AbortWithStatusJSON(403, response.Forbidden)
						return
					} else {
						//if key is valid (but no viable rate limit exists)
						keyAttemptRLMgr.Draw(keyID, -1) //revert successful auth attempt
						if defaultKeyRLMgr.Draw(keyID, 1) {
							ctx.Next()
							return
						}
					}
				}

				keyBm, ok := keyRLSuite.Get(rtLimID)
				if !ok {
					rateLimit, err := db.GetRateLimitData(rtLimID)
					if err != nil {
						if err == sql.ErrNoRows {
							//if key is valid (but no viable rate limit exists)
							keyAttemptRLMgr.Draw(keyID, -1) //revert successful auth attempt
							if defaultKeyRLMgr.Draw(keyID, 1) {
								ctx.Next()
								return
							}
							//if key is valid, no rate-limit exists, and is rate-limited
							ctx.AbortWithStatusJSON(429, response.TooManyRequests)
							return
						}
						//if key is invalid
						ctx.AbortWithStatusJSON(500, response.InternalServerError)
						return
					}

					//create and add bucket manager for rate limit level
					keyBm = rl.NewSync(rateLimit.Limit, time.Duration(rateLimit.Reset))
					keyRLSuite.Put(rtLimID, keyBm)
				}

				//draw from key rate limit
				keyAttemptRLMgr.Draw(keyID, -1) //revert successful auth attempt
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
				ctx.Next()
				return
			} else {
				ctx.AbortWithStatusJSON(429, response.TooManyRequests)
				return
			}
		}
	}
}

func KeySourceValidator(randomBytes, checksumBytes int, salt []byte) func(string) bool {
	if checksumBytes > 64 {
		log.Fatalln("KeySourceValidator: checksumBytes cannot be greater than 64 because sha512 produces 64 byte output")
	}
	const b64repMlt float64 = 1 / (6.0 / 8) //base64 representation multiplier
	size := int(math.Ceil(b64repMlt*float64(randomBytes)) + math.Ceil(b64repMlt*float64(checksumBytes)))

	return func(keyStr string) bool {
		if len(keyStr) != size {
			return false
		}

		key, err := base64.URLEncoding.DecodeString(keyStr)
		if err != nil {
			return false
		}

		data := key[:randomBytes]
		checksum := key[randomBytes:]

		sha := sha512.New()
		sha.Write(data)
		sha.Write(salt)
		shaSum := sha.Sum(nil)[:checksumBytes]
		return bytes.Equal(checksum, shaSum)
	}
}
