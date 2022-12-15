package grest

import (
	"net/url"

	"gorm.io/gorm"
)

type ModelInterface interface {
	TableVersion() string
	TableName() string
	TableAliasName() string
	SetFields()
	AddField(field_key string, field_opt map[string]any)
	GetFields() map[string]map[string]any
	SetRelations()
	AddRelation(relation_key string, relation_opt map[string]any)
	GetRelations() map[string]map[string]any
	SetFilters()
	AddFilter(filter map[string]any)
	GetFilters() []map[string]any
	SetSorts()
	AddSort(sort map[string]any)
	GetSorts() []map[string]any
	GetQuerySchema() map[string]any
	ToSQL(tx *gorm.DB, q url.Values) string
}

// model struct tag :
//
// json:
// - dot notation field to be parsed to multi-dimensional json object
// - also used for alias field when query to db
//
// db:
// - field to query to db
// - add ",group" to group the field
// - add ",hide" to hide the field on api response
//
// gorm:
// - field option for main table of the struct
// - the details can be found at : https://gorm.io/docs/models.html#Fields-Tags
// - if the field is not field of main table of the struct in the database, gorm must be setted to "-"
//
// validate:
// - validation tag to validate the data
// - "required" tag will be used as "required" on OpenAPI Specification
// - "oneof" tag will be used as "enum" on OpenAPI Specification
// - "max" tag will be used as "maximum" or "maxLength" on OpenAPI Specification based on value type
// - "min" tag will be used as "minimum" or "minLength" on OpenAPI Specification based on value type
// - the details can be found at : https://pkg.go.dev/github.com/go-playground/validator/v10
//
// title:
// - used as "title" on OpenAPI Specification
//
// note:
// - used as "description" on OpenAPI Specification
//
// default:
// - used for the default value when insert to db
// - used as "default" on OpenAPI Specification
//
// example:
// - used as "example" on OpenAPI Specification
type Model struct {
	// described in the following pattern : map[field_key]map[opt_key]opt_value
	// field_key: the field shown on json
	// opt_key, same as model struct tag + the following key :
	// - data_type
	// opt_value, same as model struct tag value
	Fields map[string]map[string]any `json:"-" gorm:"-"`

	// automaticaly created from SetFields based on struct fields, used for array filter :
	// ?array_fields.0.field.id={field_id} > where exists (select 1 from array_table at where at.parent_id = parent.id and field_id = {field_id})
	// ?array_fields.*.field.id={field_id} > same as above but the array fields response also filtered
	ArrayFields map[string]any `json:"-" gorm:"-"`

	// described in the following pattern : map[field_key]map[opt_key]opt_value
	// field_key: same as "table_alias_name" on "opt_key"
	// opt_key :
	// - type : sql join type (inner, left, etc)
	// - table_name :
	// - table_alias_name
	// - conditions : []map[string]any same as "Filters"
	Relations map[string]map[string]any `json:"-" gorm:"-"`

	// described in the following pattern : []map[opt_key]opt_value
	// opt_key :
	// - column_1 : column in the db to be filtered
	// - column_1_json_key : dot notation paths of json field in column_1
	// - operator : sql operator (=, !=, >, >=, <, <=, like, not like, in, not in, etc)
	// - column_2 : another column in the db (to compare values between columns in the db)
	// - column_2_json_key : dot notation paths of json field in column_2
	// - value : desired value to be filtered
	Filters []map[string]any `json:"-" gorm:"-"`

	// described in the following pattern : []map[opt_key]opt_value
	// opt_key :
	// - column : column in the db to be filtered
	// - json_key : dot notation paths of json field in column
	// - direction : sql order direction (asc, desc)
	// - is_required : if true, the sort will not be overridden by the client's own
	Sorts []map[string]any `json:"-" gorm:"-"`
}

// table version, used for migration flag, change the value every time there is a change in the table structure
func (m *Model) TableVersion() string {
	return "init"
}

// table name
func (m *Model) TableName() string {
	return "products"
}

// table alias name
func (m *Model) TableAliasName() string {
	return "p"
}

// move struct tag to m.Fields
func (m *Model) SetFields() {
	m.Fields = map[string]map[string]any{}
	// todo
}

// add (or replace) field to model
func (m *Model) AddField(field_key string, field_opt map[string]any) {
	if m.Fields != nil {
		m.Fields[field_key] = field_opt
	} else {
		m.Fields = map[string]map[string]any{field_key: field_opt}
	}
}

// get model field
func (m *Model) GetFields() map[string]map[string]any {
	m.SetFields()
	return m.Fields
}

// set model relation
func (m *Model) SetRelations() {
	// example :
	// m.AddRelation("pc", map[string]any{
	// 	"table_name": "product_categories",
	// 	"type":       "inner",
	// 	"conditions": []map[string]any{
	// 		{"column_1": "pc.id", "operator": "=", "column_2": "p.category_id"},
	// 	},
	// })
}

// add field to model
func (m *Model) AddRelation(relation_key string, relation_opt map[string]any) {
	if m.Relations != nil {
		m.Relations[relation_key] = relation_opt
	} else {
		m.Relations = map[string]map[string]any{relation_key: relation_opt}
	}
}

// get model relation
func (m *Model) GetRelations() map[string]map[string]any {
	m.SetRelations()
	return m.Relations
}

// set model filter
func (m *Model) SetFilters() {
	// example :
	// m.AddFilter(map[string]any{"column_1": "p.deleted_at", "operator": "=", "value": nil})
}

// add model filter
func (m *Model) AddFilter(filter map[string]any) {
	m.Filters = append(m.Filters, filter)
}

// get model filter
func (m *Model) GetFilters() []map[string]any {
	m.SetFilters()
	return m.Filters
}

// set model sort
func (m *Model) SetSorts() {
	// example :
	// m.AddSort(map[string]any{"column": "p.created_at", "direction": "desc"})
}

// add model sort
func (m *Model) AddSort(sort map[string]any) {
	m.Sorts = append(m.Sorts, sort)
}

// get model sort
func (m *Model) GetSorts() []map[string]any {
	m.SetSorts()
	return m.Sorts
}

func (m *Model) GetQuerySchema() map[string]any {
	res := map[string]any{}
	res["fields"] = m.GetFields()
	res["array_fields"] = m.ArrayFields
	res["relations"] = m.GetRelations()
	res["filters"] = m.GetFilters()
	res["sorts"] = m.GetSorts()
	return res
}

func (m *Model) ToSQL(tx *gorm.DB, q url.Values) string {
	// todo
	return ""
}

func (m *Model) OpenAPISchemaName() string {
	return ""
}

func (m *Model) GetOpenAPISchema() map[string]any {
	// todo
	return map[string]any{}
}
