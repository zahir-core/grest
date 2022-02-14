package db

import "time"

type Schema interface {
	TableVersion() string
	TableName() string
	TableAliasName() string
	SetRelation()
	SetFilter()
	SetSort()
}

// json:
// - dot notation field to be parsed to multi-dimensional json object
// - also used for alias field when query to db
//
// db:
// - field to query to db
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
//
type Model struct {
	Relation []Relation `json:"-" gorm:"-"`
	Filter   []Filter   `json:"-" gorm:"-"`
	Sort     []Sort     `json:"-" gorm:"-"`
}

type Relation struct {
	JoinType          string
	TableName         string
	TableAliasName    string
	SubQuery          Schema
	RelationCondition []Filter
}

type Filter struct {
	Column   string // direct to database
	Operator string
	Value    interface{}
	Column2  string
	JsonKey  string // dot notation paths of json field
}

type Sort struct {
	Column    string
	Direction string
	JsonKey   string
}

func NewRelation(joinType, tableName, tableAliasName string, on []Filter) Relation {
	r := Relation{}
	r.JoinType = joinType
	r.TableName = tableName
	r.TableAliasName = tableAliasName
	r.RelationCondition = on
	return r
}

func NewFilter(column, operator string, value interface{}, opt ...string) Filter {
	f := Filter{}
	f.Column = column
	f.Operator = operator
	f.Value = value
	if len(opt) > 0 {
		f.Column2 = opt[0]
	}
	if len(opt) > 1 {
		f.JsonKey = opt[1]
	}
	return f
}

func NewSort(column, direction string, jsonKey ...string) Sort {
	s := Sort{}
	s.Column = column
	s.Direction = direction
	if len(jsonKey) > 0 {
		s.JsonKey = jsonKey[0]
	}
	return s
}

type SettingTable struct {
	Key       string `gorm:"primaryKey"`
	Value     string
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (SettingTable) TableName() string {
	return "settings"
}

func (SettingTable) KeyField() string {
	return "key"
}

func (SettingTable) ValueField() string {
	return "value"
}
