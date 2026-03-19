package api

import (
	"net/http"
	"strings"

	"github.com/aaitayev/wasa-homework"
	"github.com/julienschmidt/httprouter"
)

// getGroupPhoto handles GET /groups/:groupId/photo
func (rt *_router) getGroupPhoto(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
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

	groupID := ps.ByName("groupId")
	if groupID == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// 2. Verify group exists and user is participant
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

	// 3. Get Photo from DB
	photo, contentType, err := rt.db.GetGroupPhoto(groupID)
	if err != nil {
		ctx.Logger.WithError(err).Error("error getting group photo from db")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if photo == nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	// 4. Return Photo
	if contentType == "" {
		contentType = "image/jpeg" // Fallback
	}
	w.Header().Set("Content-Type", contentType)
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(photo)
}
