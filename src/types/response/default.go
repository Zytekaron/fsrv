package response

var EmptySuccess = NewResponse[any](true, "", nil)
var EmptyError = NewResponse[any](false, "", nil)

var Forbidden = NewErrorMessage("forbidden")
var ForbiddenExpiredKey = NewErrorMessage("forbidden: expired key")
var Unauthorized = NewErrorMessage("unauthorized")
var TooManyRequests = NewErrorMessage("too many requests")
var TooManyConcurrentRequests = NewErrorMessage("too many concurrent requests")

var InternalServerError = NewErrorMessage("internal server error")
