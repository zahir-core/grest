package grest

import (
	"net/url"
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

type DBQuery struct {
	DB    *gorm.DB
	Query url.Values
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
