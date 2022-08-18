package datastore

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	"database/sql"

	_ "github.com/go-sql-driver/mysql"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/mysql"
	_ "github.com/golang-migrate/migrate/v4/source/file"
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

// DBConfig stores the configuration for the database store.
// It is loaded from the config package.
type DBConfig struct {
	Server string `mapstructure:"server"`
	Port   int    `mapstructure:"port"`
	User   string `mapstructure:"user"`
	Pass   string `mapstructure:"password"`
	Name   string `mapstructure:"database"`
}

type Datastore struct {
	db *sql.DB
}

// NewDatastore is called when a new application instance is created.
func NewDataStore(config *DBConfig) (*Datastore, error) {
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

	db.SetMaxOpenConns(maxOpenConns)
	db.SetMaxIdleConns(maxIddleConns)
	db.SetConnMaxLifetime(connMaxLifeInMinutes * time.Minute)

	if err := db.Ping(); err != nil {
		return nil, err
	}

	return &Datastore{db: db}, nil
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

// hash returns a hashed string
func Hash(s string) string {
	hash := sha256.Sum256([]byte(s))

	return hex.EncodeToString(hash[:])
}
