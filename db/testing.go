package db

import (
	"gopkg.in/DATA-DOG/go-sqlmock.v1"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func NewMockDB(isPrintQuery ...bool) (*gorm.DB, sqlmock.Sqlmock, error) {
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
	if len(isPrintQuery) > 0 && isPrintQuery[0] {
		gormDB = gormDB.Debug()
	}
	return gormDB, mock, err
}
