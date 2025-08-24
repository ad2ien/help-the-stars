package internal

import (
	"database/sql"
	"embed"
	_ "embed"
	"fmt"
	"log"
	"net/http"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/sqlite"
	"github.com/golang-migrate/migrate/v4/source/httpfs"
	_ "github.com/mattn/go-sqlite3"
	_ "modernc.org/sqlite"
)

const dbFileName = "help-stars.db"

type DbConnection struct {
	connection *sql.DB
}

func NewConnection(migrationsFs embed.FS) DbConnection {

	conn, err := sql.Open("sqlite3", dbFileName)
	if err != nil {
		log.Fatal(err)
	}

	ensureSchema(migrationsFs, conn)

	return DbConnection{conn}
}

func (dbConn *DbConnection) Close() {
	dbConn.connection.Close()
}

const schemaVersion = 1

func ensureSchema(migrations embed.FS, db *sql.DB) error {
	sourceInstance, err := httpfs.New(http.FS(migrations), "migrations")
	if err != nil {
		return fmt.Errorf("invalid source instance, %w", err)
	}
	targetInstance, err := sqlite.WithInstance(db, new(sqlite.Config))
	if err != nil {
		return fmt.Errorf("invalid target sqlite instance, %w", err)
	}
	m, err := migrate.NewWithInstance(
		"httpfs", sourceInstance, "sqlite", targetInstance)
	if err != nil {
		return fmt.Errorf("failed to initialize migrate instance, %w", err)
	}
	err = m.Migrate(schemaVersion)
	if err != nil && err != migrate.ErrNoChange {
		return err
	}
	return sourceInstance.Close()
}
