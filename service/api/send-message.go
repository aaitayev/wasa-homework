package api

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"git.sapienzaapps.it/fantasticcoffee/fantastic-coffee-decaffeinated/service/api/reqcontext"
	"github.com/gofrs/uuid"
	"github.com/julienschmidt/httprouter"
)

// sendMessage handles POST /messages
func (rt *_router) sendMessage(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
	// 1. Auth check
	authHeader := r.Header.Get("Authorization")
	if !strings.HasPrefix(authHeader, "Bearer ") {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	token := strings.TrimPrefix(authHeader, "Bearer ")
	senderName, valid := rt.validTokens[token]
	if !valid {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	// 2. Parse body
	var body struct {
		ConversationID string `json:"conversationId"`
		Text           string `json:"text"`
		IsGroup        bool   `json:"isGroup"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if body.Text == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	var conversationID string
	var conversation Conversation

	// 3. Handle Conversation Logic
	if body.ConversationID == "" {
		// Create new conversation
		uuidConf, err := uuid.NewV4()
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		conversationID = uuidConf.String()
		conversation = Conversation{
			ID:           conversationID,
			Participants: []string{senderName},
			Messages:     []Message{},
			IsGroup:      body.IsGroup,
		}
		rt.conversationsData[conversationID] = conversation

		// Link to user
		rt.conversations[senderName] = append(rt.conversations[senderName], conversationID)

	} else {
		// Existing conversation
		conversationID = body.ConversationID
		var exists bool
		conversation, exists = rt.conversationsData[conversationID]
		if !exists {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		// Check if user is participant
		isParticipant := false
		for _, p := range conversation.Participants {
			if p == senderName {
				isParticipant = true
				break
			}
		}
		if !isParticipant {
			// Spec says 404 or 403. Let's use 403 Forbidden as it's more accurate for "owned by user" check failure
			// but user asked for 404/403. Let's stick to 403.
			w.WriteHeader(http.StatusForbidden)
			return
		}
	}

	// 4. Create Message
	msgID, err := uuid.NewV4()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	msg := Message{
		ID:             msgID.String(),
		ConversationID: conversationID,
		SenderID:       senderName,
		Text:           body.Text,
		CreatedAt:      time.Now(),
	}

	// 5. Update Conversation
	conversation.Messages = append(conversation.Messages, msg)
	rt.conversationsData[conversationID] = conversation
	rt.messagesMap[msgID.String()] = conversationID

	// 6. Response
	w.WriteHeader(http.StatusCreated)
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(struct {
		ConversationID string `json:"conversationId"`
		MessageID      string `json:"messageId"`
	}{
		ConversationID: conversationID,
		MessageID:      msgID.String(),
	})
}
