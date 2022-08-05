package test

import (
	"fmt"
	"fsrv/src/database/dbimpl/sqlite"
	"fsrv/src/database/dbutil"
	"fsrv/src/database/entities"
	"fsrv/utils/serde"
	"os"
	"strings"
	"testing"
	"time"
)

func getDB() dbutil.DBInterface {
	dbFileName := "FSRV_TEST_DATABASE.sqlite"
	wd, err := os.Getwd()
	if err != nil {
		panic("FAILED TO GET WORKING DIRECTORY")
	}

	err = os.Remove(wd + "/" + dbFileName)
	if err != nil {
		if strings.Contains(err.Error(), "no such file or directory") {
			fmt.Println("created new db file")
		} else {
			panic("FAILED TO REMOVE EXISTING DB FILE: " + err.Error())

		}

	}
	_, err = os.Create(wd + "/" + dbFileName)
	if err != nil {
		panic("FAILED TO MAKE NEW TEST DB FILE")
	}

	db, err := sqlite.Create(wd + "/" + dbFileName)
	if err != nil {
		panic("FAILED TO INITIALIZE DB FILE: " + err.Error())
	}

	return db
}

func bap(t *testing.T, errs ...error) {
	for _, err := range errs {
		if err != nil {
			t.Fatalf("[TEST FAILED]: %v", errs)
		}
	}
}

func makeRoles(db dbutil.DBInterface) error {
	roleTable := map[int]string{
		100:  "stone",
		200:  "iron",
		150:  "gold",
		1000: "diamond",
	}
	for k, v := range roleTable {
		err := db.CreateRole(&entities.Role{
			ID:         v,
			Precedence: k,
		})
		if err != nil {
			return err
		}
	}
	return nil
}

func makeKeys(db dbutil.DBInterface) error {
	k1 := entities.Key{
		ID:      "q2w26DFu8dr5578x&4syd46e7",
		Comment: "rock guy",
		Roles:   []string{"stone"},
		RequestRateLimit: &entities.RateLimit{
			ID:    "DEPRECATED",
			Limit: 100,
			Reset: 60,
		},
		ExpiresAt: serde.Time(time.Now().AddDate(1, 0, 0)),
		CreatedAt: serde.Time(time.Now()),
	}

	err := db.CreateKey(&k1)
	if err != nil {
		return err
	}

	k2 := entities.Key{
		ID:      "dr476FXC8drUXe%&SR5ujr",
		Comment: "pebble person",
		Roles:   []string{"stone"},
		RequestRateLimit: &entities.RateLimit{
			ID:    "DEPRECATED",
			Limit: 100,
			Reset: 60,
		},
		ExpiresAt: serde.Time(time.Now().AddDate(0, 8, 0)),
		CreatedAt: serde.Time(time.Now()),
	}

	k3 := entities.Key{
		ID:      "dfthcr5uyers57yerd5ydr567",
		Comment: "iron ingot wingnut",
		Roles:   []string{"iron"},
		RequestRateLimit: &entities.RateLimit{
			ID:    "DEPRECATED",
			Limit: 200,
			Reset: 60,
		},
		ExpiresAt: serde.Time(time.Now().AddDate(0, 6, 0)),
		CreatedAt: serde.Time(time.Now()),
	}

	k4 := entities.Key{
		ID:      "gkfp989P$%WA$ETseTSETST$",
		Comment: "Roles: stone & diamond",
		Roles:   []string{"stone", "diamond"},
		RequestRateLimit: &entities.RateLimit{
			ID:    "DEPRECATED",
			Limit: 500,
			Reset: 60,
		},
		ExpiresAt: serde.Time(time.Now().AddDate(0, 3, 0)),
		CreatedAt: serde.Time(time.Now()),
	}

	k5 := entities.Key{
		ID:      "sdrySDRyDSrydrtyasWTT",
		Comment: "Roles: gold & iron",
		Roles:   []string{"gold", "iron"},
		RequestRateLimit: &entities.RateLimit{
			ID:    "DEPRECATED",
			Limit: 500,
			Reset: 60,
		},
		ExpiresAt: serde.Time(time.Now().AddDate(0, 0, 4)),
		CreatedAt: serde.Time(time.Now()),
	}

	//create keys
	keys := []*entities.Key{&k1, &k2, &k3, &k4, &k5}
	for _, k := range keys {
		err := db.CreateKey(k)
		if err != nil {
			return err
		}
	}

	return nil
}

func makeResources(db dbutil.DBInterface) error {
	r1 := entities.Resource{
		ID:          "",
		Flags:       0,
		ReadNodes:   nil,
		WriteNodes:  nil,
		ModifyNodes: nil,
		DeleteNodes: nil,
	}

	r2 := entities.Resource{
		ID:          "",
		Flags:       0,
		ReadNodes:   nil,
		WriteNodes:  nil,
		ModifyNodes: nil,
		DeleteNodes: nil,
	}

	r3 := entities.Resource{
		ID:          "",
		Flags:       0,
		ReadNodes:   nil,
		WriteNodes:  nil,
		ModifyNodes: nil,
		DeleteNodes: nil,
	}

	//create resources
	res := []*entities.Resource{&r1, &r2, &r3}
	for _, r := range res {
		err := db.CreateResource(r)
		if err != nil {
			return err
		}
	}

	return nil
}

func TestSQLite(t *testing.T) {
	db := getDB()
	bap(t, makeRoles(db))
	bap(t, makeResources(db))
	bap(t, makeKeys(db))
}
