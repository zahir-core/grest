package db

import (
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"net/url"
	"reflect"
	"strconv"
	"strings"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"grest.dev/grest/convert"
)

func GetOperator(key string) string {
	opt := map[string]string{
		"eq":     "=",
		"ne":     "!=",
		"gt":     ">",
		"gte":    ">=",
		"lt":     "<",
		"lte":    "<=",
		"like":   " like ",
		"ilike":  " like ",
		"nlike":  " not like ",
		"nilike": " not like ",
	}
	res, ok := opt[key]
	if !ok {
		return "="
	}
	return res

}

// in progress
func First(db *gorm.DB, dest interface{}, query url.Values) error {
	if reflect.TypeOf(dest).Kind() != reflect.Ptr {
		return errors.New("dest is not pointer")
	}
	conds := []interface{}{}
	db, conds = QueryBuilder(db, reflect.ValueOf(dest), query, conds...)

	res := map[string]interface{}{}
	db.Take(&res, conds...)
	row := map[string]interface{}{}
	for k, v := range res {
		row[strings.ReplaceAll(k, "__", ".")] = v
	}
	b, err := json.Marshal(row)
	if err != nil {
		return err
	}
	return json.Unmarshal(b, dest)
}

// in progress
func Find(db *gorm.DB, dest interface{}, query url.Values) error {
	if reflect.TypeOf(dest).Kind() != reflect.Ptr {
		return errors.New("dest is not pointer")
	}
	conds := []interface{}{}
	db, conds = QueryBuilder(db, reflect.ValueOf(dest), query, conds...)
	db = SetPagination(db, query)

	res := []map[string]interface{}{}
	db.Find(&res, conds...)
	rows := []map[string]interface{}{}
	for _, r := range res {
		row := map[string]interface{}{}
		for k, v := range r {
			row[strings.ReplaceAll(k, "__", ".")] = v
		}
		rows = append(rows, row)
	}
	b, err := json.Marshal(rows)
	if err != nil {
		return err
	}
	return json.Unmarshal(b, dest)
}

// in progress
func PaginationInfo(db *gorm.DB, dest interface{}, query url.Values) (int64, int64, int64, int64, error) {
	count := int64(0)
	conds := []interface{}{}
	db, conds = QueryBuilder(db, reflect.ValueOf(dest), query, conds...)
	db.Count(&count)

	page := 1
	perPage := 20
	if query.Get("page") != "" {
		pageTemp, _ := strconv.Atoi(query.Get("page"))
		if pageTemp > 0 {
			page = pageTemp
		}
	}
	if query.Get("per_page") != "" {
		perPageTemp, _ := strconv.Atoi(query.Get("per_page"))
		if perPageTemp > 0 {
			perPage = perPageTemp
		}
	}

	return count, int64(page), int64(perPage), int64(math.Ceil(float64(count) / float64(perPage))), nil
}

func QueryBuilder(db *gorm.DB, ptr reflect.Value, query url.Values, conds ...interface{}) (*gorm.DB, []interface{}) {
	if ptr.Elem().Kind() == reflect.Slice {
		ptr = reflect.New(ptr.Elem().Type().Elem())
	}
	db, conds = SetTable(db, ptr, query, conds...)
	db, conds = SetJoin(db, ptr, query, conds...)
	db, conds = SetWhere(db, ptr, query, conds...)
	db, conds = SetSelect(db, ptr, query, conds...)
	return db, conds
}

func SetTable(db *gorm.DB, ptr reflect.Value, query url.Values, conds ...interface{}) (*gorm.DB, []interface{}) {
	t := ptr.Type()
	v := ptr.Elem()

	// todo: from sub query
	tableName := convert.ToSnakeCase(t.Name())
	tn := ptr.MethodByName("TableName").Call([]reflect.Value{})
	if len(tn) == 0 {
		tn = v.MethodByName("TableName").Call([]reflect.Value{})
	}
	if len(tn) > 0 {
		tableName = tn[0].String()
	}

	tableAliasName := convert.ToSnakeCase(t.Name())
	tan := ptr.MethodByName("TableAliasName").Call([]reflect.Value{})
	if len(tan) == 0 {
		tan = v.MethodByName("TableAliasName").Call([]reflect.Value{})
	}
	if len(tan) > 0 {
		tableAliasName = tan[0].String()
	}

	return db.Table(tableName + " as " + tableAliasName), conds
}

func SetJoin(db *gorm.DB, ptr reflect.Value, query url.Values, conds ...interface{}) (*gorm.DB, []interface{}) {
	v := ptr.Elem()
	r, isExist := v.FieldByName("Relation").Interface().([]Relation)
	if !isExist || len(r) == 0 {
		ptr.MethodByName("SetRelation").Call([]reflect.Value{})
	}

	r, isExist = v.FieldByName("Relation").Interface().([]Relation)
	if isExist {
		for _, rel := range r {
			// inner join, left join, right join, full join, cross join
			joinStr := rel.JoinType
			if !strings.HasSuffix(strings.ToLower(joinStr), "join") {
				joinStr += " join"
			}

			// todo: join sub query
			joinStr += " " + rel.TableName

			joinStr += " as " + rel.TableAliasName

			joinConditions := []string{}
			args := []interface{}{}
			for _, rc := range rel.RelationCondition {
				jc := rc.Column
				if rc.Operator != "" {
					jc += rc.Operator
				} else {
					jc += "="
				}
				if rc.Column2 != "" {
					jc += rc.Column2
				} else if rc.Value != nil {
					jc += "?"
					args = append(args, rc.Value)
				}
				joinConditions = append(joinConditions, jc)
			}
			if len(joinConditions) > 0 {
				joinStr += " on " + strings.Join(joinConditions, " and ")
			}

			db = db.Joins(joinStr, args...)
		}
	}
	return db, conds
}

func SetWhere(db *gorm.DB, ptr reflect.Value, query url.Values, conds ...interface{}) (*gorm.DB, []interface{}) {
	v := ptr.Elem()
	f, isExist := v.FieldByName("Filter").Interface().([]Filter)
	if !isExist || len(f) == 0 {
		ptr.MethodByName("SetFilter").Call([]reflect.Value{})
	}

	f, isExist = v.FieldByName("Filter").Interface().([]Filter)
	if isExist {
		for _, w := range f {
			if w.Operator == "" {
				w.Operator = "="
			}
			if w.Column2 != "" {
				db = db.Where(db.Statement.Quote(w.Column) + w.Operator + db.Statement.Quote(w.Column2))
			} else {
				if w.JsonKey == "" {
					if w.Value != nil {
						db = db.Where(db.Statement.Quote(w.Column)+w.Operator+"?", w.Value)
					} else {
						if w.Operator == "=" {
							db = db.Where(db.Statement.Quote(w.Column) + " is null")
						} else {
							db = db.Where(db.Statement.Quote(w.Column) + " is not null")
						}
					}
				} else {
					conds = append(conds, WhereJSON(w.Column, w.JsonKey, w.Operator, w.Value))
				}
			}
		}
	}
	t := v.Type()
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		if field.Name != "Model" && field.Tag.Get("json") != "" && field.Tag.Get("json") != "-" && field.Type.Kind() != reflect.Slice {
			jsonTag := strings.Split(field.Tag.Get("json"), ",")[0]
			dbTag := strings.Split(field.Tag.Get("db"), ",")[0]
			for k, sv := range query {
				key := strings.Split(k, ".$")
				if key[0] == jsonTag {
					for _, val := range sv {
						operator := "="
						if len(key) > 1 {
							operator = GetOperator(key[1])
						}
						colVal := strings.Split(val, "$column:")
						if len(colVal) > 1 {
							db = db.Where(db.Statement.Quote(dbTag) + operator + db.Statement.Quote(colVal[1]))
						} else {
							if val != "null" {
								db = db.Where(db.Statement.Quote(dbTag)+operator+"?", val)
							} else {
								if operator == "=" {
									db = db.Where(db.Statement.Quote(dbTag) + " is null")
								} else {
									db = db.Where(db.Statement.Quote(dbTag) + " is not null")
								}
							}
						}
					}
				}

			}
		}
	}

	return db, conds
}

func SetOrder(db *gorm.DB, ptr reflect.Value, query url.Values, conds ...interface{}) (*gorm.DB, []interface{}) {
	v := ptr.Elem()
	s, isExist := v.FieldByName("Sort").Interface().([]Sort)
	if !isExist || len(s) == 0 {
		ptr.MethodByName("SetSort").Call([]reflect.Value{})
	}

	s, isExist = v.FieldByName("Sort").Interface().([]Sort)
	if isExist {
		for _, o := range s {
			if o.Direction == "" {
				o.Direction = "asc"
			}
			if o.JsonKey == "" {
				db = db.Order(o.Column + " " + o.Direction)
			} else {
				conds = append(conds, OrderByJSON(o.Column, o.JsonKey, o.Direction))
			}
		}
	}
	return db, conds
}

func SetSelect(db *gorm.DB, ptr reflect.Value, query url.Values, conds ...interface{}) (*gorm.DB, []interface{}) {
	v := ptr.Elem()
	t := v.Type()
	fields := []string{}
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		if field.Name != "Model" && field.Tag.Get("json") != "" && field.Tag.Get("json") != "-" && field.Type.Kind() != reflect.Slice {
			alias := strings.ReplaceAll(field.Tag.Get("json"), ".", "__")
			fields = append(fields, db.Statement.Quote(field.Tag.Get("db"))+" as "+alias)
		}
	}
	return db.Select(strings.Join(fields, ",")), conds
}

func SetPagination(db *gorm.DB, query url.Values) *gorm.DB {
	page := 1
	perPage := 20
	if query.Get("page") != "" {
		pageTemp, _ := strconv.Atoi(query.Get("page"))
		if pageTemp > 0 {
			page = pageTemp
		}
	}
	if query.Get("per_page") != "" {
		perPageTemp, _ := strconv.Atoi(query.Get("per_page"))
		if perPageTemp > 0 {
			perPage = perPageTemp
		}
	}
	return db.Limit(perPage).Offset((page - 1) * perPage)
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
