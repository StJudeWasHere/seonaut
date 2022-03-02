package datastore

import (
	"fmt"
	"log"
	"time"

	"database/sql"

	"github.com/mnlg/seonaut/internal/config"

	_ "github.com/go-sql-driver/mysql"
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
		"%s:%s@tcp(%s:%d)/%s?parseTime=true",
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
