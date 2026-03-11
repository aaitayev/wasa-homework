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

	// 2. Get Message ID
	messageID := ps.ByName("messageId")

	// 3. Get Message from DB
	msg, err := rt.db.GetMessage(messageID)
	if err != nil {
		ctx.Logger.WithError(err).Error("error getting message from db")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if msg == nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	// 4. Get Conversation to check participation
	conversation, err := rt.db.GetConversation(msg.ConversationID)
	if err != nil {
		ctx.Logger.WithError(err).Error("error getting conversation from db")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if conversation == nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	// 5. Check Participant
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

	if msg.Deleted {
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

	// 7. Update Message Comment in DB
	err = rt.db.UpdateMessageComment(messageID, body.Comment, time.Now())
	if err != nil {
		ctx.Logger.WithError(err).Error("error updating message comment in db")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

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

	// 2. Get Message ID
	messageID := ps.ByName("messageId")

	// 3. Get Message from DB
	msg, err := rt.db.GetMessage(messageID)
	if err != nil {
		ctx.Logger.WithError(err).Error("error getting message from db")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if msg == nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	// 4. Get Conversation to check participation
	conversation, err := rt.db.GetConversation(msg.ConversationID)
	if err != nil {
		ctx.Logger.WithError(err).Error("error getting conversation from db")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if conversation == nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	// 5. Check Participant
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

	if msg.Deleted {
		w.WriteHeader(http.StatusConflict)
		return
	}

	// 6. Remove Comment in DB
	err = rt.db.UpdateMessageComment(messageID, "", time.Time{})
	if err != nil {
		ctx.Logger.WithError(err).Error("error removing message comment in db")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

