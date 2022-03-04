package db

import (
	"encoding/json"
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

	// search query params setting
	// ex: /contacts?$search=code,name=john     => sql: select * from contacts where (lower(code) = lower('john') or lower(name) = lower('john'))
	QuerySearch = "$search"

	// field query params setting
	// useful for filter, select or sort using another field
	// ex: /products?qty_available=$field:qty_on_hand          => sql: select * from products where qty_available = qty_on_hand
	// ex: /products?qty_on_order.$gt=$field:qty_available     => sql: select * from products where qty_on_order > qty_available
	QueryField = "$field"

	// aggregation query params
	// ex: /products?$select=$count:id                         => sql: select count(id) as "count.id" from products
	// ex: /products?$select=$sum:sold                         => sql: select sum(sold) as "sum.sold" from products
	// ex: /products?$select=$min:sold                         => sql: select min(sold) as "min.sold" from products
	// ex: /products?$select=$max:sold                         => sql: select max(sold) as "max.sold" from products
	// ex: /products?$select=$avg:sold                         => sql: select avg(sold) as "avg.sold" from products
	QueryCount = "$count"
	QuerySum   = "$sum"
	QueryMin   = "$min"
	QueryMax   = "$max"
	QueryAvg   = "$avg"

	// grouping query params setting
	// ex: /products?$group=category.id                                                 => sql: select category_id from products group by category_id
	// ex: /products?$group=category.id&$select=category.id,$avg:sold                   => sql: select category_id, avg(sold) as "avg.sold" from products group by category_id
	// ex: /products?$group=category.id&$select=category.id,$sum:sold&$sum:sold.$gt=0   => sql: select category_id, sum(sold) as "sum.sold" from products group by category_id having sum(sold) > 0
	// ex: /products?$group=category.id&$select=category.id,$sum:sold&$sort:-$sum:sold  => sql: select category_id, sum(sold) as "sum.sold" from products group by category_id order by sum(sold) desc
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
	ptr := reflect.ValueOf(dest)
	if ptr.Elem().Kind() == reflect.Slice {
		ptr = reflect.New(ptr.Elem().Type().Elem())
	}
	query.Add(QueryLimit, "1")
	query.Add(QueryInclude, "all")
	rows := FindRows(db, ptr, query)
	if len(rows) > 0 {
		return &queryResult{Dest: dest, Row: rows[0]}
	}
	return &queryResult{Error: gorm.ErrRecordNotFound}
}

func Find(db *gorm.DB, dest interface{}, query url.Values) *queryResult {
	if reflect.TypeOf(dest).Kind() != reflect.Ptr {
		return &queryResult{Error: gorm.ErrInvalidValue}
	}
	ptr := reflect.ValueOf(dest)
	if ptr.Elem().Kind() == reflect.Slice {
		ptr = reflect.New(ptr.Elem().Type().Elem())
	}
	rows := FindRows(db, ptr, query)
	if len(rows) > 0 {
		return &queryResult{Dest: dest, Rows: rows}
	}
	return &queryResult{Error: gorm.ErrRecordNotFound, Rows: rows}
}

func PaginationInfo(db *gorm.DB, dest interface{}, query url.Values) (int64, int64, int64, int64, error) {
	if reflect.TypeOf(dest).Kind() != reflect.Ptr {
		return 0, 0, 0, 0, gorm.ErrInvalidValue
	}
	ptr := reflect.ValueOf(dest)
	if ptr.Elem().Kind() == reflect.Slice {
		ptr = reflect.New(ptr.Elem().Type().Elem())
	}
	count := int64(0)
	db = SetTable(db, ptr, query)
	db = SetJoin(db, ptr, query)
	db = SetWhere(db, ptr, query)
	db = SetGroup(db, ptr, query)
	db.Count(&count)
	page, limit := GetPaginationQuery(query)
	pageCount := int64(math.Ceil(float64(count) / float64(limit)))
	return count, page, limit, pageCount, nil
}

func FindRows(baseDB *gorm.DB, ptr reflect.Value, query url.Values) []map[string]interface{} {
	rows := []map[string]interface{}{}
	db := baseDB.Session(&gorm.Session{})
	db = SetTable(db, ptr, query)
	db = SetJoin(db, ptr, query)
	db = SetWhere(db, ptr, query)
	db = SetGroup(db, ptr, query)
	db = SetSelect(db, ptr, query)
	db = SetOrder(db, ptr, query)
	db = SetPagination(db, query)
	db.Find(&rows)
	for i, v := range rows {
		rows[i] = IncludeArray(baseDB, fixDataType(v, ptr), ptr, query)
	}
	return rows
}

func GetPaginationQuery(query url.Values) (int64, int64) {
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
	return int64(page), int64(limit)
}

func fixDataType(data map[string]interface{}, ptr reflect.Value) map[string]interface{} {
	t := ptr.Elem().Type()
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		dbTag := strings.Split(field.Tag.Get("db"), ",")
		if field.Type.Name() == "NullBool" {
			jsonTag := strings.Split(field.Tag.Get("json"), ",")[0]
			for key, val := range data {
				if key == jsonTag {
					b := NullBool{}
					b.Scan(val)
					data[key] = b
				}
			}
		} else if field.Type.Name() == "NullDate" {
			jsonTag := strings.Split(field.Tag.Get("json"), ",")[0]
			for key, val := range data {
				if key == jsonTag {
					b := NullDate{}
					b.Scan(val)
					data[key] = b
				}
			}
		} else if len(dbTag) > 1 && dbTag[1] == "json" {
			jsonTag := strings.Split(field.Tag.Get("json"), ",")[0]
			for key, val := range data {
				if key == jsonTag {
					var v interface{}
					var err error
					s, isString := val.(string)
					if isString {
						err = json.Unmarshal([]byte(s), &v)
					} else {
						b, _ := val.([]byte)
						err = json.Unmarshal(b, &v)
					}
					if err == nil {
						data[key] = v
					}
				}
			}
		}
	}
	return data
}

func IncludeArray(db *gorm.DB, data map[string]interface{}, ptr reflect.Value, query url.Values) map[string]interface{} {
	t := ptr.Elem().Type()
	includes := strings.Split(query.Get(QueryInclude), ",")
	isIncludeAll := len(includes) > 0 && includes[0] == "all"
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		if field.Name != "Model" && field.Tag.Get("json") != "" && field.Tag.Get("json") != "-" && field.Type.Kind() == reflect.Slice {
			jsonTag := strings.Split(field.Tag.Get("json"), ",")[0]
			dbTag := strings.Split(field.Tag.Get("db"), ",")[0]
			rel := strings.Split(dbTag, "=")
			if len(rel) > 1 && (isIncludeAll || InArray(includes, jsonTag)) {
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

func InArray(needle []string, haystack string) bool {
	for _, v := range needle {
		if v == haystack {
			return true
		}
	}
	return false
}

func CallMethod(ptr reflect.Value, methodName string, args []reflect.Value) []reflect.Value {
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

func SetTable(db *gorm.DB, ptr reflect.Value, query url.Values) *gorm.DB {
	tableName := convert.ToSnakeCase(ptr.Type().Name())
	tn := CallMethod(ptr, "TableName", []reflect.Value{})
	if len(tn) > 0 {
		tableName = tn[0].String()
	}

	tableAliasName := tableName
	tan := CallMethod(ptr, "TableAliasName", []reflect.Value{})
	if len(tan) > 0 {
		tableAliasName = tan[0].String()
	}

	// quote table name if not from sub query
	if !strings.Contains(tableName, " ") {
		tableName = Quote(db, tableName)
	}

	return db.Table(tableName + " as " + Quote(db, tableAliasName))
}

func SetJoin(db *gorm.DB, ptr reflect.Value, query url.Values) *gorm.DB {
	relations, isExist := ptr.Elem().FieldByName("Relation").Interface().([]Relation)
	if !isExist || len(relations) == 0 {
		CallMethod(ptr, "SetRelation", []reflect.Value{})
		relations, isExist = ptr.Elem().FieldByName("Relation").Interface().([]Relation)
	}
	if isExist {
		for _, rel := range relations {
			joinQuery := strings.Builder{}
			joinQuery.WriteString(rel.JoinType) // inner join, left join, right join, full join, cross join
			if !strings.HasSuffix(strings.ToLower(rel.JoinType), "join") {
				joinQuery.WriteString(" join")
			}
			if !strings.Contains(rel.TableName, " ") { // quote table name if not join sub query
				rel.TableName = Quote(db, rel.TableName)
			}
			joinQuery.WriteString(" " + rel.TableName)
			joinQuery.WriteString(" as " + Quote(db, rel.TableAliasName))
			joinConditions := []string{}
			args := []interface{}{}
			for _, rc := range rel.RelationCondition {
				joinCondition := strings.Builder{}
				joinCondition.WriteString(db.Statement.Quote(rc.Column))
				if rc.Operator != "" {
					joinCondition.WriteString(rc.Operator)
				} else {
					joinCondition.WriteString("=")
				}
				if rc.Column2 != "" {
					joinCondition.WriteString(db.Statement.Quote(rc.Column2))
				} else if rc.Value != nil {
					joinCondition.WriteString("?")
					args = append(args, rc.Value)
				}
				joinConditions = append(joinConditions, joinCondition.String())
			}
			if len(joinConditions) > 0 {
				joinQuery.WriteString(" on " + strings.Join(joinConditions, " and "))
			}
			db = db.Joins(joinQuery.String(), args...)
		}
	}
	return db
}

func SetWhere(baseDB *gorm.DB, ptr reflect.Value, query url.Values) *gorm.DB {
	db := baseDB.Session(&gorm.Session{})
	setOperator := func(key string) string {
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
	// filter from schema
	filters, isExist := ptr.Elem().FieldByName("Filter").Interface().([]Filter)
	if !isExist || len(filters) == 0 {
		CallMethod(ptr, "SetFilter", []reflect.Value{})
		filters, isExist = ptr.Elem().FieldByName("Filter").Interface().([]Filter)
	}
	if isExist {
		for _, f := range filters {
			column := f.Column
			if f.JsonKey != "" {
				column = QuoteJSON(db, column, f.JsonKey)
			} else if !strings.Contains(column, " ") && !strings.Contains(column, "'") {
				column = db.Statement.Quote(column)
			}
			if f.Operator == "" {
				f.Operator = "="
			}
			if f.Column2 != "" {
				db = db.Where(column + f.Operator + db.Statement.Quote(f.Column2))
			} else if f.Value != nil {
				db = db.Where(column+f.Operator+"?", f.Value)
			} else if f.Operator == "=" {
				db = db.Where(column + " is null")
			} else {
				db = db.Where(column + " is not null")
			}
		}
	}
	// filter from query except $search & $or
	t := ptr.Elem().Type()
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		if field.Name != "Model" && field.Type.Kind() != reflect.Slice {
			jsonTag := strings.Split(field.Tag.Get("json"), ",")[0]
			dbTag := strings.Split(field.Tag.Get("db"), ",")[0]
			for key, sv := range query {
				key, _ := url.QueryUnescape(key)
				subkey := strings.Split(key, ".")
				lastSubkey := subkey[len(subkey)-1]
				operator := setOperator(lastSubkey)
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
					} else if !strings.Contains(column, " ") && !strings.Contains(column, "'") {
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
								db = db.Where(column+operator+"(?)", strings.Split(val, ","))
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
	// filter from query $search
	qs := strings.Split(query.Get(QuerySearch), "=")
	if len(qs) > 1 {
		valSearch := strings.Builder{}
		for i, s := range strings.Split(qs[0], ",") {
			if i == 0 {
				valSearch.WriteString(s + "." + QueryOptInsensitiveLike + "=" + qs[1])
			} else {
				valSearch.WriteString(QueryOrDelimiter + s + "." + QueryOptInsensitiveLike + "=" + qs[1])
			}
		}
		if valSearch.Len() > 0 {
			query.Add(QueryOr, valSearch.String())
		}
	}
	// filter from query $or
	for orKey, sv := range query {
		if orKey == QueryOr {
			for _, orQuery := range sv {
				orDB := baseDB.Session(&gorm.Session{DryRun: true})
				orQ := strings.Split(orQuery, QueryOrDelimiter)
				for _, orStr := range orQ {
					or := strings.Split(orStr, "=")
					if len(or) > 1 {
						key, _ := url.QueryUnescape(or[0])
						subkey := strings.Split(key, ".")
						lastSubkey := subkey[len(subkey)-1]
						operator := setOperator(lastSubkey)
						val, _ := url.QueryUnescape(or[1])
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
						for i := 0; i < t.NumField(); i++ {
							field := t.Field(i)
							if field.Name != "Model" && field.Type.Kind() != reflect.Slice {
								jsonTag := strings.Split(field.Tag.Get("json"), ",")[0]
								dbTag := strings.Split(field.Tag.Get("db"), ",")[0]
								if key == jsonTag || (isDbTag && key == dbTag) || (subkey[0] == jsonTag && field.Type.Name() == "NullJSON") {
									column := dbTag
									if field.Type.Name() == "NullJSON" {
										jsonKey := strings.Join(subkey[1:], ".")
										column = QuoteJSON(db, column, strings.ReplaceAll(jsonKey, "."+lastSubkey, ""))
									} else if !strings.Contains(column, " ") && !strings.Contains(column, "'") {
										column = db.Statement.Quote(column)
									}
									colVal := strings.Split(val, QueryField+":")
									if len(colVal) > 1 {
										orDB = orDB.Or(column + operator + db.Statement.Quote(colVal[1]))
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
											orDB = orDB.Or(column+operator+"(?)", strings.Split(val, ","))
										} else {
											if strings.Contains(operator, "like") && !strings.Contains(val, "%") {
												val = "%" + val + "%"
											}
											orDB = orDB.Or(column+operator+"?", val)
										}
									} else if operator == "=" {
										orDB = orDB.Or(column + " is null")
									} else {
										orDB = orDB.Or(column + " is not null")
									}
								}
							}
						}
					}
				}
				db = db.Where(orDB)
			}
		}
	}
	return db
}

func SetGroup(db *gorm.DB, ptr reflect.Value, query url.Values) *gorm.DB {
	grouped := strings.Split(query.Get(QueryGroup), ",")
	t := ptr.Elem().Type()
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		jsonTag := strings.Split(field.Tag.Get("json"), ",")[0]
		if field.Name != "Model" && jsonTag != "" && jsonTag != "-" && field.Type.Kind() != reflect.Slice {
			dbTag := strings.Split(field.Tag.Get("db"), ",")
			if (len(dbTag) > 1 && dbTag[1] == "group") || InArray(grouped, jsonTag) {
				db = db.Group(db.Statement.Quote(dbTag[0]))
			}
		}
	}
	return db
}

func SetSelect(db *gorm.DB, ptr reflect.Value, query url.Values) *gorm.DB {
	grouped := strings.Split(query.Get(QueryGroup), ",")
	selected := strings.Split(query.Get(QuerySelect), ",")
	fields := []string{}
	t := ptr.Elem().Type()
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		jsonTag := strings.Split(field.Tag.Get("json"), ",")[0]
		if field.Name != "Model" && jsonTag != "" && jsonTag != "-" && field.Type.Kind() != reflect.Slice {
			dbTag := strings.Split(field.Tag.Get("db"), ",")[0]
			if !strings.Contains(dbTag, " ") && !strings.Contains(dbTag, "'") {
				dbTag = db.Statement.Quote(dbTag)
			}
			if (grouped[0] == "" && selected[0] == "") || InArray(grouped, jsonTag) || InArray(selected, jsonTag) {
				fields = append(fields, dbTag+" as "+Quote(db, jsonTag))
			} else if selected[0] != "" {
				for _, selectd := range selected {
					s := strings.Split(selectd, ":")
					if len(s) > 1 && s[1] == jsonTag {
						switch s[0] {
						case QueryCount:
							fields = append(fields, "count("+dbTag+") as "+Quote(db, "count."+jsonTag))
						case QuerySum:
							fields = append(fields, "sum("+dbTag+") as "+Quote(db, "sum."+jsonTag))
						case QueryMin:
							fields = append(fields, "min("+dbTag+") as "+Quote(db, "min."+jsonTag))
						case QueryMax:
							fields = append(fields, "max("+dbTag+") as "+Quote(db, "max."+jsonTag))
						case QueryAvg:
							fields = append(fields, "avg("+dbTag+") as "+Quote(db, "avg."+jsonTag))
						}
					}
				}
			}
		}
	}
	return db.Select(strings.Join(fields, ","))
}

func SetOrder(db *gorm.DB, ptr reflect.Value, query url.Values) *gorm.DB {
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
			t := ptr.Elem().Type()
			for i := 0; i < t.NumField(); i++ {
				field := t.Field(i)
				if field.Name != "Model" && field.Type.Kind() != reflect.Slice {
					jsonTag := strings.Split(field.Tag.Get("json"), ",")[0]
					dbTag := strings.Split(field.Tag.Get("db"), ",")[0]
					subKey := strings.Split(qs, ".")
					if qs == jsonTag {
						column = dbTag
						if !strings.Contains(column, " ") && !strings.Contains(column, "'") {
							column = db.Statement.Quote(column)
						}
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
	sorts, isExist := ptr.Elem().FieldByName("Sort").Interface().([]Sort)
	if !isExist || len(sorts) == 0 {
		CallMethod(ptr, "SetSort", []reflect.Value{})
		sorts, isExist = ptr.Elem().FieldByName("Sort").Interface().([]Sort)
	}
	if isExist {
		for _, s := range sorts {
			if s.Direction == "" {
				s.Direction = "asc"
			}
			if s.JsonKey == "" {
				column := s.Column
				if !strings.Contains(column, " ") && !strings.Contains(column, "'") {
					column = db.Statement.Quote(column)
				}
				db = db.Order(column + " " + s.Direction)
			} else {
				db = db.Order(QuoteJSON(db, s.Column, s.JsonKey) + " " + s.Direction)
			}
		}
	}
	return db
}

func SetPagination(db *gorm.DB, query url.Values) *gorm.DB {
	if query.Get(QueryDisablePagination) == "true" {
		return db
	}
	page, limit := GetPaginationQuery(query)
	return db.Limit(int(limit)).Offset(int((page - 1) * limit))
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
		jsonPath := strings.Builder{}
		keys := strings.Split(jsonKey, ".")
		for idx, key := range keys {
			if idx > 0 {
				jsonPath.WriteString(",")
			}
			jsonPath.WriteString("'" + key + "'")
		}
		return "json_extract_path_text(" + db.Statement.Quote(column) + "::json," + jsonPath.String() + ")"
	default:
		// unsupported json
		return db.Statement.Quote(column)
	}
}
