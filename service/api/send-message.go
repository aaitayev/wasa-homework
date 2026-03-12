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

// sendMessage handles POST /messages
func (rt *_router) sendMessage(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
	// 1. Auth check
	authHeader := r.Header.Get("Authorization")
	if !strings.HasPrefix(authHeader, "Bearer ") {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	token := strings.TrimPrefix(authHeader, "Bearer ")
	senderName, err := rt.db.GetUserByToken(token)
	if err != nil {
		ctx.Logger.WithError(err).Error("error getting user by token")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if senderName == "" {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	var body struct {
		ConversationID string   `json:"conversationId"`
		Text           string   `json:"text"`
		IsGroup        bool     `json:"isGroup"`
		Recipient      string   `json:"recipient"`
		Name           string   `json:"name"`
		Participants   []string `json:"participants"`
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
	var conversation *models.Conversation

	// 3. Handle Conversation Logic
	if body.ConversationID == "" {
		// Create new conversation
		uuidConf, err := uuid.NewV4()
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		conversationID = uuidConf.String()
		
		participants := []string{senderName}
		// Add unique participants from the list
		seen := map[string]bool{senderName: true}
		
		addParticipant := func(p string) {
			trimmed := strings.TrimSpace(p)
			if trimmed != "" && !seen[trimmed] {
				participants = append(participants, trimmed)
				seen[trimmed] = true
			}
		}

		for _, p := range body.Participants {
			addParticipant(p)
		}
		if body.Recipient != "" {
			addParticipant(body.Recipient)
		}
		
		conversation = &models.Conversation{
			ID:           conversationID,
			Participants: participants,
			Messages:     []models.Message{},
			IsGroup:      body.IsGroup,
			Name:         body.Name,
		}
		err = rt.db.CreateConversation(conversation)
		if err != nil {
			ctx.Logger.WithError(err).Error("error creating conversation in db")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

	} else {
		// Existing conversation
		conversationID = body.ConversationID
		conversation, err = rt.db.GetConversation(conversationID)
		if err != nil {
			ctx.Logger.WithError(err).Error("error getting conversation from db")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		if conversation == nil {
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
	msg := models.Message{
		ID:             msgID.String(),
		ConversationID: conversationID,
		SenderID:       senderName,
		Text:           body.Text,
		CreatedAt:      time.Now(),
	}

	// 5. Update Conversation
	err = rt.db.SaveMessage(&msg)
	if err != nil {
		ctx.Logger.WithError(err).Error("error saving message in db")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

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

