package api

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"git.sapienzaapps.it/fantasticcoffee/fantastic-coffee-decaffeinated/service/api/reqcontext"
	"github.com/julienschmidt/httprouter"
)

// commentMessage handles POST /messages/{messageId}/comment
func (rt *_router) commentMessage(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
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
		w.WriteHeader(http.StatusNotFound) // or 403
		return
	}

	// 5. Find Message
	msgIndex := -1
	for i, msg := range conversation.Messages {
		if msg.ID == messageID {
			msgIndex = i
			break
		}
	}
	if msgIndex == -1 {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	if conversation.Messages[msgIndex].Deleted {
		w.WriteHeader(http.StatusConflict) // 409 Conflict for soft-deleted message
		return
	}

	// 6. Parse Body
	var body struct {
		Comment string `json:"comment"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if body.Comment == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// 7. Update Message
	conversation.Messages[msgIndex].Comment = body.Comment
	conversation.Messages[msgIndex].CommentedAt = time.Now()
	rt.conversationsData[conversationID] = conversation

	w.WriteHeader(http.StatusNoContent)
}

// uncommentMessage handles DELETE /messages/{messageId}/comment
func (rt *_router) uncommentMessage(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
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
		w.WriteHeader(http.StatusNotFound)
		return
	}

	// 5. Find Message
	msgIndex := -1
	for i, msg := range conversation.Messages {
		if msg.ID == messageID {
			msgIndex = i
			break
		}
	}
	if msgIndex == -1 {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	if conversation.Messages[msgIndex].Deleted {
		w.WriteHeader(http.StatusConflict)
		return
	}

	// 6. Remove Comment
	conversation.Messages[msgIndex].Comment = ""
	conversation.Messages[msgIndex].CommentedAt = time.Time{}
	rt.conversationsData[conversationID] = conversation

	w.WriteHeader(http.StatusNoContent)
}
