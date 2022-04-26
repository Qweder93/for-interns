// Copyright (C) 2020 Creditor Corp. Group.
// See LICENSE for copying information.

package server

import (
	"context"
	"errors"
	"net"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/zeebo/errs"
	"golang.org/x/sync/errgroup"

	"cleanmasters/clients"
	"cleanmasters/console/consoleauth"
	"cleanmasters/internal/auth"
	"cleanmasters/internal/logger"
)

var (
	// Error is an error class that indicates internal http server error.
	Error = errs.Class("console web server error")
)

// Config contains configuration for cleanmasters api server.
type Config struct {
	Address string `help:"url of cleanmasters api web server" default:"127.0.0.1:8081"`
}

// Server represents main http server with all api endpoints.
//
// architecture: Endpoint
type Server struct {
	log    logger.Logger
	config Config

	clients *clients.Service
	auth    *consoleauth.Service

	server   http.Server
	listener net.Listener
}

// NewServer is a constructor for cleanmasters server.
func NewServer(log logger.Logger, config Config, clients *clients.Service, auth *consoleauth.Service, listener net.Listener) (*Server, error) {
	server := Server{
		log:      log,
		clients:  clients,
		config:   config,
		auth:     auth,
		listener: listener,
	}

	router := mux.NewRouter()
	router.StrictSlash(true)

	apiRouter := router.PathPrefix("/api/v0").Subrouter()

	clientsRouter := apiRouter.PathPrefix("/clients").Subrouter().StrictSlash(true)
	clientsController := NewClients(server.log, server.clients)
	clientsRouter.HandleFunc("", clientsController.UpdatePersonalData).Methods(http.MethodPatch)

	server.server = http.Server{
		Handler: router,
	}

	return &server, nil
}

// Run starts the server that host webapp and api endpoints.
func (server *Server) Run(ctx context.Context) (err error) {
	ctx, cancel := context.WithCancel(ctx)

	var group errgroup.Group

	group.Go(func() error {
		<-ctx.Done()
		return Error.Wrap(server.server.Shutdown(ctx))
	})
	group.Go(func() error {
		defer cancel()
		err := server.server.Serve(server.listener)
		if errors.Is(err, http.ErrServerClosed) {
			err = nil
		}
		return Error.Wrap(err)
	})

	return Error.Wrap(group.Wait())
}

// Close closes server and underlying listener.
func (server *Server) Close() error {
	return Error.Wrap(server.server.Close())
}

// authenticate performs initial authorization before every request.
func (server *Server) authenticate(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := r.Header.Get("Authorization")
		if len(token) == 0 {
			http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			return
		}

		ctx := auth.SetToken(r.Context(), []byte(token))

		authorization, err := server.auth.Authorize(ctx)
		if err != nil {
			http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			return
		}

		ctx = auth.SetClaims(ctx, authorization)

		handler.ServeHTTP(w, r.Clone(ctx))
	})
}
