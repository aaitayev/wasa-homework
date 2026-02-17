package api

import (
	"encoding/json"
	"github.com/gofrs/uuid"
	"net/http"
	"git.sapienzaapps.it/fantasticcoffee/fantastic-coffee-decaffeinated/service/api/reqcontext"
	"github.com/julienschmidt/httprouter"
)

// doLogin handles the POST /session endpoint.
// It reads a JSON body {"name": "..."}.
// If the name is missing or empty, it returns 400.
// If the user does not exist, it creates a new user and returns a 201 with the identifier.
// If the user exists, it returns a 201 with the existing identifier.
func (rt *_router) doLogin(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
	// Parse the request body
	var user struct {
		Name string `json:"name"`
	}
	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Validate the name
	if user.Name == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Check if the user exists
	identifier, exists := rt.users[user.Name]
	if !exists {
		// Create a new user
		ident, err := uuid.NewV4()
		if err != nil {
			ctx.Logger.WithError(err).Error("error creating uuid")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		identifier = ident.String()
		rt.users[user.Name] = identifier
		rt.validTokens[identifier] = user.Name
	} else {
		rt.validTokens[identifier] = user.Name
	}

	// Return the identifier
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(struct {
		Identifier string `json:"identifier"`
	}{
		Identifier: identifier,
	})
}
