package db

import (
	"bytes"
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

var nullBytes = []byte("null")

// NullBool is a nullable bool.
// It will marshal to null if null, not false.
// It will unmarshal to true if input value is true, "true", "True", "TRUE", "t", "T", 1, or "1"
// It will unmarshal to false if input value is false, "false", "False", "FALSE", "f", "F", 0, or "0"
// Other input value will be considered null, not false and not error.
// It supports SQL and JSON serialization.
type NullBool struct {
	sql.NullBool
}

func (n *NullBool) Set(val bool) {
	n.Valid = true
	n.Bool = val
}

func (n *NullBool) Val() bool {
	return n.Bool
}

// Scan implements the Scanner interface.
func (n *NullBool) Scan(value interface{}) error {
	if value == nil {
		n.Bool, n.Valid = false, false
		return nil
	}
	n.Valid = true

	nb := sql.NullBool{}
	err := nb.Scan(value)
	if err == nil {
		n.Bool = nb.Bool
		return nil
	}

	ni32 := sql.NullInt32{}
	err = ni32.Scan(value)
	if err == nil && ni32.Int32 == 1 {
		n.Bool = true
		return nil
	}

	ni64 := sql.NullInt64{}
	err = ni64.Scan(value)
	if err == nil && ni64.Int64 == 1 {
		n.Bool = true
		return nil
	}

	ns := sql.NullString{}
	err = ns.Scan(value)
	if err == nil && (ns.String == "1" || ns.String == "t" || ns.String == "T" || ns.String == "true" || ns.String == "True" || ns.String == "TRUE") {
		n.Bool = true
	}
	return nil
}

// Value implements the driver Valuer interface.
func (n NullBool) Value() (driver.Value, error) {
	if !n.Valid {
		return nil, nil
	}
	if n.Bool {
		return "1", nil
	}
	return "0", nil
}

// MarshalText implements encoding.TextMarshaler.
// It will encode blank if this NullBool is null, not false.
func (n NullBool) MarshalText() ([]byte, error) {
	if !n.Valid {
		return []byte{}, nil
	}
	if !n.Bool {
		return []byte("false"), nil
	}
	return []byte("true"), nil
}

// MarshalJSON implements json.Marshaler.
// It will encode null if this Bool is null, not false.
func (n NullBool) MarshalJSON() ([]byte, error) {
	if !n.Valid {
		return nullBytes, nil
	}
	return json.Marshal(n.Bool)
}

// UnmarshalText implements encoding.TextUnmarshaler.
// It will unmarshal to true if input value is true, "true", "True", "TRUE", "t", "T", 1, or "1"
// It will unmarshal to false if input value is false, "false", "False", "FALSE", "f", "F", 0, or "0"
// Other input value will be considered null, not false.
func (n *NullBool) UnmarshalText(text []byte) error {
	str := string(text)
	switch str {
	case "1", "t", "T", "true", "TRUE", "True":
		n.Bool = true
	case "0", "f", "F", "false", "FALSE", "False":
		n.Bool = false
	default:
		n.Valid = false
		return nil
	}
	n.Valid = true
	return nil
}

// UnmarshalJSON implements json.Unmarshaler.
// It will unmarshal to true if input value is true, "true", "True", "TRUE", "t", "T", 1, or "1"
// It will unmarshal to false if input value is false, "false", "False", "FALSE", "f", "F", 0, or "0"
// Other input value will be considered null, not false and not error.
func (n *NullBool) UnmarshalJSON(data []byte) error {
	if bytes.Equal(data, nullBytes) {
		n.Valid = false
		return nil
	}
	if err := json.Unmarshal(data, &n.Bool); err == nil {
		n.Valid = true
	} else {
		var str string
		if err := json.Unmarshal(data, &str); err == nil {
			n.Bool, err = strconv.ParseBool(str)
			if err == nil {
				n.Valid = true
			}
		} else {
			var integer int
			if err := json.Unmarshal(data, &integer); err == nil {
				if integer == 1 {
					n.Bool = true
				}
				if integer == 0 || integer == 1 {
					n.Valid = true
				}
			}
		}
	}
	return nil
}

// IsZero returns true for invalid bool, for omitempty support
func (n NullBool) IsZero() bool {
	return !n.Valid
}

// NullInt64 is a nullable int64.
// It supports integer number and a string that can be converted to a integer number.
// Other input value will be considered null, not 0 and not error.
// It supports SQL and JSON serialization.
type NullInt64 struct {
	sql.NullInt64
}

func (n *NullInt64) Set(val int64) {
	n.Valid = true
	n.Int64 = val
}

func (n *NullInt64) Val() int64 {
	return n.Int64
}

// MarshalText implements encoding.TextMarshaler.
// It will encode a blank string if this NullInt64 is null.
func (n NullInt64) MarshalText() ([]byte, error) {
	if !n.Valid {
		return []byte{}, nil
	}
	return []byte(strconv.FormatInt(n.Int64, 10)), nil
}

// MarshalJSON implements json.Marshaler.
// It will encode null if this NullInt64 is null.
func (n NullInt64) MarshalJSON() ([]byte, error) {
	if !n.Valid {
		return nullBytes, nil
	}
	return json.Marshal(n.Int64)
}

// UnmarshalText implements encoding.TextUnmarshaler.
// It supports integer number and a string that can be converted to a integer number.
// Other input value will be considered null, not 0 and not error.
func (n *NullInt64) UnmarshalText(text []byte) error {
	str := string(text)
	if str == "" || str == "null" {
		return nil
	}
	var err error
	n.Int64, err = strconv.ParseInt(str, 10, 64)
	if err == nil {
		n.Valid = true
	}
	return nil
}

// UnmarshalJSON implements json.Unmarshaler.
// It supports integer number and a string that can be converted to a integer number.
// Other input value will be considered null, not 0 and not error.
func (n *NullInt64) UnmarshalJSON(data []byte) error {
	if bytes.Equal(data, nullBytes) {
		return nil
	}
	if err := json.Unmarshal(data, &n.Int64); err == nil {
		n.Valid = true
	} else {
		var str string
		if err := json.Unmarshal(data, &str); err == nil {
			n.Int64, err = strconv.ParseInt(str, 10, 64)
			if err == nil {
				n.Valid = true
			}
		}
	}
	return nil
}

// IsZero returns true for invalid int64, for omitempty support
func (n NullInt64) IsZero() bool {
	return !n.Valid
}

// NullFloat64 is a nullable float64.
// It supports number and a string that can be converted to a number.
// Other input value will be considered null, not 0 and not error.
// It supports SQL and JSON serialization.
type NullFloat64 struct {
	sql.NullFloat64
}

func (n *NullFloat64) Set(val float64) {
	n.Valid = true
	n.Float64 = val
}

func (n *NullFloat64) Val() float64 {
	return n.Float64
}

// MarshalText implements encoding.TextMarshaler.
// It will encode a blank string if this NullFloat64 is null.
func (n NullFloat64) MarshalText() ([]byte, error) {
	if !n.Valid {
		return []byte{}, nil
	}
	return []byte(strconv.FormatFloat(n.Float64, 'f', -1, 64)), nil
}

// MarshalJSON implements json.Marshaler.
// It will encode null if this NullFloat64 is null.
func (n NullFloat64) MarshalJSON() ([]byte, error) {
	if !n.Valid {
		return nullBytes, nil
	}
	return json.Marshal(n.Float64)
}

// UnmarshalText implements encoding.TextUnmarshaler.
// It supports number and a string that can be converted to a number.
// Other input value will be considered null, not 0 and not error.
func (n *NullFloat64) UnmarshalText(text []byte) error {
	str := string(text)
	if str == "" || str == "null" {
		return nil
	}
	var err error
	n.Float64, err = strconv.ParseFloat(str, 64)
	if err == nil {
		n.Valid = true
	}
	return nil
}

// UnmarshalJSON implements json.Unmarshaler.
// It supports number and a string that can be converted to a number.
// Other input value will be considered null, not 0 and not error.
func (n *NullFloat64) UnmarshalJSON(data []byte) error {
	if bytes.Equal(data, nullBytes) {
		return nil
	}
	if err := json.Unmarshal(data, &n.Float64); err == nil {
		n.Valid = true
	} else {
		var str string
		if err := json.Unmarshal(data, &str); err == nil {
			n.Float64, err = strconv.ParseFloat(str, 64)
			if err == nil {
				n.Valid = true
			}
		}
	}
	return nil
}

// IsZero returns true for invalid float64, for omitempty support
func (n NullFloat64) IsZero() bool {
	return !n.Valid
}

// NullString is a nullable string.
// It supports SQL and JSON serialization.
type NullString struct {
	sql.NullString
}

func (n *NullString) Set(val string) {
	n.Valid = true
	n.String = val
}

func (n *NullString) Val() string {
	return n.String
}

// MarshalText implements encoding.TextMarshaler.
// It will encode a blank string if this NullString is null.
func (n NullString) MarshalText() ([]byte, error) {
	if !n.Valid {
		return []byte{}, nil
	}
	return []byte(n.String), nil
}

// MarshalJSON implements json.Marshaler.
// It will encode null if this NullString is null.
func (n NullString) MarshalJSON() ([]byte, error) {
	if !n.Valid {
		return nullBytes, nil
	}
	return json.Marshal(n.String)
}

// UnmarshalText implements encoding.TextUnmarshaler.
func (n *NullString) UnmarshalText(text []byte) error {
	str := string(text)
	if str == "null" {
		return nil
	}
	n.String = str
	n.Valid = true
	return nil
}

// UnmarshalJSON implements json.Unmarshaler.
func (n *NullString) UnmarshalJSON(data []byte) error {
	if bytes.Equal(data, nullBytes) {
		return nil
	}
	if err := json.Unmarshal(data, &n.String); err == nil {
		n.Valid = true
	}
	return nil
}

// IsZero returns true for invalid string, for omitempty support
func (n NullString) IsZero() bool {
	return !n.Valid
}

// GormDataType returns gorm common data type. This type is used for the field's column type.
func (NullString) GormDataType() string {
	return "varchar(255)"
}

// GormDBDataType returns gorm DB data type based on the current using database.
func (NullString) GormDBDataType(db *gorm.DB, field *schema.Field) string {
	return "varchar(255)"
}

// NullDateTime is a nullable time.Time.
// It supports SQL and JSON serialization.
type NullDateTime struct {
	sql.NullTime
}

func (n *NullDateTime) Set(val time.Time) {
	n.Valid = true
	n.Time = val
}

func (n *NullDateTime) Val() time.Time {
	return n.Time
}

// MarshalJSON implements json.Marshaler.
// It will encode null if this NullDateTime is null.
func (n NullDateTime) MarshalJSON() ([]byte, error) {
	if !n.Valid {
		return nullBytes, nil
	}
	return json.Marshal(n.Time)
}

// UnmarshalJSON implements json.Unmarshaler.
// It supports a string that can be parsed to a time.Time.
// Other input value will be considered null, not error.
func (n *NullDateTime) UnmarshalJSON(data []byte) error {
	if bytes.Equal(data, nullBytes) {
		n.Valid = false
		return nil
	}
	if err := json.Unmarshal(data, &n.Time); err == nil {
		n.Valid = true
	}
	return nil
}

// IsZero returns true for zero time, for omitempty support
func (n NullDateTime) IsZero() bool {
	return n.Time.IsZero()
}

// NullDate is a nullable date.
// It supports SQL and JSON serialization.
type NullDate struct {
	NullString
}

func (n *NullDate) Set(val string) {
	n.Valid = true
	n.String = val
}

func (n *NullDate) Val() string {
	return n.String
}

// UnmarshalJSON implements json.Unmarshaler.
// It supports a string that can be parsed to a time.Time.
// Other input value will be considered null, not error.
func (n *NullDate) UnmarshalJSON(data []byte) error {
	if bytes.Equal(data, nullBytes) {
		n.Valid = false
		return nil
	}
	if err := json.Unmarshal(data, &n.String); err == nil {
		_, err := time.Parse("2006-01-02", n.String)
		if err == nil {
			n.Valid = true
		}
	}
	return nil
}

// IsZero returns true for zero time, for omitempty support
func (n NullDate) IsZero() bool {
	return !n.Valid
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

// NullDate is a nullable date.
// It supports SQL and JSON serialization.
type NullTime struct {
	NullString
}

func (n *NullTime) Set(val string) {
	n.Valid = true
	n.String = val
}

func (n *NullTime) Val() string {
	return n.String
}

// UnmarshalJSON implements json.Unmarshaler.
// It supports a string that can be parsed to a time.Time.
// Other input value will be considered null, not error.
func (n *NullTime) UnmarshalJSON(data []byte) error {
	if bytes.Equal(data, nullBytes) {
		n.Valid = false
		return nil
	}
	if err := json.Unmarshal(data, &n.String); err == nil {
		_, err := time.Parse("15:04:05", n.String)
		if err == nil {
			n.Valid = true
		}
	}
	return nil
}

// IsZero returns true for zero time, for omitempty support
func (n NullTime) IsZero() bool {
	return !n.Valid
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

func (n *NullText) Set(val string) {
	n.Valid = true
	n.String = val
}

func (n *NullText) Val() string {
	return n.String
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

func (n *NullJSON) Set(val string) {
	n.Valid = true
	n.String = val
}

func (n *NullJSON) Val() string {
	return n.String
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

func (n *NullUUID) Set(val string) {
	n.Valid = true
	n.String = val
}

func (n *NullUUID) Val() string {
	return n.String
}

// GormDataType returns gorm common data type. This type is used for the field's column type.
func (NullUUID) GormDataType() string {
	return "char(36)"
}
