package database

import (
	"database/sql"
	"errors"
	"git.sapienzaapps.it/fantasticcoffee/fantastic-coffee-decaffeinated/service/models"
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
	rows, err := db.c.Query(`
		SELECT c.id, c.is_group, c.name FROM conversations c
		JOIN participants p ON c.id = p.conversation_id
		WHERE p.username = ?
	`, username)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var conversations []models.Conversation
	for rows.Next() {
		var c models.Conversation
		if err := rows.Scan(&c.ID, &c.IsGroup, &c.Name); err != nil {
			return nil, err
		}

		// participants will be loaded separately if needed by the handler, or we could load them here.
		// For the sake of listing, maybe we don't need all participants for every conversation.
		// But let's load them to be consistent with models.Conversation.
		pRows, err := db.c.Query("SELECT username FROM participants WHERE conversation_id = ?", c.ID)
		if err != nil {
			return nil, err
		}
		for pRows.Next() {
			var p string
			if err := pRows.Scan(&p); err != nil {
				pRows.Close()
				return nil, err
			}
			c.Participants = append(c.Participants, p)
		}
		pRows.Close()

		conversations = append(conversations, c)
	}
	return conversations, rows.Err()
}

func (db *appdbimpl) AddParticipant(conversationID string, username string) error {
	_, err := db.c.Exec("INSERT OR IGNORE INTO participants (conversation_id, username) VALUES (?, ?)", conversationID, username)
	return err
}

func (db *appdbimpl) RemoveParticipant(conversationID string, username string) error {
	_, err := db.c.Exec("DELETE FROM participants WHERE conversation_id = ? AND username = ?", conversationID, username)
	return err
}
