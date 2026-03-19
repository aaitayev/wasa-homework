package database

import (
	"database/sql"
	"errors"
	"github.com/aaitayev/wasa-homework.git"
)

func (db *appdbimpl) CreateConversation(conv *models.Conversation) error {
	tx, err := db.c.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	_, err = tx.Exec("INSERT INTO conversations (id, is_group, name) VALUES (?, ?, ?)", conv.ID, conv.IsGroup, conv.Name)
	if err != nil {
		return err
	}

	for _, p := range conv.Participants {
		_, err = tx.Exec("INSERT INTO participants (conversation_id, username) VALUES (?, ?)", conv.ID, p)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (db *appdbimpl) GetConversation(id string) (*models.Conversation, error) {
	var conv models.Conversation
	err := db.c.QueryRow("SELECT id, is_group, name FROM conversations WHERE id = ?", id).Scan(&conv.ID, &conv.IsGroup, &conv.Name)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	// Get participants
	rows, err := db.c.Query("SELECT username FROM participants WHERE conversation_id = ?", id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var p string
		if err := rows.Scan(&p); err != nil {
			return nil, err
		}
		conv.Participants = append(conv.Participants, p)
	}

	return &conv, rows.Err()
}

func (db *appdbimpl) UpdateConversationName(id string, name string) error {
	_, err := db.c.Exec("UPDATE conversations SET name = ? WHERE id = ?", name, id)
	return err
}

func (db *appdbimpl) GetUserConversations(username string) ([]models.Conversation, error) {
	// 1. Get IDs of conversations the user is in
	rows, err := db.c.Query(`
		SELECT c.id, c.is_group, c.name 
		FROM conversations c
		JOIN participants p ON c.id = p.conversation_id
		WHERE p.username = ?
	`, username)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var conversations []models.Conversation
	convMap := make(map[string]int) // Map ID to index in slice

	for rows.Next() {
		var c models.Conversation
		if err := rows.Scan(&c.ID, &c.IsGroup, &c.Name); err != nil {
			return nil, err
		}
		c.Participants = []string{}
		convMap[c.ID] = len(conversations)
		conversations = append(conversations, c)
	}

	if len(conversations) == 0 {
		return conversations, nil
	}

	// 2. Fetch ALL participants for these conversations in one go
	pRows, err := db.c.Query(`
		SELECT p.conversation_id, p.username 
		FROM participants p
		WHERE p.conversation_id IN (
			SELECT conversation_id FROM participants WHERE username = ?
		)
	`, username)
	if err != nil {
		return nil, err
	}
	defer pRows.Close()

	for pRows.Next() {
		var cid, p string
		if err := pRows.Scan(&cid, &p); err != nil {
			return nil, err
		}
		if idx, ok := convMap[cid]; ok {
			conversations[idx].Participants = append(conversations[idx].Participants, p)
		}
	}

	return conversations, pRows.Err()
}

func (db *appdbimpl) AddParticipant(conversationID string, username string) error {
	_, err := db.c.Exec("INSERT OR IGNORE INTO participants (conversation_id, username) VALUES (?, ?)", conversationID, username)
	return err
}

func (db *appdbimpl) RemoveParticipant(conversationID string, username string) error {
	_, err := db.c.Exec("DELETE FROM participants WHERE conversation_id = ? AND username = ?", conversationID, username)
	return err
}
