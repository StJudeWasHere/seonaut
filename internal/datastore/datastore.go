package datastore

import (
	"crypto/sha256"
	"encoding/hex"

	"database/sql"

	_ "github.com/go-sql-driver/mysql"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/mysql"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

type Datastore struct {
	db *sql.DB
}

// NewDatastore is called when a new application instance is created.
func NewDataStore(db *sql.DB) *Datastore {
	return &Datastore{db: db}
}

// Migrate is called when the app is launched to apply the database migrations.
func (ds *Datastore) Migrate() error {
	driver, err := mysql.WithInstance(ds.db, &mysql.Config{})
	if err != nil {
		return err
	}

	m, err := migrate.NewWithDatabaseInstance(
		"file://migrations",
		"mysql",
		driver,
	)
	if err != nil {
		return err
	}

	m.Up()

	return nil
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
