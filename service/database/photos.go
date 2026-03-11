package database

import (
	"database/sql"
	"errors"
)

func (db *appdbimpl) SetUserPhoto(username string, photo []byte) error {
	_, err := db.c.Exec("INSERT INTO user_photos (username, photo) VALUES (?, ?) ON CONFLICT(username) DO UPDATE SET photo=excluded.photo", username, photo)
	return err
}

func (db *appdbimpl) GetUserPhoto(username string) ([]byte, error) {
	var photo []byte
	err := db.c.QueryRow("SELECT photo FROM user_photos WHERE username = ?", username).Scan(&photo)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	return photo, err
}

func (db *appdbimpl) SetGroupPhoto(groupID string, photo []byte) error {
	_, err := db.c.Exec("INSERT INTO group_photos (group_id, photo) VALUES (?, ?) ON CONFLICT(group_id) DO UPDATE SET photo=excluded.photo", groupID, photo)
	return err
}

func (db *appdbimpl) GetGroupPhoto(groupID string) ([]byte, error) {
	var photo []byte
	err := db.c.QueryRow("SELECT photo FROM group_photos WHERE group_id = ?", groupID).Scan(&photo)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	return photo, err
}
