package api

import (
	"io"
	"net/http"
	"strings"

	"git.sapienzaapps.it/fantasticcoffee/fantastic-coffee-decaffeinated/service/api/reqcontext"
	"github.com/julienschmidt/httprouter"
)


// setGroupPhoto handles the PUT /groups/:groupId/photo endpoint.
func (rt *_router) setGroupPhoto(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
	// Extract the token
	authHeader := r.Header.Get("Authorization")
	if !strings.HasPrefix(authHeader, "Bearer ") {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	token := strings.TrimPrefix(authHeader, "Bearer ")

	// 1. Auth check
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

	groupID := ps.ByName("groupId")
	if groupID == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// 2. Verify group exists and is a group
	group, err := rt.db.GetConversation(groupID)
	if err != nil {
		ctx.Logger.WithError(err).Error("error getting group from db")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if group == nil || !group.IsGroup {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	// 3. Verify user is participant
	isParticipant := false
	for _, p := range group.Participants {
		if p == username {
			isParticipant = true
			break
		}
	}
	if !isParticipant {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	// 4. Validate Content-Type
	contentType := r.Header.Get("Content-Type")
	if contentType != "image/jpeg" && contentType != "image/png" {
		w.WriteHeader(http.StatusUnsupportedMediaType)
		return
	}

	// 5. Read body up to 5MB
	const maxMemory = 5 * 1024 * 1024
	r.Body = http.MaxBytesReader(w, r.Body, maxMemory)

	photoBytes, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusRequestEntityTooLarge)
		return
	}

	// 6. Store photo in DB
	err = rt.db.SetGroupPhoto(groupID, photoBytes, contentType)
	if err != nil {
		ctx.Logger.WithError(err).Error("error setting group photo in db")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

