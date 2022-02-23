package db

import (
	"encoding/json"
	"errors"

	"gorm.io/gorm"
)

var dbMigration = map[string]map[string]interface{}{}

type MigrationTableInterface interface {
	TableName() string
	KeyField() string
	ValueField() string
	MigrationKey() string
}

type MigrationTable struct {
	SettingTable
}

func (MigrationTable) MigrationKey() string {
	return "table_versions"
}

func RegisterTable(connName string, tableStruct interface{}) error {
	t, ok := tableStruct.(interface {
		TableName() string
	})
	if !ok {
		return errors.New("RegisterTable: tableStruct has no method TableName")
	}
	cfg, ok := dbMigration[connName]
	if ok {
		cfg[t.TableName()] = tableStruct
	} else {
		cfg = map[string]interface{}{t.TableName(): tableStruct}
	}

	dbMigration[connName] = cfg

	return nil
}

func Migrate(tx *gorm.DB, connName string, migrationTable MigrationTableInterface) error {
	tableName := Quote(tx, migrationTable.TableName())
	keyField := Quote(tx, migrationTable.KeyField())
	valueField := Quote(tx, migrationTable.ValueField())
	migrationKey := migrationTable.MigrationKey()

	err := tx.AutoMigrate(&migrationTable)
	if err != nil {
		return err
	}

	migrationData := map[string]interface{}{}
	tx.Table(tableName).
		Where(keyField+" = ?", migrationKey).
		Select(valueField + " as " + Quote(tx, "value")).
		Take(&migrationData)

	migrationMap := map[string]string{}
	migrationJsonString, isMigrationStringExist := migrationData["value"]
	if isMigrationStringExist {
		json.Unmarshal([]byte(migrationJsonString.(string)), &migrationMap)
	}

	dbMigrations, isDbMigrationExist := dbMigration[connName]
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
					return err
				}
				migrationMap[tableName] = tableVersion
			}
		}
		migrationJson, err := json.Marshal(migrationMap)
		if err == nil {
			if isMigrationStringExist {
				tx.Table(tableName).
					Where(keyField+" = ?", migrationKey).
					Update(valueField, string(migrationJson))
			} else {
				migrationData[keyField] = migrationKey
				migrationData[valueField] = string(migrationJson)
				tx.Table(tableName).Create(migrationData)
			}
		}
	}

	return nil
}
