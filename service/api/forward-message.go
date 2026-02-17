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

// forwardMessage handles POST /messages/{messageId}/forward
func (rt *_router) forwardMessage(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
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

	// 2. Validate Source Message
	sourceMessageID := ps.ByName("messageId")

	// Find source conversation
	sourceConversationID, exists := rt.messagesMap[sourceMessageID]
	if !exists {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	sourceConversation, exists := rt.conversationsData[sourceConversationID]
	if !exists {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	// Check if user is participant in source conversation
	isSourceParticipant := false
	for _, p := range sourceConversation.Participants {
		if p == username {
			isSourceParticipant = true
			break
		}
	}
	if !isSourceParticipant {
		w.WriteHeader(http.StatusNotFound) 
		return
	}

	// Find source message details
	var sourceMessage Message
	found := false
	for _, msg := range sourceConversation.Messages {
		if msg.ID == sourceMessageID {
			sourceMessage = msg
			found = true
			break
		}
	}
	if !found {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	// Check if source message is deleted
	if sourceMessage.Deleted {
		w.WriteHeader(http.StatusConflict) // 409 Conflict
		return
	}

	// 3. Parse Body for Target Conversation
	var body struct {
		ConversationID string `json:"conversationId"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if body.ConversationID == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// 4. Validate Target Conversation
	targetConversationID := body.ConversationID
	targetConversation, exists := rt.conversationsData[targetConversationID]
	if !exists {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	// Check if user is participant in target conversation
	isTargetParticipant := false
	for _, p := range targetConversation.Participants {
		if p == username {
			isTargetParticipant = true
			break
		}
	}
	if !isTargetParticipant {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	// 5. Create New Message
	newMessageID, err := uuid.NewV4()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	newMessage := Message{
		ID:             newMessageID.String(),
		ConversationID: targetConversationID,
		SenderID:       username, // The forwarder is the new sender
		Text:           sourceMessage.Text,
		CreatedAt:      time.Now(),
		ForwardedFrom:  sourceMessageID,
	}

	// 6. Update Target Conversation and Maps
	targetConversation.Messages = append(targetConversation.Messages, newMessage)
	rt.conversationsData[targetConversationID] = targetConversation
	rt.messagesMap[newMessageID.String()] = targetConversationID

	// 7. Response
	w.WriteHeader(http.StatusCreated)
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(struct {
		ConversationID string `json:"conversationId"`
		MessageID      string `json:"messageId"`
	}{
		ConversationID: targetConversationID,
		MessageID:      newMessageID.String(),
	})
}
