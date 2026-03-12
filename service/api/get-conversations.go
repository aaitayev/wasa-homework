package api

import (
	"encoding/json"
	"net/http"
	"sort"
	"strings"
	"time"

	"https://github.com/aaitayev/wasa-homework.git"
	"github.com/julienschmidt/httprouter"
)


type ConversationSummary struct {
	ID              string    `json:"id"`
	IsGroup         bool      `json:"isGroup"`
	Name            string    `json:"name"`
	Participants    []string  `json:"participants"`
	LastMessageAt   time.Time `json:"lastMessageAt"`
	LastMessageText string    `json:"lastMessageText"`
}

// getMyConversations handles the GET /conversations endpoint.
func (rt *_router) getMyConversations(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
	// Extract the token from the Authorization header
	authHeader := r.Header.Get("Authorization")
	if !strings.HasPrefix(authHeader, "Bearer ") {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	token := strings.TrimPrefix(authHeader, "Bearer ")

	// Validate the token
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

	// Get the conversations for the user from DB
	dbConvs, err := rt.db.GetUserConversations(username)
	if err != nil {
		ctx.Logger.WithError(err).Error("error getting user conversations from db")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	summaries := make([]ConversationSummary, 0, len(dbConvs))
	for _, conv := range dbConvs {
		// Load messages to get the last one
		messages, err := rt.db.GetMessages(conv.ID)
		if err != nil {
			ctx.Logger.WithError(err).Error("error getting messages for summary")
			continue // Skip or handle error
		}

		var lastMsgAt time.Time
		var lastMsgText string
		if len(messages) > 0 {
			lastMsg := messages[len(messages)-1]
			lastMsgAt = lastMsg.CreatedAt
			lastMsgText = lastMsg.Text
		}

		summaries = append(summaries, ConversationSummary{
			ID:              conv.ID,
			IsGroup:         conv.IsGroup,
			Name:            conv.Name,
			Participants:    conv.Participants,
			LastMessageAt:   lastMsgAt,
			LastMessageText: lastMsgText,
		})
	}

	// Sort reverse-chronological (newest first)
	sort.Slice(summaries, func(i, j int) bool {
		return summaries[i].LastMessageAt.After(summaries[j].LastMessageAt)
	})

	// Return the conversations
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(summaries)
}

