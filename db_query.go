package grest

import (
	"net/url"
	"strings"

	"gorm.io/gorm"
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
