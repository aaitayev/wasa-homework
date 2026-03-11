/*
Package database is the middleware between the app database and the code. All data (de)serialization (save/load) from a
persistent database are handled here. Database specific logic should never escape this package.

To use this package you need to apply migrations to the database if needed/wanted, connect to it (using the database
data source name from config), and then initialize an instance of AppDatabase from the DB connection.

For example, this code adds a parameter in `webapi` executable for the database data source name (add it to the
main.WebAPIConfiguration structure):

	DB struct {
		Filename string `conf:""`
	}

This is an example on how to migrate the DB and connect to it:

	// Start Database
	logger.Println("initializing database support")
	db, err := sql.Open("sqlite3", "./foo.db")
	if err != nil {
		logger.WithError(err).Error("error opening SQLite DB")
		return fmt.Errorf("opening SQLite: %w", err)
	}
	defer func() {
		logger.Debug("database stopping")
		_ = db.Close()
	}()

Then you can initialize the AppDatabase and pass it to the api package.
*/
package database

import (
	"database/sql"
	"errors"
	"fmt"
	"git.sapienzaapps.it/fantasticcoffee/fantastic-coffee-decaffeinated/service/models"
	"time"
)

// AppDatabase is the high level interface for the DB
type AppDatabase interface {
	// User operations
	CreateUser(name string, token string) error
	GetUserByName(name string) (*models.User, error)
	GetUserByToken(token string) (string, error)
	UpdateUserName(oldName string, newName string) error
	SearchUsers(query string) ([]string, error)

	// Conversation operations
	CreateConversation(conv *models.Conversation) error
	GetConversation(id string) (*models.Conversation, error)
	UpdateConversationName(id string, name string) error
	GetUserConversations(username string) ([]models.Conversation, error)

	// Message operations
	SaveMessage(msg *models.Message) error
	GetMessage(id string) (*models.Message, error)
	RemoveParticipant(conversationID string, username string) error
	DeleteMessage(id string) error
	GetMessages(conversationID string) ([]models.Message, error)
	UpdateMessageComment(id string, comment string, commentedAt time.Time) error

	// Participant operations
	AddParticipant(conversationID string, username string) error

	// Photo operations
	SetUserPhoto(username string, photo []byte, contentType string) error
	GetUserPhoto(username string) ([]byte, string, error)
	SetGroupPhoto(groupID string, photo []byte, contentType string) error
	GetGroupPhoto(groupID string) ([]byte, string, error)

	Ping() error
}

type appdbimpl struct {
	c *sql.DB
}

// New returns a new instance of AppDatabase based on the SQLite connection `db`.
// `db` is required - an error will be returned if `db` is `nil`.
func New(db *sql.DB) (AppDatabase, error) {
	if db == nil {
		return nil, errors.New("database is required when building a AppDatabase")
	}

	// Create tables if they don't exist
	tables := []string{
		`CREATE TABLE IF NOT EXISTS users (
			name TEXT PRIMARY KEY,
			token TEXT NOT NULL
		);`,
		`CREATE TABLE IF NOT EXISTS conversations (
			id TEXT PRIMARY KEY,
			is_group BOOLEAN NOT NULL DEFAULT 0,
			name TEXT
		);`,
		`CREATE TABLE IF NOT EXISTS participants (
			conversation_id TEXT NOT NULL,
			username TEXT NOT NULL,
			PRIMARY KEY (conversation_id, username),
			FOREIGN KEY (conversation_id) REFERENCES conversations(id) ON DELETE CASCADE,
			FOREIGN KEY (username) REFERENCES users(name) ON DELETE CASCADE ON UPDATE CASCADE
		);`,
		`CREATE TABLE IF NOT EXISTS messages (
			id TEXT PRIMARY KEY,
			conversation_id TEXT NOT NULL,
			sender TEXT NOT NULL,
			text TEXT NOT NULL,
			created_at DATETIME NOT NULL,
			deleted BOOLEAN NOT NULL DEFAULT 0,
			comment TEXT,
			commented_at DATETIME,
			forwarded_from TEXT,
			FOREIGN KEY (conversation_id) REFERENCES conversations(id) ON DELETE CASCADE
		);`,
		`CREATE TABLE IF NOT EXISTS user_photos (
			username TEXT PRIMARY KEY,
			photo BLOB,
			content_type TEXT NOT NULL DEFAULT 'image/jpeg',
			FOREIGN KEY (username) REFERENCES users(name) ON DELETE CASCADE ON UPDATE CASCADE
		);`,

		`CREATE TABLE IF NOT EXISTS group_photos (
			group_id TEXT PRIMARY KEY,
			photo BLOB,
			content_type TEXT NOT NULL DEFAULT 'image/jpeg',
			FOREIGN KEY (group_id) REFERENCES conversations(id) ON DELETE CASCADE
		);`,
	}

	for _, stmt := range tables {
		_, err := db.Exec(stmt)
		if err != nil {
			return nil, fmt.Errorf("error creating database structure: %w\nStatement: %s", err, stmt)
		}
	}

	// Add content_type column if it doesn't exist (migration)
	_, _ = db.Exec("ALTER TABLE user_photos ADD COLUMN content_type TEXT NOT NULL DEFAULT 'image/jpeg';")
	_, _ = db.Exec("ALTER TABLE group_photos ADD COLUMN content_type TEXT NOT NULL DEFAULT 'image/jpeg';")

	// Enable foreign keys
	_, err := db.Exec("PRAGMA foreign_keys = ON;")
	if err != nil {
		return nil, fmt.Errorf("error enabling foreign keys: %w", err)
	}

	return &appdbimpl{
		c: db,
	}, nil
}

func (db *appdbimpl) Ping() error {
	return db.c.Ping()
}

