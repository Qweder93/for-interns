// Copyright (C) 2020 Creditor Corp. Group.
// See LICENSE for copying information.

package database

import (
	"context"
	"database/sql"

	_ "github.com/lib/pq" // postgres driver.
	"github.com/zeebo/errs"

	"cleanmasters"
	"cleanmasters/adminportal/managers"
	"cleanmasters/clients"
)

var (
	// Error is database error type.
	Error = errs.Class("database error")
)

// ensures that database implements cleanmasters.DB.
var _ cleanmasters.DB = (*database)(nil)

// database is a Postgres implementation of the admin DB.
type database struct {
	conn *sql.DB
}

// Close closes underlying db connection.
func (db *database) Close() error {
	return Error.Wrap(db.conn.Close())
}

// CreateSchema create tables.
func (db *database) CreateSchema(ctx context.Context) (err error) {
	createTableQuery :=
		`
		CREATE TABLE IF NOT EXISTS clients (
            id                  BYTEA NOT NULL,
            phone               TEXT  NOT NULL,
            first_name          TEXT,
            last_name           TEXT,
            email               TEXT,
            email_normalized    TEXT,
            created_at          timestamp with time zone NOT NULL,
            PRIMARY KEY(id),
            UNIQUE (phone),
            UNIQUE (email_normalized)
		);
		CREATE TABLE IF NOT EXISTS managers (
            id                  BYTEA NOT NULL,
            first_name          TEXT  NOT NULL,
            last_name           TEXT  NOT NULL,
            email               TEXT  NOT NULL,
            email_normalized    TEXT  NOT NULL,
            password_hash       BYTEA NOT NULL,
            created_at          timestamp with time zone NOT NULL,
            PRIMARY KEY(id),
            UNIQUE (email_normalized)
		);
		`

	_, err = db.conn.ExecContext(ctx, createTableQuery)

	return Error.Wrap(err)
}

// Open returns cleanmasters.DB postgresql implementation.
func Open(databaseURL string) (cleanmasters.DB, error) {
	conn, err := sql.Open("postgres", databaseURL)
	if err != nil {
		return nil, Error.Wrap(err)
	}

	return &database{conn: conn}, nil
}

// Clients provides access to Clients store.
func (db *database) Clients() clients.DB {
	return &clientsdb{conn: db.conn}
}

// Managers provides access to Managers database.
func (db *database) Managers() managers.DB {
	return &managersdb{conn: db.conn}
}
