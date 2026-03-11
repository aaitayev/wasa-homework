package api

import (
	"io"
	"net/http"
	"strings"

	"git.sapienzaapps.it/fantasticcoffee/fantastic-coffee-decaffeinated/service/api/reqcontext"
	"github.com/julienschmidt/httprouter"
)

// setGroupPhoto handles the PUT /groups/:groupId/photo endpoint.
// Requires valid Bearer token.
// Requires groupId to exist and be a group (isGroup=true).
// Requires the authenticated user to be a participant of the group.
// Accepts image/png or image/jpeg up to 5MB. Returns 204 No Content.
func (rt *_router) setGroupPhoto(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
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

	groupId := ps.ByName("groupId")
	if groupId == "" {
		w.WriteHeader(http.StatusBadRequest) // or NotFound depending on framework handling
		return
	}

	// Verify group exists and is a group
	group, exists := rt.conversationsData[groupId]
	if !exists || !group.IsGroup {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	// Verify user is participant
	isParticipant := false
	for _, p := range group.Participants {
		if p == username {
			isParticipant = true
			break
		}
	}
	if !isParticipant {
		w.WriteHeader(http.StatusNotFound)
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
	rt.groupPhotos[groupId] = photoBytes

	w.WriteHeader(http.StatusNoContent)
}
