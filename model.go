package grest

import (
	"net/http"

	"grest.dev/grest/swagger"
)

type Schema interface {
	ConnName() string
	TableVersion() string
	TableName() string
	TableAliasName() string
	SetRelation()
	GetRelation() []Relation
	ModifyRelation([]Relation)
	SetFilter()
	GetFilter() []Filter
	ModifyFilter([]Filter)
	SetSort()
	GetSort() []Sort
	ModifySort([]Sort)
}

type Model struct {
	Relation []Relation `gorm:"-" json:"-"`
	Filter   []Filter   `gorm:"-" json:"-"`
	Sort     []Sort     `gorm:"-" json:"-"`
}

type Relation struct {
	JoinType          string
	TableName         string
	TableAliasName    string
	RelationCondition []Filter
}

type Filter struct {
	Column   string
	JsonKey  string
	Operator string
	Column2  string
	Value    interface{}
}

type Sort struct {
	Column    string
	JsonKey   string
	Direction string
}

func (Model) ConnName() string {
	return "default"
}

func (Model) TableVersion() string {
	return "init"
}

// default same with gorm
// func (m *Model) TableName() string {
// 	return ""
// }

// TableAliasName used by gREST to generate query
func (m *Model) TableAliasName() string {
	return "m"
}

// set the relation of the model
func (m *Model) SetRelation() {
	m.Relation = []Relation{}
}

// GetRelation used by gREST to generate query
func (m *Model) GetRelation() []Relation {
	return m.Relation
}

// ModifyRelation used in case you need to modify relation by context without url.Values query
func (m *Model) ModifyRelation(r []Relation) {
	m.Relation = r
}

// set the filter of the model
func (m *Model) SetFilter() {
	m.Filter = []Filter{}
}

// GetFilter used by gREST to generate query
func (m *Model) GetFilter() []Filter {
	return m.Filter
}

// ModifyFilter use in case you need to modify filter by context without url.Values query
func (m *Model) ModifyFilter(f []Filter) {
	m.Filter = f
}

// set the sort of the model
func (m *Model) SetSort() {
	m.Sort = []Sort{}
}

// GetSort used by gREST to generate query
func (m *Model) GetSort() []Sort {
	return m.Sort
}

// ModifySort use in case you need to modify sort by context without url.Values query
func (m *Model) ModifySort(s []Sort) {
	m.Sort = s
}

// used by swagger api documentation generator
func (Model) Tags() []string {
	return []string{}
}

// used by swagger api documentation generator
func (Model) Summary() string {
	return ""
}

// used by swagger api documentation generator
func (Model) Description() string {
	return ""
}

// used by swagger api documentation generator
func (Model) Accept() string {
	return "json"
}

// used by swagger api documentation generator
func (Model) Produce() string {
	return "json"
}

// used by swagger api documentation generator
func (Model) Security() []string {
	return []string{}
}

// used by swagger api documentation generator
func (m Model) RequestBody() interface{} {
	return m
}

// used by swagger api documentation generator
// https://github.com/OAI/OpenAPI-Specification/issues/270
func (m Model) SuccessResponses() []swagger.Response {
	res := []swagger.Response{}
	res = append(res, swagger.Response{StatusCode: http.StatusOK, Body: m})
	// res = append(res, Response{StatusCode: http.StatusCreated, Body: m})
	// res = append(res, Response{StatusCode: http.StatusOK, Body: grest.ListBodyStruct(m)})
	return res
}

// used by swagger api documentation generator
// https://github.com/OAI/OpenAPI-Specification/issues/270
func (Model) FailureResponses() []swagger.Response {
	res := []swagger.Response{}
	res = append(res, GetErrorResponse(NewError(http.StatusBadRequest, "A validation exception occurred.")))
	res = append(res, GetErrorResponse(NewError(http.StatusUnauthorized, "Invalid authorization credentials.")))
	res = append(res, GetErrorResponse(NewError(http.StatusForbidden, "User doesn't have permission to access the resource.")))
	return res
}

// used by swagger api documentation generator
func (Model) ExternalDoc() swagger.ExternalDoc {
	return swagger.ExternalDoc{}
}
