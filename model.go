package grest

import (
	"encoding/json"
	"reflect"
	"strings"
)

type ModelInterface interface {
	TableVersion() string
	TableName() string
	TableSchema() map[string]any
	TableAliasName() string
	SetFields(any)
	AddField(fieldKey string, fieldOpt map[string]any)
	GetFields() map[string]map[string]any
	AddArrayField(fieldKey string, fieldOpt map[string]any)
	GetArrayFields() map[string]map[string]any
	AddGroup(fieldKey string, fieldName string)
	GetGroups() map[string]string
	AddRelation(joinType string, tableName any, tableAliasName string, conditions []map[string]any) map[string]any
	GetRelations() map[string]map[string]any
	AddFilter(filter map[string]any)
	GetFilters() []map[string]any
	AddSort(sort map[string]any)
	GetSorts() []map[string]any
	SetSchema(ModelInterface) map[string]any
	GetSchema() map[string]any
	OpenAPISchemaName() string
	SetOpenAPISchema(ModelInterface) map[string]any
	GetOpenAPISchema() map[string]any
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
	// described in the following pattern : map[fieldKey]map[optKey]optValue
	// fieldKey: the field shown on json
	// optKey, same as model struct tag + the following key :
	// - dataType
	// optValue, same as model struct tag value
	Fields map[string]map[string]any `json:"-" gorm:"-"`

	// hold sql group by field data
	Groups map[string]string `json:"-" gorm:"-"`

	// hold array fields schema and filter (based on relation) :
	// ?arrayFields.0.field.id={field_id} > where exists (select 1 from array_table at where at.parent_id = parent.id and field_id = {field_id})
	// ?arrayFields.*.field.id={field_id} > same as above but the array fields response also filtered
	ArrayFields map[string]map[string]any `json:"-" gorm:"-"`

	// described in the following pattern : map[fieldKey]map[optKey]optValue
	// fieldKey: same as "tableAliasName" on "optKey"
	// optKey :
	// - type : sql join type (inner, left, etc)
	// - tableName : table name to join
	// - tableSchema : schema for dynamic "join sub query" based on client's query params
	// - tableAliasName : table alias name
	// - conditions : []map[string]any same as "Filters"
	Relations map[string]map[string]any `json:"-" gorm:"-"`

	// described in the following pattern : []map[optKey]optValue
	// optKey :
	// - column1 : column in the db to be filtered
	// - column1jsonKey : dot notation paths of json field in column1
	// - operator : sql operator (=, !=, >, >=, <, <=, like, not like, in, not in, etc)
	// - column2 : another column in the db (to compare values between columns in the db)
	// - column2jsonKey : dot notation paths of json field in column2
	// - value : desired value to be filtered
	Filters []map[string]any `json:"-" gorm:"-"`

	// described in the following pattern : []map[optKey]optValue
	// optKey :
	// - column : column in the db to be filtered
	// - jsonKey : dot notation paths of json field in column
	// - direction : sql order direction (asc, desc)
	// - isRequired : if true, the sort will not be overridden by the client's own
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

// you can set `TableSchema` if you need "from sub query" with dynamic query based on client's query params
//
// example "from sub query" :
//
//	SELECT
//		"u"."id" "user.id",
//		"u"."name" "user.name",
//		"ur"."total_review" "total_review"
//	FROM
//	  (SELECT
//	    "user_id" "user_id",
//	    COUNT("user_id") "total_review"
//	  FROM
//	    "user_reviews"
//	  WHERE
//	    "rate" >= 4
//	  GROUP BY
//	    "user_id"
//	  ) as "ur"
//	  JOIN "users" "u" on "u"."id" = "ur"."user_id"
//
// for example :
//
//	func (m *Model) TableSchema() map[string]any {
//		ur := &UserReviewTotal{}
//		return ur.GetSchema()
//	}
func (m *Model) TableSchema() map[string]any {
	return nil
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
		jsonTag := strings.Split(field.Tag.Get("json"), ",")[0]
		if jsonTag != "" && jsonTag != "-" {
			dbTag := field.Tag.Get("db")
			dbTags := strings.Split(dbTag, ",")
			isHide := len(dbTags) > 1 && dbTags[1] == "hide"
			isGroup := len(dbTags) > 1 && dbTags[1] == "group"
			if isGroup {
				m.AddGroup(jsonTag, dbTags[0])
			}
			if isHide || isGroup {
				dbTag = dbTags[0]
			}
			isArray := field.Type.Kind() == reflect.Slice
			if isArray {
				gqs := m.callMethod(reflect.New(field.Type.Elem()), "GetSchema", []reflect.Value{})
				if len(gqs) > 0 {
					arraySchemaTemp := gqs[0].Interface()
					arraySchema, _ := arraySchemaTemp.(map[string]any)
					m.AddArrayField(jsonTag, map[string]any{"schema": arraySchema, "filter": dbTag})
				}
			} else {
				fieldOpt := map[string]any{
					"db":      dbTag,
					"as":      jsonTag,
					"type":    field.Type.Name(),
					"isHide":  isHide,
					"isGroup": isGroup,
				}
				if field.Tag.Get("gorm") != "" {
					fieldOpt["gorm"] = field.Tag.Get("gorm")
				}
				if field.Tag.Get("validate") != "" {
					fieldOpt["validate"] = field.Tag.Get("validate")
				}
				if field.Tag.Get("title") != "" {
					fieldOpt["title"] = field.Tag.Get("title")
				}
				if field.Tag.Get("note") != "" {
					fieldOpt["note"] = field.Tag.Get("note")
				}
				if field.Tag.Get("default") != "" {
					fieldOpt["default"] = field.Tag.Get("default")
				}
				if field.Tag.Get("example") != "" {
					fieldOpt["example"] = field.Tag.Get("example")
				}
				m.AddField(jsonTag, fieldOpt)
			}
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

// add field to model if not exists
func (m *Model) AddField(fieldKey string, fieldOpt map[string]any) {
	if m.Fields != nil {
		if m.Fields[fieldKey] == nil {
			m.Fields[fieldKey] = fieldOpt
		}
	} else {
		m.Fields = map[string]map[string]any{fieldKey: fieldOpt}
	}
}

// get model field, use SetFields to automatically setted by struct tag using SetFields, but you can add or override this. expected key :
//
//	column : column to filter, based on field in the db (or raw query)
//	as : column to filter, based on field in the db (or raw query)
//	column2 : another column to filter, based on field in the db (or raw query), used to compare 2 column
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

// add array field to model if not exists
func (m *Model) AddArrayField(fieldKey string, fieldOpt map[string]any) {
	if m.ArrayFields != nil {
		if m.ArrayFields[fieldKey] == nil {
			m.ArrayFields[fieldKey] = fieldOpt
		}
	} else {
		m.ArrayFields = map[string]map[string]any{fieldKey: fieldOpt}
	}
}

func (m *Model) GetArrayFields() map[string]map[string]any {
	return m.ArrayFields
}

// add group to model if not exists
func (m *Model) AddGroup(fieldKey string, fieldName string) {
	if m.Groups != nil {
		if m.Groups[fieldKey] == "" {
			m.Groups[fieldKey] = fieldName
		}
	} else {
		m.Groups = map[string]string{fieldKey: fieldName}
	}
}

func (m *Model) GetGroups() map[string]string {
	return m.Groups
}

// add relation to model
//
//	joinType : sql join type (inner, left, etc)
//	tableName : column in the db (or raw subquery) to be joined or model schema (to auto generate sub query filtered based on client's filter)
//	tableAliasName : table alias name on sql join, also used as relation key
//	conditions : []map[string]any same as "Filters"
func (m *Model) AddRelation(joinType string, tableName any, tableAliasName string, conditions []map[string]any) map[string]any {
	relation := map[string]any{
		"tableName":      tableName,
		"tableAliasName": tableAliasName,
		"type":           joinType,
		"conditions":     conditions,
	}
	tableSchema, isSchema := tableName.(map[string]any)
	if isSchema {
		relation["tableSchema"] = tableSchema
	}
	if m.Relations != nil {
		m.Relations[tableAliasName] = relation
	} else {
		m.Relations = map[string]map[string]any{tableAliasName: relation}
	}
	return relation
}

// get model relation, expected key :
//
//	type : sql join type (inner, left, etc)
//	tableName :
//	tableAliasName
//	conditions : []map[string]any same as "Filters"
//
// example :
//
//	func (m *Model) GetRelations() map[string]map[string]any {
//		m.AddRelation("left", "product_categories", "pc", []map[string]any{{"column1": "pc.id", "operator": "=", "column2": "p.category_id"}})
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
//	column1 : column (or raw query) in the db to be filtered
//	column1jsonKey : dot notation paths of json field in column1
//	operator : sql operator (=, !=, >, >=, <, <=, like, not like, in, not in, etc)
//	column2 : another column (or raw query) in the db (to compare values between columns in the db)
//	column2jsonKey : dot notation paths of json field in column2
//	value : desired value to be filtered
//
// example :
//
//	func (m *Model) GetFilters() []map[string]any {
//		m.AddFilter(map[string]any{"column1": "p.deleted_at", "operator": "=", "value": nil})
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
//	jsonKey : dot notation paths of json field in column
//	direction : sql order direction (asc, desc)
//	isRequired : if true, the sort will not be overridden by the client's own
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

func (m *Model) SetSchema(model ModelInterface) map[string]any {
	return map[string]any{
		"tableName":      model.TableName(),
		"tableSchema":    model.TableSchema(),
		"tableAliasName": model.TableAliasName(),
		"fields":         model.GetFields(),
		"arrayFields":    model.GetArrayFields(),
		"groups":         model.GetGroups(),
		"relations":      model.GetRelations(),
		"filters":        model.GetFilters(),
		"sorts":          model.GetSorts(),
		"isFlat":         model.IsFlat(),
	}
}

func (m *Model) GetSchema() map[string]any {
	return m.SetSchema(m)
}

func (m *Model) OpenAPISchemaName() string {
	return ""
}

// todo
func (m *Model) SetOpenAPISchema(model ModelInterface) map[string]any {
	return map[string]any{}
}

func (m *Model) GetOpenAPISchema() map[string]any {
	return m.SetOpenAPISchema(m)
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
