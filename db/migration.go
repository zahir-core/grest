package db

import (
	"encoding/json"
	"errors"
)

var dbMigration = map[string]map[string]interface{}{}

type MigrationTable interface {
	TableName() string
	KeyField() string
	ValueField() string
	MigrationKey() string
}

func RegisterTable(param interface{}) error {
	t, ok := param.(interface {
		ConnName() string
		TableName() string
		TableVersion() string
	})
	if !ok {
		return errors.New("RegisterTable: param has no method ConnName or TableName")
	}
	cfg, _ := dbMigration[t.ConnName()]
	cfg[t.TableName()] = param

	dbMigration[t.ConnName()] = cfg

	return nil
}

func Migrate(connName string) error {
	conn, ok := dbConn[connName]
	if ok {
		return errors.New("DB connection " + connName + " is not found")
	}

	db, err := DB(connName)
	if err != nil {
		return err
	}

	mt := conn.migrationTable
	where := map[string]interface{}{
		mt.KeyField(): mt.MigrationKey(),
	}

	migrationData := map[string]interface{}{}
	db.Table(mt.TableName()).Select(mt.ValueField() + " as migration_key").Where(where).First(&migrationData)

	migrationMap := map[string]string{}
	migrationJsonString, skOK := migrationData["migration_key"]
	if skOK {
		json.Unmarshal([]byte(migrationJsonString.(string)), &migrationMap)
	}

	dm, dmOK := dbMigration[connName]
	if dmOK {
		migrationDb, err := DB(connName)
		if err != nil {
			return err
		}
		for tableName, tableStruct := range dm {
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
				err := migrationDb.AutoMigrate(&tableStruct)
				if err != nil {
					return err
				}
				migrationMap[tableName] = tableVersion
			}
		}
		migrationJson, err := json.Marshal(migrationMap)
		if err == nil {
			if skOK {
				db.Table(mt.TableName()).Where(where).Update(mt.ValueField(), string(migrationJson))
			} else {
				migrationData[mt.KeyField()] = mt.MigrationKey()
				migrationData[mt.ValueField()] = string(migrationJson)
				db.Table(mt.TableName()).Create(migrationData)
			}
		}
	}

	return nil
}
