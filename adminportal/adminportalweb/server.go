// Copyright (C) 2020 Creditor Corp. Group.
// See LICENSE for copying information.

package adminportalweb

import (
	"context"
	"errors"
	"net"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/zeebo/errs"
	"golang.org/x/sync/errgroup"

	"cleanmasters/adminportal/adminauth"
	"cleanmasters/adminportal/managers"
	"cleanmasters/clients"
	"cleanmasters/internal/auth"
	"cleanmasters/internal/logger"
)

var (
	// Error is an error class that indicates internal admin http server error.
	Error = errs.Class("admin server error")
)

// Config contains configuration for server.
type Config struct {
	Address   string
	StaticDir string
}

// Server represents main admin portal http server with all endpoints.
//
// architecture: Endpoint
type Server struct {
	log    logger.Logger
	config Config

	managers   *managers.Service
	clients    *clients.Service
	service    *adminauth.Service
	cookieAuth *auth.Cookie

	server   http.Server
	listener net.Listener
}

// NewServer returns new instance of Admin Portal HTTP Server.
func NewServer(log logger.Logger, config Config, authService *adminauth.Service, clients *clients.Service, managers *managers.Service, listener net.Listener) *Server {
	// TODO: take this values from config.
	cookieAuth := auth.NewCookie(
		auth.CookieSettings{
			Name: "cleanmasters_manager_cookie",
			Path: "/",
		},
	)

	server := Server{
		log:        log,
		config:     config,
		service:    authService,
		clients:    clients,
		managers:   managers,
		cookieAuth: cookieAuth,
		listener:   listener,
	}

	router := mux.NewRouter()

	managersRouter := router.PathPrefix("/managers").Subrouter()
	managersRouter.Use(server.withAuth)
	managersController := NewManagers(log, config, server.managers)
	managersRouter.HandleFunc("", managersController.List).Methods(http.MethodGet, http.MethodPost)
	managersRouter.HandleFunc("/create", managersController.Create).Methods(http.MethodGet, http.MethodPost)
	managersRouter.HandleFunc("/{id}/update", managersController.Update).Methods(http.MethodGet, http.MethodPost)
	managersRouter.HandleFunc("/{id}/delete", managersController.Delete).Methods(http.MethodGet)

	authRouter := router.PathPrefix("/authorize").Subrouter()
	authController := NewAuth(log, config, server.service, server.cookieAuth)
	authRouter.HandleFunc("", authController.Authorize).Methods(http.MethodGet, http.MethodPost)

	clientsRouter := router.PathPrefix("/clients").Subrouter()
	clientsRouter.Use(server.withAuth)
	clientsController := NewClients(log, server.config, server.clients)
	clientsRouter.HandleFunc("", clientsController.List).Methods(http.MethodGet)
	clientsRouter.HandleFunc("/create", clientsController.Create).Methods(http.MethodGet, http.MethodPost)
	clientsRouter.HandleFunc("/{id}/update", clientsController.Update).Methods(http.MethodGet, http.MethodPost)
	clientsRouter.HandleFunc("/{id}/delete", clientsController.Delete).Methods(http.MethodGet)

	server.server = http.Server{
		Handler: router,
	}

	return &server
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

// withAuth performs initial authorization before every request.
func (server *Server) withAuth(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		token, err := server.cookieAuth.GetToken(r)
		if err != nil {
			// httputil.Redirect(w, r, "/authorize", "GET")
			r = r.Clone(ctx)
			r.Method = http.MethodGet
			http.Redirect(w, r, "/authorize", http.StatusMovedPermanently)
			return
		}

		ctx = auth.SetToken(ctx, []byte(token))

		authorization, err := server.service.Authorize(ctx)
		if err != nil {
			r = r.Clone(ctx)
			r.Method = http.MethodGet
			http.Redirect(w, r, "/authorize", http.StatusMovedPermanently)
			return
		}

		ctx = auth.SetClaims(ctx, authorization)

		handler.ServeHTTP(w, r.Clone(ctx))
	})
}
