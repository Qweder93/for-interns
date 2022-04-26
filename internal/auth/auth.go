package auth

import (
	"context"

	"github.com/zeebo/errs"
)

// Error is an error class for auth errors.
var Error = errs.Class("auth error")

type key string

const (
	// keyClaims is a key to receive auth result from context - claims or error.
	keyClaims key = "claims"
	// keyToken is a key to receive auth token from context.
	keyToken key = "token"
)

// SetClaims creates new context with Claims.
func SetClaims(ctx context.Context, claims Claims) context.Context {
	return context.WithValue(ctx, keyClaims, claims)
}

// GetClaims gets Claims from context.
func GetClaims(ctx context.Context) (Claims, error) {
	value := ctx.Value(keyClaims)

	if auth, ok := value.(Claims); ok {
		return auth, nil
	}

	if err, ok := value.(error); ok {
		return Claims{}, Error.Wrap(err)
	}

	return Claims{}, Error.New("could not get auth or error from context")
}

// SetToken creates context with auth token.
func SetToken(ctx context.Context, key []byte) context.Context {
	return context.WithValue(ctx, keyToken, key)
}

// GetToken returns auth token from context is exists.
func GetToken(ctx context.Context) ([]byte, error) {
	key, ok := ctx.Value(keyToken).([]byte)
	if !ok {
		return nil, Error.New("could not take token from context")
	}
	return key, nil
}
