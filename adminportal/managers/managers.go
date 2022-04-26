// Copyright (C) 2020 Creditor Corp. Group.
// See LICENSE for copying information.

package managers

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/zeebo/errs"
)

// ErrNoManager special error class that indicates that manager not exist.
var ErrNoManager = errs.Class("manager does not exist")

// DB exposes methods to manage Managers database.
//
// architecture: Database
type DB interface {
	// Add is a method for inserting new Manager to the database.
	Add(ctx context.Context, manager Manager) error
	// Remove is a method for deleting a Manager from the database.
	Remove(ctx context.Context, id uuid.UUID) error
	// Update is a method for updating a Manager in the database.
	Update(ctx context.Context, manager Manager) error
	// List is used to return all managers.
	List(ctx context.Context) ([]Manager, error)
	// Get is used to return manager by id.
	Get(ctx context.Context, id uuid.UUID) (Manager, error)
	// GetByEmail is used to return manager by email.
	GetByEmail(ctx context.Context, email string) (Manager, error)
}

// Manager describes manager entity.
//
// Manager is responsible for accepting/declining orders,
// managing services, goods, etc.
type Manager struct {
	ID           uuid.UUID
	FirstName    string
	LastName     string
	Email        string
	PasswordHash []byte
	CreatedAt    time.Time
}

// TODO: create IsValid method for Manager and use it in admin service.

// ManagerUpdateFields contains all fields that could be updated in Manager entity.
type ManagerUpdateFields struct {
	FirstName string
	LastName  string
	Email     string
	Password  string
}
