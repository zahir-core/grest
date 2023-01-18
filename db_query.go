package grest

import (
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"gorm.io/gorm"
)

// GREST support a common way for pagination, selecting fields, filtering, sorting, searching and other using URL query params
// this is the default of query parameter setting, you can override this with your own preferences
var (
	// pagination query params setting
	// if QueryLimit not setted, QueryDefaultLimit will be applied
	// if QueryLimit setted to 0, only PaginationInfo will be executed (only run a counting query against the database).
	// if QueryLimit setted greater than QueryMaxLimit, QueryDefaultLimit will be applied, Set QueryDefaultLimit to 0 to allow unlimited QueryLimit.
	// if QueryDisablePagination setted to true, QueryMaxLimit will be ignored and PaginationInfo will not be executed
	QueryDefaultLimit      = 10                       // Sets the default number of items when $per_page is not set
	QueryMaxLimit          = 100                      // Sets the maximum allowed number of items per page (even if the QueryLimit query parameter is set higher)
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

func Find(db *gorm.DB, model ModelInterface, query url.Values) ([]map[string]any, error) {
	q := &DBQuery{
		DB:     db,
		Model:  model,
		Schema: model.GetSchema(),
		Query:  query,
	}
	return q.Find(q.Schema)
}

type DBQuery struct {
	DB     *gorm.DB
	Model  ModelInterface
	Schema map[string]any
	Query  url.Values
	Data   []map[string]any
	Err    error
}

func (q *DBQuery) Prepare(schema map[string]any, db *gorm.DB) (*gorm.DB, error) {
	var err error
	if db == nil {
		db = q.DB.Session(&gorm.Session{})
	}
	db = q.SetTable(db, schema)
	db = q.SetJoin(db, schema)
	db = q.SetWhere(db, schema)
	db = q.SetGroup(db, schema)
	fmt.Println("todo")
	return db, err
}

func (q *DBQuery) Find(schema map[string]any) ([]map[string]any, error) {
	rows := []map[string]any{}
	db, err := q.Prepare(schema, nil)
	db = q.SetSelect(db, schema)
	db = q.SetOrder(db, schema)
	db = q.SetPagination(db)
	if err != nil {
		return rows, NewError(http.StatusInternalServerError, err.Error())
	}
	err = db.Find(&rows).Error
	if err != nil {
		return rows, NewError(http.StatusInternalServerError, err.Error())
	}
	return rows, nil
}

func (q *DBQuery) ToSQL(schema map[string]any) string {
	return q.DB.ToSQL(func(tx *gorm.DB) *gorm.DB {
		db, _ := q.Prepare(schema, tx)
		rows := []map[string]any{}
		return db.Find(&rows)
	})
}

func (q *DBQuery) SetTable(db *gorm.DB, schema map[string]any) *gorm.DB {
	tableName, _ := schema["tableName"].(string)
	tableAliasName, _ := schema["tableAliasName"].(string)
	if tableAliasName == "" {
		tableAliasName = tableName
	}

	// dynamic from sub query based on client's query params
	tableSchema, _ := schema["tableSchema"].(map[string]any)
	if len(tableSchema) > 0 {
		tableName = "(" + q.ToSQL(tableSchema) + ")"
	}

	// quote table name if not from sub query
	if !strings.Contains(tableName, " ") {
		tableName = q.Quote(tableName)
	}

	if tableName != "" {
		fromSQL := strings.Builder{}
		fromSQL.WriteString(tableName)
		fromSQL.WriteString(" AS ")
		fromSQL.WriteString(q.Quote(tableAliasName))
		db = db.Table(fromSQL.String())
	}

	return db
}

func (q *DBQuery) SetJoin(db *gorm.DB, schema map[string]any) *gorm.DB {
	relations, _ := schema["relations"].(map[string]map[string]any)
	if len(relations) > 0 {
		for key, rel := range relations {
			joinType, _ := rel["type"].(string)
			joinType = strings.ToUpper(joinType)
			if !strings.HasSuffix(joinType, "JOIN") {
				if joinType != "" {
					joinType += " "
				}
				joinType += "JOIN "
			}

			tableName, _ := rel["tableName"].(string)
			tableAliasName, _ := rel["tableAliasName"].(string)
			if tableAliasName == "" {
				tableAliasName = tableName
			}

			// dynamic from sub query based on client's query params
			tableSchema, _ := rel["tableSchema"].(map[string]any)
			if len(tableSchema) > 0 {
				subQuery := q.ToSQL(tableSchema)
				if subQuery != "" {
					tableName = "(" + subQuery + ")"
				}
			}

			if tableName != "" {

				// quote table name if not from sub query
				if !strings.Contains(tableName, " ") {
					tableName = q.Quote(tableName)
				}

				args := []any{}
				joinConditions := []string{}
				conditions, _ := rel["conditions"].([]map[string]any)
				for _, cond := range conditions {
					if len(cond) > 0 {
						joinCondition, arg := q.toWhereSQL(cond)
						joinConditions = append(joinConditions, joinCondition)
						args = append(args, arg)
					}
				}

				joinSQL := strings.Builder{}
				joinSQL.WriteString(joinType)
				joinSQL.WriteString(tableName)
				joinSQL.WriteString(" AS ")
				joinSQL.WriteString(q.Quote(key))
				if len(joinConditions) > 0 {
					joinSQL.WriteString(" ON " + strings.Join(joinConditions, " AND "))
				}
				db = db.Joins(joinSQL.String(), args...)
			}
		}
	}
	return db
}

func (q *DBQuery) toWhereSQL(cond map[string]any) (string, any) {
	where := strings.Builder{}

	operator, _ := cond["operator"].(string)
	if operator == "" {
		operator = "="
	}
	isOperatorIN := strings.Contains(strings.ToUpper(operator), "IN")
	isOperatorLIKE := strings.Contains(strings.ToUpper(operator), "LIKE")
	isCaseInsensitive, _ := cond["isCaseInsensitive"].(bool)

	column1, _ := cond["column1"].(string)
	column1jsonKey, _ := cond["column1jsonKey"].(string)
	column1type, _ := cond["column1type"].(string)
	column1type = strings.ToLower(column1type)
	column1isBool := strings.Contains(column1type, "bool")
	column1isDateTime := strings.Contains(column1type, "date") || strings.Contains(column1type, "time")
	if column1jsonKey != "" {
		column1 = q.QuoteJSON(column1, column1jsonKey)
	}
	if column1 != "" {
		// quote table name if not from sub query
		if !strings.Contains(column1, " ") {
			column1 = q.DB.Statement.Quote(column1)
		}
		if isCaseInsensitive && !column1isDateTime {
			column1 = "LOWER(" + column1 + ")"
		}
		if column1isDateTime && isOperatorLIKE {
			column1 = "CAST(" + column1 + " AS CHAR)"
		}
		where.WriteString(column1)
	}

	arg, isValueExists := cond["value"]
	argStr, _ := arg.(string)
	if isValueExists && (arg == nil || strings.ToLower(argStr) == "null") {
		sqlNullOpt := " IS "
		if operator != "=" {
			sqlNullOpt += "NOT "
		}
		where.WriteString(sqlNullOpt)
	} else if isOperatorIN || isOperatorLIKE {
		sqlInLikeOpt := strings.ToUpper(operator)
		if !strings.Contains(operator, " ") {
			sqlInLikeOpt = " " + sqlInLikeOpt + " "
		}
		where.WriteString(sqlInLikeOpt)
	} else {
		where.WriteString(operator)
	}

	column2, _ := cond["column2"].(string)
	column2jsonKey, _ := cond["column2jsonKey"].(string)
	column2type, _ := cond["column2type"].(string)
	column2type = strings.ToLower(column2type)
	column2isDateTime := strings.Contains(column2type, "date") || strings.Contains(column2type, "time")
	if column2jsonKey != "" {
		column2 = q.QuoteJSON(column2, column2jsonKey)
	}
	if column2 != "" {
		// quote table name if not from sub query
		if !strings.Contains(column2, " ") {
			column2 = q.DB.Statement.Quote(column2)
		}
		if isCaseInsensitive && !column2isDateTime {
			column2 = "LOWER(" + column1 + ")"
		}
		if column2isDateTime && isOperatorLIKE {
			column2 = "CAST(" + column2 + " AS CHAR)"
		}
		where.WriteString(column2)
	} else if isOperatorIN {
		where.WriteString("(?)")
		if argStr != "" {
			arg = strings.Split(argStr, ",")
		}
	} else {
		where.WriteString("?")
		if argStr != "" {
			if isCaseInsensitive || column1isBool {
				argStr = strings.ToLower(argStr)
			}
			if column1isBool {
				if argStr == "true" || argStr == "t" || argStr == "1" {
					argStr = "1"
				} else {
					argStr = "0"
				}
			}
			if isOperatorLIKE && !strings.Contains(argStr, "%") {
				argStr = "%" + argStr + "%"
			}
			arg = argStr
		}
	}

	return where.String(), arg
}

func (q *DBQuery) toOptSQL(key string) string {
	opt := map[string]string{
		QueryOptEqual:              "=",
		QueryOptNotEqual:           "!=",
		QueryOptGreaterThan:        ">",
		QueryOptGreaterThanOrEqual: ">=",
		QueryOptLowerThan:          "<",
		QueryOptLowerThanOrEqual:   "<=",
		QueryOptLike:               "LIKE",
		QueryOptNotLike:            "NOT LIKE",
		QueryOptInsensitiveLike:    "LIKE",
		QueryOptInsensitiveNotLike: "NOT LIKE",
		QueryOptIn:                 "IN",
		QueryOptNotIn:              "NOT IN",
	}
	res, _ := opt[key]
	return res
}

func (q *DBQuery) keyValQueryToCond(key, val string, fields map[string]map[string]any, arrayFields map[string]map[string]any) map[string]any {
	cond := map[string]any{}
	key, _ = url.QueryUnescape(key)
	subkey := strings.Split(key, ".")
	lastSubkey := subkey[len(subkey)-1]

	operator := q.toOptSQL(lastSubkey)
	cond["operator"] = operator

	if lastSubkey == QueryOptInsensitiveLike || lastSubkey == QueryOptInsensitiveNotLike {
		cond["isCaseInsensitive"] = true
	}

	if subkey[0] == QueryDbField {
		cond["column1"] = strings.ReplaceAll(key, QueryDbField+".", "")
	}
	if operator != "" {
		key = strings.ReplaceAll(key, "."+lastSubkey, "")
	}
	if fields[key] != nil {
		cond["column1"], _ = fields[key]["db"].(string)
		cond["column1type"], _ = fields[key]["type"].(string)
	} else {
		// todo : arrayFields, json, etc
	}

	colVal := strings.Split(val, QueryField+":")
	if len(colVal) > 1 {
		cond["column2"] = colVal[1]
	} else {
		vUnescape, err := url.QueryUnescape(val)
		if err != nil {
			cond["value"] = val
		} else {
			cond["value"] = vUnescape
		}
	}
	return cond
}

func (q *DBQuery) SetWhere(db *gorm.DB, schema map[string]any) *gorm.DB {

	// filter from schema
	filters, _ := schema["filters"].([]map[string]any)
	if len(filters) > 0 {
		for _, cond := range filters {
			whereSQL, arg := q.toWhereSQL(cond)
			if strings.Contains(whereSQL, "?") {
				db = db.Where(whereSQL, arg)
			} else {
				db = db.Where(whereSQL)
			}
		}
	}

	// filter from query except $search & $or
	fields, _ := schema["fields"].(map[string]map[string]any)
	arrayFields, _ := schema["arrayFields"].(map[string]map[string]any)
	for key, val := range q.Query {
		cond := q.keyValQueryToCond(key, val[0], fields, arrayFields)
		if cond["column1"] != nil {
			whereSQL, arg := q.toWhereSQL(cond)
			if strings.Contains(whereSQL, "?") {
				db = db.Where(whereSQL, arg)
			} else {
				db = db.Where(whereSQL)
			}
		}
	}

	// filter from query $search
	qs := strings.Split(q.Query.Get(QuerySearch), "=")
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
			valSearchStr := valSearch.String()
			orVal, isOrValExists := q.Query[QueryOr]
			isSearchValExists := !isOrValExists
			if !isSearchValExists {
				for _, ov := range orVal {
					if ov == valSearchStr {
						isSearchValExists = true
					}
				}
			}
			if !isSearchValExists {
				q.Query.Add(QueryOr, valSearch.String())
			}
		}
	}

	// filter from query $or
	orVal, isOrValExists := q.Query[QueryOr]
	if isOrValExists {
		for _, ov := range orVal {
			orDB := q.DB.Session(&gorm.Session{})
			orQueries := strings.Split(ov, QueryOrDelimiter)
			for _, orQuery := range orQueries {
				orQ := strings.Split(orQuery, "=")
				val := ""
				if len(orQ) > 0 {
					val = orQ[1]
				}
				cond := q.keyValQueryToCond(orQ[0], val, fields, arrayFields)
				if cond["column1"] != nil {
					whereSQL, arg := q.toWhereSQL(cond)
					if strings.Contains(whereSQL, "?") {
						orDB = orDB.Or(whereSQL, arg)
					} else {
						orDB = orDB.Or(whereSQL)
					}
				}
			}
			db = db.Where(orDB)
		}
	}
	return db
}

func (q *DBQuery) SetGroup(db *gorm.DB, schema map[string]any) *gorm.DB {
	fields, _ := schema["fields"].(map[string]map[string]any)
	groups, _ := schema["groups"].(map[string]string)
	queryGroups := strings.Split(q.Query.Get(QueryGroup), ",")
	for _, qg := range queryGroups {
		group, ok := fields[qg]["db"].(string)
		if ok {
			if groups != nil {
				groups[qg] = group
			} else {
				groups = map[string]string{qg: group}
			}
		}
	}
	for _, group := range groups {
		// quote table name if not from sub query
		if !strings.Contains(group, " ") {
			group = q.DB.Statement.Quote(group)
		}
		db = db.Group(group)
	}
	return db
}

func (q *DBQuery) addSelect(selectedFields []string, field, alias string) []string {
	// quote table name if not from sub query
	if !strings.Contains(field, " ") {
		field = q.DB.Statement.Quote(field)
	}
	return append(selectedFields, field+" AS "+q.DB.Statement.Quote(alias))
}

func (q *DBQuery) toAggSQL(key string) string {
	opt := map[string]string{
		QueryCount: "count",
		QuerySum:   "sum",
		QueryMin:   "min",
		QueryMax:   "max",
		QueryAvg:   "avg",
	}
	res, _ := opt[key]
	return res
}

func (q *DBQuery) SetSelect(db *gorm.DB, schema map[string]any) *gorm.DB {
	selectedFields := []string{}
	fields, _ := schema["fields"].(map[string]map[string]any)
	querySelect := strings.Split(q.Query.Get(QuerySelect), ",")
	querySelect = append(querySelect, strings.Split(q.Query.Get(QueryGroup), ",")...)
	if len(querySelect) > 0 {
		for _, k := range querySelect {
			field, ok := fields[k]["db"].(string)
			if ok {
				selectedFields = q.addSelect(selectedFields, field, k)
			} else {
				aggregationQuery := strings.Split(k, "=")
				if aggregationQuery[0] == QueryCount {

				}
				fmt.Println("=================================================")
				fmt.Println("TODO")
				fmt.Println(k)
				fmt.Println("=================================================")
			}
		}
	} else {
		for k, f := range fields {
			field, ok := f["db"].(string)
			if ok {
				selectedFields = q.addSelect(selectedFields, field, k)
			}
		}
	}
	return db.Select(strings.Join(selectedFields, ", "))
}

func (q *DBQuery) SetOrder(db *gorm.DB, schema map[string]any) *gorm.DB {
	return db
}

func (q *DBQuery) SetPagination(db *gorm.DB) *gorm.DB {
	page, limit := q.GetPaginationQuery()
	if limit == 0 {
		return db
	}
	return db.Limit(int(limit)).Offset(int((page - 1) * limit))
}

func (q *DBQuery) GetPaginationQuery() (int, int) {
	if q.Query.Get(QueryDisablePagination) == "true" {
		return 0, 0
	}
	page := 1
	limit := QueryDefaultLimit
	if q.Query.Get(QueryPage) != "" {
		pageTemp, _ := strconv.Atoi(q.Query.Get(QueryPage))
		if pageTemp > 0 {
			page = pageTemp
		}
	}
	if q.Query.Get(QueryLimit) != "" {
		limitTemp, _ := strconv.Atoi(q.Query.Get(QueryLimit))
		if limitTemp > 0 {
			if limitTemp > QueryMaxLimit {
				limit = QueryMaxLimit
			} else {
				limit = limitTemp
			}
		}
	}
	return page, limit
}

func (q DBQuery) Quote(text string) string {
	switch q.DB.Dialector.Name() {
	case "sqlite", "mysql":
		return "`" + text + "`"
	case "postgres", "sqlserver", "firebird":
		return `"` + text + `"`
	default:
		return `"` + text + `"`
	}
}

func (q DBQuery) QuoteJSON(column, jsonKey string) string {
	switch q.DB.Dialector.Name() {
	case "mysql", "sqlite":
		return "JSON_EXTRACT(" + q.DB.Statement.Quote(column) + ",$." + jsonKey + ")"
	case "sqlserver":
		return "JSON_VALUE(" + q.DB.Statement.Quote(column) + ",$." + jsonKey + ")"
	case "postgres":
		jsonPath := strings.Builder{}
		keys := strings.Split(jsonKey, ".")
		for idx, key := range keys {
			if idx > 0 {
				jsonPath.WriteString(",")
			}
			jsonPath.WriteString("'" + key + "'")
		}
		return "json_extract_path_text(" + q.DB.Statement.Quote(column) + "::json," + jsonPath.String() + ")"
	default:
		// unsupported json
		return q.DB.Statement.Quote(column)
	}
}

func (q DBQuery) NewUUIDSQL() string {
	switch q.DB.Dialector.Name() {
	case "sqlite":
		return "lower(hex(randomblob(4))) || '-' || lower(hex(randomblob(2))) || '-4' || substr(lower(hex(randomblob(2))),2) || '-' || substr('89ab',abs(random()) % 4 + 1, 1) || substr(lower(hex(randomblob(2))),2) || '-' || lower(hex(randomblob(6)))"
	case "mysql":
		return "UUID()"
	case "postgres":
		return "md5(random()::text || clock_timestamp()::text)::uuid"
	case "sqlserver":
		return "LOWER(CAST(NEWID() AS CHAR(36)))"
	case "firebird":
		return "LOWER(UUID_TO_CHAR(GEN_UUID()))"
	default:
		return ""
	}
}
