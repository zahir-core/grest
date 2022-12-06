package grest

import (
	"net/url"
	"strconv"
	"strings"
	"time"
)

type DBConfig struct {
	Driver   string
	Host     string
	Port     int
	User     string
	Password string
	DbName   string

	Protocol     string
	Charset      string
	TimeZone     *time.Location // see https://pkg.go.dev/time#LoadLocation for details
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	SslMode      string

	OtherParams map[string]string
}

func (d *DBConfig) DSN() string {
	switch d.Driver {
	case "mysql":
		return d.MySqlDSN()
	case "postgres":
		return d.PostgreSqlDSN()
	case "sqlserver":
		return d.SqlServerDSN()
	case "firebird":
		return d.FirebirdDSN()
	case "clickhouse":
		return d.ClickhouseDSN()
	case "sqlite":
		return d.SqliteDSN()
	default:
		return d.PostgreSqlDSN()
	}
}

// https://github.com/go-sql-driver/mysql#dsn-data-source-name
// example : "user:pass@tcp(127.0.0.1:3306)/dbname?parseTime=true&loc=Asia/Jakarta"
func (d *DBConfig) MySqlDSN() string {
	var s strings.Builder

	if d.User != "" {
		s.WriteString(d.User)
		if d.Password != "" {
			s.WriteByte(':')
			s.WriteString(d.Password)
		}
		s.WriteByte('@')
	}

	if d.Protocol == "" {
		d.Protocol = "tcp"
	}
	s.WriteString(d.Protocol)
	if d.Host != "" {
		s.WriteByte('(')
		s.WriteString(d.Host)
		if d.Port != 0 {
			s.WriteByte(':')
			s.WriteString(strconv.Itoa(d.Port))
		}
		s.WriteByte(')')
	}

	s.WriteByte('/')
	s.WriteString(d.DbName)
	s.WriteByte('?')

	qs := url.Values{}
	qs.Add("parseTime", "true")
	if d.TimeZone != nil {
		qs.Add("loc", d.TimeZone.String())
	}
	if d.OtherParams != nil {
		for key, value := range d.OtherParams {
			qs.Add(key, value)
		}
	}
	s.WriteString(qs.Encode())

	return s.String()
}

// https://www.postgresql.org/docs/current/libpq-connect.html#LIBPQ-PARAMKEYWORDS
// example: "host=localhost port=9920 user=postgres password=postgres dbname=postgres sslmode=disable TimeZone=Asia/Jakarta"
func (d *DBConfig) PostgreSqlDSN() string {
	var s strings.Builder

	s.WriteString("dbname=")
	s.WriteString(d.DbName)

	if d.Host != "" {
		d.writeParam(&s, "host", d.Host)
	}

	if d.Port != 0 {
		d.writeParam(&s, "port", strconv.Itoa(d.Port))
	}

	if d.User != "" {
		d.writeParam(&s, "user", d.User)
	}

	if d.Password != "" {
		d.writeParam(&s, "password", d.Password)
	}

	if d.SslMode == "" {
		d.SslMode = "disable"
	}
	d.writeParam(&s, "sslmode", d.SslMode)

	if d.TimeZone != nil {
		d.writeParam(&s, "TimeZone", d.TimeZone.String())
	}

	if d.OtherParams != nil {
		for key, value := range d.OtherParams {
			d.writeParam(&s, key, value)
		}
	}

	return s.String()
}

func (*DBConfig) writeParam(s *strings.Builder, key, value string) {
	s.WriteByte(' ')
	s.WriteString(key)
	s.WriteByte('=')
	s.WriteString(value)
}

// https://github.com/denisenkom/go-mssqldb#connection-parameters-and-dsn
// example: "sqlserver://username:password@localhost:9930?database=gorm"
func (d *DBConfig) SqlServerDSN() string {
	var s strings.Builder

	s.WriteString("sqlserver://")

	// [username[:password]@]
	if d.User != "" {
		s.WriteString(d.User)
		if d.Password != "" {
			s.WriteByte(':')
			s.WriteString(d.Password)
		}
		s.WriteByte('@')
	}

	s.WriteString(d.DbName)

	if d.Host != "" {
		s.WriteString(d.Host)
	}

	if d.Port != 0 {
		s.WriteByte(':')
		s.WriteString(strconv.Itoa(d.Port))
	}
	s.WriteByte('?')

	qs := url.Values{}
	qs.Add("database", d.DbName)
	if d.OtherParams != nil {
		for key, value := range d.OtherParams {
			qs.Add(key, value)
		}
	}
	s.WriteString(qs.Encode())

	return s.String()
}

// https://github.com/nakagami/firebirdsql#connection-string
// example: "SYSDBA:masterkey@127.0.0.1:3050/path/to/db_file_or_alias?charset=utf8"
func (d *DBConfig) FirebirdDSN() string {
	var s strings.Builder

	if d.User != "" {
		s.WriteString(d.User)
		if d.Password != "" {
			s.WriteByte(':')
			s.WriteString(d.Password)
		}
		s.WriteByte('@')
	}

	s.WriteString(d.DbName)

	if d.Host != "" {
		s.WriteString(d.Host)
	}

	if d.Port != 0 {
		s.WriteByte(':')
		s.WriteString(strconv.Itoa(d.Port))
	}

	s.WriteByte('/')
	s.WriteString(d.DbName)
	s.WriteByte('?')

	qs := url.Values{}
	if d.Charset == "" {
		d.Charset = "UTF8"
	}
	qs.Add("charset", d.Charset)
	if d.OtherParams != nil {
		for key, value := range d.OtherParams {
			qs.Add(key, value)
		}
	}
	s.WriteString(qs.Encode())

	return s.String()
}

// https://github.com/ClickHouse/clickhouse-go#dsn
// example: "tcp://localhost:9000?database=gorm&username=gorm&password=gorm&read_timeout=10&write_timeout=20"
func (d *DBConfig) ClickhouseDSN() string {
	var s strings.Builder

	if d.Protocol == "" {
		d.Protocol = "tcp"
	}
	s.WriteString("://")

	if d.Host != "" {
		s.WriteString(d.Host)
	}

	if d.Port != 0 {
		s.WriteByte(':')
		s.WriteString(strconv.Itoa(d.Port))
	}
	s.WriteByte('?')

	qs := url.Values{}
	qs.Add("database", d.DbName)
	qs.Add("username", d.User)
	qs.Add("password", d.Password)
	if d.ReadTimeout > 0 {
		qs.Add("read_timeout", d.ReadTimeout.String())
	}
	if d.WriteTimeout > 0 {
		qs.Add("write_timeout", d.WriteTimeout.String())
	}
	if d.OtherParams != nil {
		for key, value := range d.OtherParams {
			qs.Add(key, value)
		}
	}
	s.WriteString(qs.Encode())

	return s.String()
}

func (d *DBConfig) SqliteDSN() string {
	return d.DbName
}
