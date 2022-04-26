// Copyright (C) 2020 Creditor Corp. Group.
// See LICENSE for copying information.

package clients

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/zeebo/errs"
)

var (
	// ErrNotExist indicates that client is not exist in database.
	ErrNotExist = errs.Class("client does not exist")
)

// DB exposes methods to manage Clients database.
//
// architecture: Database
type DB interface {
	// Add is a method for inserting new Client to the database.
	Add(ctx context.Context, client Client) error
	// Register is a method for inserting new Client to the database.
	Register(ctx context.Context, phone string) (uuid.UUID, error)
	// Update is a method for updating a Client in the database.
	Update(ctx context.Context, client Client) error
	// List is used to return all clients.
	List(ctx context.Context) ([]Client, error)
	// Get is used to return Client by id.
	Get(ctx context.Context, id uuid.UUID) (Client, error)
	// GetByPhone is used to return Client by phone number.
	GetByPhone(ctx context.Context, phone string) (Client, error)
	// Delete deletes specified client.
	Delete(ctx context.Context, id uuid.UUID) error
}

// Client describes cleanmasters client.
type Client struct {
	ID        uuid.UUID
	Email     string
	Phone     string
	FirstName string
	LastName  string
	CreatedAt time.Time
}

// ClientUpdateFields contains all fields that could be updated in Client entity.
type ClientUpdateFields struct {
	FirstName string
	LastName  string
	Email     string
	Phone     string
}
