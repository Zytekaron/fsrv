package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/pkg/xattr"
	"net/http"
	"path/filepath"
	"strings"
)

const xAttributeNS = "user.fsrv."
const xAttributeResource = xAttributeNS + "resourceid"

func extractKey(ctx *gin.Context) (string, bool) {
	auth := ctx.GetHeader("authorization")
	if len(auth) > 0 {
		return auth, true
	}
	return ctx.GetQuery("key")
}

func extractResPath(ctx *gin.Context) string {
	return ctx.Request.URL.Path
}

func extractResourceID(dir *http.Dir, path string) (string, string, bool) {
	for {
		//check if valid
		_, err := dir.Open(string(*dir) + path)

		if err != nil {
			return "", "", false
		}

		resourceID, err := xattr.Get(path, xAttributeResource)
		//if resource id attribute is found
		if err == nil {
			return string(resourceID), path, true
		}

		path = filepath.Dir(strings.TrimSuffix(path, "/")) //get parent directory
	}
}
