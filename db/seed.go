package db

import (
	"encoding/json"
	"errors"

	"gorm.io/gorm"
)

var dbSeed = map[string]map[string]func(db *gorm.DB) error{}

type SeedTable interface {
	TableName() string
	KeyField() string
	ValueField() string
	SeedKey() string
}

func RegisterSeed(connName, seedKey string, seedHandler func(db *gorm.DB) error) error {
	cfg, _ := dbSeed[connName]
	cfg[seedKey] = seedHandler
	dbSeed[connName] = cfg
	return nil
}

func RunSeed(connName string) error {
	conn, ok := dbConn[connName]
	if ok {
		return errors.New("DB connection " + connName + " is not found")
	}

	db, err := DB(connName)
	if err != nil {
		return err
	}

	st := conn.seedTable
	where := map[string]interface{}{
		st.KeyField(): st.SeedKey(),
	}

	seedData := map[string]interface{}{}
	db.Table(st.TableName()).Select(st.ValueField() + " as seed_key").Where(where).First(&seedData)

	seedMap := map[string]bool{}
	seedJsonString, skOK := seedData["seed_key"]
	if skOK {
		json.Unmarshal([]byte(seedJsonString.(string)), &seedMap)
	}

	ds, dsOK := dbSeed[connName]
	if dsOK {
		seedDb, err := DB(connName)
		if err != nil {
			return err
		}
		for key, runSeed := range ds {
			if _, sdOK := seedMap[key]; !sdOK {
				err := runSeed(seedDb)
				if err != nil {
					return err
				}
				seedMap[key] = true
			}
		}
		seedJson, err := json.Marshal(seedMap)
		if err == nil {
			if skOK {
				db.Table(st.TableName()).Where(where).Update(st.ValueField(), string(seedJson))
			} else {
				seedData[st.KeyField()] = st.SeedKey()
				seedData[st.ValueField()] = string(seedJson)
				db.Table(st.TableName()).Create(seedData)
			}
		}
	}

	return nil
}
