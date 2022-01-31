package db

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

var nullBytes = []byte("null")

type NullBool struct {
	sql.NullBool
}

func (n NullBool) MarshalJSON() ([]byte, error) {
	if !n.Valid {
		return nullBytes, nil
	}
	return json.Marshal(n.Bool)
}

func (n *NullBool) UnmarshalJSON(data []byte) error {
	if bytes.Equal(data, nullBytes) {
		n.Valid = false
		return nil
	}

	if err := json.Unmarshal(data, &n.Bool); err != nil {
		return fmt.Errorf("couldn't unmarshal JSON: %w", err)
	}

	n.Valid = true
	return nil
}

type NullInt64 struct {
	sql.NullInt64
}

func (n NullInt64) MarshalJSON() ([]byte, error) {
	if !n.Valid {
		return nullBytes, nil
	}
	return json.Marshal(n.Int64)
}

func (n *NullInt64) UnmarshalJSON(data []byte) error {
	if bytes.Equal(data, nullBytes) {
		n.Valid = false
		return nil
	}

	if err := json.Unmarshal(data, &n.Int64); err != nil {
		return fmt.Errorf("couldn't unmarshal JSON: %w", err)
	}

	n.Valid = true
	return nil
}

type NullFloat64 struct {
	sql.NullFloat64
}

func (n NullFloat64) MarshalJSON() ([]byte, error) {
	if !n.Valid {
		return nullBytes, nil
	}
	return json.Marshal(n.Float64)
}

func (n *NullFloat64) UnmarshalJSON(data []byte) error {
	if bytes.Equal(data, nullBytes) {
		n.Valid = false
		return nil
	}

	if err := json.Unmarshal(data, &n.Float64); err != nil {
		return fmt.Errorf("couldn't unmarshal JSON: %w", err)
	}

	n.Valid = true
	return nil
}

type NullString struct {
	sql.NullString
}

func (n NullString) MarshalJSON() ([]byte, error) {
	if !n.Valid {
		return nullBytes, nil
	}
	return json.Marshal(n.String)
}

func (n *NullString) UnmarshalJSON(data []byte) error {
	if bytes.Equal(data, nullBytes) {
		n.Valid = false
		return nil
	}

	if err := json.Unmarshal(data, &n.String); err != nil {
		return fmt.Errorf("couldn't unmarshal JSON: %w", err)
	}

	n.Valid = true
	return nil
}

type NullDateTime struct {
	sql.NullTime
}

func (n NullDateTime) MarshalJSON() ([]byte, error) {
	if !n.Valid {
		return nullBytes, nil
	}
	return json.Marshal(n.Time)
}

func (n *NullDateTime) UnmarshalJSON(data []byte) error {
	if bytes.Equal(data, nullBytes) {
		n.Valid = false
		return nil
	}

	if err := json.Unmarshal(data, &n.Time); err != nil {
		return fmt.Errorf("couldn't unmarshal JSON: %w", err)
	}

	n.Valid = true
	return nil
}

type NullDate struct {
	NullString
}

// GormDataType returns gorm common data type. This type is used for the field's column type.
func (NullDate) GormDataType() string {
	return "date"
}

// GormDBDataType returns gorm DB data type based on the current using database.
func (NullDate) GormDBDataType(db *gorm.DB, field *schema.Field) string {
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
func (d *NullDate) Scan(value interface{}) error {
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

type NullTime struct {
	NullString
}

// GormDataType returns gorm common data type. This type is used for the field's column type.
func (NullTime) GormDataType() string {
	return "time"
}

// GormDBDataType returns gorm DB data type based on the current using database.
func (NullTime) GormDBDataType(db *gorm.DB, field *schema.Field) string {
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
func (t *NullTime) Scan(value interface{}) error {
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

type NullText struct {
	NullString
}

// GormDataType returns gorm common data type. This type is used for the field's column type.
func (NullText) GormDataType() string {
	return "text"
}

// GormDBDataType returns gorm DB data type based on the current using database.
func (NullText) GormDBDataType(db *gorm.DB, field *schema.Field) string {
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

type NullJSON struct {
	NullString
}

// GormDataType returns gorm common data type. This type is used for the field's column type.
func (NullJSON) GormDataType() string {
	return "json"
}

// GormDBDataType returns gorm DB data type based on the current using database.
func (NullJSON) GormDBDataType(db *gorm.DB, field *schema.Field) string {
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

type NullUUID struct {
	NullString
}

// GormDataType returns gorm common data type. This type is used for the field's column type.
func (NullUUID) GormDataType() string {
	return "char(36)"
}
