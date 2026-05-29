package errors

import "fmt"

const (
	ErrCodeSuccess   = 0
	ErrCodeNotFound  = 1002
	ErrCodeNotLogin  = 2001
	ErrCodeRealname  = 2002
	ErrCodeVehicle   = 2003
	ErrCodeAccount   = 2005
	ErrCodeOnline    = 2006
	ErrCodeOngoing   = 2007
)

type BusinessError struct {
	Code int `json:"code"`
	Message string `json:"message"`
}

func (e *BusinessError) Error() string {
	return fmt.Sprintf("[%d] %s", e.Code, e.Message)
}

func NewBusinessError(code int, msg string) *BusinessError {
	return &BusinessError{Code: code, Message: msg}
}

func NewOnlineCheckError(code int, msg string) *BusinessError {
	return &BusinessError{Code: code, Message: msg}
}

func NewDispatchRejectError(code int, msg string) *BusinessError {
	return &BusinessError{Code: code, Message: msg}
}
