package db

import (
	"encoding/json"
	"fmt"
	"math"
	"net/url"
	"reflect"
	"strconv"
	"strings"

	"gorm.io/gorm"

	"grest.dev/grest/convert"
)

// GREST support a common way for pagination, selecting fields, filtering, sorting, searching and other using URL query params
// this is the default of query parameter setting, you can override this with your own preferences
var (
	// pagination query params setting
	// if QueryLimit not setted, DefaultLimit will be applied
	// if QueryLimit setted to 0, only PaginationInfo will be executed (only run a counting query against the database).
	// if QueryLimit setted greater than MaxLimit, DefaultLimit will be applied, Set DefaultLimit to 0 to allow unlimited QueryLimit.
	// if QueryDisablePagination setted to true, MaxLimit will be ignored and PaginationInfo will not be executed
	DefaultLimit           = 10                       // Sets the default number of items when $per_page is not set
	MaxLimit               = 100                      // Sets the maximum allowed number of items per page (even if the QueryLimit query parameter is set higher)
	QueryLimit             = "$per_page"              // ex: /contacts?$per_page=20                  => sql: select * from contacts limit 20
	QueryOffset            = "$offset"                // ex: /contacts?$offset=20                    => sql: select * from contacts offset 20
	QueryPage              = "$page"                  // ex: /contacts?$page=3&per_page=10           => sql: select * from contacts limit 10 offset 20
	QueryDisablePagination = "$is_disable_pagination" // ex: /contacts?$is_disable_pagination=true   => sql: select * from contacts

	// selection query params setting
	// it can be setted by multiple fields, separated by comma
	// ex: /contacts?$select=id,code,name    => sql: select id, code, name from contacts
	QuerySelect = "$select"

	// filtering query params setting
	QueryOptEqual              = "$eq"     // ex: /contacts?gender.$eq=male            => sql: select * from contacts where gender = 'male'                       => same with /contacts?gender=male
	QueryOptNotEqual           = "$ne"     // ex: /contacts?phone.$ne=null             => sql: select * from contacts where phone is not null
	QueryOptGreaterThan        = "$gt"     // ex: /contacts?age.$gt=18                 => sql: select * from contacts where age > 18
	QueryOptGreaterThanOrEqual = "$gte"    // ex: /contacts?age.$gte=21                => sql: select * from contacts where age >= 21
	QueryOptLowerThan          = "$lt"     // ex: /contacts?age.$lt=17                 => sql: select * from contacts where age < 17
	QueryOptLowerThanOrEqual   = "$lte"    // ex: /contacts?age.$lte=15                => sql: select * from contacts where age <= 15
	QueryOptLike               = "$like"   // ex: /contacts?name.$like=john%           => sql: select * from contacts where name like 'john%'
	QueryOptNotLike            = "$nlike"  // ex: /contacts?name.$nlike=john%          => sql: select * from contacts where name not like 'john%'
	QueryOptInsensitiveLike    = "$ilike"  // ex: /contacts?name.$ilike=john%          => sql: select * from contacts where lower(name) like lower('john%')
	QueryOptInsensitiveNotLike = "$nilike" // ex: /contacts?name.$nilike=john%         => sql: select * from contacts where lower(name) not like lower('john%')
	QueryOptIn                 = "$in"     // ex: /contacts?age.$in=17,21,34           => sql: select * from contacts where age in (17,21,34)
	QueryOptNotIn              = "$nin"    // ex: /contacts?age.$nin=17,21,34          => sql: select * from contacts where age not in (17,21,34)
	QuerySearch                = "$search" // ex: /contacts?$search=code,name=john     => sql: select * from contacts where (lower(code) = lower('john') or lower(name) = lower('john'))

	// sorting query params setting
	// default is ascending
	// it can be setted by multiple fields, separated by comma
	// add prefix - to sort descending
	// add sufix :i to sort case insensitive
	// ex: /contacts?$sort=gender,-age,-name:i   => sql: select * from contacts order by gender, age desc, lower(name) desc
	QuerySort = "$sort"

	// ===== Advance Query Params =====
	// it combined by another query params

	// or query params setting
	// ex: /contacts?$or=gender=female||age.$lt=10&$or=is_salesman=true||is_employee=true  => sql: select * from contacts where (gender = 'female' or age < 10) and (is_salesman = '1' or is_employee = '1')
	QueryOr          = "$or"
	QueryOrDelimiter = "||"

	// field query params setting
	// useful for filter, select or sort using another field
	// ex: /products?qty_available=$field:qty_on_hand          => sql: select * from products where qty_available = qty_on_hand
	// ex: /products?qty_on_order.$gt=$field:qty_available     => sql: select * from products where qty_on_order > qty_available
	QueryField = "$field"

	// aggregation query params
	// ex: /products?$select=$count:id                         => sql: select count(id) as "count_id" from products
	// ex: /products?$select=$sum:sold                         => sql: select sum(sold) as "sum_sold" from products
	// ex: /products?$select=$min:sold                         => sql: select min(sold) as "min_sold" from products
	// ex: /products?$select=$max:sold                         => sql: select max(sold) as "max_sold" from products
	// ex: /products?$select=$avg:sold                         => sql: select avg(sold) as "avg_sold" from products
	QueryCount = "$count"
	QuerySum   = "$sum"
	QueryMin   = "$min"
	QueryMax   = "$max"
	QueryAvg   = "$avg"

	// grouping query params setting
	// if QueryGroup is setted, QuerySelect is required
	// ex: /products?$group=$field:category_id&$select=category_id                             => sql: select category_id from products group by category_id
	// ex: /products?$group=$field:category_id&$select=category_id,$avg:sold                   => sql: select category_id, avg(sold) as "avg_sold" from products group by category_id
	// ex: /products?$group=$field:category_id&$select=category_id,$sum:sold&$sum:sold.$gt=0   => sql: select category_id, sum(sold) as "sum_sold" from products group by category_id having sum(sold) > 0
	// ex: /products?$group=$field:category_id&$select=category_id,$sum:sold&$sort:-$sum:sold  => sql: select category_id, sum(sold) as "sum_sold" from products group by category_id order by sum(sold) desc
	QueryGroup = "$group"

	// include query params setting
	// for First method, by default query for all array fields is executed
	// but for Find method, by default query for array fields (has many or many to many) is not executed for optimum performance
	// to execute array fields query on Find method, you can add using QueryInclude
	// it can be setted by multiple fields, separated by comma
	// if QueryInclude setted to all, query for all array fields is executed
	// ex: /contacts?$include=families,friends,phones            => include array fields: families, friends, and phones
	// ex: /contacts?$include=all                                => include all array fields
	// ex: /contacts/{id}                                        => same as /contacts?id={id}&$include=all
	QueryInclude = "$include"
	QueryDbField = "$db_field"
)

type queryResult struct {
	Dest  interface{}              // pointer of struct or slice
	Row   map[string]interface{}   // first result
	Rows  []map[string]interface{} // find result
	Error error                    // error
}

func (q *queryResult) Marshal() ([]byte, error) {
	if q.Row != nil {
		return json.Marshal(q.Row)
	} else if q.Rows != nil {
		return json.Marshal(q.Rows)
	}
	return []byte{}, q.Error
}

func (q *queryResult) Unmarshal(v ...interface{}) error {
	if q.Error != nil {
		return q.Error
	}
	b, err := q.Marshal()
	if err != nil {
		return err
	}
	dest := q.Dest
	if len(v) > 0 {
		dest = v[0]
	}
	return json.Unmarshal(b, dest)
}

func First(db *gorm.DB, dest interface{}, query url.Values) *queryResult {
	if reflect.TypeOf(dest).Kind() != reflect.Ptr {
		return &queryResult{Error: gorm.ErrInvalidValue}
	}
	query.Add(QueryLimit, "1")
	query.Add(QueryInclude, "all")
	rows := FindRows(db, reflect.ValueOf(dest), query)
	if len(rows) > 0 {
		return &queryResult{Dest: dest, Row: rows[0]}
	}
	return &queryResult{Error: gorm.ErrRecordNotFound}
}

func Find(db *gorm.DB, dest interface{}, query url.Values) *queryResult {
	if reflect.TypeOf(dest).Kind() != reflect.Ptr {
		return &queryResult{Error: gorm.ErrInvalidValue}
	}
	rows := FindRows(db, reflect.ValueOf(dest), query)
	if len(rows) > 0 {
		return &queryResult{Dest: dest, Rows: rows}
	}
	return &queryResult{Error: gorm.ErrRecordNotFound, Rows: rows}
}

func PaginationInfo(db *gorm.DB, dest interface{}, query url.Values) (int64, int64, int64, int64, error) {
	count := int64(0)
	db = QueryBuilder(db, reflect.ValueOf(dest), query)
	db.Count(&count)
	page, limit := GetPaginationQuery(query)
	return count, int64(page), int64(limit), int64(math.Ceil(float64(count) / float64(limit))), nil
}

func FindRows(db *gorm.DB, rv reflect.Value, query url.Values) []map[string]interface{} {
	rows := []map[string]interface{}{}
	tx := db.Session(&gorm.Session{})
	tx = QueryBuilder(tx, rv, query)
	tx = SetOrder(tx, rv, query)
	tx = SetPagination(tx, query)
	tx.Find(&rows)
	for i, v := range rows {
		rows[i] = IncludeArray(db, FixDataType(v, rv), rv, query)
	}
	return rows
}

func QueryBuilder(db *gorm.DB, ptr reflect.Value, query url.Values) *gorm.DB {
	if ptr.Elem().Kind() == reflect.Slice {
		ptr = reflect.New(ptr.Elem().Type().Elem())
	}
	db = SetTable(db, ptr, query)
	db = SetJoin(db, ptr, query)
	db = SetWhere(db, ptr, query)
	db = SetSelect(db, ptr, query)
	return db
}

func FixDataType(data map[string]interface{}, ptr reflect.Value) map[string]interface{} {
	if ptr.Elem().Kind() == reflect.Slice {
		ptr = reflect.New(ptr.Elem().Type().Elem())
	}
	v := ptr.Elem()
	t := v.Type()
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		if field.Type.Name() == "NullBool" {
			for key, val := range data {
				if field.Tag.Get("json") == key {
					b := NullBool{}
					b.Scan(val)
					data[key] = b
				}
			}
		}
	}
	return data
}

func IncludeArray(db *gorm.DB, data map[string]interface{}, ptr reflect.Value, query url.Values) map[string]interface{} {
	if ptr.Elem().Kind() == reflect.Slice {
		ptr = reflect.New(ptr.Elem().Type().Elem())
	}
	v := ptr.Elem()
	t := v.Type()
	includes := strings.Split(query.Get(QueryInclude), ",")
	isIncludeAll := len(includes) > 0 && includes[0] == "all"
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		if field.Name != "Model" && field.Tag.Get("json") != "" && field.Tag.Get("json") != "-" && field.Type.Kind() == reflect.Slice {
			jsonTag := strings.Split(field.Tag.Get("json"), ",")[0]
			dbTag := strings.Split(field.Tag.Get("db"), ",")[0]
			rel := strings.Split(dbTag, "=")
			if len(rel) > 1 && (isIncludeAll || inArray(includes, jsonTag)) {
				if val, isExist := data[rel[1]]; isExist {
					if valString, isOk := val.(string); isOk {
						q := url.Values{}
						q.Add(QueryDbField+"."+rel[0], valString)
						q.Add(QueryInclude, "all")
						q.Add(QueryDisablePagination, "true")
						data[jsonTag] = FindRows(db, reflect.New(field.Type.Elem()), q)
					}
				}
			} else {
				data[jsonTag] = []map[string]interface{}{}
			}
		}
	}
	return data
}

func inArray(needle []string, haystack string) bool {
	for _, v := range needle {
		if v == haystack {
			return true
		}
	}
	return false
}

func SetTable(db *gorm.DB, ptr reflect.Value, query url.Values) *gorm.DB {
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

	return db.Table(Quote(db, tableName) + " as " + Quote(db, tableAliasName))
}

func SetJoin(db *gorm.DB, ptr reflect.Value, query url.Values) *gorm.DB {
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
			joinStr += " " + Quote(db, rel.TableName)

			joinStr += " as " + Quote(db, rel.TableAliasName)

			joinConditions := []string{}
			args := []interface{}{}
			for _, rc := range rel.RelationCondition {
				jc := db.Statement.Quote(rc.Column)
				if rc.Operator != "" {
					jc += rc.Operator
				} else {
					jc += "="
				}
				if rc.Column2 != "" {
					jc += db.Statement.Quote(rc.Column2)
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
	return db
}

func SetWhere(db *gorm.DB, ptr reflect.Value, query url.Values) *gorm.DB {
	v := ptr.Elem()
	f, isExist := v.FieldByName("Filter").Interface().([]Filter)
	if !isExist || len(f) == 0 {
		setFilter := ptr.MethodByName("SetFilter")
		if setFilter.IsValid() {
			setFilter.Call([]reflect.Value{})
		}
	}

	f, isExist = v.FieldByName("Filter").Interface().([]Filter)
	if isExist {
		for _, w := range f {
			column := w.Column
			if w.JsonKey == "" {
				column = db.Statement.Quote(column)
			} else {
				column = QuoteJSON(db, column, w.JsonKey)
			}
			if w.Operator == "" {
				w.Operator = "="
			}
			if w.Column2 != "" {
				db = db.Where(column + w.Operator + db.Statement.Quote(w.Column2))
			} else if w.Value != nil {
				db = db.Where(column+w.Operator+"?", w.Value)
			} else if w.Operator == "=" {
				db = db.Where(column + " is null")
			} else {
				db = db.Where(column + " is not null")
			}
		}
	}
	qs := strings.Split(query.Get(QuerySearch), "=")
	if len(qs) > 1 {
		valSearch := ""
		for i, s := range strings.Split(qs[0], ",") {
			if i == 0 {
				valSearch += s + "." + QueryOptInsensitiveLike + "=" + qs[1]
			} else {
				valSearch += QueryOrDelimiter + s + "." + QueryOptInsensitiveLike + "=" + qs[1]
			}
		}
		if valSearch != "" {
			query.Add(QueryOr, valSearch)
			b, _ := json.MarshalIndent(query, "", "  ")
			fmt.Println(string(b))
		}
	}

	t := v.Type()
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		if field.Name != "Model" && field.Type.Kind() != reflect.Slice {
			jsonTag := strings.Split(field.Tag.Get("json"), ",")[0]
			dbTag := strings.Split(field.Tag.Get("db"), ",")[0]
			for key, sv := range query {
				key, _ := url.QueryUnescape(key)
				subkey := strings.Split(key, ".")
				lastSubkey := subkey[len(subkey)-1]
				operator := GetOperator(lastSubkey)
				if operator == "" {
					operator = "="
				} else {
					key = strings.ReplaceAll(key, "."+lastSubkey, "")
				}
				isDbTag := false

				if subkey[0] == QueryDbField {
					isDbTag = true
					key = strings.ReplaceAll(key, QueryDbField+".", "")
				}
				if key == jsonTag || (isDbTag && key == dbTag) || (subkey[0] == jsonTag && field.Type.Name() == "NullJSON") {
					column := dbTag
					if field.Type.Name() == "NullJSON" {
						jsonKey := strings.Join(subkey[1:], ".")
						column = QuoteJSON(db, column, strings.ReplaceAll(jsonKey, "."+lastSubkey, ""))
					} else {
						column = db.Statement.Quote(column)
					}
					for _, val := range sv {
						colVal := strings.Split(val, QueryField+":")
						if len(colVal) > 1 {
							db = db.Where(column + operator + db.Statement.Quote(colVal[1]))
						} else if val != "null" {
							if field.Type.Name() == "NullBool" {
								if val == "true" {
									val = "1"
								} else if val == "false" {
									val = "0"
								}
							}
							if lastSubkey == QueryOptInsensitiveLike || lastSubkey == QueryOptInsensitiveNotLike {
								column = "lower(" + column + ")"
								val = strings.ToLower(val)
							}
							if lastSubkey == QueryOptIn || lastSubkey == QueryOptNotIn {
								args := strings.Split(val, ",")
								vars := "("
								for i := range args {
									if i == 0 {
										vars += "?"
									} else {
										vars += ",?"
									}
								}
								vars += ")"
								db = db.Where(column+operator+vars, args)
							} else {
								if strings.Contains(operator, "like") && !strings.Contains(val, "%") {
									val = "%" + val + "%"
								}
								db = db.Where(column+operator+"?", val)
							}
						} else if operator == "=" {
							db = db.Where(column + " is null")
						} else {
							db = db.Where(column + " is not null")
						}
					}
				}
			}
		}
	}

	return db
}

func GetOperator(key string) string {
	opt := map[string]string{
		QueryOptEqual:              "=",
		QueryOptNotEqual:           "!=",
		QueryOptGreaterThan:        ">",
		QueryOptGreaterThanOrEqual: ">=",
		QueryOptLowerThan:          "<",
		QueryOptLowerThanOrEqual:   "<=",
		QueryOptLike:               " like ",
		QueryOptNotLike:            " not like ",
		QueryOptInsensitiveLike:    " like ",
		QueryOptInsensitiveNotLike: " not like ",
		QueryOptIn:                 " in ",
		QueryOptNotIn:              " not in ",
	}
	res, _ := opt[key]
	return res
}

func SetOrder(db *gorm.DB, ptr reflect.Value, query url.Values) *gorm.DB {
	if ptr.Elem().Kind() == reflect.Slice {
		ptr = reflect.New(ptr.Elem().Type().Elem())
	}
	v := ptr.Elem()
	qSorts := strings.Split(query.Get(QuerySort), ",")
	if len(qSorts) > 0 && qSorts[0] != "" {
		for _, qs := range qSorts {
			direction := "asc"
			if qs[0:1] == "-" {
				qs = qs[1:]
				direction = "desc"
			}
			isCaseInsensitive := false
			ci := strings.Split(qs, ":")
			if len(ci) > 1 && ci[1] == "i" {
				qs = ci[0]
				isCaseInsensitive = true
			}
			column := ""
			t := v.Type()
			for i := 0; i < t.NumField(); i++ {
				field := t.Field(i)
				if field.Name != "Model" && field.Type.Kind() != reflect.Slice {
					jsonTag := strings.Split(field.Tag.Get("json"), ",")[0]
					dbTag := strings.Split(field.Tag.Get("db"), ",")[0]
					subKey := strings.Split(qs, ".")
					if qs == jsonTag {
						column = db.Statement.Quote(dbTag)
					} else if field.Type.Name() == "NullJSON" && subKey[0] == jsonTag {
						column = QuoteJSON(db, dbTag, strings.Join(subKey[1:], "."))
					}
				}
			}
			if column != "" {
				if isCaseInsensitive {
					column = "lower(" + column + ")"
				}
				db = db.Order(column + " " + direction)
			}
		}
		return db
	}
	if ptr.Elem().Kind() == reflect.Slice {
		ptr = reflect.New(ptr.Elem().Type().Elem())
	}
	s, isExist := v.FieldByName("Sort").Interface().([]Sort)
	if !isExist || len(s) == 0 {
		setSort := ptr.MethodByName("SetSort")
		if setSort.IsValid() {
			setSort.Call([]reflect.Value{})
		}
	}

	s, isExist = v.FieldByName("Sort").Interface().([]Sort)
	if isExist {
		for _, o := range s {
			if o.Direction == "" {
				o.Direction = "asc"
			}
			if o.JsonKey == "" {
				db = db.Order(db.Statement.Quote(o.Column) + " " + o.Direction)
			} else {
				db = db.Order(QuoteJSON(db, o.Column, o.JsonKey) + " " + o.Direction)
			}
		}
	}
	return db
}

func SetSelect(db *gorm.DB, ptr reflect.Value, query url.Values) *gorm.DB {
	v := ptr.Elem()
	t := v.Type()
	fields := []string{}
	selected := strings.Split(query.Get(QuerySelect), ",")
	hasSelected := func(field string) bool {
		for _, s := range selected {
			if field == s {
				return true
			}
		}
		return false
	}
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		if field.Name != "Model" && field.Tag.Get("json") != "" && field.Tag.Get("json") != "-" && field.Type.Kind() != reflect.Slice {
			if len(selected) == 0 || selected[0] == "" || hasSelected(field.Tag.Get("json")) {
				fields = append(fields, db.Statement.Quote(field.Tag.Get("db"))+" as "+Quote(db, field.Tag.Get("json")))
			}
		}
	}
	return db.Select(strings.Join(fields, ","))
}

func SetPagination(db *gorm.DB, query url.Values) *gorm.DB {
	if query.Get(QueryDisablePagination) == "true" {
		return db
	}
	page, limit := GetPaginationQuery(query)
	return db.Limit(limit).Offset((page - 1) * limit)
}

func GetPaginationQuery(query url.Values) (int, int) {
	page := 1
	limit := DefaultLimit
	if query.Get(QueryPage) != "" {
		pageTemp, _ := strconv.Atoi(query.Get(QueryPage))
		if pageTemp > 0 {
			page = pageTemp
		}
	}
	if query.Get(QueryLimit) != "" {
		limitTemp, _ := strconv.Atoi(query.Get(QueryLimit))
		if limitTemp > 0 {
			if limitTemp > MaxLimit {
				limit = MaxLimit
			} else {
				limit = limitTemp
			}
		}
	}
	return page, limit
}

func Quote(db *gorm.DB, text string) string {
	switch db.Dialector.Name() {
	case "sqlite", "mysql":
		return "`" + text + "`"
	case "postgres", "sqlserver", "firebird":
		return `"` + text + `"`
	default:
		return `"` + text + `"`
	}
}

func QuoteJSON(db *gorm.DB, column, jsonKey string) string {
	switch db.Dialector.Name() {
	case "mysql", "sqlite":
		return "JSON_EXTRACT(" + db.Statement.Quote(column) + ",$." + jsonKey + ")"
	case "sqlserver":
		return "JSON_VALUE(" + db.Statement.Quote(column) + ",$." + jsonKey + ")"
	case "postgres":
		keys := strings.Split(jsonKey, ".")
		jsonPath := ""
		for idx, key := range keys {
			if idx > 0 {
				jsonPath += ","
			}
			jsonPath += "'" + key + "'"
		}
		return "json_extract_path_text(" + db.Statement.Quote(column) + "::json," + jsonPath + ")"
	default:
		// unsupported json
		return db.Statement.Quote(column)
	}
}
