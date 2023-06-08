package grest

import (
	"encoding/json"
	"reflect"
	"strconv"
	"strings"
	"time"
)

type ModelInterface interface {
	TableVersion() string
	TableName() string
	TableSchema() map[string]any
	TableAliasName() string
	SetFields(any)
	AddField(fieldKey string, fieldOpt map[string]any)
	GetFields() map[string]map[string]any
	GetFieldOrder() []string
	AddArrayField(fieldKey string, fieldOpt map[string]any)
	GetArrayFields() map[string]map[string]any
	GetArrayFieldOrder() []string
	AddGroup(fieldKey string, fieldName string)
	GetGroups() map[string]string
	AddRelation(joinType string, tableName any, tableAliasName string, conditions []map[string]any)
	GetRelationOrder() []string
	GetRelations() map[string]map[string]any
	AddFilter(filter map[string]any)
	GetFilters() []map[string]any
	AddSort(sort map[string]any)
	GetSorts() []map[string]any
	SetSchema(ModelInterface) map[string]any
	GetSchema() map[string]any
	OpenAPISchemaName() string
	SetOpenAPISchema(any) map[string]any
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
	Fields     map[string]map[string]any `json:"-" gorm:"-"`
	FieldOrder []string                  `json:"-" gorm:"-"`

	// hold sql group by field data
	Groups map[string]string `json:"-" gorm:"-"`

	// hold array fields schema and filter (based on relation) :
	// ?arrayFields.0.field.id={field_id} > where exists (select 1 from array_table at where at.parent_id = parent.id and field_id = {field_id})
	// ?arrayFields.*.field.id={field_id} > same as above but the array fields response also filtered
	ArrayFields     map[string]map[string]any `json:"-" gorm:"-"`
	ArrayFieldOrder []string                  `json:"-" gorm:"-"`

	// described in the following pattern : map[fieldKey]map[optKey]optValue
	// fieldKey: same as "tableAliasName" on "optKey"
	// optKey :
	// - type : sql join type (inner, left, etc)
	// - tableName : table name to join
	// - tableSchema : schema for dynamic "join sub query" based on client's query params
	// - tableAliasName : table alias name
	// - conditions : []map[string]any same as "Filters"
	Relations     map[string]map[string]any `json:"-" gorm:"-"`
	RelationOrder []string                  `json:"-" gorm:"-"`

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
	// - isCaseInsensitive : if true, the sort will case insensitive
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
			isHide := dbTag == "-" || (len(dbTags) > 1 && dbTags[1] == "hide")
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
				fieldType := field.Type.Name()
				if isNullJSON(field.Type) {
					fieldType = "NullJSON"
				}
				fieldOpt := map[string]any{
					"db":      dbTag,
					"as":      jsonTag,
					"type":    fieldType,
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
	m.FieldOrder = append(m.FieldOrder, fieldKey)
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

func (m *Model) GetFieldOrder() []string {
	return m.FieldOrder
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
	m.ArrayFieldOrder = append(m.ArrayFieldOrder, fieldKey)
}

func (m *Model) GetArrayFields() map[string]map[string]any {
	return m.ArrayFields
}

func (m *Model) GetArrayFieldOrder() []string {
	return m.ArrayFieldOrder
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
func (m *Model) AddRelation(joinType string, tableName any, tableAliasName string, conditions []map[string]any) {
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
	m.RelationOrder = append(m.RelationOrder, tableAliasName)
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

func (m *Model) GetRelationOrder() []string {
	return m.RelationOrder
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
//	isCaseInsensitive : if true, the sort will case insensitive
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
		"tableName":       model.TableName(),
		"tableSchema":     model.TableSchema(),
		"tableAliasName":  model.TableAliasName(),
		"fields":          model.GetFields(),
		"fieldOrder":      model.GetFieldOrder(),
		"arrayFields":     model.GetArrayFields(),
		"arrayFieldOrder": model.GetArrayFieldOrder(),
		"groups":          model.GetGroups(),
		"relations":       model.GetRelations(),
		"relationOrder":   model.GetRelationOrder(),
		"filters":         model.GetFilters(),
		"sorts":           model.GetSorts(),
		"isFlat":          model.IsFlat(),
	}
}

func (m *Model) GetSchema() map[string]any {
	return m.SetSchema(m)
}

func (m *Model) OpenAPISchemaName() string {
	return ""
}

func (m *Model) GetOpenAPISchema() map[string]any {
	return m.SetOpenAPISchema(m)
}

func (m *Model) SetOpenAPISchema(p any) map[string]any {
	openAPISchema := map[string]any{}
	ptr := reflect.ValueOf(p)
	t := ptr.Elem().Type()
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		jsonTag := strings.Split(field.Tag.Get("json"), ",")[0]
		if jsonTag != "" && jsonTag != "-" {
			dbTag := field.Tag.Get("db")
			dbTags := strings.Split(dbTag, ",")
			isHide := dbTag == "-" || (len(dbTags) > 1 && dbTags[1] == "hide")
			if !isHide {
				isArray := field.Type.Kind() == reflect.Slice
				if isArray {
					gqs := m.callMethod(reflect.New(field.Type.Elem()), "GetOpenAPISchema", []reflect.Value{})
					if len(gqs) > 0 {
						openAPISchema[jsonTag] = map[string]any{"type": "array", "items": gqs[0].Interface()}
					}
				} else {
					fieldType := field.Type.Name()
					if isNullJSON(field.Type) {
						fieldType = "NullJSON"
					}
					fieldOpt := m.getJSONSchema(fieldType, field.Tag)
					// todo : if isNullJSON, override type and format based on NullJSON.Data struct
					openAPISchema[jsonTag] = fieldOpt
				}
			}
		}
	}
	flat, ok := p.(interface{ IsFlat() bool })
	if ok {
		if !flat.IsFlat() {
			openAPISchema = m.nestedOpenAPISchema(openAPISchema)
		}
	}
	return map[string]any{
		"type":       "object",
		"properties": openAPISchema,
	}
}

// JSON Schema Spec : https://datatracker.ietf.org/doc/html/draft-bhutton-json-schema-validation-01
//
//	type: string               # null, boolean, string, integer, number, object, array
//	format: string             # int32, int64, float, double, uuid, email, password, date, time, date-time, duration
//	default: string            # default
//	enum: [string]             # enum
//	description: string        # description
//	deprecated: boolean        # true
//	multipleOf: number         # 1
//	maximum: number            # 1
//	exclusiveMaximum: boolean  # true
//	minimum: number            # 1
//	exclusiveMinimum: boolean  # true
//	maxLength: integer         # 1
//	minLength: integer         # 1
//	pattern: string            # pattern
//	maxItems: integer          # 1
//	minItems: integer          # 1
//	uniqueItems: boolean       # true
//	maxProperties: integer     # 1
//	minProperties: integer     # 1
//	required: [string]         # array of field name
//	writeOnly: boolean         # true
//	readOnly: boolean          # true
//	nullable: boolean          # true
//	oneOf: [string]            # oneOf
//	anyOf: [string]            # anyOf
//	allOf: [string]            # allOf
//	examples: [string]         # examples
//
//	NullBool field name :
//		type: boolean
//
//	NullUUID field name :
//		type: string
//		format: uuid
//
//	NullString field name :
//		type: string
//
//	object field name :
//		type: object
//		properties:
//			NullString field name :
//				type: string
//
//	array of string field name :
//		type: array
//		items:
//			type: string
//
//	array of object field name :
//		type: array
//		items:
//			type: object
//			properties:
//				NullString field name :
//					type: string
func (m *Model) getJSONSchema(typeName string, tag reflect.StructTag) map[string]any {
	f := map[string]any{}
	switch typeName {
	case "NullBool":
		f["type"] = "boolean"
	case "NullInt64", "NullUnixTime", "int", "int8", "int16", "int32", "int64", "uint", "uint8", "uint16", "uint32", "uint64", "complex64", "complex128":
		f["type"] = "integer"
	case "NullFloat64", "float32", "float64":
		f["type"] = "number"
	case "NullString", "NullText", "NullDateTime", "NullDate", "NullTime", "NullUUID", "NullJSON", "string":
		f["type"] = "string"
	default:
		f["type"] = "string"
	}

	switch typeName {
	case "NullUnixTime":
		f["example"] = time.Now().Unix()
	case "NullDateTime":
		f["format"] = "date-time"
	case "NullDate":
		f["format"] = "date"
	case "NullTime":
		f["format"] = "time"
	case "NullUUID":
		f["format"] = "uuid"
	}

	if tag.Get("note") != "" {
		f["description"] = tag.Get("note")
	}
	for _, k := range []string{"title", "description", "default", "example"} {
		if tag.Get(k) != "" {
			f[k] = tag.Get(k)
		}
	}
	for _, k := range []string{"deprecated", "exclusiveMaximum", "exclusiveMinimum", "uniqueItems", "writeOnly", "readOnly", "nullable"} {
		if tag.Get(k) != "" {
			f[k] = tag.Get(k) == "true"
		}
	}
	for _, k := range []string{"enum", "oneOf", "anyOf", "allOf"} {
		if tag.Get(k) != "" {
			f[k] = strings.Split(tag.Get(k), ",")
		}
	}
	for _, k := range []string{"multipleOf", "maximum", "minimum", "maxLength", "minLength", "maxItems", "minItems", "maxProperties", "minProperties"} {
		if tag.Get(k) != "" {
			v, err := strconv.ParseInt(tag.Get(k), 10, 64)
			if err == nil {
				f[k] = v
			}
		}
	}
	// parse go-playground validation tag to OAS validation
	if tag.Get("validate") != "" {
		for _, vk := range strings.Split(tag.Get("validate"), ",") {
			k, v, _ := strings.Cut(vk, ",")
			if k == "email" {
				f["format"] = "email"
			}
			if k == "oneof" {
				f["oneOf"] = strings.Split(v, " ")
			}
			if k == "max" {
				max, err := strconv.ParseInt(v, 10, 64)
				if err == nil {
					if f["type"] == "string" {
						f["maxLength"] = max
					} else {
						f["maximum"] = max
					}
				}
			}
			if k == "min" {
				min, err := strconv.ParseInt(v, 10, 64)
				if err == nil {
					if f["type"] == "string" {
						f["minLength"] = min
					} else {
						f["minimum"] = min
					}
				}
			}
		}
	}
	return f
}

func (m Model) nestedOpenAPISchema(flatSchema map[string]any) map[string]any {
	nested := map[string]any{}
	for k, v := range flatSchema {
		keys := strings.Split(k, ".")
		if len(keys) > 1 {
			for i := len(keys) - 1; i >= 1; i-- {
				if i == len(keys)-1 {
					v = map[string]any{
						keys[i]: v,
					}
				} else {
					v = map[string]any{
						keys[i]: map[string]any{
							"type":       "object",
							"properties": v,
						},
					}
				}
			}
			nested[keys[0]] = map[string]any{
				"type":       "object",
				"properties": m.fillOpenAPISchema(nested, keys[0], v),
			}
		} else {
			nested[k] = v
		}
	}
	return nested
}

func (m Model) fillOpenAPISchema(data map[string]any, key string, val any) any {
	d, exist := data[key].(map[string]any)
	if exist {
		p, ok := d["properties"].(map[string]any)
		if ok {
			temp, _ := val.(map[string]any)
			for k, v := range temp {
				p[k] = m.fillOpenAPISchema(p, k, v)
			}
			return p
		}
	}
	return val
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
