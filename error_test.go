package grest

import (
	"net/http"
	"testing"
)

func TestError(t *testing.T) {
	statusCode := http.StatusBadRequest
	message := "A validation exception occurred."

	err := NewError(statusCode, message)
	res := GetErrorResponse(err)

	if res.StatusCode != statusCode {
		t.Errorf("Expected StatusCode [%v], got [%v]", statusCode, res.StatusCode)
	}

	b, ok := res.Body.(Error)
	if !ok {
		t.Error("res.Body is not Error")
	}

	if b.Err.Message != message {
		t.Errorf("Expected res.Body.Err.Message [%v], got [%v]", message, b.Err.Message)
	}
}
