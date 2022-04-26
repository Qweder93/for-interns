// Copyright (C) 2020 Creditor Corp. Group.
// See LICENSE for copying information.

package database

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/zeebo/errs"

	"cleanmasters/adminportal/managers"
)

// ensures that managers implements cleanmasters.Managers.
var _ managers.DB = (*managersdb)(nil)

// ErrManagersDB in the error class that indicates about ManagersDB error.
var ErrManagersDB = errs.Class("ManagersDB error")

// managersdb is an Postrgres implementation of cleanmasters.Managers.
//
// architecture: Database
type managersdb struct {
	conn *sql.DB
}

// Get is used to return manager by id.
func (repository *managersdb) Get(ctx context.Context, id uuid.UUID) (managers.Manager, error) {
	statement := `SELECT password_hash, first_name, last_name, email, created_at FROM managers WHERE id = $1;`

	manager := managers.Manager{
		ID: id,
	}

	row := repository.conn.QueryRowContext(ctx, statement, id)

	if err := row.Scan(&manager.PasswordHash, &manager.FirstName, &manager.LastName, &manager.Email, &manager.CreatedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return managers.Manager{}, managers.ErrNoManager.Wrap(err)
		}
		return managers.Manager{}, ErrManagersDB.Wrap(err)
	}

	return manager, nil
}

// GetByEmail is used to return manager by id.
func (repository *managersdb) GetByEmail(ctx context.Context, email string) (managers.Manager, error) {
	statement := `SELECT id, first_name, last_name, password_hash, created_at FROM managers WHERE email_normalized = $1;`

	manager := managers.Manager{
		Email: email,
	}

	row := repository.conn.QueryRowContext(ctx, statement, normalizeEmail(email))

	var id []byte
	err := row.Scan(&id, &manager.FirstName, &manager.LastName, &manager.PasswordHash, &manager.CreatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return managers.Manager{}, managers.ErrNoManager.Wrap(err)
		}
		return managers.Manager{}, ErrManagersDB.Wrap(err)
	}

	manager.ID, err = uuid.ParseBytes(id)
	if err != nil {
		return managers.Manager{}, ErrManagersDB.Wrap(err)
	}

	return manager, nil
}

// Update is a method for updating a Manager in the database.
func (repository *managersdb) Update(ctx context.Context, manager managers.Manager) error {
	statement := `UPDATE managers 
					SET password_hash = $1,
						first_name = $2,
						last_name = $3,
						email = $4
					WHERE id = $5`

	_, err := repository.conn.ExecContext(ctx, statement, manager.PasswordHash, manager.FirstName, manager.LastName, manager.Email, manager.ID)

	return ErrManagersDB.Wrap(err)
}

// Add is a method for inserting new Manager to the database.
func (repository *managersdb) Add(ctx context.Context, manager managers.Manager) error {
	statement := `INSERT INTO managers (id, password_hash, first_name, last_name, email, email_normalized, created_at) VALUES ($1, $2, $3, $4, $5, $6, $7);`

	_, err := repository.conn.ExecContext(ctx, statement, manager.ID, manager.PasswordHash, manager.FirstName, manager.LastName, manager.Email, normalizeEmail(manager.Email), time.Now().UTC())

	return ErrManagersDB.Wrap(err)
}

// Remove is a method for deleting a Manager from the database.
func (repository *managersdb) Remove(ctx context.Context, id uuid.UUID) error {
	statement := `DELETE FROM managers WHERE id = $1;`

	_, err := repository.conn.ExecContext(ctx, statement, id)

	return ErrManagersDB.Wrap(err)
}

// List is used to return all managers.
func (repository *managersdb) List(ctx context.Context) (managerList []managers.Manager, err error) {
	statement := `SELECT id, password_hash, first_name, last_name, email, created_at FROM managers;`

	rows, err := repository.conn.QueryContext(ctx, statement)
	if err != nil {
		return nil, ErrManagersDB.Wrap(err)
	}
	defer func() { err = errs.Combine(err, rows.Close()) }()

	for rows.Next() {
		manager := managers.Manager{}

		var id []byte
		if err := rows.Scan(&id, &manager.PasswordHash, &manager.FirstName, &manager.LastName, &manager.Email, &manager.CreatedAt); err != nil {
			return nil, ErrManagersDB.Wrap(err)
		}

		manager.ID, err = uuid.ParseBytes(id)
		if err != nil {
			return nil, ErrManagersDB.Wrap(err)
		}

		managerList = append(managerList, manager)
	}
	if err = rows.Err(); err != nil {
		return nil, ErrManagersDB.Wrap(err)
	}

	return managerList, nil
}

// normalizeEmail normalizes email for more elegant storing and checking.
func normalizeEmail(email string) string {
	return strings.ToUpper(email)
}
