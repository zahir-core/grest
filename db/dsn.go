package db

import (
	"net/url"
	"strconv"
	"strings"
)

func (c *Config) DSN() string {
	switch c.Driver {
	case "mysql":
		return c.MySqlDSN()
	case "postgres":
		return c.PostgreSqlDSN()
	case "sqlserver":
		return c.SqlServerDSN()
	case "firebird":
		return c.FirebirdDSN()
	case "clickhouse":
		return c.ClickhouseDSN()
	case "sqlite":
		return c.SqliteDSN()
	default:
		return c.MySqlDSN()
	}
}

// https://github.com/go-sql-driver/mysql#dsn-data-source-name
// example : "user:pass@tcp(127.0.0.1:3306)/dbname?parseTime=true&loc=Asia/Jakarta"
func (c *Config) MySqlDSN() string {
	var s strings.Builder

	if c.User != "" {
		s.WriteString(c.User)
		if c.Password != "" {
			s.WriteByte(':')
			s.WriteString(c.Password)
		}
		s.WriteByte('@')
	}

	if c.Protocol == "" {
		c.Protocol = "tcp"
	}
	s.WriteString(c.Protocol)
	if c.Host != "" {
		s.WriteByte('(')
		s.WriteString(c.Host)
		if c.Port != 0 {
			s.WriteByte(':')
			s.WriteString(strconv.Itoa(c.Port))
		}
		s.WriteByte(')')
	}

	s.WriteByte('/')
	s.WriteString(c.DbName)
	s.WriteByte('?')

	qs := url.Values{}
	qs.Add("parseTime", "true")
	if c.TimeZone != nil {
		qs.Add("loc", c.TimeZone.String())
	}
	if c.OtherParams != nil {
		for key, value := range c.OtherParams {
			qs.Add(key, value)
		}
	}
	s.WriteString(qs.Encode())

	return s.String()
}

// https://www.postgresql.org/docs/current/libpq-connect.html#LIBPQ-PARAMKEYWORDS
// example: "host=localhost port=9920 user=postgres password=postgres dbname=postgres sslmode=disable TimeZone=Asia/Jakarta"
func (c *Config) PostgreSqlDSN() string {
	var s strings.Builder

	s.WriteString("dbname=")
	s.WriteString(c.DbName)

	if c.Host != "" {
		writeParam(&s, "host", c.Host)
	}

	if c.Port != 0 {
		writeParam(&s, "port", strconv.Itoa(c.Port))
	}

	if c.User != "" {
		writeParam(&s, "user", c.User)
	}

	if c.Password != "" {
		writeParam(&s, "password", c.Password)
	}

	if c.SslMode == "" {
		c.SslMode = "disable"
	}
	writeParam(&s, "sslmode", c.SslMode)

	if c.TimeZone != nil {
		writeParam(&s, "TimeZone", c.TimeZone.String())
	}

	if c.OtherParams != nil {
		for key, value := range c.OtherParams {
			writeParam(&s, key, value)
		}
	}

	return s.String()
}

func writeParam(s *strings.Builder, key, value string) {
	s.WriteByte(' ')
	s.WriteString(key)
	s.WriteByte('=')
	s.WriteString(value)
}

// https://github.com/denisenkom/go-mssqldb#connection-parameters-and-dsn
// example: "sqlserver://username:password@localhost:9930?database=gorm"
func (c *Config) SqlServerDSN() string {
	var s strings.Builder

	s.WriteString("sqlserver://")

	// [username[:password]@]
	if c.User != "" {
		s.WriteString(c.User)
		if c.Password != "" {
			s.WriteByte(':')
			s.WriteString(c.Password)
		}
		s.WriteByte('@')
	}

	s.WriteString(c.DbName)

	if c.Host != "" {
		s.WriteString(c.Host)
	}

	if c.Port != 0 {
		s.WriteByte(':')
		s.WriteString(strconv.Itoa(c.Port))
	}
	s.WriteByte('?')

	qs := url.Values{}
	qs.Add("database", c.DbName)
	if c.OtherParams != nil {
		for key, value := range c.OtherParams {
			qs.Add(key, value)
		}
	}
	s.WriteString(qs.Encode())

	return s.String()
}

// https://github.com/nakagami/firebirdsql#connection-string
// example: "SYSDBA:masterkey@127.0.0.1:3050/path/to/db_file_or_alias?charset=utf8"
func (c *Config) FirebirdDSN() string {
	var s strings.Builder

	if c.User != "" {
		s.WriteString(c.User)
		if c.Password != "" {
			s.WriteByte(':')
			s.WriteString(c.Password)
		}
		s.WriteByte('@')
	}

	s.WriteString(c.DbName)

	if c.Host != "" {
		s.WriteString(c.Host)
	}

	if c.Port != 0 {
		s.WriteByte(':')
		s.WriteString(strconv.Itoa(c.Port))
	}

	s.WriteByte('/')
	s.WriteString(c.DbName)
	s.WriteByte('?')

	qs := url.Values{}
	if c.Charset == "" {
		c.Charset = "UTF8"
	}
	qs.Add("charset", c.Charset)
	if c.OtherParams != nil {
		for key, value := range c.OtherParams {
			qs.Add(key, value)
		}
	}
	s.WriteString(qs.Encode())

	return s.String()
}

// https://github.com/ClickHouse/clickhouse-go#dsn
// example: "tcp://localhost:9000?database=gorm&username=gorm&password=gorm&read_timeout=10&write_timeout=20"
func (c *Config) ClickhouseDSN() string {
	var s strings.Builder

	if c.Protocol == "" {
		c.Protocol = "tcp"
	}
	s.WriteString("://")

	if c.Host != "" {
		s.WriteString(c.Host)
	}

	if c.Port != 0 {
		s.WriteByte(':')
		s.WriteString(strconv.Itoa(c.Port))
	}
	s.WriteByte('?')

	qs := url.Values{}
	qs.Add("database", c.DbName)
	qs.Add("username", c.User)
	qs.Add("password", c.Password)
	if c.ReadTimeout > 0 {
		qs.Add("read_timeout", c.ReadTimeout.String())
	}
	if c.WriteTimeout > 0 {
		qs.Add("write_timeout", c.WriteTimeout.String())
	}
	if c.OtherParams != nil {
		for key, value := range c.OtherParams {
			qs.Add(key, value)
		}
	}
	s.WriteString(qs.Encode())

	return s.String()
}

func (c *Config) SqliteDSN() string {
	return c.DbName
}
