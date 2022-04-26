package auth

import (
	"bytes"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/zeebo/errs"
)

// ErrClaims is an error class for Claims errors.
var ErrClaims = errs.Class("auth claims error")

// Claims represents data signed by server and used for authentication.
type Claims struct {
	ID        uuid.UUID `json:"id"`
	ExpiresAt time.Time `json:"expiresAt"`
}

// JSON returns json representation of Claims.
func (c *Claims) JSON() ([]byte, error) {
	buffer := bytes.NewBuffer(nil)

	err := json.NewEncoder(buffer).Encode(c)
	return buffer.Bytes(), ErrClaims.Wrap(err)
}

// FromJSON returns Claims instance, parsed from JSON.
func FromJSON(data []byte) (*Claims, error) {
	claims := new(Claims)

	err := json.NewDecoder(bytes.NewReader(data)).Decode(claims)
	if err != nil {
		return nil, ErrClaims.Wrap(err)
	}

	return claims, nil
}
