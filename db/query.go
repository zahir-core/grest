package db

import (
	"fmt"
	"net/url"
	"strings"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// in progress
func First(db *gorm.DB, dest interface{}, query url.Values) error {
	return nil
}

// in progress
func Find(db *gorm.DB, dest interface{}, query url.Values) error {
	return nil
}

// in progress
func PaginationInfo(db *gorm.DB, dest interface{}, query url.Values) (int, int, int, int, error) {
	return 0, 0, 0, 0, nil
}

// WhereExpression implements clause.Expression interface to use as querier
type WhereExpression struct {
	column   string
	operator string
	value    interface{}
}

func Where(column, operator string, value interface{}) *WhereExpression {
	return &WhereExpression{
		column:   column,
		operator: operator,
		value:    value,
	}
}

// Build implements clause.Expression
func (w *WhereExpression) Build(builder clause.Builder) {
	if stmt, ok := builder.(*gorm.Statement); ok {
		builder.WriteString(stmt.Quote(w.column))

		stringVal, stringValOK := w.value.(string)
		isNilValue := w.value == nil || (stringValOK && strings.ToLower(stringVal) == "null")
		if isNilValue {
			if w.operator == "=" {
				w.operator = " IS "
			} else if w.operator == "!=" {
				w.operator = " IS NOT "
			}
		}
		builder.WriteString(w.operator)

		if isNilValue {
			builder.WriteString("NULL")
		} else {
			builder.AddVar(stmt, w.value)
		}
	}
}

// WhereJSONExpression json query expression, implements clause.Expression interface to use as querier
type WhereJSONExpression struct {
	column   string
	jsonKey  string
	operator string
	value    interface{}
}

// WhereJSON query column as json
func WhereJSON(column, jsonKey, operator string, value interface{}) *WhereJSONExpression {
	return &WhereJSONExpression{
		column:   column,
		jsonKey:  jsonKey,
		operator: operator,
		value:    value,
	}
}

// Build implements clause.Expression
func (w *WhereJSONExpression) Build(builder clause.Builder) {
	if stmt, ok := builder.(*gorm.Statement); ok {
		switch stmt.Dialector.Name() {
		case "mysql", "sqlite":
			builder.WriteString("JSON_EXTRACT(" + stmt.Quote(w.column) + ",")
			builder.AddVar(stmt, "$."+w.jsonKey)
			builder.WriteByte(')')
		case "sqlserver":
			builder.WriteString("JSON_VALUE(" + stmt.Quote(w.column) + ",")
			builder.AddVar(stmt, "$."+w.jsonKey)
			builder.WriteByte(')')
		case "postgres":
			builder.WriteString(fmt.Sprintf("json_extract_path_text(%v::json,", stmt.Quote(w.column)))
			keys := strings.Split(w.jsonKey, ".")
			for idx, key := range keys {
				if idx > 0 {
					builder.WriteByte(',')
				}
				stmt.AddVar(builder, key)
			}
			builder.WriteByte(')')
		default:
			// unsupported json query
			builder.WriteString(stmt.Quote(w.column))
		}

		stringVal, stringValOK := w.value.(string)
		isNilValue := w.value == nil || (stringValOK && strings.ToLower(stringVal) == "null")
		if isNilValue {
			if w.operator == "=" {
				w.operator = " IS "
			} else if w.operator == "!=" {
				w.operator = " IS NOT "
			}
		}
		builder.WriteString(w.operator)

		if isNilValue {
			builder.WriteString("NULL")
		} else {
			builder.AddVar(stmt, fmt.Sprintf("%v", w.value))
		}
	}
}

// OrderByExpression implements clause.Expression interface to use as querier
type OrderByExpression struct {
	column    string
	direction string
}

func OrderBy(column, direction string) *OrderByExpression {
	return &OrderByExpression{
		column:    column,
		direction: direction,
	}
}

// Build implements clause.Expression
func (o *OrderByExpression) Build(builder clause.Builder) {
	if stmt, ok := builder.(*gorm.Statement); ok {
		builder.WriteString(stmt.Quote(o.column))
		builder.WriteByte(' ')
		builder.WriteString(o.direction)
	}
}

// OrderByJSONExpression implements clause.Expression interface to use as querier
type OrderByJSONExpression struct {
	column    string
	jsonKey   string
	direction string
}

func OrderByJSON(column, jsonKey, direction string) *OrderByJSONExpression {
	return &OrderByJSONExpression{
		column:    column,
		jsonKey:   jsonKey,
		direction: direction,
	}
}

// Build implements clause.Expression
func (o *OrderByJSONExpression) Build(builder clause.Builder) {
	if stmt, ok := builder.(*gorm.Statement); ok {
		switch stmt.Dialector.Name() {
		case "mysql", "sqlite":
			builder.WriteString("JSON_EXTRACT(" + stmt.Quote(o.column) + ",")
			builder.AddVar(stmt, "$."+o.jsonKey)
			builder.WriteByte(')')
		case "sqlserver":
			builder.WriteString("JSON_VALUE(" + stmt.Quote(o.column) + ",")
			builder.AddVar(stmt, "$."+o.jsonKey)
			builder.WriteByte(')')
		case "postgres":
			builder.WriteString(fmt.Sprintf("json_extract_path_text(%v::json,", stmt.Quote(o.column)))
			keys := strings.Split(o.jsonKey, ".")
			for idx, key := range keys {
				if idx > 0 {
					builder.WriteByte(',')
				}
				stmt.AddVar(builder, key)
			}
			builder.WriteByte(')')
		default:
			// unsupported json query
			builder.WriteString(stmt.Quote(o.column))
		}

		builder.WriteByte(' ')
		builder.WriteString(o.direction)
	}
}
