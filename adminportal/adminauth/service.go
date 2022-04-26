// Copyright (C) 2020 Creditor Corp. Group.
// See LICENSE for copying information.

package adminauth

import (
	"context"
	"crypto/subtle"
	"time"

	"github.com/zeebo/errs"
	"golang.org/x/crypto/bcrypt"

	"cleanmasters/adminportal/managers"
	"cleanmasters/internal/auth"
)

var (
	// Error in an internal error for admin service.
	Error = errs.Class("admin service error")
)

// Service is exposing all business logic of managers portal.
//
// architecture: Service
type Service struct {
	signer   *auth.TokenSigner
	managers *managers.Service
}

// NewService is a constructor for admin Service.
func NewService(signer *auth.TokenSigner, managers *managers.Service) *Service {
	return &Service{
		signer:   signer,
		managers: managers,
	}
}

// Token authenticates manager by credentials and returns auth token.
func (service *Service) Token(ctx context.Context, email string, password string) (token auth.Token, err error) {
	manager, err := service.managers.GetByEmail(ctx, email)
	if err != nil {
		return auth.Token{}, Error.Wrap(err)
	}

	err = bcrypt.CompareHashAndPassword(manager.PasswordHash, []byte(password))
	if err != nil {
		return auth.Token{}, Error.Wrap(err)
	}

	claims := auth.Claims{
		ID: manager.ID,
		// TODO: place duration to some const
		ExpiresAt: time.Now().Add(time.Hour * 24),
	}

	token, err = service.signer.CreateToken(ctx, claims)

	return token, Error.Wrap(err)
}

// Authorize validates token from context and returns authorized Authorization.
func (service *Service) Authorize(ctx context.Context) (auth.Claims, error) {
	tokenS, err := auth.GetToken(ctx)
	if err != nil {
		return auth.Claims{}, Error.Wrap(err)
	}

	token, err := auth.FromBase64URLString(string(tokenS))
	if err != nil {
		return auth.Claims{}, Error.Wrap(err)
	}

	claims, err := service.authenticate(token)
	if err != nil {
		return auth.Claims{}, Error.Wrap(err)
	}

	err = service.authorize(ctx, claims)
	if err != nil {
		return auth.Claims{}, Error.Wrap(err)
	}

	return *claims, nil
}

// authenticate validates token signature.
func (service *Service) authenticate(token auth.Token) (_ *auth.Claims, err error) {
	signature := token.Signature

	err = service.signer.SignToken(&token)
	if err != nil {
		return nil, err
	}

	if subtle.ConstantTimeCompare(signature, token.Signature) != 1 {
		return nil, Error.New("incorrect signature")
	}

	return auth.FromJSON(token.Payload)
}

// authorize checks claims.
func (service *Service) authorize(ctx context.Context, claims *auth.Claims) (err error) {
	if !claims.ExpiresAt.IsZero() && claims.ExpiresAt.Before(time.Now()) {
		return Error.Wrap(err)
	}

	_, err = service.managers.Get(ctx, claims.ID)

	return Error.Wrap(err)
}
