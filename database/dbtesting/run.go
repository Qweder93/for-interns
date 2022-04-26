// Copyright (C) 2020 Creditor Corp. Group.
// See LICENSE for copying information.

package dbtesting

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"testing"

	"cleanmasters"
	"cleanmasters/database"
	"cleanmasters/internal/tempdb"
)

// Database describes a test database.
type Database struct {
	Name string
	URL  string
}

// tempMasterDB is a cleanmasters.DB-implementing type that cleans up after itself when closed.
type tempMasterDB struct {
	cleanmasters.DB
	tempDB *tempdb.TempDatabase
}

// DefaultTestConn default test conn string that is expected to work with postgres server.
const DefaultTestConn = "postgres://postgres:123456@localhost/cleanmastersdb_test?sslmode=disable"

// Run method will establish connection with db, create tables in random schema, run tests.
func Run(t *testing.T, test func(ctx context.Context, t *testing.T, db cleanmasters.DB)) {
	t.Run("Postgres", func(t *testing.T) {
		ctx := context.Background()

		masterDB := Database{
			Name: "Postgres",
			URL:  DefaultTestConn,
		}

		db, err := CreateMasterDB(ctx, t.Name(), "Test", 0, masterDB)
		if err != nil {
			t.Fatal(err)
		}
		defer func() {
			err := db.Close()
			if err != nil {
				t.Fatal(err)
			}
		}()
		err = db.CreateSchema(ctx)
		if err != nil {
			t.Fatal(err)
		}

		test(ctx, t, db)
	})
}

// CreateMasterDB creates a new cleanmasters database for testing.
func CreateMasterDB(ctx context.Context, name string, category string, index int, dbInfo Database) (db cleanmasters.DB, err error) {
	if dbInfo.URL == "" {
		return nil, fmt.Errorf("database %s connection string not provided", dbInfo.Name)
	}

	schemaSuffix := tempdb.CreateRandomTestingSchemaName(6)
	schema := SchemaName(name, category, index, schemaSuffix)

	tempDB, err := tempdb.OpenUnique(ctx, dbInfo.URL, schema)
	if err != nil {
		return nil, err
	}

	return CreateMasterDBOnTopOf(tempDB)
}

// SchemaName returns a properly formatted schema string.
func SchemaName(testname, category string, index int, schemaSuffix string) string {
	// postgres has a maximum schema length of 64
	// we need additional 6 bytes for the random suffix
	//    and 4 bytes for the index "/S0/""

	indexStr := strconv.Itoa(index)

	var maxTestNameLen = 64 - len(category) - len(indexStr) - len(schemaSuffix) - 2
	if len(testname) > maxTestNameLen {
		testname = testname[:maxTestNameLen]
	}

	if schemaSuffix == "" {
		return strings.ToLower(testname + "/" + category + indexStr)
	}

	return strings.ToLower(testname + "/" + schemaSuffix + "/" + category + indexStr)
}

// CreateMasterDBOnTopOf creates a new cleanmasters database on top of an already existing
// temporary database.
func CreateMasterDBOnTopOf(tempDB *tempdb.TempDatabase) (db cleanmasters.DB, err error) {
	masterDB, err := database.Open(tempDB.ConnStr)
	return &tempMasterDB{DB: masterDB, tempDB: tempDB}, err
}
