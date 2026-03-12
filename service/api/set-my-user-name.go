package api

import (
	"encoding/json"
	"net/http"
	"strings"
	"https://github.com/aaitayev/wasa-homework.git"
	"github.com/julienschmidt/httprouter"
)


// setMyUserName handles the PUT /me/name endpoint.
func (rt *_router) setMyUserName(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
	// Extract the token
	authHeader := r.Header.Get("Authorization")
	if !strings.HasPrefix(authHeader, "Bearer ") {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	token := strings.TrimPrefix(authHeader, "Bearer ")

	// 1. Auth check and get old name
	oldName, err := rt.db.GetUserByToken(token)
	if err != nil {
		ctx.Logger.WithError(err).Error("error getting user by token")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if oldName == "" {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	// 2. Parse the request body
	var body struct {
		Name string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// 3. Validate name
	if body.Name == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if body.Name == oldName {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	// 4. Check if new name is taken
	existing, err := rt.db.GetUserByName(body.Name)
	if err != nil {
		ctx.Logger.WithError(err).Error("error checking name availability")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if existing != nil {
		w.WriteHeader(http.StatusBadRequest) // Name taken
		return
	}

	// 5. Update Name in DB
	err = rt.db.UpdateUserName(oldName, body.Name)
	if err != nil {
		ctx.Logger.WithError(err).Error("error updating username in db")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

