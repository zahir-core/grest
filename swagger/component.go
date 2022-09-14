package swagger

import "net/http"

type Component interface {
	Tags() []string
	Summary() string
	Description() string
	Accept() string
	Produce() string
	Security() []string
	RequestBody() any
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
	Body       any
}

// // used by swagger api documentation generator
// // https://github.com/OAI/OpenAPI-Specification/issues/270
// func (m Model) SuccessResponses() []Response {
// 	res := []Response{}
// 	res = append(res, Response{StatusCode: http.StatusOK, Body: m})
// 	// res = append(res, Response{StatusCode: http.StatusCreated, Body: m})
// 	// res = append(res, Response{StatusCode: http.StatusOK, Body: grest.ListBodyStruct(m)})
// 	return res
// }

// // used by swagger api documentation generator
// // https://github.com/OAI/OpenAPI-Specification/issues/270
// func (Model) FailureResponses() []Response {
// 	res := []Response{}
// 	res = append(res, GetErrorResponse(NewError(http.StatusBadRequest, "A validation exception occurred.")))
// 	res = append(res, GetErrorResponse(NewError(http.StatusUnauthorized, "Invalid authorization credentials.")))
// 	res = append(res, GetErrorResponse(NewError(http.StatusForbidden, "User doesn't have permission to access the resource.")))
// 	return res
// }
