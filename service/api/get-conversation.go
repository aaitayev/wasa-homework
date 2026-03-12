package api

import (
	"encoding/json"
	"net/http"
	"strings"

	"https://github.com/aaitayev/wasa-homework.git"
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

	// 2. Get Conversation ID
	conversationID := ps.ByName("conversationId")

	// 3. Get Conversation from DB
	conversation, err := rt.db.GetConversation(conversationID)
	if err != nil {
		ctx.Logger.WithError(err).Error("error getting conversation from db")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if conversation == nil {
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

	// 5. Load messages
	messages, err := rt.db.GetMessages(conversationID)
	if err != nil {
		ctx.Logger.WithError(err).Error("error getting messages from db")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	conversation.Messages = messages

	// 6. Response
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(conversation)
}

