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

	"cleanmasters/adminportal/managers"
	"cleanmasters/internal/logger"
)

var (
	// ManagersError is an internal error type for managers controller.
	ManagersError = errs.Class("managers controller error")
)

// ManagerTemplates holds templates needed for managers controller.
type ManagerTemplates struct {
	List   *template.Template
	Add    *template.Template
	Update *template.Template
}

// Managers is a web api controller.
// Exposes functionality and web views to manage manager entity.
type Managers struct {
	log       logger.Logger
	config    Config
	managers  *managers.Service
	templates ManagerTemplates
}

// NewManagers is a constructor for managers controller.
func NewManagers(log logger.Logger, config Config, managers *managers.Service) *Managers {
	managersController := &Managers{
		log:      log,
		managers: managers,
		config:   config,
	}

	// TODO: process error.
	err := managersController.initializeTemplates()
	if err != nil {
		panic(err)
	}

	return managersController
}

// initializeTemplates initializes and caches templates for managers controller.
func (controller *Managers) initializeTemplates() (err error) {
	controller.templates.List, err = template.ParseFiles(filepath.Join(controller.config.StaticDir, "managers", "list.html"))
	if err != nil {
		return err
	}

	controller.templates.Add, err = template.ParseFiles(filepath.Join(controller.config.StaticDir, "managers", "create.html"))
	if err != nil {
		return err
	}

	controller.templates.Update, err = template.ParseFiles(filepath.Join(controller.config.StaticDir, "managers", "update.html"))

	return err
}

// Create is an endpoint that handles create manager web page on GET request and
// tries to create manager on POST request.
func (controller *Managers) Create(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	switch r.Method {
	case http.MethodGet:
		err := controller.templates.Add.Execute(w, nil)
		if err != nil {
			controller.log.Error("can not execute add managers template", ManagersError.Wrap(err))
			http.Error(w, ManagersError.Wrap(err).Error(), http.StatusInternalServerError)
			return
		}
	case http.MethodPost:
		err := r.ParseForm()
		if err != nil {
			controller.log.Error("can not parse html form while post create.html template", ManagersError.Wrap(err))
			http.Error(w, ManagersError.Wrap(err).Error(), http.StatusBadRequest)
			return
		}

		// TODO: create separate function - parse manager from Form.
		password := r.Form["password"]
		if len(password) == 0 {
			http.Error(w, ClientsError.New("password parameter is not found").Error(), http.StatusBadRequest)
			return
		}
		firstName := r.Form["first-name"]
		if len(firstName) == 0 {
			http.Error(w, ClientsError.New("firstName parameter is not found").Error(), http.StatusBadRequest)
			return
		}
		lastName := r.Form["last-name"]
		if len(lastName) == 0 {
			http.Error(w, ClientsError.New("lastName parameter is not found").Error(), http.StatusBadRequest)
			return
		}
		email := r.Form["email"]
		if len(email) == 0 {
			http.Error(w, ClientsError.New("email parameter is not found").Error(), http.StatusBadRequest)
			return
		}

		err = controller.managers.Create(ctx, password[0], firstName[0], lastName[0], email[0])
		if err != nil {
			//if adminportal.ValidationError.Has(err) {
			//	controller.log.Error("can not create manager", zap.Error(ManagersError.Wrap(err)))
			//	http.Error(w, ManagersError.Wrap(err).Error(), http.StatusBadRequest)
			//	return
			//}

			controller.log.Error("can not create manager", ManagersError.Wrap(err))
			http.Error(w, ManagersError.Wrap(err).Error(), http.StatusInternalServerError)
			return
		}

		r = r.Clone(ctx)
		r.Method = http.MethodGet
		http.Redirect(w, r, "/managers", http.StatusMovedPermanently)
	}
}

// Update is an endpoint that handles update manager web page on GET request and
// tries to update manager on POST request.
func (controller *Managers) Update(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	params := mux.Vars(r)
	idParam, ok := params["id"]
	if !ok {
		http.Error(w, ManagersError.New("error parsing segment parameters. ID expected").Error(), http.StatusBadRequest)
		return
	}

	managerID, err := uuid.Parse(idParam)
	if err != nil {
		http.Error(w, ManagersError.New("error parsing segment parameters. Id is not valid.").Error(), http.StatusBadRequest)
		return
	}

	switch r.Method {
	case http.MethodGet:
		manager, err := controller.managers.Get(ctx, managerID)
		if err != nil {
			controller.log.Error("could not get manager", ManagersError.Wrap(err))
			http.Error(w, ManagersError.Wrap(err).Error(), http.StatusNotFound)
			return
		}

		err = controller.templates.Update.Execute(w, manager)
		if err != nil {
			controller.log.Error("can not execute update managers template", ManagersError.Wrap(err))
			http.Error(w, ManagersError.Wrap(err).Error(), http.StatusInternalServerError)
			return
		}
	case http.MethodPost:
		err := r.ParseForm()
		if err != nil {
			controller.log.Error("can not parse html form while post update.html template", ManagersError.Wrap(err))
			http.Error(w, ManagersError.Wrap(err).Error(), http.StatusBadRequest)
			return
		}

		// TODO: process form in a better way
		password := r.Form["password"][0]
		firstName := r.Form["first-name"][0]
		lastName := r.Form["last-name"][0]
		email := r.Form["email"][0]

		manager := managers.ManagerUpdateFields{
			FirstName: firstName,
			LastName:  lastName,
			Email:     email,
			Password:  password,
		}

		err = controller.managers.Update(ctx, managerID, manager)
		if err != nil {
			controller.log.Error("can not update manager", ManagersError.Wrap(err))
			//if adminportal.ValidationError.Has(err) {
			//	http.Error(w, ManagersError.Wrap(err).Error(), http.StatusBadRequest)
			//	return
			//}

			http.Error(w, ManagersError.Wrap(err).Error(), http.StatusInternalServerError)
			return
		}

		r = r.Clone(ctx)
		r.Method = http.MethodGet
		http.Redirect(w, r, "/managers", http.StatusMovedPermanently)
	}
}

// List is an endpoint that will provide a web page with all managers.
func (controller *Managers) List(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	managers, err := controller.managers.List(ctx)
	if err != nil {
		controller.log.Error("can not list managers", ManagersError.Wrap(err))
		http.Error(w, ManagersError.Wrap(err).Error(), http.StatusInternalServerError)
		return
	}

	err = controller.templates.List.Execute(w, managers)
	if err != nil {
		controller.log.Error("can not execute list managers template", ManagersError.Wrap(err))
		http.Error(w, ManagersError.Wrap(err).Error(), http.StatusInternalServerError)
		return
	}
}

// Delete is an endpoint that will delete manager.
func (controller *Managers) Delete(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	params := mux.Vars(r)
	idParam, ok := params["id"]
	if !ok {
		http.Error(w, ManagersError.New("error parsing segment parameters. Id expected").Error(), http.StatusBadRequest)
		return
	}

	managerID, err := uuid.Parse(idParam)
	if err != nil {
		http.Error(w, ManagersError.New("error parsing segment parameters. Id is not valid.").Error(), http.StatusBadRequest)
		return
	}

	err = controller.managers.Delete(ctx, managerID)
	if err != nil {
		controller.log.Error("could not delete manager", ManagersError.Wrap(err))
		http.Error(w, ManagersError.Wrap(err).Error(), http.StatusBadRequest)
		return
	}

	r = r.Clone(ctx)
	r.Method = http.MethodGet
	http.Redirect(w, r, "/managers", http.StatusMovedPermanently)
}
