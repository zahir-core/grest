package grest

import (
	"encoding/json"
	"reflect"
	"strings"
)

type ModelInterface interface {
	TableVersion() string
	TableName() string
	TableAliasName() string
	SetFields(any)
	GetArrayFields() map[string]map[string]any
	GetGroups() map[string]string
	AddField(field_key string, field_opt map[string]any)
	GetFields() map[string]map[string]any
	AddRelation(join_type string, table_name any, table_alias_name string, conditions []map[string]any) map[string]any
	GetRelations() map[string]map[string]any
	AddFilter(filter map[string]any)
	GetFilters() []map[string]any
	AddSort(sort map[string]any)
	GetSorts() []map[string]any
	GetSchema() map[string]any
	SetSchema(ModelInterface) map[string]any
	SetOpenAPISchema(ModelInterface) map[string]any
	IsFlat() bool
}

// model struct tag :
//
// json :
//   - dot notation field to be parsed to multi-dimensional json object
//   - also used for alias field when query to db
//
// db :
//   - field to query to db
//   - add ",group" to group the field
//   - add ",hide" to hide the field on api response
//
// gorm :
//   - field option for main table of the struct
//   - the details can be found at : https://gorm.io/docs/models.html#Fields-Tags
//   - if the field is not field of main table of the struct in the database, gorm must be setted to "-"
//
// validate :
//   - validation tag to validate the data
//   - "required" tag will be used as "required" on OpenAPI Specification
//   - "oneof" tag will be used as "enum" on OpenAPI Specification
//   - "max" tag will be used as "maximum" or "maxLength" on OpenAPI Specification based on value type
//   - "min" tag will be used as "minimum" or "minLength" on OpenAPI Specification based on value type
//   - the details can be found at : https://pkg.go.dev/github.com/go-playground/validator/v10
//
// title :
//   - used as "title" on OpenAPI Specification
//
// note :
//   - used as "description" on OpenAPI Specification
//
// default :
//   - used for the default value when insert to db
//   - used as "default" on OpenAPI Specification
//
// example :
//   - used as "example" on OpenAPI Specification
type Model struct {
	// described in the following pattern : map[field_key]map[opt_key]opt_value
	// field_key: the field shown on json
	// opt_key, same as model struct tag + the following key :
	// - data_type
	// opt_value, same as model struct tag value
	Fields map[string]map[string]any `json:"-" gorm:"-"`

	// hold sql group by field data
	Groups map[string]string `json:"-" gorm:"-"`

	// hold array fields schema and filter (based on relation) :
	// ?array_fields.0.field.id={field_id} > where exists (select 1 from array_table at where at.parent_id = parent.id and field_id = {field_id})
	// ?array_fields.*.field.id={field_id} > same as above but the array fields response also filtered
	ArrayFields map[string]map[string]any `json:"-" gorm:"-"`

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
func (m *Model) SetFields(p any) {
	ptr := reflect.ValueOf(p)
	t := ptr.Elem().Type()
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		jsonTag := strings.Split(field.Tag.Get("json"), ",")
		if jsonTag[0] != "" && jsonTag[0] != "-" {
			dbTag := strings.Split(field.Tag.Get("db"), ",")
			isGroup := len(dbTag) > 1 && dbTag[1] == "group"
			if isGroup {
				m.AddGroup(jsonTag[0], dbTag[0])
			}
			isArray := field.Type.Kind() == reflect.Slice
			if isArray {
				gqs := m.callMethod(reflect.New(field.Type.Elem()), "GetSchema", []reflect.Value{})
				if len(gqs) > 0 {
					arraySchemaTemp := gqs[0].Interface()
					arraySchema, _ := arraySchemaTemp.(map[string]any)
					m.AddArrayField(jsonTag[0], map[string]any{"schema": arraySchema, "filter": dbTag[0]})
				}
			}
			m.AddField(jsonTag[0], map[string]any{
				"db":       dbTag[0],
				"as":       jsonTag[0],
				"gorm":     field.Tag.Get("gorm"),
				"validate": field.Tag.Get("validate"),
				"title":    field.Tag.Get("title"),
				"note":     field.Tag.Get("note"),
				"default":  field.Tag.Get("default"),
				"example":  field.Tag.Get("example"),
				"type":     field.Type.Name(),
				"is_hide":  len(dbTag) > 1 && dbTag[1] == "hide",
				"is_group": isGroup,
				"is_array": isArray,
			})
		}
	}
}

func (m *Model) callMethod(ptr reflect.Value, methodName string, args []reflect.Value) []reflect.Value {
	val := []reflect.Value{}
	if m := ptr.Elem().MethodByName(methodName); m.IsValid() {
		val = m.Call(args)
	}
	if len(val) == 0 {
		if m := ptr.MethodByName(methodName); m.IsValid() {
			val = m.Call(args)
		}
	}
	return val
}

// add (or replace) field to model
func (m *Model) AddField(field_key string, field_opt map[string]any) {
	if m.Fields != nil {
		m.Fields[field_key] = field_opt
	} else {
		m.Fields = map[string]map[string]any{field_key: field_opt}
	}
}

// add (or replace) field to model
func (m *Model) AddArrayField(field_key string, field_opt map[string]any) {
	if m.ArrayFields != nil {
		m.ArrayFields[field_key] = field_opt
	} else {
		m.ArrayFields = map[string]map[string]any{field_key: field_opt}
	}
}

// add (or replace) field to model
func (m *Model) AddGroup(field_key string, field_name string) {
	if m.Groups != nil {
		m.Groups[field_key] = field_name
	} else {
		m.Groups = map[string]string{field_key: field_name}
	}
}

// get model field, by default it automatically setted by struct tag using SetFields, but you can add or override this. expected key :
//
//	column : column to filter, based on field in the db (or raw query)
//	as : column to filter, based on field in the db (or raw query)
//	column_2 : another column to filter, based on field in the db (or raw query), used to compare 2 column
//	operator : operator to compare, if not set the default is "="
//	value : value to compare
//
// example :
//
//	func (m *Model) GetFields() map[string]map[string]any {
//		m.SetFields(m)
//		m.AddField(map[string]any{"column": "lower(p.name)", "as": "name_lower", "type": NullString})
//	}
func (m *Model) GetFields() map[string]map[string]any {
	m.SetFields(m)
	return m.Fields
}

// add relation to model
//
//	join_type : sql join type (inner, left, etc)
//	table_name : column in the db (or raw subquery) to be joined or model schema (to auto generate sub query filtered based on client's filter)
//	table_alias_name : table alias name on sql join, also used as relation key
//	conditions : []map[string]any same as "Filters"
func (m *Model) AddRelation(join_type string, table_name any, table_alias_name string, conditions []map[string]any) map[string]any {
	relation := map[string]any{
		"table_name":       table_name,
		"table_alias_name": table_alias_name,
		"type":             join_type,
		"conditions":       conditions,
	}
	table_schema, is_schema := table_name.(map[string]any)
	if is_schema {
		relation["table_schema"] = table_schema
	}
	if m.Relations != nil {
		m.Relations[table_alias_name] = relation
	} else {
		m.Relations = map[string]map[string]any{table_alias_name: relation}
	}
	return relation
}

// get model relation, expected key :
//
//	type : sql join type (inner, left, etc)
//	table_name :
//	table_alias_name
//	conditions : []map[string]any same as "Filters"
//
// example :
//
//	func (m *Model) GetRelations() map[string]map[string]any {
//		m.AddRelation("left", "product_categories", "pc", []map[string]any{{"column_1": "pc.id", "operator": "=", "column_2": "p.category_id"}})
//		return m.Relations
//	}
func (m *Model) GetRelations() map[string]map[string]any {
	return m.Relations
}

// add model filter
func (m *Model) AddFilter(filter map[string]any) {
	m.Filters = append(m.Filters, filter)
}

// get model filter, expected key :
//
//	column_1 : column (or raw query) in the db to be filtered
//	column_1_json_key : dot notation paths of json field in column_1
//	operator : sql operator (=, !=, >, >=, <, <=, like, not like, in, not in, etc)
//	column_2 : another column (or raw query) in the db (to compare values between columns in the db)
//	column_2_json_key : dot notation paths of json field in column_2
//	value : desired value to be filtered
//
// example :
//
//	func (m *Model) GetFilters() []map[string]any {
//		m.AddFilter(map[string]any{"column_1": "p.deleted_at", "operator": "=", "value": nil})
//		return m.Filters
//	}
func (m *Model) GetFilters() []map[string]any {
	return m.Filters
}

// add model sort
func (m *Model) AddSort(sort map[string]any) {
	m.Sorts = append(m.Sorts, sort)
}

// get model sort, expected key :
//
//	column : column in the db to be filtered
//	json_key : dot notation paths of json field in column
//	direction : sql order direction (asc, desc)
//	is_required : if true, the sort will not be overridden by the client's own
//
// example :
//
//	func (m *Model) GetSorts() []map[string]any {
//		m.AddSort(map[string]any{"column": "p.created_at", "direction": "desc"})
//		return m.Sorts
//	}
func (m *Model) GetSorts() []map[string]any {
	return m.Sorts
}

func (m *Model) GetArrayFields() map[string]map[string]any {
	return m.ArrayFields
}

func (m *Model) GetGroups() map[string]string {
	return m.Groups
}

func (m *Model) SetSchema(model ModelInterface) map[string]any {
	return map[string]any{
		"fields":       model.GetFields(),
		"array_fields": model.GetArrayFields(),
		"groups":       model.GetGroups(),
		"relations":    model.GetRelations(),
		"filters":      model.GetFilters(),
		"sorts":        model.GetSorts(),
		"is_flat":      model.IsFlat(),
	}
}

func (m *Model) GetSchema() map[string]any {
	return m.SetSchema(m)
}

// todo
func (m *Model) SetOpenAPISchema(model ModelInterface) map[string]any {
	return map[string]any{}
}

func (m *Model) IsFlat() bool {
	return false
}

// bind json byte to struct
func (m *Model) Bind(data []byte) error {
	if m.IsFlat() {
		return json.Unmarshal(data, m)
	}
	return NewJSON(data).ToFlat().Unmarshal(m)
}

// get data after query to db
func (m *Model) GetData() any {
	if m.IsFlat() {
		return m
	}
	return NewJSON(m).ToStructured().Data
}

func (m *Model) OpenAPISchemaName() string {
	return ""
}

func (m *Model) OpenAPISchema(schema map[string]any) map[string]any {
	// todo
	return map[string]any{}
}
