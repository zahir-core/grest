package db

import (
	"os"

	"gopkg.in/DATA-DOG/go-sqlmock.v1"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func NewMockDB() (*gorm.DB, sqlmock.Sqlmock, error) {
	db, mock, err := sqlmock.New()
	if err != nil {
		return nil, mock, err
	}

	gormDB, err := gorm.Open(postgres.New(
		postgres.Config{
			Conn:                 db,
			PreferSimpleProtocol: true,
		},
	), &gorm.Config{})
	if os.Getenv("IS_PRINT_SQL") == "true" {
		gormDB = gormDB.Debug()
	}
	return gormDB, mock, err
}
