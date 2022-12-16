package grest

import (
	"fmt"
	"os"
	"runtime"
	"strings"
)

type ErrorInterface interface {
	Error() string
	StatusCode() int
	Body() map[string]any
	Trace() []map[string]any
	TraceSimple() map[string]string
}

// Error is an implementation of error with trace & other detail.
type Error struct {
	Code    int
	Message string
	Detail  any
	PCs     []uintptr
}

func (e *Error) Error() string {
	return e.Message
}

func (e *Error) StatusCode() int {
	return e.Code
}

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

// NewError returns an error that formats as the given text with statusCode and detail if needed.
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
