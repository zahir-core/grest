package grest

import (
	"net/http"
	"strings"
	"testing"
)

func TestError(t *testing.T) {
	statusCode := http.StatusBadRequest
	message := "A validation exception occurred."
	traceSimple := "grest.TestError🔹 error_test.go:17"
	traceFunc := "grest.dev/grest.TestError"
	traceFile := "error_test.go"
	traceLine := 17

	e := NewError(statusCode, message)

	var err error
	err = e
	if err.Error() != e.Error() {
		t.Errorf("e is not an error")
	}

	if e.StatusCode() != statusCode {
		t.Errorf("Expected e.StatusCode() [%v], got [%v]", statusCode, e.StatusCode())
	}

	if e.Error() != message {
		t.Errorf("Expected e.Error() [%v], got [%v]", message, e.Error())
	}

	ts, ok := e.TraceSimple()["#00"]
	if !ok || ts != traceSimple {
		t.Errorf("Expected e.TraceSimple()[\"#00\"] [%v], got [%v]", traceSimple, ts)
	}

	traceTemp := e.Trace()
	if len(traceTemp) == 0 {
		t.Errorf("e.Trace() is empty")
	} else {
		funcName, ok := traceTemp[0]["func"].(string)
		if !ok || funcName != traceFunc {
			t.Errorf("Expected e.Trace()[0][func] [%v], got [%v]", traceFunc, funcName)
		}

		fileName, ok := traceTemp[0]["file"].(string)
		if !ok || !strings.HasSuffix(fileName, traceFile) {
			t.Errorf("Expected e.Trace()[0][file] [%v], got [%v]", traceFile, fileName)
		}

		line, ok := traceTemp[0]["line"].(int)
		if !ok || line != traceLine {
			t.Errorf("Expected e.Trace()[0][line] [%v], got [%v]", traceLine, line)
		}
	}
}
