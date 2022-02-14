package db

import (
	"errors"

	"gorm.io/gorm"
)

var dbConn = map[string]*gorm.DB{}

func Configure(connName string, db *gorm.DB) {
	dbConn[connName] = db
}

func DB(connName string) (*gorm.DB, error) {
	db, ok := dbConn[connName]
	if ok {
		return db, nil
	}
	return nil, errors.New("DB connection " + connName + " is not found")
}

func Close() {
	for _, db := range dbConn {
		sqlDB, err := db.DB()
		if err == nil {
			sqlDB.Close()
		}
	}
}
