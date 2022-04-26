// Copyright (C) 2020 Creditor Corp. Group.
// See LICENSE for copying information.

package cleanmasters

import (
	"context"
	"errors"
	"net"

	"golang.org/x/sync/errgroup"

	"cleanmasters/adminportal/adminauth"
	"cleanmasters/adminportal/adminportalweb"
	"cleanmasters/adminportal/managers"
	"cleanmasters/clients"
	consoleserver "cleanmasters/console/server"
	"cleanmasters/internal/auth"
	"cleanmasters/internal/logger"
)

// DB provides access to all databases and database related functionality.
//
// architecture: Master Database
type DB interface {
	// Clients provides access to the clients database.
	Clients() clients.DB
	// Managers provides access to the managers database.
	Managers() managers.DB

	// Close closes underlying db connection.
	Close() error
	// CreateSchema create tables.
	CreateSchema(ctx context.Context) (err error)
}

// Config is the global configuration for cleanmasters service.
type Config struct {
	Console struct {
		Endpoint     consoleserver.Config
		SignerSecret string
	}
	AdminPortal struct {
		Endpoint     adminportalweb.Config
		SignerSecret string
	}
}

// Peer is the representation of a cleanmasters bank service.
type Peer struct {
	Config   Config
	Listener net.Listener
	Log      logger.Logger
	Database DB

	// contains logic of clients domain.
	Clients struct {
		Service *clients.Service
	}

	// Web server with web api.
	Console struct {
		Listener net.Listener
		Endpoint *consoleserver.Server
		Signer   *auth.TokenSigner
	}

	// Administrator portal mor managers to manage everything.
	AdminPortal struct {
		Signer         *auth.TokenSigner
		Authentication *adminauth.Service
		Managers       *managers.Service
		Listener       net.Listener
		Endpoint       *adminportalweb.Server
	}
}

// NewPeer is a constructor for cleanmasters Peer.
func NewPeer(log logger.Logger, db DB, config Config) (peer *Peer, err error) {
	peer = &Peer{
		Log:      log,
		Database: db,
		Config:   config,
	}

	{ // clients setup
		peer.Clients.Service = clients.NewService(
			peer.Database.Clients(),
		)
	}

	{ // managers setup
		peer.Clients.Service = clients.NewService(
			peer.Database.Clients(),
		)
	}

	{ // console setup
		peer.Console.Listener, err = net.Listen("tcp", config.Console.Endpoint.Address)
		if err != nil {
			return nil, err
		}

		peer.Console.Endpoint, err = consoleserver.NewServer(
			peer.Log,
			config.Console.Endpoint,
			peer.Clients.Service,
			nil, // TODO: add auth service
			peer.Console.Listener,
		)
		if err != nil {
			return nil, err
		}
	}

	{ // admin portal setup
		peer.AdminPortal.Listener, err = net.Listen("tcp", config.AdminPortal.Endpoint.Address)
		if err != nil {
			return nil, err
		}

		peer.AdminPortal.Signer = auth.NewTokenSigner(peer.Config.AdminPortal.SignerSecret)

		peer.AdminPortal.Managers = managers.NewService(peer.Database.Managers())

		peer.AdminPortal.Authentication = adminauth.NewService(peer.AdminPortal.Signer, peer.AdminPortal.Managers)

		peer.AdminPortal.Endpoint = adminportalweb.NewServer(
			peer.Log,
			peer.Config.AdminPortal.Endpoint,
			peer.AdminPortal.Authentication,
			peer.Clients.Service,
			peer.AdminPortal.Managers,
			peer.AdminPortal.Listener,
		)
	}

	return peer, nil
}

// Run runs cleanmasters console Peer until it's either closed or it errors.
func (peer *Peer) Run(ctx context.Context) error {
	group, ctx := errgroup.WithContext(ctx)

	// start cleanmasters web api server as a separate goroutine.
	group.Go(func() error {
		return ignoreCancel(peer.Console.Endpoint.Run(ctx))
	})

	return group.Wait()
}

// RunAdmin runs cleanmasters admin panel Peer until it's either closed or it errors.
func (peer *Peer) RunAdmin(ctx context.Context) error {
	group, ctx := errgroup.WithContext(ctx)

	// start cleanmasters web api server as a separate goroutine.
	group.Go(func() error {
		return ignoreCancel(peer.AdminPortal.Endpoint.Run(ctx))
	})

	return group.Wait()
}

// Close closes all the resources.
func (peer *Peer) Close() error {
	if peer.Console.Endpoint != nil {
		return peer.Console.Endpoint.Close()
	}

	if peer.AdminPortal.Endpoint != nil {
		return peer.AdminPortal.Endpoint.Close()
	}

	return nil
}

// we ignore cancellation and stopping errors since they are expected.
func ignoreCancel(err error) error {
	if errors.Is(err, context.Canceled) {
		return nil
	}

	return err
}
