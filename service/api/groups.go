package api

import (
	"encoding/json"
	"net/http"
	"strings"

	"git.sapienzaapps.it/fantasticcoffee/fantastic-coffee-decaffeinated/service/api/reqcontext"
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
	username, valid := rt.validTokens[token]
	if !valid {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	// 2. Get Group ID
	groupID := ps.ByName("groupId")
	conversation, exists := rt.conversationsData[groupID]
	if !exists {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	if !conversation.IsGroup {
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
		w.WriteHeader(http.StatusForbidden) // 403 or 404
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

	// 5. Check if user exists (simplification: check internal map)
	// We only track "users" map which maps username -> token.
	// If the requirement says "memberId" is the username, we check if it exists in "users".
	// But "users" map is keyed by username.
	if _, userExists := rt.users[body.MemberID]; !userExists {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	// 6. Add Member
	// Check if already in group
	alreadyIn := false
	for _, p := range conversation.Participants {
		if p == body.MemberID {
			alreadyIn = true
			break
		}
	}

	if !alreadyIn {
		conversation.Participants = append(conversation.Participants, body.MemberID)
		rt.conversationsData[groupID] = conversation
		// Update reverse index
		rt.conversations[body.MemberID] = append(rt.conversations[body.MemberID], groupID)
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
	username, valid := rt.validTokens[token]
	if !valid {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	// 2. Get Group ID
	groupID := ps.ByName("groupId")
	conversation, exists := rt.conversationsData[groupID]
	if !exists {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	// Spec doesn't strictly say it must be IsGroup=true, but implies it.
	// "If group not found or not isGroup => 404" was for addToGroup.
	// For leaveGroup, usually applies too, or generally for any conversation.
	if !conversation.IsGroup {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	// 3. Check/Remove Participant
	isParticipant := false
	participantIndex := -1
	for i, p := range conversation.Participants {
		if p == username {
			isParticipant = true
			participantIndex = i
			break
		}
	}
	if !isParticipant {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	// Remove from Participants
	conversation.Participants = append(conversation.Participants[:participantIndex], conversation.Participants[participantIndex+1:]...)
	rt.conversationsData[groupID] = conversation

	// Remove from reverse index
	userConvs := rt.conversations[username]
	for i, cID := range userConvs {
		if cID == groupID {
			rt.conversations[username] = append(userConvs[:i], userConvs[i+1:]...)
			break
		}
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
	username, valid := rt.validTokens[token]
	if !valid {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	// 2. Get Group ID
	groupID := ps.ByName("groupId")
	conversation, exists := rt.conversationsData[groupID]
	if !exists {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	if !conversation.IsGroup {
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
		w.WriteHeader(http.StatusNotFound) // or 403
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
	conversation.Name = body.Name
	rt.conversationsData[groupID] = conversation

	w.WriteHeader(http.StatusNoContent)
}
