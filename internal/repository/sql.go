package repository

import (
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/stjudewashere/seonaut/internal/config"
)

const (
	// paginationMax is the maximum number of items allowed in paginated lists
	paginationMax = 25

	// maxOpenConns is the maximum number of open connections to the database.
	// Use 0 for unlimited connections.
	maxOpenConns = 25

	// maxIddleConns is the maximum number of connections in the idle connection pool.
	// Use 0 for no idle connections retained.
	maxIddleConns = 25

	// connMaxLifeInMinutes is the maximum amount of time a connection may be reused.
	// Use 0 to not close connections due to it's age.
	connMaxLifeInMinutes = 5
)

// SqlConnect creates a new SQL connection with the provided configuration.
func SqlConnect(config *config.DBConfig) (*sql.DB, error) {
	db, err := sql.Open("mysql", fmt.Sprintf(
		"%s:%s@tcp(%s:%d)/%s?parseTime=true&multiStatements=true",
		config.User,
		config.Pass,
		config.Server,
		config.Port,
		config.Name,
	))

	if err != nil {
		return nil, err
	}

	// Set maximum number of open connections to the database.
	db.SetMaxOpenConns(maxOpenConns)

	// Set maximum number of idle connections to the database.
	db.SetMaxIdleConns(maxIddleConns)

	// Set maximum lifetime for each connection to the database.
	db.SetConnMaxLifetime(connMaxLifeInMinutes * time.Minute)

	// Ping the database to check if the connection is successful.
	if err := db.Ping(); err != nil {
		return nil, err
	}

	return db, nil
}

// Hash returns a hashed string.
func Hash(s string) string {
	hash := sha256.Sum256([]byte(s))

	return hex.EncodeToString(hash[:])
}

// Truncate a string to the requiered length.
func Truncate(s string, length int) string {
	text := []rune(s)
	if len(text) > length {
		s = string(text[:length-3]) + "..."
	}

	return s
}
