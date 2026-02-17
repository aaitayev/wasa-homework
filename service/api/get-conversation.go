package api

import (
	"encoding/json"
	"net/http"
	"strings"

	"git.sapienzaapps.it/fantasticcoffee/fantastic-coffee-decaffeinated/service/api/reqcontext"
	"github.com/julienschmidt/httprouter"
)

// getConversation handles GET /conversations/{conversationId}
func (rt *_router) getConversation(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
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

	// 2. Get Conversation ID
	conversationID := ps.ByName("conversationId")

	// 3. Check Existence
	conversation, exists := rt.conversationsData[conversationID]
	if !exists {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	// 4. Check Membership
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

	// 5. Response
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(struct {
		ConversationID string    `json:"conversationId"`
		Messages       []Message `json:"messages"`
	}{
		ConversationID: conversation.ID,
		Messages:       conversation.Messages,
	})
}
