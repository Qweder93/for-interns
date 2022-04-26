// Copyright (C) 2020 Creditor Corp. Group.
// See LICENSE for copying information.

package server

import (
	"encoding/json"
	"net/http"

	"github.com/zeebo/errs"

	"cleanmasters/clients"
	"cleanmasters/internal/auth"
	"cleanmasters/internal/logger"
)

var (
	// ErrClients is an internal error type for clients controller.
	ErrClients = errs.Class("clients controller error")
)

// Clients is a web api controller.
type Clients struct {
	log     logger.Logger
	clients *clients.Service
}

// NewClients is a constructor for clients controller.
func NewClients(log logger.Logger, clients *clients.Service) *Clients {
	return &Clients{
		log:     log,
		clients: clients,
	}
}

// UpdateClientRequest holds all needed data to update client.
type UpdateClientRequest struct {
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
	Email     string `json:"email"`
	Phone     string `json:"phone"`
}

// UpdatePersonalData is a method for updating accounts fields in the database.
func (controller *Clients) UpdatePersonalData(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	w.Header().Add("Content-Type", "application/json")

	claims, err := auth.GetClaims(ctx)
	if err != nil {
		controller.serveError(w, http.StatusUnauthorized, ErrClients.Wrap(err))
		return
	}

	request := UpdateClientRequest{}
	err = json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		controller.serveError(w, http.StatusBadRequest, ErrClients.Wrap(err))
		return
	}

	err = controller.clients.Update(ctx, clients.Client{
		ID:        claims.ID,
		FirstName: request.FirstName,
		LastName:  request.LastName,
		Email:     request.Email,
		Phone:     request.Phone,
	})
	if err != nil {
		controller.log.Error("couldn't update client", ErrClients.Wrap(err))
		controller.serveError(w, http.StatusInternalServerError, ErrClients.Wrap(err))
		return
	}
}

// serveError set http statuses and send json error.
func (controller *Clients) serveError(w http.ResponseWriter, status int, err error) {
	w.WriteHeader(status)

	var response struct {
		Error string `json:"error"`
	}

	response.Error = err.Error()

	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		controller.log.Error("failed to write json error response", ErrClients.Wrap(err))
	}
}
