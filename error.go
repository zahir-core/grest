package grest

import (
	"net/http"

	"github.com/gofiber/fiber/v2"

	"grest.dev/grest/swagger"
)

// Error is an implementation of error.
type Error struct {
	Err struct {
		Code    int         `json:"code"`
		Message string      `json:"message"`
		Detail  interface{} `json:"detail,omitempty"`
	} `json:"error"`
}

func (e Error) Error() string {
	return e.Err.Message
}

// NewError returns an error that formats as the given text with statusCode and detail if needed.
func NewError(statusCode int, message string, detail ...interface{}) error {
	err := Error{}
	err.Err.Code = statusCode
	err.Err.Message = message
	if len(detail) > 0 {
		err.Err.Detail = detail[0]
	}
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
