package api

import (
	"encoding/json"
	"net/http"
	"strings"
	"git.sapienzaapps.it/fantasticcoffee/fantastic-coffee-decaffeinated/service/api/reqcontext"
	"github.com/julienschmidt/httprouter"
)

// getMyConversations handles the GET /conversations endpoint.
// It requires a valid Bearer token in the Authorization header.
// It returns a list of conversations for the authenticated user.
func (rt *_router) getMyConversations(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
	// Extract the token from the Authorization header
	authHeader := r.Header.Get("Authorization")
	if !strings.HasPrefix(authHeader, "Bearer ") {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	token := strings.TrimPrefix(authHeader, "Bearer ")

	// Validate the token
	username, valid := rt.validTokens[token]
	if !valid {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	// Get the conversations for the user
	conversations, exists := rt.conversations[username]
	if !exists {
		conversations = []string{}
	}

	// Return the conversations
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(conversations)
}
