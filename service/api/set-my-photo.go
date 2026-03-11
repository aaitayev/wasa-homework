package api

import (
	"io"
	"net/http"
	"strings"

	"git.sapienzaapps.it/fantasticcoffee/fantastic-coffee-decaffeinated/service/api/reqcontext"
	"github.com/julienschmidt/httprouter"
)

// setMyPhoto handles the PUT /me/photo endpoint.
// It requires a valid Bearer token.
// It accepts an image/png or image/jpeg up to 5MB.
// It stores the image in the user's record and returns 204 No Content.
func (rt *_router) setMyPhoto(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
	// Extract the token
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

	// Validate Content-Type
	contentType := r.Header.Get("Content-Type")
	if contentType != "image/jpeg" && contentType != "image/png" {
		w.WriteHeader(http.StatusUnsupportedMediaType)
		return
	}

	// Read body up to 5MB
	const maxMemory = 5 * 1024 * 1024
	r.Body = http.MaxBytesReader(w, r.Body, maxMemory)
	
	photoBytes, err := io.ReadAll(r.Body)
	if err != nil {
		// If MaxBytesReader hits the limit, it returns an error and we should respond with 413.
		w.WriteHeader(http.StatusRequestEntityTooLarge)
		return
	}

	// Store photo
	rt.userPhotos[username] = photoBytes

	w.WriteHeader(http.StatusNoContent)
}
