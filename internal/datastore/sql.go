package datastore

import (
	"fmt"
	"time"

	"database/sql"

	_ "github.com/go-sql-driver/mysql"
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
