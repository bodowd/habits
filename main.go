package main

import (
	"log"
	"os"

	"github.com/bodowd/habits/data"
	"github.com/bodowd/habits/pages"
	tea "github.com/charmbracelet/bubbletea"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func openSQLite() *gorm.DB {
	var dbName string
	if os.Getenv("DEMO") == "true" {
		dbName = "demo.db"
	} else {
		dbName = "habits.db"
	}

	db, err := gorm.Open(sqlite.Open(dbName),
		&gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		log.Fatalf("unable to open database: %v", err)
	}

	err = db.AutoMigrate(&data.Habit{})
	if err != nil {
		log.Fatal(err)
	}
	return db
}

func main() {

	db := openSQLite()

	p := tea.NewProgram(pages.NewList(db))

	if _, err := p.Run(); err != nil {
		log.Fatal(err)
		os.Exit(1)
	}
}
