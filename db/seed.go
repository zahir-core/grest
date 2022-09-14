package db

import (
	"encoding/json"

	"gorm.io/gorm"
)

var dbSeed = map[string]map[string]SeedHandler{}

type SeedHandler func(db *gorm.DB) error

type SeedTableInterface interface {
	TableName() string
	KeyField() string
	ValueField() string
	SeedKey() string
}

type SeedTable struct {
	SettingTable
}

func (SeedTable) SeedKey() string {
	return "executed_seeds"
}

func RegisterSeed(connName, seedKey string, seedHandler SeedHandler) error {
	cfg, _ := dbSeed[connName]
	cfg[seedKey] = seedHandler
	dbSeed[connName] = cfg
	return nil
}

func RunSeed(tx *gorm.DB, connName string, seedTable SeedTableInterface) error {
	tableName := Quote(tx, seedTable.TableName())
	keyField := Quote(tx, seedTable.KeyField())
	valueField := Quote(tx, seedTable.ValueField())
	seedKey := seedTable.SeedKey()

	seedData := map[string]any{}
	tx.Table(tableName).
		Where(keyField+" = ?", seedKey).
		Select(valueField + " as " + Quote(tx, "value")).
		Take(&seedData)

	seedMap := map[string]bool{}
	seedJsonString, isSeedStringExist := seedData["value"]
	if isSeedStringExist {
		json.Unmarshal([]byte(seedJsonString.(string)), &seedMap)
	}

	registeredSeeds, isRegisteredSeedExist := dbSeed[connName]
	if isRegisteredSeedExist {
		for key, runSeed := range registeredSeeds {
			if _, sdOK := seedMap[key]; !sdOK {
				err := runSeed(tx)
				if err != nil {
					return err
				}
				seedMap[key] = true
			}
		}
		seedJson, err := json.Marshal(seedMap)
		if err == nil {
			if isSeedStringExist {
				tx.Table(tableName).
					Where(keyField+" = ?", seedKey).
					Update(valueField, string(seedJson))
			} else {
				seedData[keyField] = seedKey
				seedData[valueField] = string(seedJson)
				tx.Table(tableName).Create(seedData)
			}
		}
	}

	return nil
}
