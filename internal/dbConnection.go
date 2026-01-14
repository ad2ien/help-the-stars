package internal

import (
	"database/sql"
	"embed"
	"fmt"
	"net/http"

	"github.com/charmbracelet/log"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/sqlite"
	"github.com/golang-migrate/migrate/v4/source/httpfs"
	_ "github.com/mattn/go-sqlite3"
	_ "modernc.org/sqlite"
)

const dbFileName = "db/help-the-stars.db"

type DbConnection struct {
	Connection *sql.DB
}

func NewConnection(migrationsFs embed.FS) DbConnection {

	configDbFile := GetSettings().DBFile
	if configDbFile == "" {
		configDbFile = dbFileName
	}
	log.Info("DB using file : " + configDbFile)

	conn, err := sql.Open("sqlite3", configDbFile)
	if err != nil {
		log.Fatal(err)
	}

	err = ensureSchema(migrationsFs, conn)
	if err != nil {
		log.Fatal(err)
	}

	return DbConnection{conn}
}

func (dbConn *DbConnection) Close() {
	err := dbConn.Connection.Close()
	if err != nil {
		log.Error("error closing connection", "error", err)
	}
}

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

	// Get the latest version from the migrations directory
	latestVersion, err := getLatestMigrationVersion(migrations)
	if err != nil {
		return fmt.Errorf("failed to get latest migration version, %w", err)
	}

	err = m.Migrate(latestVersion)
	if err != nil && err != migrate.ErrNoChange {
		return err
	}
	return sourceInstance.Close()
}

func getLatestMigrationVersion(migrations embed.FS) (uint, error) {
	entries, err := migrations.ReadDir("migrations")
	if err != nil {
		return 0, fmt.Errorf("failed to read migrations directory, %w", err)
	}

	var maxVersion uint = 0
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		var version uint
		_, err := fmt.Sscanf(entry.Name(), "%06d_", &version)
		if err != nil {
			continue
		}
		if version > maxVersion {
			maxVersion = version
		}
	}

	if maxVersion == 0 {
		return 0, fmt.Errorf("no valid migrations found")
	}

	return maxVersion, nil
}
