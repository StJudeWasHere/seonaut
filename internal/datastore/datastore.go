package datastore

import (
	"fmt"
	"log"
	"time"

	"database/sql"

	"github.com/stjudewashere/seonaut/internal/config"

	_ "github.com/go-sql-driver/mysql"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/mysql"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

type Datastore struct {
	db *sql.DB
}

const (
	paginationMax        = 25
	maxOpenConns         = 25
	maxIddleConns        = 25
	connMaxLifeInMinutes = 5
)

func NewDataStore(config config.DBConfig) (*Datastore, error) {
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
		log.Printf("Unable to reach database: %v\n", err)
		return nil, err
	}

	return &Datastore{db: db}, nil
}

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
