package database

import (
	"database/sql"
	"errors"
)

func (db *appdbimpl) SetUserPhoto(username string, photo []byte, contentType string) error {
	_, err := db.c.Exec("INSERT INTO user_photos (username, photo, content_type) VALUES (?, ?, ?) ON CONFLICT(username) DO UPDATE SET photo=excluded.photo, content_type=excluded.content_type", username, photo, contentType)
	return err
}

func (db *appdbimpl) GetUserPhoto(username string) ([]byte, string, error) {
	var photo []byte
	var contentType string
	err := db.c.QueryRow("SELECT photo, content_type FROM user_photos WHERE username = ?", username).Scan(&photo, &contentType)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, "", nil
	}
	return photo, contentType, err
}

func (db *appdbimpl) SetGroupPhoto(groupID string, photo []byte, contentType string) error {
	_, err := db.c.Exec("INSERT INTO group_photos (group_id, photo, content_type) VALUES (?, ?, ?) ON CONFLICT(group_id) DO UPDATE SET photo=excluded.photo, content_type=excluded.content_type", groupID, photo, contentType)
	return err
}

func (db *appdbimpl) GetGroupPhoto(groupID string) ([]byte, string, error) {
	var photo []byte
	var contentType string
	err := db.c.QueryRow("SELECT photo, content_type FROM group_photos WHERE group_id = ?", groupID).Scan(&photo, &contentType)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, "", nil
	}
	return photo, contentType, err
}
