// Copyright © 2020 Krishna Iyer Easwaran
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package middleware provides the http middleware functions.
package middleware

import (
	"crypto/subtle"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
)

// BasicAuth contains http basic auth credentials
type BasicAuth struct {
	Credentials map[string]string
}

// NewBasicAuth returns a new Basic auth.
func NewBasicAuth(username, password string) *BasicAuth {
	return &BasicAuth{
		Credentials: make(map[string]string),
	}
}

// Add adds a new set of credentials.
func (auth BasicAuth) Add(username, password string) {
	auth.Credentials[username] = password
}

// Validate validates the provided credentials.
func (auth BasicAuth) Validate(username, password string) bool {
	return (subtle.ConstantTimeCompare([]byte(password), []byte(auth.Credentials[username])) == 1)
}

// Middleware returns a middleware that validates the request for basic auth.
func (auth BasicAuth) Middleware(realm string) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			username, password, ok := r.BasicAuth()
			if !ok || !auth.Validate(username, password) {
				w.Header().Add("WWW-Authenticate", fmt.Sprintf(`Basic realm="%s"`, realm))
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
