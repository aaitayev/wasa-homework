package api

import (
	"encoding/json"
	"net/http"
	"strings"

	"https://github.com/aaitayev/wasa-homework.git"
	"github.com/julienschmidt/httprouter"
)


// addToGroup handles POST /groups/{groupId}/members
func (rt *_router) addToGroup(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
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

	// 2. Get Group ID
	groupID := ps.ByName("groupId")
	conversation, err := rt.db.GetConversation(groupID)
	if err != nil {
		ctx.Logger.WithError(err).Error("error getting conversation from db")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if conversation == nil || !conversation.IsGroup {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	// 3. Check Requester Participation
	isParticipant := false
	for _, p := range conversation.Participants {
		if p == username {
			isParticipant = true
			break
		}
	}
	if !isParticipant {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	// 4. Parse Body
	var body struct {
		MemberID string `json:"memberId"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if body.MemberID == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// 5. Check if member-to-be-added exists
	existingUser, err := rt.db.GetUserByName(body.MemberID)
	if err != nil {
		ctx.Logger.WithError(err).Error("error checking user existence in db")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if existingUser == nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	// 6. Add Member
	err = rt.db.AddParticipant(groupID, body.MemberID)
	if err != nil {
		ctx.Logger.WithError(err).Error("error adding participant to group in db")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// leaveGroup handles POST /groups/{groupId}/leave
func (rt *_router) leaveGroup(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
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

	// 2. Get Group ID
	groupID := ps.ByName("groupId")
	conversation, err := rt.db.GetConversation(groupID)
	if err != nil {
		ctx.Logger.WithError(err).Error("error getting conversation from db")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if conversation == nil || !conversation.IsGroup {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	// 3. Check/Remove Participant
	isParticipant := false
	for _, p := range conversation.Participants {
		if p == username {
			isParticipant = true
			break
		}
	}
	if !isParticipant {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	// Remove Participant
	err = rt.db.RemoveParticipant(groupID, username)
	if err != nil {
		ctx.Logger.WithError(err).Error("error removing participant from db")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// setGroupName handles PUT /groups/{groupId}/name
func (rt *_router) setGroupName(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
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

	// 2. Get Group ID
	groupID := ps.ByName("groupId")
	conversation, err := rt.db.GetConversation(groupID)
	if err != nil {
		ctx.Logger.WithError(err).Error("error getting conversation from db")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if conversation == nil || !conversation.IsGroup {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	// 3. Check Requester Participation
	isParticipant := false
	for _, p := range conversation.Participants {
		if p == username {
			isParticipant = true
			break
		}
	}
	if !isParticipant {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	// 4. Parse Body
	var body struct {
		Name string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if body.Name == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// 5. Update Name
	err = rt.db.UpdateConversationName(groupID, body.Name)
	if err != nil {
		ctx.Logger.WithError(err).Error("error updating conversation name in db")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

