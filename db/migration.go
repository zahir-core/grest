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

func RegisterTable(connName string, tableStruct interface{}) error {
	t, ok := tableStruct.(interface{ TableName() string })
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

func Migrate(connName string, migrationKey ...string) error {
	conn, ok := dbConn[connName]
	if !ok {
		return errors.New("DB connection " + connName + " is not found")
	}

	mt := conn.migrationTable
	err := conn.db.AutoMigrate(&mt)
	if err != nil {
		return err
	}

	where := map[string]interface{}{
		mt.KeyField(): mt.MigrationKey(),
	}

	migrationData := map[string]interface{}{}
	conn.db.Table(mt.TableName()).Select(mt.ValueField() + " as value").Where(where).Take(&migrationData)

	migrationMap := map[string]string{}
	migrationJsonString, skOK := migrationData["value"]
	if skOK {
		json.Unmarshal([]byte(migrationJsonString.(string)), &migrationMap)
	}

	mKey := connName
	if len(migrationKey) > 0 {
		mKey = migrationKey[0]
	}
	dm, dmOK := dbMigration[mKey]
	if dmOK {
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
				err := conn.db.AutoMigrate(&tableStruct)
				if err != nil {
					return err
				}
				migrationMap[tableName] = tableVersion
			}
		}
		migrationJson, err := json.Marshal(migrationMap)
		if err == nil {
			if skOK {
				conn.db.Table(mt.TableName()).Where(where).Update(mt.ValueField(), string(migrationJson))
			} else {
				migrationData[mt.KeyField()] = mt.MigrationKey()
				migrationData[mt.ValueField()] = string(migrationJson)
				conn.db.Table(mt.TableName()).Create(migrationData)
			}
		}
	}

	return nil
}
