package database

import (
	"database/sql"
	"errors"
	"git.sapienzaapps.it/fantasticcoffee/fantastic-coffee-decaffeinated/service/models"
)

func (db *appdbimpl) CreateUser(name string, token string) error {
	_, err := db.c.Exec("INSERT INTO users (name, token) VALUES (?, ?)", name, token)
	return err
}

func (db *appdbimpl) GetUserByName(name string) (*models.User, error) {
	var user models.User
	err := db.c.QueryRow("SELECT name, token FROM users WHERE name = ?", name).Scan(&user.Name, &user.Token)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	return &user, err
}

func (db *appdbimpl) GetUserByToken(token string) (string, error) {
	var name string
	err := db.c.QueryRow("SELECT name FROM users WHERE token = ?", token).Scan(&name)
	if errors.Is(err, sql.ErrNoRows) {
		return "", nil
	}
	return name, err
}

func (db *appdbimpl) UpdateUserName(oldName string, newName string) error {
	_, err := db.c.Exec("UPDATE users SET name = ? WHERE name = ?", newName, oldName)
	return err
}

func (db *appdbimpl) SearchUsers(query string) ([]string, error) {
	rows, err := db.c.Query("SELECT name FROM users WHERE name LIKE ?", "%"+query+"%")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, err
		}
		users = append(users, name)
	}
	return users, rows.Err()
}
