package test

import (
	"fmt"
	"fsrv/src/database"
	"fsrv/src/database/dbimpl/sqlite"
	"fsrv/src/database/entities"
	"fsrv/utils/serde"
	"os"
	"strings"
	"testing"
	"time"
)

func getDB() database.DBInterface {
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

func makeRoles(db database.DBInterface) error {
	roleTable := map[int]string{
		100:  "stone",
		200:  "iron",
		250:  "gold",
		1500: "obsidian",
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

func makeKeys(db database.DBInterface) error {
	k1 := entities.Key{
		ID:      "key_q2w26DFu8dr5578x&4syd46e7",
		Comment: "rock guy",
		Roles:   []string{"stone"},
		RequestRateLimit: &entities.RateLimit{
			ID:     "DEPRECATED",
			Limit:  100,
			Refill: 60,
		},
		ExpiresAt: serde.Time(time.Now().AddDate(1, 0, 0)),
		CreatedAt: serde.Time(time.Now()),
	}

	k2 := entities.Key{
		ID:      "key_dr476FXC8drUXe%&SR5ujr",
		Comment: "pebble person",
		Roles:   []string{"stone"},
		RequestRateLimit: &entities.RateLimit{
			ID:     "DEPRECATED",
			Limit:  100,
			Refill: 60,
		},
		ExpiresAt: serde.Time(time.Now().AddDate(0, 8, 0)),
		CreatedAt: serde.Time(time.Now()),
	}

	k3 := entities.Key{
		ID:      "key_dfthcr5uyers57yerd5ydr567",
		Comment: "iron ingot wingnut",
		Roles:   []string{"iron"},
		RequestRateLimit: &entities.RateLimit{
			ID:     "DEPRECATED",
			Limit:  200,
			Refill: 60,
		},
		ExpiresAt: serde.Time(time.Now().AddDate(0, 6, 0)),
		CreatedAt: serde.Time(time.Now()),
	}

	k4 := entities.Key{
		ID:      "key_gkfp989P$%WA$ETseTSETST$",
		Comment: "Roles: stone & diamond",
		Roles:   []string{"stone", "diamond"},
		RequestRateLimit: &entities.RateLimit{
			ID:     "DEPRECATED",
			Limit:  500,
			Refill: 60,
		},
		ExpiresAt: serde.Time(time.Now().AddDate(0, 3, 0)),
		CreatedAt: serde.Time(time.Now()),
	}

	k5 := entities.Key{
		ID:      "key_sdrySDRyDSrydrtyasWTT",
		Comment: "Roles: gold & iron",
		Roles:   []string{"gold", "iron"},
		RequestRateLimit: &entities.RateLimit{
			ID:     "DEPRECATED",
			Limit:  500,
			Refill: 60,
		},
		ExpiresAt: serde.Time(time.Now().AddDate(0, 0, 4)),
		CreatedAt: serde.Time(time.Now()),
	}

	k6 := entities.Key{
		ID:      "key_ancientEvil",
		Comment: "ancient evil must be contained through strict ratelimits and permission control",
		Roles:   []string{"obsidian"},
		RequestRateLimit: &entities.RateLimit{
			ID:     "STRICT_LIMIT",
			Limit:  10,
			Refill: 60,
		},
		ExpiresAt: serde.Time(time.Now().AddDate(9999, 0, 0)),
		CreatedAt: serde.Time(time.Now()),
	}

	//create keys
	keys := []*entities.Key{&k1, &k2, &k3, &k4, &k5, &k6}
	for _, k := range keys {
		err := db.CreateKey(k)
		if err != nil {
			return err
		}
	}

	return nil
}

func makeResources(db database.DBInterface) error {
	res := []*entities.Resource{
		{
			ID:          "res_stoneWorld:WYRSRYssysysySrysrur6i98",
			Flags:       0,
			ReadNodes:   map[string]bool{"stone": true},
			WriteNodes:  map[string]bool{"stone": true},
			ModifyNodes: map[string]bool{"stone": true},
			DeleteNodes: map[string]bool{"stone": true},
		},

		{
			ID:          "res_publicReadAndModDiamondDelete:pu9ipuijpj0m0uji0ji0j0ji0",
			Flags:       entities.FlagAuthedRead | entities.FlagAuthedModify,
			ReadNodes:   nil,
			WriteNodes:  nil,
			ModifyNodes: nil,
			DeleteNodes: map[string]bool{"diamond": true},
		},

		{
			ID:          "res_READ:DiamondAllowStoneDeny,WRITE:IronAllowGoldDeny:qwdxqwdxawqxqwdqwd",
			Flags:       0,
			ReadNodes:   map[string]bool{"diamond": true, "stone": false},
			WriteNodes:  map[string]bool{"gold": false, "iron": true},
			ModifyNodes: nil,
			DeleteNodes: nil,
		},

		{
			ID:          "res_READ:ObsidianDenyStoneAllowStoneDeny,WRITE:Gold",
			Flags:       0,
			ReadNodes:   map[string]bool{"obsidian": false, "stone": true, "gold": false},
			WriteNodes:  map[string]bool{"gold": true},
			ModifyNodes: nil,
			DeleteNodes: nil,
		},
	}

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
