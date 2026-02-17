package api

import (
	"encoding/json"
	"net/http"
	"strings"
	"git.sapienzaapps.it/fantasticcoffee/fantastic-coffee-decaffeinated/service/api/reqcontext"
	"github.com/julienschmidt/httprouter"
)

// setMyUserName handles the PUT /me/name endpoint.
// It requires a valid Bearer token.
// It updates the username of the authenticated user.
func (rt *_router) setMyUserName(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
	// Extract the token
	authHeader := r.Header.Get("Authorization")
	if !strings.HasPrefix(authHeader, "Bearer ") {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	token := strings.TrimPrefix(authHeader, "Bearer ")

	// Validate the token
	oldName, valid := rt.validTokens[token]
	if !valid {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	// Parse the request body
	var body struct {
		Name string `json:"name"`
	}
	err := json.NewDecoder(r.Body).Decode(&body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Validate the new name
	if body.Name == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Check if the new name is already taken
	if _, exists := rt.users[body.Name]; exists {
		// New name is taken (even if it is by the same user, we can consider it a conflict or just no-op)
		// If it's the same user, it's a no-op 204.
		if body.Name == oldName {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		// Otherwise 400 Bad Request (Conflict)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Create atomic-like update
	// Delete old user mapping
	delete(rt.users, oldName)
	// Add new user mapping
	rt.users[body.Name] = token
	// Update token mapping
	rt.validTokens[token] = body.Name
	// Move conversations
	if convs, exists := rt.conversations[oldName]; exists {
		rt.conversations[body.Name] = convs
		delete(rt.conversations, oldName)
	}

	w.WriteHeader(http.StatusNoContent)
}
