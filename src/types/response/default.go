package response

var EmptySuccess = NewResponse[any](true, "", nil)
var EmptyError = NewResponse[any](false, "", nil)

var Forbidden = NewErrorMessage("forbidden")
var Unauthorized = NewErrorMessage("unauthorized")
var UnauthorizedExpired = NewErrorMessage("unauthorized: expired key")
var TooManyRequests = NewErrorMessage("too many requests")

var InternalServerError = NewErrorMessage("internal server error")
