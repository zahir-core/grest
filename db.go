package grest

import (
	"encoding/json"
	"net/http"
	"os"

	"gopkg.in/DATA-DOG/go-sqlmock.v1"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type DB struct {
	Conns      map[string]*gorm.DB
	Migrations map[string]map[string]Table
	Seeders    map[string]map[string]any
}

type Table interface {
	TableName() string
}

type SettingTable interface {
	Table
	KeyField() string
	ValueField() string
}

type MigrationTable interface {
	SettingTable
	MigrationKey() string
}

type SeederTable interface {
	SettingTable
	SeederKey() string
}

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

func (db *DB) RegisterConn(connName string, conn *gorm.DB) {
	if db.Conns != nil {
		db.Conns[connName] = conn
	} else {
		db.Conns = map[string]*gorm.DB{connName: conn}
	}
}

func (db *DB) Conn(connName string) (*gorm.DB, error) {
	conn, ok := db.Conns[connName]
	if ok {
		return conn, nil
	}
	return nil, NewError(http.StatusInternalServerError, "DB connection "+connName+" is not found")
}

func (db *DB) CloseConn(connName string) error {
	conn, ok := db.Conns[connName]
	if ok {
		dbSQL, err := conn.DB()
		if err == nil {
			dbSQL.Close()
		}
		delete(db.Conns, connName)
		return err
	}
	return NewError(http.StatusInternalServerError, "DB connection "+connName+" is not found")
}

func (db *DB) Close() {
	for _, conn := range db.Conns {
		dbSQL, err := conn.DB()
		if err == nil {
			dbSQL.Close()
		}
	}
	db.Conns = map[string]*gorm.DB{}
}

func (db *DB) RegisterTable(connName string, t Table) error {
	m, ok := db.Migrations[connName]
	if ok {
		m[t.TableName()] = t
	} else {
		m = map[string]Table{t.TableName(): t}
	}
	if db.Migrations != nil {
		db.Migrations[connName] = m
	} else {
		db.Migrations = map[string]map[string]Table{connName: m}
	}
	return nil
}

func (db *DB) MigrateTable(tx *gorm.DB, connName string, mTable MigrationTable) error {
	q := DBQuery{DB: tx}
	tableName := q.Quote(mTable.TableName())
	keyField := q.Quote(mTable.KeyField())
	valueField := q.Quote(mTable.ValueField())
	migrationKey := mTable.MigrationKey()

	err := tx.AutoMigrate(&mTable)
	if err != nil {
		return NewError(http.StatusInternalServerError, err.Error())
	}

	mData := map[string]any{}
	tx.Table(tableName).
		Where(keyField+" = ?", migrationKey).
		Select(valueField + " as " + q.Quote("value")).
		Take(&mData)

	migrationMap := map[string]string{}
	migrationJsonString, isMigrationStringExist := mData["value"].(string)
	if isMigrationStringExist {
		json.Unmarshal([]byte(migrationJsonString), &migrationMap)
	}

	dbMigrations, isDbMigrationExist := db.Migrations[connName]
	if isDbMigrationExist {
		for tableName, tableStruct := range dbMigrations {
			tableVersion := "init"
			tv, tvOK := tableStruct.(interface {
				TableVersion() string
			})
			if tvOK {
				tableVersion = tv.TableVersion()
			}

			existingTableVersion := ""
			md, mdOK := migrationMap[tableName]
			if mdOK {
				existingTableVersion = md
			}

			if tableVersion != existingTableVersion {
				err := tx.AutoMigrate(&tableStruct)
				if err != nil {
					return NewError(http.StatusInternalServerError, err.Error())
				}
				migrationMap[tableName] = tableVersion
			}
		}
		migrationJson, err := json.Marshal(migrationMap)
		if err == nil {
			if isMigrationStringExist {
				tx.Table(mTable.TableName()).
					Where(keyField+" = ?", migrationKey).
					Update(valueField, string(migrationJson))
			} else {
				mData[mTable.KeyField()] = migrationKey
				mData[mTable.ValueField()] = string(migrationJson)
				tx.Table(mTable.TableName()).Create(mData)
			}
		}
	}
	return nil
}

func (db *DB) RegisterSeeder(connName, seederKey string, seederHandler any) error {
	sh, ok := db.Seeders[connName]
	if ok {
		sh[seederKey] = seederHandler
	} else {
		sh = map[string]any{seederKey: seederHandler}
	}

	if db.Seeders != nil {
		db.Seeders[connName] = sh
	} else {
		db.Seeders = map[string]map[string]any{connName: sh}
	}

	return nil
}

func (db *DB) RunSeeder(tx *gorm.DB, connName string, seedTable SeederTable) error {
	q := DBQuery{DB: tx}
	tableName := q.Quote(seedTable.TableName())
	keyField := q.Quote(seedTable.KeyField())
	valueField := q.Quote(seedTable.ValueField())
	seederKey := seedTable.SeederKey()

	seedData := map[string]any{}
	tx.Table(tableName).
		Where(keyField+" = ?", seederKey).
		Select(valueField + " as " + q.Quote("value")).
		Take(&seedData)

	seedMap := map[string]bool{}
	seedJsonString, isSeedStringExist := seedData["value"].(string)
	if isSeedStringExist {
		json.Unmarshal([]byte(seedJsonString), &seedMap)
	}

	registeredSeeders, isRegisteredSeederExist := db.Seeders[connName]
	if isRegisteredSeederExist {
		for key, seeder := range registeredSeeders {
			if _, sdOK := seedMap[key]; !sdOK {
				seederWithTx, ok := seeder.(func(db *gorm.DB) error)
				if ok {
					err := seederWithTx(tx)
					if err != nil {
						return NewError(http.StatusInternalServerError, err.Error())
					}
				} else {
					seeder, ok := seeder.(func() error)
					if ok {
						err := seeder()
						if err != nil {
							return NewError(http.StatusInternalServerError, err.Error())
						}
					}
				}
				seedMap[key] = true
			}
		}
		seedJson, err := json.Marshal(seedMap)
		if err == nil {
			if isSeedStringExist {
				tx.Table(seedTable.TableName()).
					Where(keyField+" = ?", seederKey).
					Update(valueField, string(seedJson))
			} else {
				seedData[seedTable.KeyField()] = seederKey
				seedData[seedTable.ValueField()] = string(seedJson)
				tx.Table(seedTable.TableName()).Create(seedData)
			}
		}
	}

	return nil
}
