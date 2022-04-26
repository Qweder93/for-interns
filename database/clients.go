// Copyright (C) 2020 Creditor Corp. Group.
// See LICENSE for copying information.

package database

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/zeebo/errs"

	"cleanmasters/clients"
)

// ensures that clients implements cleanmasters.Clients.
var _ clients.DB = (*clientsdb)(nil)

// ErrClientsBD in the error class that indicates about ClientsDB error.
var ErrClientsBD = errs.Class("ClientsDB error")

// clientsdb is an Postrgres implementation of cleanmasters.Clients.
//
// architecture: database
type clientsdb struct {
	conn *sql.DB
}

// Add is a method for inserting new Client to the database.
func (repository *clientsdb) Add(ctx context.Context, client clients.Client) error {

	statement := `INSERT INTO clients (id, email, phone, first_name, last_name, created_at) VALUES ($1, $2, $3, $4, $5, $6);`

	_, err := repository.conn.ExecContext(ctx, statement, client.ID, client.Email, client.Phone, client.FirstName, client.LastName, time.Now().UTC())

	return ErrClientsBD.Wrap(err)
}

// Register is a method for inserting new Client to the database.
func (repository *clientsdb) Register(ctx context.Context, phone string) (uuid.UUID, error) {
	id := uuid.New()

	statement := `INSERT INTO clients (id, email, phone, first_name, last_name, created_at) VALUES ($1, $2, $3, $4, $5, $6);`

	_, err := repository.conn.ExecContext(ctx, statement, id, "", phone, "", "", time.Now().UTC())

	return id, ErrClientsBD.Wrap(err)
}

// Update is a method for updating a Client in the database.
func (repository *clientsdb) Update(ctx context.Context, client clients.Client) error {
	statement := `UPDATE clients 
					SET phone = $1,
						first_name = $2,
						last_name = $3,
						email = $4
					WHERE id = $5`

	_, err := repository.conn.ExecContext(ctx, statement, client.Phone, client.FirstName, client.LastName, client.Email, client.ID)

	return ErrClientsBD.Wrap(err)
}

// List is used to return all clients.
func (repository *clientsdb) List(ctx context.Context) (clientList []clients.Client, err error) {
	statement := `SELECT id, email, phone, first_name, last_name, created_at FROM clients;`

	rows, err := repository.conn.QueryContext(ctx, statement)
	if err != nil {
		return nil, ErrClientsBD.Wrap(err)
	}
	defer func() { err = errs.Combine(err, rows.Close()) }()

	for rows.Next() {
		client := clients.Client{}

		var id []byte
		if err := rows.Scan(&id, &client.Email, &client.Phone, &client.FirstName, &client.LastName, &client.CreatedAt); err != nil {
			return nil, ErrClientsBD.Wrap(err)
		}

		client.ID, err = uuid.ParseBytes(id)
		if err != nil {
			return nil, ErrClientsBD.Wrap(err)
		}

		clientList = append(clientList, client)
	}
	if err = rows.Err(); err != nil {
		return nil, ErrClientsBD.Wrap(err)
	}

	return clientList, nil
}

// GetByID is used to return client by id.
func (repository *clientsdb) Get(ctx context.Context, id uuid.UUID) (clients.Client, error) {
	statement := `SELECT phone, first_name, last_name, email FROM clients WHERE id = $1;`

	client := clients.Client{
		ID: id,
	}

	row := repository.conn.QueryRowContext(ctx, statement, id)

	if err := row.Scan(&client.Phone, &client.FirstName, &client.LastName, &client.Email); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return clients.Client{}, clients.ErrNotExist.Wrap(err)
		}
		return clients.Client{}, ErrClientsBD.Wrap(err)
	}

	return client, nil
}

// GetByPhone is used to return Client by phone number.
func (repository *clientsdb) GetByPhone(ctx context.Context, phone string) (clients.Client, error) {
	statement := `SELECT id, email, first_name, last_name FROM clients WHERE phone = $1;`

	client := clients.Client{
		Phone: phone,
	}

	row := repository.conn.QueryRowContext(ctx, statement, phone)

	err := row.Scan(&client.ID, &client.Email, &client.FirstName, &client.LastName)

	return client, ErrClientsBD.Wrap(err)
}

// Delete deletes specified client.
func (repository *clientsdb) Delete(ctx context.Context, id uuid.UUID) error {
	statement := `DELETE FROM clients WHERE id = $1`

	_, err := repository.conn.ExecContext(ctx, statement, id)

	return ErrClientsBD.Wrap(err)
}
