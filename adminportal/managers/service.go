// Copyright (C) 2020 Creditor Corp. Group.
// See LICENSE for copying information.

package managers

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/zeebo/errs"
	"golang.org/x/crypto/bcrypt"
)

var (
	// Error in an internal error for managers service.
	Error = errs.Class("managers service error")
	// ValidationError indicates that input data was incorrect or that entity is invalid.
	ValidationError = errs.Class("managers service validation error")
)

// Service exposes all managers related functionality.
//
// architecture: Service
type Service struct {
	db DB
}

// NewService initializes new instance of managers service.
func NewService(db DB) *Service {
	return &Service{
		db: db,
	}
}

// Create is used to create new manager.
func (service *Service) Create(ctx context.Context, password, firstName, lastName, email string) error {
	// TODO: validate manager
	if password == "" {
		return ValidationError.New("password is incorrect")
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return Error.Wrap(err)
	}

	manager := Manager{
		ID:           uuid.New(),
		FirstName:    firstName,
		LastName:     lastName,
		Email:        email,
		PasswordHash: passwordHash,
		CreatedAt:    time.Now().UTC(),
	}

	return Error.Wrap(service.db.Add(ctx, manager))
}

// Get returns manager by ID.
func (service *Service) Get(ctx context.Context, id uuid.UUID) (Manager, error) {
	manager, err := service.db.Get(ctx, id)

	return manager, Error.Wrap(err)
}

// GetByEmail returns manager by normalized email (lowercase).
func (service *Service) GetByEmail(ctx context.Context, email string) (Manager, error) {
	manager, err := service.db.GetByEmail(ctx, email)

	return manager, Error.Wrap(err)
}

// Update is used to update manager.
func (service *Service) Update(ctx context.Context, id uuid.UUID, fields ManagerUpdateFields) error {
	manager, err := service.db.Get(ctx, id)
	if err != nil {
		return Error.Wrap(err)
	}

	// TODO: generate and check password string in fields.

	manager.LastName = fields.LastName
	manager.FirstName = fields.FirstName
	manager.Email = fields.Email

	return Error.Wrap(service.db.Update(ctx, manager))
}

// List is used to return all managers.
func (service *Service) List(ctx context.Context) (result []Manager, err error) {
	result, err = service.db.List(ctx)

	return result, Error.Wrap(err)
}

// Delete will remove manager from DB by id.
func (service *Service) Delete(ctx context.Context, id uuid.UUID) error {
	return Error.Wrap(service.db.Remove(ctx, id))
}
