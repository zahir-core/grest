package grest

import (
	"fmt"
	"net/http"
	"os"
	"reflect"
	"runtime"
	"strings"
)

// Error is an implementation of the error interface with trace and other details.
type Error struct {
	Code    int
	Message string
	Detail  any
	PCs     []uintptr
}

// NewError returns an error with the specified status code, message, and optional detail.
func NewError(statusCode int, message string, detail ...any) *Error {
	var pcs [32]uintptr
	err := &Error{
		Code:    statusCode,
		Message: message,
		PCs:     pcs[0:runtime.Callers(2, pcs[:])],
	}
	if len(detail) > 0 {
		err.Detail = detail[0]
	}

	return err
}

// New return new *Error
func (e *Error) New(code int, message string, detail ...any) *Error {
	return NewError(code, message, detail...)
}

// GetError return *Error from any type err
func (e *Error) GetError(err any) *Error {
	if er, ok := err.(*Error); ok {
		return er
	} else if er, ok := err.(error); ok {
		code := http.StatusInternalServerError
		rv := reflect.ValueOf(err)
		if rv.Kind() == reflect.Pointer {
			rv = rv.Elem()
		}
		if rv.Kind() == reflect.Struct {
			if rvCode := rv.FieldByName("Code"); rvCode.Kind() == reflect.Int {
				code = int(rvCode.Int())
			}
		}
		return NewError(code, er.Error())
	}
	return NewError(http.StatusInternalServerError, fmt.Sprintf("%v", err))
}

// Error returns the error message, makes it compatible with the `error` interface.
func (e *Error) Error() string {
	return e.Message
}

// StatusCode returns the HTTP status code associated with the error.
func (e *Error) StatusCode() int {
	return e.Code
}

// Body returns the error details in a structured format suitable for response bodies.
func (e *Error) Body() map[string]any {
	body := map[string]any{
		"code":    e.Code,
		"message": e.Message,
	}
	if e.Detail != nil {
		body["detail"] = e.Detail
	}
	return map[string]any{
		"error": body,
	}
}

// Trace returns a slice of maps containing trace information about the error's origin.
func (e *Error) Trace() []map[string]any {
	trace := []map[string]any{}
	for _, pc := range e.PCs {
		pc = pc - 1
		fn := runtime.FuncForPC(pc)
		if fn != nil {
			funcName := fn.Name()
			fileName, lineNo := fn.FileLine(pc)
			trace = append(trace, map[string]any{
				"func": funcName,
				"file": fileName,
				"line": lineNo,
			})
		}
	}
	return trace
}

// TraceSimple returns a simplified map of trace information for displaying error traces.
func (e *Error) TraceSimple() map[string]string {
	trace := map[string]string{}
	for i, pc := range e.PCs {
		pc = pc - 1
		fn := runtime.FuncForPC(pc)
		if fn != nil {
			funcName := fn.Name()
			fileName, lineNo := fn.FileLine(pc)
			wd, _ := os.Getwd()
			if wd != "" {
				projectFile := strings.Split(fileName, wd+"/")
				if len(projectFile) > 1 {
					fileName = projectFile[1]
				}
			}
			modFile := strings.Split(fileName, "/pkg/mod/")
			if len(modFile) > 1 {
				fileName = modFile[1]
			}
			projectFunc := strings.Split(funcName, "/")
			funcName = projectFunc[len(projectFunc)-1]
			projectFunc = strings.Split(funcName, ".")
			if len(projectFunc) > 2 {
				funcName = projectFunc[len(projectFunc)-2] + "." + projectFunc[len(projectFunc)-1]
			}
			trace[fmt.Sprintf("#%02d", i)] = fmt.Sprintf("%s🔹 %s:%d", funcName, fileName, lineNo)
		}
	}
	return trace
}

// OriginalMessage return error message or error.detail.message if exists
func (e *Error) OriginalMessage() string {
	msg := e.Error()
	if mapError, ok := e.Body()["error"].(map[string]any); ok {
		if mapDetail, ok := mapError["detail"]; ok {
			if mpDetail, ok := mapDetail.(map[string]string); ok {
				if errMsg, ok := mpDetail["message"]; ok {
					msg = errMsg
				}
			} else if mpDetail, ok := mapDetail.(map[string]any); ok {
				if errMsg, ok := mpDetail["message"].(string); ok {
					msg = errMsg
				}
			}
		}
	}
	return msg
}
