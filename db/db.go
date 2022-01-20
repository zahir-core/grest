package db

import (
	"errors"
	"time"

	"gorm.io/gorm"
)

var (
	dbConfig     = map[string]Config{}
	dbConnection = map[string]*gorm.DB{}
)

type Config struct {
	Driver   string
	Host     string
	Port     int
	User     string
	Password string
	DbName   string

	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
	ConnMaxIdleTime time.Duration

	Protocol     string
	Charset      string
	TimeZone     *time.Location // see https://pkg.go.dev/time#LoadLocation for details
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	SslMode      string

	OtherParams map[string]string

	dialector  gorm.Dialector
	gormConfig *gorm.Config

	MigrationTable interface{}
	SeedTable      interface{}

	IsDebug bool
}

type SettingTable struct {
	Key       string `gorm:"primaryKey"`
	Value     string
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (SettingTable) TableName() string {
	return "settings"
}

func (SettingTable) KeyField() string {
	return "key"
}

func (SettingTable) ValueField() string {
	return "value"
}

func (SettingTable) MigrationKey() string {
	return "table_versions"
}

func (SettingTable) SeedKey() string {
	return "executed_seeds"
}

func Configure(connName string, c Config, dialector gorm.Dialector, gormConfig *gorm.Config) {
	c.dialector = dialector
	c.gormConfig = gormConfig
	// todo: validate if response of KeyField() & ValueField() is exist on struct field
	if _, ok := c.MigrationTable.(MigrationTabler); !ok {
		c.MigrationTable = SettingTable{}
	}
	// todo: validate if response of KeyField() & ValueField() is exist on struct field
	if _, ok := c.SeedTable.(SeedTabler); !ok {
		c.SeedTable = SettingTable{}
	}
	dbConfig[connName] = c
}

func DB(connName string) (*gorm.DB, error) {
	dbConn, ok := dbConnection[connName]
	if ok {
		return dbConn, nil
	}
	dbConf, ok := dbConfig[connName]
	if ok {
		return Connect(connName, dbConf)
	}
	return nil, errors.New("DB config for " + connName + " is not found")
}

func Connect(connName string, c Config) (*gorm.DB, error) {
	db, err := gorm.Open(c.dialector, c.gormConfig)
	if err != nil {
		return nil, err
	}

	if c.IsDebug {
		db = db.Debug()
	}

	dbConn, err := db.DB()
	if err != nil {
		return nil, err
	}
	if c.MaxOpenConns != 0 {
		dbConn.SetMaxOpenConns(c.MaxOpenConns)
	}
	if c.MaxIdleConns != 0 {
		dbConn.SetMaxIdleConns(c.MaxIdleConns)
	}
	if c.ConnMaxLifetime != 0 {
		dbConn.SetConnMaxLifetime(c.ConnMaxLifetime)
	}
	if c.ConnMaxIdleTime != 0 {
		dbConn.SetConnMaxIdleTime(c.ConnMaxIdleTime)
	}

	dbConnection[connName] = db
	return db, nil
}

func Close() {
	for _, db := range dbConnection {
		dbConn, err := db.DB()
		if err == nil {
			dbConn.Close()
		}
	}
}
