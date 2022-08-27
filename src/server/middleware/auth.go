package middleware

import (
	"context"
	"fsrv/src/config"
	"fsrv/src/database/dbutil"
	"fsrv/src/database/entities"
	"fsrv/src/types"
	"fsrv/src/types/response"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
	"path/filepath"
	"strings"
	"time"
)

// Auth verifies that the issuer a request has authority to take a given action on the resource in question
func Auth(db dbutil.DBInterface, cfg *config.FileManager) gin.HandlerFunc {
	root := cfg.Path
	return func(ctx *gin.Context) {
		c, cancel := context.WithTimeout(ctx, 10*time.Second)
		defer cancel()
		authHandler(ctx, db, root, extractResPath(ctx))
		c.Done()
	}
}

func authHandler(ctx *gin.Context, db dbutil.DBInterface, root, path string) {
	dir := http.Dir(root)

	//get resource data
	resID, ok := extractResourceID(&dir, path)
	if !ok {
		ctx.AbortWithStatusJSON(403, response.Unauthorized)
		return
	}

	res, err := db.GetResourceData(resID)
	if err != nil {
		ctx.AbortWithStatusJSON(403, response.Unauthorized)
		return
	}

	//get key string
	authKey, keyGiven := extractKey(ctx)

	if !keyGiven {
		//evaluate access based on resource flags
		switch res.Flags {
		case entities.FlagPublicRead:
			if ctx.Request.Method != http.MethodGet {
				ctx.Next()
				return
			}
		default:
			log.Fatalf("Bad resource flag value \"%d\"", res.Flags)
		}

		ctx.Set("Resource", res)
		ctx.Next()
		return
	}

	key, err := db.GetKeyData(authKey)
	if err != nil {
		ctx.AbortWithStatusJSON(500, response.InternalServerError)
		return
	}

	//check key expiry
	if key.IsExpired() {
		ctx.AbortWithStatusJSON(401, response.UnauthorizedExpired)
		return
	}

	//evaluate access based on roles
	status := res.CheckAccess(key, getAccessType(ctx))
	switch status {
	case entities.AccessAllowed:
		ctx.Set("Resource", res)
		ctx.Set("Key", key)
		ctx.Next()
	case entities.AccessDenied:
		ctx.AbortWithStatusJSON(401, response.UnauthorizedExpired)
		return
	case entities.AccessNeutral:
		authHandler(ctx, db, root, filepath.Dir(strings.TrimSuffix(root, "/")))
	}
}

func getAccessType(ctx *gin.Context) types.OperationType {
	switch ctx.Request.Method {
	case http.MethodGet:
		return types.OperationRead
	default:
		log.Fatalf("Bad access type \"%s\"", ctx.Request.Method)
	}
	return -1
}
