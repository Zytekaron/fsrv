package response

type Response[T any] struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
	Data    *T     `json:"data,omitempty"`
}

func NewResponse[T any](success bool, message string, data *T) *Response[T] {
	return &Response[T]{
		Success: success,
		Message: message,
		Data:    data,
	}
}

func NewSuccess[T any](message string, data T) *Response[T] {
	return NewResponse[T](true, message, &data)
}

func NewSuccessMessage[T any](message string) *Response[T] {
	return NewResponse[T](true, message, nil)
}

func NewSuccessData[T any](data T) *Response[T] {
	return NewResponse[T](true, "", &data)
}

func NewError[T any](message string, data T) *Response[T] {
	return NewResponse[T](false, message, &data)
}

func NewErrorMessage(message string) *Response[any] {
	return NewResponse[any](false, message, nil)
}

func NewErrorData[T any](data T) *Response[T] {
	return NewResponse[T](false, "", &data)
}
