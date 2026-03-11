package api

import (
	"net/http"
	"strings"

	"git.sapienzaapps.it/fantasticcoffee/fantastic-coffee-decaffeinated/service/api/reqcontext"
	"github.com/julienschmidt/httprouter"
)

// getMyPhoto handles GET /me/photo
func (rt *_router) getMyPhoto(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
	// 1. Auth check
	authHeader := r.Header.Get("Authorization")
	if !strings.HasPrefix(authHeader, "Bearer ") {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	token := strings.TrimPrefix(authHeader, "Bearer ")
	username, err := rt.db.GetUserByToken(token)
	if err != nil {
		ctx.Logger.WithError(err).Error("error getting user by token")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if username == "" {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	// 2. Get Photo from DB
	photo, err := rt.db.GetUserPhoto(username)
	if err != nil {
		ctx.Logger.WithError(err).Error("error getting user photo from db")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if photo == nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	// 3. Return Photo
	// Note: We don't store MIME type, but we usually default to image/jpeg or detect it
	w.Header().Set("Content-Type", "image/jpeg")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(photo)
}
