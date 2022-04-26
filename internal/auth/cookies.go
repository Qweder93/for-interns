// Copyright (C) 2020 Creditor Corp. Group.
// See LICENSE for copying information.

package auth

import (
	"net/http"
	"time"
)

// cookieDuration holds cookie expiration.
var cookieDuration = time.Hour * 24

// CookieSettings variable cookie settings.
type CookieSettings struct {
	Name string
	Path string
}

// Cookie handles cookie authorization.
type Cookie struct {
	settings CookieSettings
}

// NewCookie create new cookie authorization with provided settings.
func NewCookie(settings CookieSettings) *Cookie {
	return &Cookie{
		settings: settings,
	}
}

// GetToken retrieves token from request.
func (cookieAuth *Cookie) GetToken(r *http.Request) (string, error) {
	cookie, err := r.Cookie(cookieAuth.settings.Name)
	if err != nil {
		return "", err
	}

	return cookie.Value, nil
}

// SetToken sets parametrized token cookie that is not accessible from js.
func (cookieAuth *Cookie) SetToken(w http.ResponseWriter, token string) {
	http.SetCookie(w, &http.Cookie{
		Name:     cookieAuth.settings.Name,
		Value:    token,
		Path:     cookieAuth.settings.Path,
		Expires:  time.Now().Add(cookieDuration),
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
	})
}

// RemoveToken removes auth cookie that is not accessible from js.
func (cookieAuth *Cookie) RemoveToken(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     cookieAuth.settings.Name,
		Value:    "",
		Path:     "/",
		Expires:  time.Unix(0, 0),
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
	})
}
