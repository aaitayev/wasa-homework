package api

import (
	"encoding/json"
	"net/http"
	"strings"

	"https://github.com/aaitayev/wasa-homework.git"
	"github.com/julienschmidt/httprouter"
)


// searchUsers handles GET /users endpoint
func (rt *_router) searchUsers(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
	// 1. Auth check
	authHeader := r.Header.Get("Authorization")
	if !strings.HasPrefix(authHeader, "Bearer ") {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	token := strings.TrimPrefix(authHeader, "Bearer ")
	callingUser, err := rt.db.GetUserByToken(token)
	if err != nil {
		ctx.Logger.WithError(err).Error("error getting calling user by token")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if callingUser == "" {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	// 2. Read query string
	searchQuery := r.URL.Query().Get("search")

	// 3. Search in DB
	users, err := rt.db.SearchUsers(searchQuery)
	if err != nil {
		ctx.Logger.WithError(err).Error("error searching users in db")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// 4. Transform and Filter
	type UserResponse struct {
		Name string `json:"name"`
	}
	results := make([]UserResponse, 0)
	for _, username := range users {
		if username == callingUser {
			continue // Omit the requester
		}
		results = append(results, UserResponse{Name: username})
	}

	// 5. Return results
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(results)
}


