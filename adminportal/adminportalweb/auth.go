// Copyright (C) 2020 Creditor Corp. Group.
// See LICENSE for copying information.

package adminportalweb

import (
	"html/template"
	"net/http"
	"path/filepath"

	"github.com/zeebo/errs"

	"cleanmasters/adminportal/adminauth"
	"cleanmasters/internal/auth"
	"cleanmasters/internal/logger"
)

var (
	// ErrAuth is an internal error type for auth controller.
	ErrAuth = errs.Class("auth controller error")
)

// Auth is a web api controller.
// Exposes functionality and web views to authorize in admin portal.
type Auth struct {
	log    logger.Logger
	config Config

	authentication *adminauth.Service
	cookieAuth     *auth.Cookie

	authorizeTemplate *template.Template
}

// NewAuth is a constructor for auth controller.
func NewAuth(log logger.Logger, config Config, service *adminauth.Service, cookieAuth *auth.Cookie) *Auth {
	authController := &Auth{
		log:            log,
		config:         config,
		authentication: service,
		cookieAuth:     cookieAuth,
	}

	err := authController.initializeTemplates()
	if err != nil {
		panic(err)
	}

	return authController
}

// initializeTemplates initializes and caches templates for managers controller.
func (controller *Auth) initializeTemplates() (err error) {
	controller.authorizeTemplate, err = template.ParseFiles(filepath.Join(controller.config.StaticDir, "authorize", "authorize.html"))
	if err != nil {
		return err
	}

	return nil
}

// Authorize is an endpoint to authorize admin and set auth cookie in browser.
func (controller *Auth) Authorize(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var err error

	switch r.Method {
	case http.MethodGet:
		err = controller.authorizeTemplate.Execute(w, nil)
		if err != nil {
			controller.log.Error("could not execute authorize template", ErrAuth.Wrap(err))
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
	case http.MethodPost:
		err = r.ParseForm()
		if err != nil {
			controller.log.Error("could not parse form with credentials", ErrAuth.Wrap(err))
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}

		// TODO: process form in a better way
		email := r.Form["email"]
		password := r.Form["password"]
		response, err := controller.authentication.Token(ctx, email[0], password[0])
		if err != nil {
			controller.log.Error("could not issue auth token", ErrAuth.Wrap(err))
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		controller.cookieAuth.SetToken(w, response.String())

		r = r.Clone(ctx)
		r.Method = http.MethodGet
		http.Redirect(w, r, "/managers", http.StatusMovedPermanently)
	}
}
