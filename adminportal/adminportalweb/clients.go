// Copyright (C) 2021 Creditor Corp. Group.
// See LICENSE for copying information.

package adminportalweb

import (
	"html/template"
	"net/http"
	"path/filepath"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/zeebo/errs"

	"cleanmasters/clients"
	"cleanmasters/internal/logger"
)

var (
	// ClientsError is an internal error type for clients controller.
	ClientsError = errs.Class("clients controller error")
)

// ClientTemplates holds templates needed for clients controller.
type ClientTemplates struct {
	List   *template.Template
	Add    *template.Template
	Update *template.Template
}

// Clients is a web api controller.
// Exposes functionality and web views to manage client entity.
type Clients struct {
	log       logger.Logger
	config    Config
	clients   *clients.Service
	templates ClientTemplates
}

// NewClients is a constructor for clients controller.
func NewClients(log logger.Logger, config Config, clients *clients.Service) *Clients {
	controller := &Clients{
		log:     log,
		clients: clients,
		config:  config,
	}

	// TODO: process error.
	err := controller.InitializeTemplates()
	if err != nil {
		panic(err)
	}

	return controller
}

// InitializeTemplates initializes and caches templates for clients controller.
func (controller *Clients) InitializeTemplates() (err error) {
	controller.templates.List, err = template.ParseFiles(filepath.Join(controller.config.StaticDir, "clients", "list.html"))
	if err != nil {
		return err
	}

	controller.templates.Add, err = template.ParseFiles(filepath.Join(controller.config.StaticDir, "clients", "create.html"))
	if err != nil {
		return err
	}

	controller.templates.Update, err = template.ParseFiles(filepath.Join(controller.config.StaticDir, "clients", "update.html"))
	if err != nil {
		return err
	}

	return nil
}

// Create is an endpoint that handles create client web page on GET request and
// try to create client on POST request.
func (controller *Clients) Create(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	switch r.Method {
	case http.MethodGet:
		err := controller.templates.Add.Execute(w, nil)
		if err != nil {
			controller.log.Error("can not execute list clients template", ClientsError.Wrap(err))
			http.Error(w, ClientsError.Wrap(err).Error(), http.StatusInternalServerError)
			return
		}
	case http.MethodPost:
		err := r.ParseForm()
		if err != nil {
			controller.log.Error("can not parse html form while post create.html template", ClientsError.Wrap(err))
			http.Error(w, ClientsError.Wrap(err).Error(), http.StatusBadRequest)
			return
		}

		firstName := r.Form["first-name"]
		if len(firstName) == 0 {
			http.Error(w, ClientsError.New("first-name parameter is not found").Error(), http.StatusBadRequest)
			return
		}
		lastName := r.Form["last-name"]
		if len(lastName) == 0 {
			http.Error(w, ClientsError.New("last-name parameter is not found").Error(), http.StatusBadRequest)
			return
		}
		email := r.Form["email"]
		if len(email) == 0 {
			http.Error(w, ClientsError.New("email parameter is not found").Error(), http.StatusBadRequest)
			return
		}
		phone := r.Form["phone"]
		if len(phone) == 0 {
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}

		err = controller.clients.Create(ctx, email[0], phone[0], firstName[0], lastName[0])
		if err != nil {
			controller.log.Error("can not create client", ClientsError.Wrap(err))
			http.Error(w, ClientsError.Wrap(err).Error(), http.StatusInternalServerError)
			return
		}

		r = r.Clone(ctx)
		r.Method = http.MethodGet
		http.Redirect(w, r, "/clients", http.StatusMovedPermanently)
	}
}

// Update is an endpoint that handles update client web page on GET request and
// try to update client on POST request.
func (controller *Clients) Update(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	params := mux.Vars(r)
	idParam, ok := params["id"]
	if !ok {
		http.Error(w, ClientsError.New("error parsing segment parameters. ID expected").Error(), http.StatusBadRequest)
		return
	}

	clientID, err := uuid.Parse(idParam)
	if err != nil {
		http.Error(w, ClientsError.New("error parsing segment parameters. Id is not valid.").Error(), http.StatusBadRequest)
		return
	}

	switch r.Method {
	case http.MethodGet:
		client, err := controller.clients.Get(ctx, clientID)
		if err != nil {
			controller.log.Error("could not get client", ClientsError.Wrap(err))
			http.Error(w, ClientsError.Wrap(err).Error(), http.StatusNotFound)
			return
		}

		err = controller.templates.Update.Execute(w, client)
		if err != nil {
			controller.log.Error("can not execute list clients template", ClientsError.Wrap(err))
			http.Error(w, ClientsError.Wrap(err).Error(), http.StatusInternalServerError)
			return
		}
	case http.MethodPost:
		err := r.ParseForm()
		if err != nil {
			controller.log.Error("can not parse html form while post update.html template", ClientsError.Wrap(err))
			http.Error(w, ClientsError.Wrap(err).Error(), http.StatusBadRequest)
			return
		}

		firstName := r.Form["first-name"][0]
		lastName := r.Form["last-name"][0]
		email := r.Form["email"][0]
		phone := r.Form["phone"][0]

		client := clients.Client{
			ID:        clientID,
			FirstName: firstName,
			LastName:  lastName,
			Email:     email,
			Phone:     phone,
		}

		err = controller.clients.Update(ctx, client)
		if err != nil {
			controller.log.Error("can not update client", ClientsError.Wrap(err))
			//if adminportal.ValidationError.Has(err) {
			//	http.Error(w, ClientsError.Wrap(err).Error(), http.StatusBadRequest)
			//	return
			//}

			http.Error(w, ClientsError.Wrap(err).Error(), http.StatusInternalServerError)
			return
		}

		r = r.Clone(ctx)
		r.Method = http.MethodGet
		http.Redirect(w, r, "/clients", http.StatusMovedPermanently)
	}
}

// List is an endpoint that will provide a web page with all clients.
func (controller *Clients) List(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	clients, err := controller.clients.List(ctx)
	if err != nil {
		controller.log.Error("can not list clients", ClientsError.Wrap(err))
		http.Error(w, ClientsError.Wrap(err).Error(), http.StatusInternalServerError)
		return
	}

	err = controller.templates.List.Execute(w, clients)
	if err != nil {
		controller.log.Error("can not execute list clients template", ClientsError.Wrap(err))
		http.Error(w, ClientsError.Wrap(err).Error(), http.StatusInternalServerError)
		return
	}
}

// Delete is an endpoint that will delete client.
func (controller *Clients) Delete(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	params := mux.Vars(r)
	idParam, ok := params["id"]
	if !ok {
		http.Error(w, ClientsError.New("error parsing segment parameters. Id expected").Error(), http.StatusBadRequest)
		return
	}

	clientID, err := uuid.Parse(idParam)
	if err != nil {
		http.Error(w, ClientsError.New("error parsing segment parameters. Id is not valid.").Error(), http.StatusBadRequest)
		return
	}

	err = controller.clients.Delete(ctx, clientID)
	if err != nil {
		controller.log.Error("could not delete client", ClientsError.Wrap(err))
		http.Error(w, ClientsError.Wrap(err).Error(), http.StatusBadRequest)
		return
	}

	r = r.Clone(ctx)
	r.Method = http.MethodGet
	http.Redirect(w, r, "/clients", http.StatusMovedPermanently)
}
