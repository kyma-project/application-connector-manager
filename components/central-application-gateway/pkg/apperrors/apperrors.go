package apperrors

import "fmt"

const (
	CodeInternal                 = 1
	CodeNotFound                 = 2
	CodeAlreadyExists            = 3
	CodeWrongInput               = 4
	CodeUpstreamServerCallFailed = 5
)

type AppError interface {
	Code() int
	Error() string
}

type appError struct {
	code    int
	message string
}

func errorf(code int, format string, a ...interface{}) AppError {
	return appError{code: code, message: fmt.Sprintf(format, a...)}
}

func Internalf(format string, a ...interface{}) AppError {
	return errorf(CodeInternal, format, a...)
}

func Internal(message string) AppError {
	return appError{code: CodeInternal, message: message}
}

func NotFoundf(format string, a ...interface{}) AppError {
	return errorf(CodeNotFound, format, a...)
}

func NotFound(message string) AppError {
	return appError{code: CodeNotFound, message: message}
}

func AlreadyExists(format string, a ...interface{}) AppError {
	return errorf(CodeAlreadyExists, format, a...)
}

func WrongInputf(format string, a ...interface{}) AppError {
	return errorf(CodeWrongInput, format, a...)
}

func WrongInput(message string) AppError {
	return appError{code: CodeWrongInput, message: message}
}

func UpstreamServerCallFailed(format string, a ...interface{}) AppError {
	return errorf(CodeUpstreamServerCallFailed, format, a...)
}

func (ae appError) Code() int {
	return ae.code
}

func (ae appError) Error() string {
	return ae.message
}
