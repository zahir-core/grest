package grest

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/go-playground/validator/v10"
	"golang.org/x/exp/slices"
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
	// exclude query params setting
	// for First method and Find method, by default query for all fields defined in struct is executed
	// except for fields that explicitly hide by db:"-,hide" struct tags
	// to exclude some fields from executed in query you can hide it using QueryExclude
	// it can be setted by multiple fields, separated by comma
	// ex: /contacts?$exclude=families,friends,phones            => exclude fields: families, friends, and phones
	QueryExclude = "$exclude"
	QueryDbField = "$db_field"
)

// Find finds all records matching given conditions conds from model and query params
func Find(db *gorm.DB, model ModelInterface, query url.Values) ([]map[string]any, error) {
	q := &DBQuery{
		DB:     db,
		Model:  model,
		Schema: model.GetSchema(),
		Query:  query,
	}
	return q.Find(q.Schema, query)
}

// DBQuery DBQuery definition for querying with model & query params
type DBQuery struct {
	DB     *gorm.DB
	Model  ModelInterface
	Schema map[string]any
	Query  url.Values
	Data   []map[string]any
	Err    error
}

// Find finds all records matching given conditions conds from schema and query params
func (q *DBQuery) Find(schema map[string]any, qry ...url.Values) ([]map[string]any, error) {
	rows := []map[string]any{}
	query := q.Query
	if len(qry) > 0 {
		query = qry[0]
	}
	db, err := q.Prepare(nil, schema, query)
	db = q.SetSelect(db, schema, query)
	db = q.SetOrder(db, schema, query)
	db = q.SetPagination(db, query)
	if err != nil {
		return rows, NewError(http.StatusInternalServerError, err.Error())
	}
	err = db.Find(&rows).Error
	if err != nil {
		return rows, NewError(http.StatusInternalServerError, err.Error())
	}
	rows = q.fixDataType(schema, rows)
	return q.includeArray(schema, rows)
}

// fixDataType from db
func (q *DBQuery) fixDataType(schema map[string]any, rows []map[string]any) []map[string]any {
	isNeedFiDataType := false
	boolKeys := []string{}
	jsonKeys := []string{}
	fields, _ := schema["fields"].(map[string]map[string]any)
	for k, f := range fields {
		dataType, _ := f["type"].(string)
		dataType = strings.ToLower(dataType)
		if strings.Contains(dataType, "bool") {
			isNeedFiDataType = true
			boolKeys = append(boolKeys, k)
		} else if strings.Contains(dataType, "json") {
			isNeedFiDataType = true
			jsonKeys = append(jsonKeys, k)
		}
	}
	if isNeedFiDataType {
		for i, row := range rows {
			for k, v := range row {
				for _, bk := range boolKeys {
					if k == bk {
						switch fmt.Sprintf("%v", v) {
						case "1":
							rows[i][k] = true
						case "0":
							rows[i][k] = false
						}
					}
				}
				for _, jk := range jsonKeys {
					if k == jk {
						err := json.Unmarshal([]byte(fmt.Sprintf("%v", v)), &v)
						if err == nil {
							rows[i][k] = v
						}
					}
				}
			}
		}
	}
	return rows
}

// includeArray include array fields to rows
func (q *DBQuery) includeArray(schema map[string]any, rows []map[string]any) ([]map[string]any, error) {
	includeArray := strings.Split(q.Query.Get(QueryInclude), ",")
	if includeArray[0] != "" {
		arrayFieldOrder, _ := schema["arrayFieldOrder"].([]string)
		arrayFields, _ := schema["arrayFields"].(map[string]map[string]any)
		for i, row := range rows {
			if includeArray[0] == "all" {
				for _, arrayKey := range arrayFieldOrder {
					arrayField, _ := arrayFields[arrayKey]
					arrayRows, err := q.getArrayRows(arrayKey, arrayField, row)
					if err != nil {
						return arrayRows, err
					}
					rows[i][arrayKey] = arrayRows
				}
			} else {
				for _, k := range includeArray {
					arrayField, ok := arrayFields[k]
					if ok {
						arrayRows, err := q.getArrayRows(k, arrayField, row)
						if err != nil {
							return arrayRows, err
						}
						rows[i][k] = arrayRows
					}
				}
			}
		}
	}

	return rows, nil
}

// getArrayRows include array fields to rows
func (q *DBQuery) getArrayRows(arrayKey string, arrayField map[string]any, row map[string]any) ([]map[string]any, error) {
	arraySchema, ok := arrayField["schema"].(map[string]any)
	if ok {
		arrayFilter, _ := arrayField["filter"].(string)
		vars := String{}.GetVars(arrayFilter, "{", "}")
		for _, v := range vars {
			val, ok := row[v]
			if ok {
				arrayFilter = strings.ReplaceAll(arrayFilter, "{"+v+"}", fmt.Sprintf("%v", val))
			}
		}
		arrayQuery, _ := url.ParseQuery(arrayFilter)
		for key, qs := range q.Query {
			if strings.HasPrefix(key, arrayKey+".*.") {
				for _, qv := range qs {
					arrayQuery.Add(strings.ReplaceAll(key, arrayKey+".*.", ""), qv)
				}
			}
		}
		arrayQuery.Add(QueryDisablePagination, "true")
		return q.Find(arraySchema, arrayQuery)
	}
	return []map[string]any{}, nil
}

// ToSQL generate SQL string from schema and query params
func (q *DBQuery) ToSQL(schema map[string]any, qry ...url.Values) string {
	return q.DB.ToSQL(func(tx *gorm.DB) *gorm.DB {
		rows := []map[string]any{}
		query := q.Query
		if len(qry) > 0 {
			query = qry[0]
		}
		db, _ := q.Prepare(tx, schema, query)
		schema["is_skip_query_select"] = true
		db = q.SetSelect(db, schema, query)
		db = q.SetOrder(db, schema, query)
		return db.Find(&rows)
	})
}

// Prepare prepare gorm.DB for querying with schema & query params
func (q *DBQuery) Prepare(db *gorm.DB, schema map[string]any, query url.Values) (*gorm.DB, error) {
	var err error
	if db == nil {
		db = q.DB.Session(&gorm.Session{})
	}
	db = q.SetTable(db, schema, query)
	db = q.SetJoin(db, schema, query)
	db = q.SetWhere(db, schema, query)
	db = q.SetGroup(db, schema, query)
	return db, err
}

// SetTable specify the table you would like to run db operations
func (q *DBQuery) SetTable(db *gorm.DB, schema map[string]any, query url.Values) *gorm.DB {
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

// SetJoin specify the join method when querying
func (q *DBQuery) SetJoin(db *gorm.DB, schema map[string]any, query url.Values) *gorm.DB {
	relationOrder, _ := schema["relationOrder"].([]string)
	if len(relationOrder) > 0 {
		relations, _ := schema["relations"].(map[string]map[string]any)
		type joinStr struct {
			Str  string
			Args []any
		}
		joins := []joinStr{}
		for _, key := range relationOrder {
			rel, _ := relations[key]
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
					tableName = "( " + subQuery + " )"
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
						joinCondition, arg := q.condToWhereSQL(cond)
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
				joins = append(joins, joinStr{Str: joinSQL.String(), Args: args})
			}
		}
		for _, j := range joins {
			db = db.Joins(j.Str, j.Args...)
		}
	}
	return db
}

// SetWhere specify the where method when querying
func (q *DBQuery) SetWhere(db *gorm.DB, schema map[string]any, query url.Values) *gorm.DB {

	// filter from schema
	filters, _ := schema["filters"].([]map[string]any)
	if len(filters) > 0 {
		for _, cond := range filters {
			whereSQL, arg := q.condToWhereSQL(cond)
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
	for key, val := range query {
		cond := q.qsToCond(key, val[0], fields, arrayFields)
		if cond["column1"] != nil {
			whereSQL, arg := q.condToWhereSQL(cond)
			if strings.Contains(whereSQL, "?") {
				db = db.Where(whereSQL, arg)
			} else {
				db = db.Where(whereSQL)
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
			valSearchStr := valSearch.String()
			orVal, isOrValExists := query[QueryOr]
			isSearchValExists := !isOrValExists
			if !isSearchValExists {
				for _, ov := range orVal {
					if ov == valSearchStr {
						isSearchValExists = true
					}
				}
			}
			if !isSearchValExists {
				query.Add(QueryOr, valSearch.String())
			}
		}
	}

	// filter from query $or
	orVal, isOrValExists := query[QueryOr]
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
				cond := q.qsToCond(orQ[0], val, fields, arrayFields)
				if cond["column1"] != nil {
					whereSQL, arg := q.condToWhereSQL(cond)
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

// qsToCond convert key val query params to schema conditions
func (q *DBQuery) qsToCond(key, val string, fields map[string]map[string]any, arrayFields map[string]map[string]any) map[string]any {
	cond := map[string]any{}
	key, _ = url.QueryUnescape(key)
	subkey := strings.Split(key, ".")
	lastSubkey := subkey[len(subkey)-1]

	operator := q.qsToOptSQL(lastSubkey)
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
		for k, v := range fields {
			if strings.HasPrefix(key, k) {
				cond["column1"], _ = v["db"].(string)
				fType, _ := v["type"].(string)
				cond["column1type"] = fType
				if strings.Contains(strings.ToLower(fType), "json") {
					cond["column1jsonKey"] = strings.Replace(key, k+".", "", 1)
				}
			}
		}
		if cond["column1"] == nil {
			// todo : filter from array fields
			// ?arrayFields.0.field.id={field_id} > where exists (select 1 from array_table at where at.parent_id = parent.id and field_id = {field_id})
			// ?arrayFields.*.field.id={field_id} > same as above but the array fields response also filtered
			// for k, v := range arrayFields {
			// 	// todo
			// }
		}
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

// qsToOptSQL return sql operator from part of query params key
func (q *DBQuery) qsToOptSQL(key string) string {
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

// condToWhereSQL convert schema conditions to where method SQL string
func (q *DBQuery) condToWhereSQL(cond map[string]any) (string, any) {
	where := strings.Builder{}

	column1, _ := cond["column1"].(string)
	column2, _ := cond["column2"].(string)
	operator, _ := cond["operator"].(string)
	if operator == "" {
		operator = "="
	}
	arg, isValueExists := cond["value"]
	argStr, _ := arg.(string)
	isNullSQL := false

	isOperatorIN := strings.Contains(strings.ToUpper(operator), "IN")
	isOperatorLIKE := strings.Contains(strings.ToUpper(operator), "LIKE")
	isCaseInsensitive, _ := cond["isCaseInsensitive"].(bool)

	column1jsonKey, _ := cond["column1jsonKey"].(string)
	column1type, _ := cond["column1type"].(string)
	column1type = strings.ToLower(column1type)
	column1isBool := strings.Contains(column1type, "bool")
	column1isDateTime := strings.Contains(column1type, "date") || strings.Contains(column1type, "time")

	isInvalidUUID := strings.Contains(column1type, "uuid") && argStr != "" && validator.New().Var(argStr, "uuid") != nil
	if column1jsonKey != "" {
		column1 = q.QuoteJSON(column1, column1jsonKey)
	}
	if column1 != "" {
		// quote table name if not from sub query
		if column1jsonKey == "" && !strings.Contains(column1, " ") && !strings.Contains(column1, "(") {
			column1 = q.DB.Statement.Quote(column1)
		}
		if (column1isDateTime && isOperatorLIKE) || isInvalidUUID {
			column1 = "CAST(" + column1 + " AS CHAR(36))"
		}
		if isCaseInsensitive {
			column1 = "LOWER(" + column1 + ")"
		}
		where.WriteString(column1)
	}

	if isValueExists && (arg == nil || strings.ToLower(argStr) == "null") {
		isNullSQL = true
		where.WriteString(" IS")
		if operator != "=" {
			where.WriteString(" NOT")
		}
		where.WriteString(" NULL")
	} else if isOperatorIN || isOperatorLIKE {
		sqlInLikeOpt := strings.ToUpper(operator)
		if !strings.Contains(operator, " ") {
			sqlInLikeOpt = " " + sqlInLikeOpt + " "
		}
		where.WriteString(sqlInLikeOpt)
	} else {
		where.WriteString(operator)
	}

	column2jsonKey, _ := cond["column2jsonKey"].(string)
	column2type, _ := cond["column2type"].(string)
	column2type = strings.ToLower(column2type)
	column2isDateTime := strings.Contains(column2type, "date") || strings.Contains(column2type, "time")
	if column2jsonKey != "" {
		column2 = q.QuoteJSON(column2, column2jsonKey)
	}
	if column2 != "" {
		// quote table name if not from sub query
		if column2jsonKey == "" && !strings.Contains(column2, " ") && !strings.Contains(column2, "(") {
			column2 = q.DB.Statement.Quote(column2)
		}
		if column2isDateTime && isOperatorLIKE {
			column2 = "CAST(" + column2 + " AS CHAR(36))"
		}
		if isCaseInsensitive {
			column2 = "LOWER(" + column1 + ")"
		}
		where.WriteString(column2)
	} else if isOperatorIN {
		where.WriteString("(?)")
		if argStr != "" {
			arg = strings.Split(argStr, ",")
		}
	} else if !isNullSQL {
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

// SetGroup specify the group method when querying
func (q *DBQuery) SetGroup(db *gorm.DB, schema map[string]any, query url.Values) *gorm.DB {
	fields, _ := schema["fields"].(map[string]map[string]any)

	// group from schema
	groups, _ := schema["groups"].(map[string]string)

	// group from query $group
	queryGroups := strings.Split(query.Get(QueryGroup), ",")
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

	// group from query $select with aggregation
	querySelect := strings.Split(query.Get(QuerySelect), ",")
	if len(querySelect) > 0 {
		isAggFunc := false
		for _, k := range querySelect {
			agg := strings.Split(k, ":")
			aggFunc := q.qsToAggFuncSQL(agg[0])
			if aggFunc != "" {
				isAggFunc = true
			}
		}
		if isAggFunc {
			for _, k := range querySelect {
				group, ok := fields[k]["db"].(string)
				if ok {
					if groups != nil {
						groups[k] = group
					} else {
						groups = map[string]string{k: group}
					}
				}
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

// SetSelect specify fields that you want when querying
func (q *DBQuery) SetSelect(db *gorm.DB, schema map[string]any, query url.Values) *gorm.DB {
	selectedFields := []string{}
	fieldOrder, _ := schema["fieldOrder"].([]string)
	fields, _ := schema["fields"].(map[string]map[string]any)
	querySelect := strings.Split(query.Get(QuerySelect), ",")
	querySelect = append(querySelect, strings.Split(query.Get(QueryGroup), ",")...)
	if len(querySelect) > 0 && schema["is_skip_query_select"] == nil {
		for _, k := range querySelect {
			field, ok := fields[k]["db"].(string)
			if ok {
				selectedFields = q.addSelect(selectedFields, field, k)
			} else {
				agg := strings.Split(k, ":")
				aggFunc := q.qsToAggFuncSQL(agg[0])
				if aggFunc != "" {
					aggField := ""
					aggAlias := aggFunc
					if len(agg) > 1 {
						field, ok = fields[agg[1]]["db"].(string)
						if ok {
							aggField = aggFunc + "(" + field + ")"
							aggAlias = strings.ToLower(aggFunc) + "_" + agg[1]
						}
					} else if agg[0] == QueryCount {
						aggField = "COUNT(*)"
						aggAlias = "count_all"
					}
					if aggField != "" {
						selectedFields = q.addSelect(selectedFields, aggField, aggAlias)
					}
				}
			}
		}
	}
	if len(selectedFields) == 0 {
		queryExclude := strings.Split(query.Get(QueryExclude), ",")
		for _, k := range fieldOrder {
			if slices.Contains(queryExclude, k) {
				continue
			}
			f, _ := fields[k]
			field, ok := f["db"].(string)
			isHide, _ := f["isHide"].(bool)
			if ok && !isHide {
				selectedFields = q.addSelect(selectedFields, field, k)
			}
		}
	}
	return db.Select(strings.Join(selectedFields, ", "))
}

// qsToAggFuncSQL return aggregate function SQL string from part of query params value
func (q *DBQuery) qsToAggFuncSQL(key string) string {
	opt := map[string]string{
		QueryCount: "COUNT",
		QuerySum:   "SUM",
		QueryMin:   "MIN",
		QueryMax:   "MAX",
		QueryAvg:   "AVG",
	}
	aggFuncSQL, _ := opt[key]
	return aggFuncSQL
}

// addSelect append quoted select SQL string
func (q *DBQuery) addSelect(selectedFields []string, field, alias string) []string {
	// quote table name if not from sub query
	if !strings.Contains(field, " ") && !strings.Contains(field, "(") {
		field = q.DB.Statement.Quote(field)
	}
	return append(selectedFields, field+" AS "+q.Quote(alias))
}

// SetOrder specify order method when retrieve records
func (q *DBQuery) SetOrder(db *gorm.DB, schema map[string]any, query url.Values) *gorm.DB {
	fields, _ := schema["fields"].(map[string]map[string]any)
	sorts, _ := schema["sorts"].([]map[string]any)
	hasQuerySort := false
	querySorts := strings.Split(query.Get(QuerySort), ",")
	for _, s := range querySorts {
		srt := map[string]any{"direction": "asc"}
		if strings.HasPrefix(s, "-") {
			s = s[1:]
			srt["direction"] = "desc"
		}
		if strings.HasSuffix(s, ":i") {
			s = s[0 : len(s)-2]
			srt["isCaseInsensitive"] = true
		}
		field, ok := fields[s]["db"].(string)
		if ok {
			srt["column"] = field
		} else {
			for k, v := range fields {
				if strings.HasPrefix(s, k) {
					srt["column"] = k
					fType, ok := v["type"].(string)
					if ok && strings.Contains(strings.ToLower(fType), "json") {
						srt["jsonKey"] = strings.Replace(s, k+".", "", 1)
					}
				}
			}
		}
		if srt["column"] != nil {
			hasQuerySort = true
			orderBySQL := q.sortToOrderBySQL(srt)
			if orderBySQL != "" {
				db = db.Order(orderBySQL)
			}
		}
	}

	for _, srt := range sorts {
		isRequired, _ := srt["isRequired"].(bool)
		if isRequired || !hasQuerySort {
			orderBySQL := q.sortToOrderBySQL(srt)
			if orderBySQL != "" {
				db = db.Order(orderBySQL)
			}
		}
	}

	return db
}

// sortToOrderBySQL convert schema sorts to order by method SQL string
func (q *DBQuery) sortToOrderBySQL(srt map[string]any) string {
	column, _ := srt["column"].(string)
	if column == "" {
		return column
	}
	jsonKey, _ := srt["jsonKey"].(string)
	if jsonKey != "" {
		column = q.QuoteJSON(column, jsonKey)
	} else if !strings.Contains(column, " ") && !strings.Contains(column, "(") {
		column = q.DB.Statement.Quote(column)
	}
	isCaseInsensitive, _ := srt["isCaseInsensitive"].(bool)
	if isCaseInsensitive {
		column = "LOWER(" + column + ")"
	}
	direction, _ := srt["direction"].(string)
	if direction == "" {
		direction = "ASC"
	}

	return column + " " + strings.ToUpper(direction)
}

// SetPagination specify limit & offset method when retrieve records
func (q *DBQuery) SetPagination(db *gorm.DB, query url.Values) *gorm.DB {
	page, limit := q.GetPageLimit(query)
	if limit == 0 {
		return db
	}
	return db.Limit(limit).Offset((page - 1) * limit)
}

// GetPageLimit return desire page & limit from query params
func (q *DBQuery) GetPageLimit(qry ...url.Values) (int, int) {
	query := q.Query
	if len(qry) > 0 {
		query = qry[0]
	}

	if query.Get(QueryDisablePagination) == "true" {
		return 0, 0
	}
	page := 1
	limit := QueryDefaultLimit
	if query.Get(QueryPage) != "" {
		pageTemp, _ := strconv.Atoi(query.Get(QueryPage))
		if pageTemp > 0 {
			page = pageTemp
		}
	}
	if query.Get(QueryLimit) != "" {
		limitTemp, _ := strconv.Atoi(query.Get(QueryLimit))
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

// Quote returns quoted SQL string
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

// QuoteJSON returns quoted json extract SQL string for json column with json key
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

// NewUUIDSQL returns uuid SQL string
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
