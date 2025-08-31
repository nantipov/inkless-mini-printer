package internal

import (
	"database/sql"
	"log"
	"os"
	"path"
	"sync"

	_ "github.com/mattn/go-sqlite3"
)

const (
	dbFilename = "printing_service.db"
)

var (
	db      *sql.DB
	dbMutex sync.Mutex
)

func InitDatabase() {
	settings := GetSettings()

	err := os.MkdirAll(settings.DataDir, os.ModeDir)
	if err != nil {
		log.Fatalf("could not create directories for a database: %s: %s", settings.DataDir, err.Error())
	}

	dbFile := path.Join(settings.DataDir, dbFilename)
	log.Printf("connect database from file %s", dbFile)
	db, err = sql.Open("sqlite3", dbFile)
	if err != nil {
		log.Fatalf("could not open database: %s", err.Error())
	}

	applySchema(path.Join(settings.AppDir, settings.DBMigrationsDir))
}

func applySchema(schemaFile string) {
	dbMutex.Lock()
	defer dbMutex.Unlock()

	//todo: support db migrations (existing library?)
	log.Printf("apply database schema from file %s", schemaFile)
	queryBytes, err := os.ReadFile(schemaFile)
	if err != nil {
		log.Fatalf("could not read schema for a database: %s: %s", schemaFile, err.Error())
	}

	_, err = db.Exec(string(queryBytes))
	if err != nil {
		log.Fatalf("could not apply schema to a database: %s: %s", schemaFile, err.Error())
	}
}

func GetDatabaseConnection() *sql.DB {
	return db
}
