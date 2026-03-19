package database

import (
	"database/sql"
	"errors"
	"github.com/aaitayev/wasa-homework.git"
	"time"
)

func (db *appdbimpl) SaveMessage(msg *models.Message) error {
	_, err := db.c.Exec(`
		INSERT INTO messages (id, conversation_id, sender, text, created_at, deleted, comment, commented_at, forwarded_from)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, msg.ID, msg.ConversationID, msg.SenderID, msg.Text, msg.CreatedAt.Format(time.RFC3339), msg.Deleted, msg.Comment, msg.CommentedAt.Format(time.RFC3339), msg.ForwardedFrom)
	return err
}

func (db *appdbimpl) GetMessage(id string) (*models.Message, error) {
	var msg models.Message
	var commentedAt sql.NullString
	var comment sql.NullString
	var forwardedFrom sql.NullString
	var createdAtStr string

	err := db.c.QueryRow(`
		SELECT id, conversation_id, sender, text, created_at, deleted, comment, commented_at, forwarded_from
		FROM messages WHERE id = ?
	`, id).Scan(&msg.ID, &msg.ConversationID, &msg.SenderID, &msg.Text, &createdAtStr, &msg.Deleted, &comment, &commentedAt, &forwardedFrom)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	msg.CreatedAt, _ = time.Parse(time.RFC3339, createdAtStr)
	if comment.Valid {
		msg.Comment = comment.String
	}
	if commentedAt.Valid {
		msg.CommentedAt, _ = time.Parse(time.RFC3339, commentedAt.String)
	}
	if forwardedFrom.Valid {
		msg.ForwardedFrom = forwardedFrom.String
	}

	return &msg, nil
}

func (db *appdbimpl) DeleteMessage(id string) error {
	_, err := db.c.Exec("UPDATE messages SET deleted = 1 WHERE id = ?", id)
	return err
}

func (db *appdbimpl) GetMessages(conversationID string) ([]models.Message, error) {
	rows, err := db.c.Query(`
		SELECT id, conversation_id, sender, text, created_at, deleted, comment, commented_at, forwarded_from
		FROM messages WHERE conversation_id = ? ORDER BY created_at ASC
	`, conversationID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messages []models.Message
	for rows.Next() {
		var msg models.Message
		var commentedAt sql.NullString
		var comment sql.NullString
		var forwardedFrom sql.NullString
		var createdAtStr string

		err := rows.Scan(&msg.ID, &msg.ConversationID, &msg.SenderID, &msg.Text, &createdAtStr, &msg.Deleted, &comment, &commentedAt, &forwardedFrom)
		if err != nil {
			return nil, err
		}

		msg.CreatedAt, _ = time.Parse(time.RFC3339, createdAtStr)
		if comment.Valid {
			msg.Comment = comment.String
		}
		if commentedAt.Valid {
			msg.CommentedAt, _ = time.Parse(time.RFC3339, commentedAt.String)
		}
		if forwardedFrom.Valid {
			msg.ForwardedFrom = forwardedFrom.String
		}

		messages = append(messages, msg)
	}
	return messages, rows.Err()
}

func (db *appdbimpl) UpdateMessageComment(id string, comment string, commentedAt time.Time) error {
	_, err := db.c.Exec("UPDATE messages SET comment = ?, commented_at = ? WHERE id = ?", comment, commentedAt.Format(time.RFC3339), id)
	return err
}
