package server

import (
	"database/sql"
	"log"
	"os"

	_ "github.com/mattn/go-sqlite3"
)

func Run() error {
	settings := loadSettings()
	if err := os.MkdirAll(settings.DataDir, 0755); err != nil {
		return err
	}
	isNewDB := false
	if _, err := os.Stat(settings.DBPath); err != nil {
		if os.IsNotExist(err) {
			isNewDB = true
		} else {
			return err
		}
	}
	db, err := sql.Open("sqlite3", settings.DBPath+"?_busy_timeout=5000&_foreign_keys=on")
	if err != nil {
		return err
	}
	app := &App{db: db, settings: settings}
	if err := app.migrate(isNewDB); err != nil {
		return err
	}

	go func() {
		if err := app.startSMTP(); err != nil {
			log.Fatalf("smtp server failed: %v", err)
		}
	}()

	r := app.router()
	log.Printf("admin api listening on %s, smtp relay listening on %s", settings.HTTPAddr, settings.SMTPAddr)
	return r.Run(settings.HTTPAddr)
}
