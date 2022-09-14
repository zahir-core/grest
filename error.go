package grest

import (
	"fmt"
	"net/http"
	"os"
	"runtime"
	"strings"

	"github.com/gofiber/fiber/v2"

	"grest.dev/grest/swagger"
)

// Error is an implementation of error.
type Error struct {
	Err struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
		Detail  any    `json:"detail,omitempty"`
	} `json:"error"`
	PCs []uintptr `json:"-"`
}

type Trace struct {
	FunctionName string `json:"func"`
	FileName     string `json:"file"`
	LineNumber   int    `json:"line"`
}

func (e Error) Code() int {
	return e.Err.Code
}

func (e Error) Error() string {
	return e.Err.Message
}

func (e Error) Trace() []Trace {
	trace := []Trace{}
	for _, pc := range e.PCs {
		pc = pc - 1
		fn := runtime.FuncForPC(pc)
		if fn != nil {
			funcName := fn.Name()
			fileName, lineNo := fn.FileLine(pc)
			trace = append(trace, Trace{
				FunctionName: funcName,
				FileName:     fileName,
				LineNumber:   lineNo,
			})
		}
	}
	return trace
}

func (e Error) TraceSimple() map[string]string {
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
func NewError(statusCode int, message string, detail ...any) error {
	err := Error{}
	err.Err.Code = statusCode
	err.Err.Message = message
	if len(detail) > 0 {
		err.Err.Detail = detail[0]
	}

	var pcs [32]uintptr
	err.PCs = pcs[0:runtime.Callers(2, pcs[:])]
	return err
}

// GetErrorResponse returns a Response with original Error
func GetErrorResponse(err error) swagger.Response {
	e, ok := err.(Error)
	if !ok {
		e.Err.Message = err.Error()
	}
	if e.Err.Code < 400 || e.Err.Code > 599 {
		e.Err.Code = http.StatusInternalServerError
	}
	return swagger.Response{StatusCode: e.Err.Code, Body: e}
}

func NewErrorHandler() fiber.ErrorHandler {
	return func(c *fiber.Ctx, err error) error {
		res := GetErrorResponse(err)
		return c.Status(res.StatusCode).JSON(res.Body)
	}
}

func NewNotFoundHandler() fiber.Handler {
	return func(c *fiber.Ctx) error {
		err := Error{}
		err.Err.Code = http.StatusNotFound
		err.Err.Message = "The resource you have specified cannot be found."
		return c.Status(err.Err.Code).JSON(err)
	}
}
