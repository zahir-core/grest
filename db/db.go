package db

import (
	"errors"
	"time"

	"gorm.io/gorm"
)

var (
	dbConn = map[string]Conn{}
)

type Conn struct {
	db             *gorm.DB
	migrationTable MigrationTable
	seedTable      SeedTable
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

func Configure(connName string, db *gorm.DB, migrationTable MigrationTable, seedTable SeedTable) {
	c := Conn{}
	c.db = db
	c.migrationTable = SettingTable{}
	if migrationTable != nil {
		c.migrationTable = migrationTable
	}
	c.seedTable = SettingTable{}
	if seedTable != nil {
		c.seedTable = seedTable
	}
	dbConn[connName] = c
}

func DB(connName string) (*gorm.DB, error) {
	conn, ok := dbConn[connName]
	if ok {
		return conn.db, nil
	}
	return nil, errors.New("DB connection " + connName + " is not found")
}

func Close() {
	for _, conn := range dbConn {
		db, err := conn.db.DB()
		if err == nil {
			db.Close()
		}
	}
}
