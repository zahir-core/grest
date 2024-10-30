package grest

import (
	"encoding/json"
	"net/http"
	"os"
	"sync"

	"gopkg.in/DATA-DOG/go-sqlmock.v1"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// DB is a DB utility to manage database connections, migrations, and seeders.
type DB struct {
	Conns      map[string]*gorm.DB
	Migrations map[string]map[string]Table
	Seeders    map[string]map[string]any
	mu         sync.RWMutex
}

// Table is an interface for database table models.
type Table interface {
	TableName() string
}

// SettingTable is an interface for setting-related tables.
type SettingTable interface {
	Table
	KeyField() string
	ValueField() string
}

// MigrationTable is an interface for migration-related tables.
type MigrationTable interface {
	SettingTable
	MigrationKey() string
}

// SeederTable is an interface for seeder-related tables.
type SeederTable interface {
	SettingTable
	SeederKey() string
}

// NewMockDB creates a new mock database and returns the gorm.DB instance.
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

// RegisterConn registers a database connection.
func (db *DB) RegisterConn(connName string, conn *gorm.DB) {
	db.mu.Lock()
	defer db.mu.Unlock()
	if db.Conns != nil {
		db.Conns[connName] = conn
	} else {
		db.Conns = map[string]*gorm.DB{connName: conn}
	}
}

// Conn retrieves a registered database connection.
func (db *DB) Conn(connName string) (*gorm.DB, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()
	conn, ok := db.Conns[connName]
	if ok {
		return conn, nil
	}
	return nil, NewError(http.StatusInternalServerError, "DB connection "+connName+" is not found")
}

// CloseConn closes a registered database connection.
func (db *DB) CloseConn(connName string) error {
	db.mu.Lock()
	defer db.mu.Unlock()
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

// Close closes all registered database connections.
func (db *DB) Close() {
	db.mu.Lock()
	defer db.mu.Unlock()
	for _, conn := range db.Conns {
		dbSQL, err := conn.DB()
		if err == nil {
			dbSQL.Close()
		}
	}
	db.Conns = map[string]*gorm.DB{}
}

// RegisterTable registers a table for migration.
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

// MigrateTable performs migrations for a specific table.
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

// RegisterSeeder registers a seeder for a specific connection.
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

// RunSeeder runs seeders for a specific connection and seeder table.
func (db *DB) RunSeeder(tx *gorm.DB, connName string, seederTable SeederTable) error {
	q := DBQuery{DB: tx}
	tableName := q.Quote(seederTable.TableName())
	keyField := q.Quote(seederTable.KeyField())
	valueField := q.Quote(seederTable.ValueField())
	seederKey := seederTable.SeederKey()

	seederData := map[string]any{}
	tx.Table(tableName).
		Where(keyField+" = ?", seederKey).
		Select(valueField + " as " + q.Quote("value")).
		Take(&seederData)

	seederMap := map[string]bool{}
	seedJsonString, isSeedStringExist := seederData["value"].(string)
	if isSeedStringExist {
		json.Unmarshal([]byte(seedJsonString), &seederMap)
	}

	registeredSeeders, isRegisteredSeederExist := db.Seeders[connName]
	if isRegisteredSeederExist {
		for key, seeder := range registeredSeeders {
			if _, sdOK := seederMap[key]; !sdOK {
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
				seederMap[key] = true
			}
		}
		for key := range seederMap {
			if _, sdOK := registeredSeeders[key]; !sdOK {
				delete(seederMap, key)
			}
		}
		seederJSON, err := json.Marshal(seederMap)
		if err == nil {
			if isSeedStringExist {
				tx.Table(seederTable.TableName()).
					Where(keyField+" = ?", seederKey).
					Update(valueField, string(seederJSON))
			} else {
				seederData[seederTable.KeyField()] = seederKey
				seederData[seederTable.ValueField()] = string(seederJSON)
				tx.Table(seederTable.TableName()).Create(seederData)
			}
		}
	}

	return nil
}
