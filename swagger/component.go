package swagger

import "net/http"

type Component interface {
	Tags() []string
	Summary() string
	Description() string
	Accept() string
	Produce() string
	Security() []string
	RequestBody() interface{}
	SuccessResponses() []Response
	FailureResponses() []Response
	ExternalDoc() ExternalDoc
}

type Route struct {
	Method    string
	Path      string
	Component Component
}

type Response struct {
	StatusCode int
	Header     http.Header
	Body       interface{}
}
