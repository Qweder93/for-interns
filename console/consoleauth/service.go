// Copyright (C) 2020 Creditor Corp. Group.
// See LICENSE for copying information.

package consoleauth

import (
	"context"
	"crypto/subtle"
	"time"

	"github.com/zeebo/errs"

	"cleanmasters/clients"
	"cleanmasters/internal/auth"
)

var (
	// Error is an internal error class for auth service.
	Error = errs.Class("console authentication error")
)

const (
	// AuthTokenDuration is an expiration duration for auth token.
	AuthTokenDuration = 24 * time.Hour
)

// Service exposes all console authentication rules.
type Service struct {
	clients clients.Service
	signer  auth.TokenSigner
}

// Login validates token from context and returns authorized Authorization.
func (service *Service) Login(ctx context.Context, token, phone string) (auth.Token, error) {
	// TODO: add firebase token check
	//err := validateToken(token)
	//if err != nil {
	//	return auth.Token{}, Error.Wrap(err)
	//}

	client, err := service.clients.GetByPhone(ctx, phone)
	if err != nil {
		if !clients.ErrNotExist.Has(err) {
			return auth.Token{}, Error.Wrap(err)
		}

		id, err := service.clients.Register(ctx, phone)
		if err != nil {
			return auth.Token{}, Error.Wrap(err)
		}
		client.ID = id
	}

	claims := auth.Claims{
		ID:        client.ID,
		ExpiresAt: time.Now().Add(AuthTokenDuration),
	}

	authToken, err := service.signer.CreateToken(ctx, claims)

	return authToken, Error.Wrap(err)
}

// Authorize validates token from context and returns authorized Authorization.
func (service *Service) Authorize(ctx context.Context) (auth.Claims, error) {
	tokenBytes, err := auth.GetToken(ctx)
	if err != nil {
		return auth.Claims{}, Error.Wrap(err)
	}

	token, err := auth.FromBase64URLString(string(tokenBytes))
	if err != nil {
		return auth.Claims{}, Error.Wrap(err)
	}

	claims, err := service.Authenticate(ctx, token)
	if err != nil {
		return auth.Claims{}, Error.Wrap(err)
	}

	err = service.authorize(ctx, claims)
	if err != nil {
		return auth.Claims{}, err
	}

	return *claims, nil
}

// Authenticate validates token signature.
func (service *Service) Authenticate(ctx context.Context, token auth.Token) (_ *auth.Claims, err error) {
	signature := token.Signature

	err = service.signer.SignToken(&token)
	if err != nil {
		return nil, err
	}

	if subtle.ConstantTimeCompare(signature, token.Signature) != 1 {
		return nil, Error.New("incorrect signature")
	}

	claims, err := auth.FromJSON(token.Payload)
	if err != nil {
		return nil, err
	}

	return claims, nil
}

// authorize checks claims.
func (service *Service) authorize(ctx context.Context, claims *auth.Claims) (err error) {
	if !claims.ExpiresAt.IsZero() && claims.ExpiresAt.Before(time.Now()) {
		return Error.Wrap(err)
	}

	_, err = service.clients.Get(ctx, claims.ID)
	if err != nil {
		return Error.Wrap(err)
	}

	return nil
}
