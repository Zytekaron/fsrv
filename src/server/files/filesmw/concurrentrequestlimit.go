package filesmw

import (
	"fsrv/src/database/entities"
	"fsrv/src/types"
	"fsrv/src/types/response"
	"fsrv/utils/syncmap"
	"github.com/gin-gonic/gin"
)

// ConcurrentRequestLimit
//
//	Middleware Dependencies:
//	 GetIP
func ConcurrentRequestLimit(readLimit, writeLimit int) gin.HandlerFunc {
	readMap := syncmap.New[string, int]()
	writeMap := syncmap.New[string, int]()

	return func(ctx *gin.Context) {
		// limit and counts map for this request
		var limit int
		var counts *syncmap.CountMap[string, int]
		switch crlGetMode(ctx) {
		case types.OperationRead:
			limit = readLimit
			counts = readMap
		case types.OperationWrite, types.OperationModify, types.OperationDelete:
			limit = writeLimit
			counts = writeMap
		}

		// client id (key id or ip)
		var id string
		key, ok := ctx.Get("key")
		if ok {
			id = key.(*entities.Key).ID
		} else {
			id = ctx.GetString("ip")
		}

		// check if the count is less than the limit, and
		// increment it if so. otherwise, 429 and exit.
		if !counts.CompareLessAndIncrement(id, limit) {
			ctx.AbortWithStatusJSON(429, response.TooManyConcurrentRequests)
			return
		}

		// run the other request handlers
		ctx.Next()

		// decrement the counter when done
		counts.Decrement(id)
	}
}

// TODO: get the mode based on the information about the request.
//
//	maybe involve another prerequisite adminmw to assign the
//	request type to the context for use within here and ratelimiting.
func crlGetMode(ctx *gin.Context) types.OperationType {
	return types.OperationRead
}
