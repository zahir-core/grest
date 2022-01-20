package db

import (
	"testing"
	"time"
)

var testDSNs = []struct {
	Config   *Config
	Expected string
}{{
	&Config{Driver: "mysql", Host: "127.0.0.1", Port: 3306, User: "username", Password: "password", DbName: "dbname", OtherParams: map[string]string{"param": "value"}},
	"username:password@tcp(127.0.0.1:3306)/dbname?param=value&parseTime=true",
}, {
	&Config{Driver: "mysql", Protocol: "unix", Host: "/path/to/socket", User: "username", Password: "password", DbName: "dbname", OtherParams: map[string]string{"param": "value"}},
	"username:password@unix(/path/to/socket)/dbname?param=value&parseTime=true",
}, {
	&Config{Protocol: "unix", Host: "/tmp/mysql.sock", OtherParams: map[string]string{"arg": "/some/path.ext"}, TimeZone: time.UTC},
	"unix(/tmp/mysql.sock)/?arg=%2Fsome%2Fpath.ext&loc=UTC&parseTime=true",
}, {
	&Config{User: "user", Password: "p@ss(word)", Protocol: "tcp", Host: "[de:ad:be:ef::ca:fe]", DbName: "dbname", TimeZone: time.Local},
	"user:p@ss(word)@tcp([de:ad:be:ef::ca:fe])/dbname?loc=Local&parseTime=true",
},
}

func TestDbConfigDSN(t *testing.T) {
	for _, dsn := range testDSNs {
		result := dsn.Config.DSN()
		if result != dsn.Expected {
			t.Errorf("Expected DSN [%v], got [%v]", dsn.Expected, result)
		}
	}
}

func BenchmarkDSN(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		for _, t := range testDSNs {
			t.Config.DSN()
		}
	}
}
