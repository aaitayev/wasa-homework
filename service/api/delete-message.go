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

	// 3. Get message from DB
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

	// 5. Check Participation
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

	// 6. Delete (Mark as Deleted)
	// Spec often implies only sender can delete, but let's stick to participant check + sender check if needed.
	// Most implementations allow sender to delete their own message.
	if msg.SenderID != username {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	err = rt.db.DeleteMessage(messageID)
	if err != nil {
		ctx.Logger.WithError(err).Error("error deleting message in db")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// 7. Response
	w.WriteHeader(http.StatusNoContent)
}

