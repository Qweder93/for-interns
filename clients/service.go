// Copyright (C) 2020 Creditor Corp. Group.
// See LICENSE for copying information.

package clients

import (
	"context"

	"github.com/google/uuid"
	"github.com/zeebo/errs"
)

var (
	// Error in an internal error for clients service.
	Error = errs.Class("clients service error")
)

// Service exposes all clients related functionality.
//
// architecture: Service
type Service struct {
	db DB
}

// NewService is a constructor for clients service.
func NewService(db DB) *Service {
	return &Service{
		db: db,
	}
}

// Create is used by manager to create new client.
func (clients *Service) Create(ctx context.Context, email, phone, firstName, lastName string) error {
	client := Client{
		ID:        uuid.New(),
		Email:     email,
		Phone:     phone,
		FirstName: firstName,
		LastName:  lastName,
	}

	return Error.Wrap(clients.db.Add(ctx, client))
}

// Register is used by client to register an account.
func (clients *Service) Register(ctx context.Context, phone string) (uuid.UUID, error) {
	id, err := clients.db.Register(ctx, phone)
	return id, Error.Wrap(err)
}

// Update is used to update client.
func (clients *Service) Update(ctx context.Context, newClient Client) error {
	client, err := clients.db.Get(ctx, newClient.ID)
	if err != nil {
		return Error.Wrap(err)
	}

	client.LastName = newClient.LastName
	client.FirstName = newClient.FirstName
	client.Email = newClient.Email
	client.Phone = newClient.Phone

	return Error.Wrap(clients.db.Update(ctx, newClient))
}

// List is used to return all clients.
func (clients *Service) List(ctx context.Context) ([]Client, error) {
	result, err := clients.db.List(ctx)

	return result, Error.Wrap(err)
}

// Get returns client by ID.
func (clients *Service) Get(ctx context.Context, clientID uuid.UUID) (Client, error) {
	client, err := clients.db.Get(ctx, clientID)

	return client, Error.Wrap(err)
}

// GetByPhone returns client by phone number.
func (clients *Service) GetByPhone(ctx context.Context, phone string) (Client, error) {
	client, err := clients.db.GetByPhone(ctx, phone)

	return client, Error.Wrap(err)
}

// Delete deletes specified client.
func (clients *Service) Delete(ctx context.Context, id uuid.UUID) error {
	return Error.Wrap(clients.db.Delete(ctx, id))
}
