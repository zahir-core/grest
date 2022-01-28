package db

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

type Bool struct {
	sql.NullBool
}

type Int64 struct {
	sql.NullInt64
}

type Float64 struct {
	sql.NullFloat64
}

type String struct {
	sql.NullString
}

type DateTime struct {
	sql.NullTime
}

type Date struct {
	sql.NullString
}

// GormDataType returns gorm common data type. This type is used for the field's column type.
func (Date) GormDataType() string {
	return "date"
}

// GormDBDataType returns gorm DB data type based on the current using database.
func (Date) GormDBDataType(db *gorm.DB, field *schema.Field) string {
	switch db.Dialector.Name() {
	case "sqlite":
		return "TEXT"
	case "mysql":
		return "DATE"
	case "postgres":
		return "DATE"
	case "sqlserver":
		return "DATE"
	case "firebird":
		return "DATE"
	default:
		return ""
	}
}

// Scan implements sql.Scanner interface and scans value into Date
func (d *Date) Scan(value interface{}) error {
	if value == nil {
		d.String, d.Valid = "", false
		return nil
	}

	switch v := value.(type) {
	case []byte:
		d.String, d.Valid = string(v), true
	case string:
		d.String, d.Valid = v, true
	case time.Time:
		d.String, d.Valid = v.Format("2006-01-02"), true
	default:
		return errors.New(fmt.Sprintf("failed to scan value: %v", v))
	}

	return nil
}

type Time struct {
	sql.NullString
}

// GormDataType returns gorm common data type. This type is used for the field's column type.
func (Time) GormDataType() string {
	return "time"
}

// GormDBDataType returns gorm DB data type based on the current using database.
func (Time) GormDBDataType(db *gorm.DB, field *schema.Field) string {
	switch db.Dialector.Name() {
	case "sqlite":
		return "TEXT"
	case "mysql":
		return "TIME"
	case "postgres":
		return "TIME"
	case "sqlserver":
		return "TIME"
	case "firebird":
		return "TIME"
	default:
		return ""
	}
}

// Scan implements sql.Scanner interface and scans value into Time
func (t *Time) Scan(value interface{}) error {
	if value == nil {
		t.String, t.Valid = "", false
		return nil
	}

	switch v := value.(type) {
	case []byte:
		t.String, t.Valid = string(v), true
	case string:
		t.String, t.Valid = v, true
	case time.Time:
		t.String, t.Valid = v.Format("15:04:05"), true
	default:
		return errors.New(fmt.Sprintf("failed to scan value: %v", v))
	}

	return nil
}

type Text struct {
	sql.NullString
}

// GormDataType returns gorm common data type. This type is used for the field's column type.
func (Text) GormDataType() string {
	return "text"
}

// GormDBDataType returns gorm DB data type based on the current using database.
func (Text) GormDBDataType(db *gorm.DB, field *schema.Field) string {
	switch db.Dialector.Name() {
	case "sqlite":
		return "TEXT"
	case "mysql":
		return "TEXT"
	case "postgres":
		return "TEXT"
	case "sqlserver":
		return "NVARCHAR(MAX)"
	case "firebird":
		return "BLOB SUB_TYPE TEXT"
	default:
		return ""
	}
}

// Scan implements sql.Scanner interface and scans value into Text
func (t *Text) Scan(value interface{}) error {
	if value == nil {
		t.String, t.Valid = "", false
		return nil
	}

	switch v := value.(type) {
	case []byte:
		t.String, t.Valid = string(v), true
	case string:
		t.String, t.Valid = v, true
	default:
		return errors.New(fmt.Sprintf("failed to scan value: %v", v))
	}

	return nil
}

type JSON struct {
	sql.NullString
}

// GormDataType returns gorm common data type. This type is used for the field's column type.
func (JSON) GormDataType() string {
	return "json"
}

// GormDBDataType returns gorm DB data type based on the current using database.
func (JSON) GormDBDataType(db *gorm.DB, field *schema.Field) string {
	switch db.Dialector.Name() {
	case "sqlite":
		return "JSON"
	case "mysql":
		return "JSON"
	case "postgres":
		return "JSONB"
	case "sqlserver":
		return "NVARCHAR(MAX)"
	case "firebird":
		return "BLOB SUB_TYPE TEXT"
	default:
		return ""
	}
}

// Scan implements sql.Scanner interface and scans value into JSON
func (j *JSON) Scan(value interface{}) error {
	if value == nil {
		j.String, j.Valid = "", false
		return nil
	}

	switch v := value.(type) {
	case []byte:
		j.String, j.Valid = string(v), true
	case string:
		j.String, j.Valid = v, true
	default:
		return errors.New(fmt.Sprintf("failed to scan value: %v", v))
	}

	return nil
}

type UUID struct {
	sql.NullString
}

// GormDataType returns gorm common data type. This type is used for the field's column type.
func (UUID) GormDataType() string {
	return "char(36)"
}
