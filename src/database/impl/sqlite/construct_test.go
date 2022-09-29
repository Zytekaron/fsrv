package sqlite

import (
	"fmt"
	"fsrv/src/database"
	"fsrv/src/database/entities"
	"fsrv/src/types"
	"fsrv/utils/serde"
	"os"
	"strings"
	"testing"
	"time"
)

func getDB() *SQLiteDB {
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

	db, err := Create(wd + "/" + dbFileName)
	if err != nil {
		panic("FAILED TO INITIALIZE DB FILE: " + err.Error())
	}

	return db
}

func bap(t *testing.T, errs ...error) {
	for _, err := range errs {
		if err != nil {
			t.Errorf("[TEST FAILED]: %v", errs)
			t.Fail()
		}
	}
}

func makeRoles(db *SQLiteDB) error {
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

func makeKeys(db *SQLiteDB) error {
	k1 := entities.Key{
		ID:          "key_q2w26DFu8dr5578x&4syd46e7",
		Comment:     "rock guy",
		Roles:       []string{"stone"},
		RateLimitID: "DEFAULT",
		ExpiresAt:   serde.Time(time.Now().AddDate(1, 0, 0)),
		CreatedAt:   serde.Time(time.Now()),
	}

	k2 := entities.Key{
		ID:          "key_dr476FXC8drUXe%&SR5ujr",
		Comment:     "pebble person",
		Roles:       []string{"stone"},
		RateLimitID: "DEFAULT",
		ExpiresAt:   serde.Time(time.Now().AddDate(0, 8, 0)),
		CreatedAt:   serde.Time(time.Now()),
	}

	k3 := entities.Key{
		ID:          "key_dfthcr5uyers57yerd5ydr567",
		Comment:     "iron ingot wingnut",
		Roles:       []string{"iron"},
		RateLimitID: "DEFAULT",
		ExpiresAt:   serde.Time(time.Now().AddDate(0, 6, 0)),
		CreatedAt:   serde.Time(time.Now()),
	}

	k4 := entities.Key{
		ID:          "key_gkfp989P$%WA$ETseTSETST$",
		Comment:     "Roles: stone & diamond",
		Roles:       []string{"stone", "diamond"},
		RateLimitID: "high limit",
		ExpiresAt:   serde.Time(time.Now().AddDate(0, 3, 0)),
		CreatedAt:   serde.Time(time.Now()),
	}

	k5 := entities.Key{
		ID:          "key_sdrySDRyDSrydrtyasWTT",
		Comment:     "Roles: gold & iron",
		Roles:       []string{"gold", "iron"},
		RateLimitID: "LowLimitFastReset",
		ExpiresAt:   serde.Time(time.Now().AddDate(0, 0, 4)),
		CreatedAt:   serde.Time(time.Now()),
	}

	k6 := entities.Key{
		ID:          "key_ancientEvil",
		Comment:     "ancient evil must be contained through strict ratelimits and permission control",
		Roles:       []string{"obsidian"},
		RateLimitID: "STRICT_LIMIT",
		ExpiresAt:   serde.Time(time.Now().AddDate(9999, 0, 0)),
		CreatedAt:   serde.Time(time.Now()),
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

func makeResources(db *SQLiteDB) error {
	res := []*entities.Resource{
		{
			ID:    "res_stoneWorld:WYRSRYssysysySrysrur6i98",
			Flags: 0,
			OperationNodes: map[entities.ResourceOperationAccess]bool{
				{"stone", types.OperationRead}:   true,
				{"stone", types.OperationWrite}:  false,
				{"stone", types.OperationModify}: false,
				{"stone", types.OperationDelete}: true,
			},
		},

		{
			ID:    "res_publicReadAndModDiamondDelete:pu9ipuijpj0m0uji0ji0j0ji0",
			Flags: 0,
			OperationNodes: map[entities.ResourceOperationAccess]bool{
				{"diamond", types.OperationDelete}: true,
			},
		},

		{
			ID:    "res_READ:DiamondAllowStoneDeny,WRITE:IronAllowGoldDeny:qwdxqwdxawqxqwdqwd",
			Flags: 0,
			OperationNodes: map[entities.ResourceOperationAccess]bool{
				{"diamond", types.OperationRead}: true,
				{"stone", types.OperationRead}:   false,
				{"gold", types.OperationWrite}:   false,
				{"iron", types.OperationWrite}:   true,
			},
		},

		{
			ID:    "res_READ:ObsidianDenyStoneAllowStoneDeny,WRITE:Gold",
			Flags: 0,
			OperationNodes: map[entities.ResourceOperationAccess]bool{
				{"obsidian", types.OperationRead}: false,
				{"stone", types.OperationRead}:    true,
				{"gold", types.OperationRead}:     false,
				{"gold", types.OperationWrite}:    true,
			},
		},

		{
			ID:    "res_READ:fake_roles,WRITE:fakeAndReal",
			Flags: 0,
			OperationNodes: map[entities.ResourceOperationAccess]bool{
				{"glass", types.OperationRead}:    false,
				{"coal", types.OperationRead}:     true,
				{"steel", types.OperationRead}:    false,
				{"mithril", types.OperationWrite}: true,
				{"diamond", types.OperationWrite}: true,
			},
		},
		{
			ID:    "res_C/R/U/D:[key: key_ancientEvil]",
			Flags: 0,
			OperationNodes: map[entities.ResourceOperationAccess]bool{
				{"key_ancientEvil", types.OperationRead}:   false,
				{"key_ancientEvil", types.OperationWrite}:  true,
				{"key_ancientEvil", types.OperationModify}: false,
				{"key_ancientEvil", types.OperationDelete}: true,
			},
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

func grantPermissionsPostHoc(db *SQLiteDB) (errs []error) {
	roleperms := map[*entities.Permission][]string{
		&entities.Permission{
			ResourceID: "res_READ:fake_roles,WRITE:fakeAndReal",
			TypeRWMD:   0,
			Status:     true,
		}: {"diamond"},
		&entities.Permission{
			ResourceID: "res_READ:fake_roles,WRITE:fakeAndReal",
			TypeRWMD:   0,
			Status:     true,
		}: {"iron, gold, pyrite"},
		&entities.Permission{
			ResourceID: "Bad Resource id",
			TypeRWMD:   2,
			Status:     false,
		}: {"gold"},
		&entities.Permission{
			ResourceID: "res_READ:fake_roles,WRITE:fakeAndReal",
			TypeRWMD:   6, //todo: fix accepting bad permission type
			Status:     true,
		}: {"diamond, obsidian"},
		&entities.Permission{ //grant per key permission
			ResourceID: "",
			TypeRWMD:   0,
			Status:     true,
		}: {"key_dfthcr5uyers57yerd5ydr567"},
	}

	for k, v := range roleperms {
		err := db.GrantPermission(k, v...)
		if err != nil {
			errs = append(errs, err)
		}
	}
	return errs
}

func getResourcePermData(t *testing.T, db *SQLiteDB, ids []string) (errs []error) {
	for i, id := range ids {
		data, err := db.GetResourceData(id)
		t.Logf("%d: [FLAGS=%d, ID=%s]\n", i, data.Flags, id)
		if err == nil {
			//t.Logf("ReadNodes:")
			//for k, v := range data.ReadNodes {
			//	t.Logf("%s:%t", k, v)
			//}
			//t.Logf("\nWriteNodes:")
			//for k, v := range data.WriteNodes {
			//	t.Logf("%s:%t", k, v)
			//}
			//t.Logf("\nModifyNodes:")
			//for k, v := range data.ModifyNodes {
			//	t.Logf("%s:%t", k, v)
			//}
			//t.Logf("\nDeleteNodes:")
			//for k, v := range data.DeleteNodes {
			//	t.Logf("%s:%t", k, v)
			//}
		} else {
			errs = append(errs, err)
		}
	}
	return errs
}

func getRatelimitsForAllKeys(t *testing.T, db *SQLiteDB, ids []string) (errs []error) {
	for _, id := range ids {
		limID, err := db.GetKeyRateLimitID(id)
		if err != nil {
			errs = append(errs, err)
		}
		lim, err := db.GetRateLimitData(limID)
		if err != nil {
			errs = append(errs, err)
		} else {
			t.Logf("KeyID: %s, ratelimit:{id:%s, lim:%d, reset:%d }", id, lim.ID, lim.Limit, lim.Refill)
		}
	}
	return errs
}

func giveRoles(t *testing.T, db *SQLiteDB) (errs []error) {

	return errs
}

func takeRoles(t *testing.T, db *SQLiteDB) (errs []error) {

	return errs
}

func createRateLimits(t *testing.T, db *SQLiteDB) (errs []error) {
	limits := []entities.RateLimit{
		{"DEFAULT", 20, 60},
		{"high limit", 200, 60},
		{"LowLimitFastReset", 2, 1},
		{"STRICT_LIMIT", 1, 60},
	}
	var err error
	for _, lim := range limits {
		err = db.CreateRateLimit(&lim)
		if err != nil {
			errs = append(errs, err)
		}
	}

	return errs
}

func TestSQLite(t *testing.T) {
	db := getDB()
	bap(t, makeRoles(db))
	bap(t, makeResources(db))
	bap(t, createRateLimits(t, db)...)
	bap(t, makeKeys(db))
	bap(t, grantPermissionsPostHoc(db)...)
	resources, err := db.GetResourceIDs(1000, 0)
	bap(t, err)
	bap(t, getResourcePermData(t, db, resources)...)
	keys, err := db.GetKeyIDs(1000, 0)
	bap(t, err)
	bap(t, getRatelimitsForAllKeys(t, db, keys)...)
	bap(t, giveRoles(t, db)...)
	bap(t, takeRoles(t, db)...)
	_, err = db.GetKeyRateLimitID("idontexist")
	t.Logf("Correct Error? %t", err == database.ErrKeyMissing)
}
