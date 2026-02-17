package api

import (
	"net/http"
	"strings"

	"git.sapienzaapps.it/fantasticcoffee/fantastic-coffee-decaffeinated/service/api/reqcontext"
	"github.com/julienschmidt/httprouter"
)

// deleteMessage handles DELETE /messages/{messageId}
func (rt *_router) deleteMessage(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
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

	// 2. Get Message ID
	messageID := ps.ByName("messageId")

	// 3. Find Conversation
	conversationID, exists := rt.messagesMap[messageID]
	if !exists {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	conversation, exists := rt.conversationsData[conversationID]
	if !exists {
		// Should not happen if messagesMap is consistent, but safety first
		w.WriteHeader(http.StatusNotFound)
		return
	}

	// 4. Check Participant
	isParticipant := false
	for _, p := range conversation.Participants {
		if p == username {
			isParticipant = true
			break
		}
	}
	if !isParticipant {
		w.WriteHeader(http.StatusForbidden) // or 404 per requirements
		return
	}

	// 5. Delete (Mark as Deleted)
	found := false
	for i, msg := range conversation.Messages {
		if msg.ID == messageID {
			conversation.Messages[i].Deleted = true
			found = true
			break
		}
	}

	if !found {
		// This might happen if we had map entry but message was already hard-deleted or logic mismatch
		w.WriteHeader(http.StatusNotFound)
		return
	}

	// Save back
	rt.conversationsData[conversationID] = conversation

	// 6. Response
	w.WriteHeader(http.StatusNoContent)
}
