package api

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"https://github.com/aaitayev/wasa-homework.git"
	"https://github.com/aaitayev/wasa-homework.git"
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

	// 2. Validate Source Message
	sourceMessageID := ps.ByName("messageId")

	// Get source message from DB
	sourceMessage, err := rt.db.GetMessage(sourceMessageID)
	if err != nil {
		ctx.Logger.WithError(err).Error("error getting source message from db")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if sourceMessage == nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	// Get source conversation to check participation
	sourceConversation, err := rt.db.GetConversation(sourceMessage.ConversationID)
	if err != nil {
		ctx.Logger.WithError(err).Error("error getting source conversation from db")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if sourceConversation == nil {
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
		w.WriteHeader(http.StatusForbidden)
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
	targetConversation, err := rt.db.GetConversation(targetConversationID)
	if err != nil {
		ctx.Logger.WithError(err).Error("error getting target conversation from db")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if targetConversation == nil {
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
		w.WriteHeader(http.StatusForbidden)
		return
	}

	// 5. Create New Message
	newMessageID, err := uuid.NewV4()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	newMessage := models.Message{
		ID:             newMessageID.String(),
		ConversationID: targetConversationID,
		SenderID:       username, // The forwarder is the new sender
		Text:           sourceMessage.Text,
		CreatedAt:      time.Now(),
		ForwardedFrom:  sourceMessageID,
	}

	// 6. Save in DB
	err = rt.db.SaveMessage(&newMessage)
	if err != nil {
		ctx.Logger.WithError(err).Error("error saving forwarded message in db")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

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

