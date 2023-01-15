package data

import (
	"log"
	"testing"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestCreateHabit(t *testing.T) {
	db := setup(t)
	g := Database{DB: db}

	wantName := "eat"

	hab, _ := g.CreateHabit(wantName)

	var habit Habit
	db.First(&habit, hab.ID)

	gotName := habit.Name
	gotCreatedAt := habit.CreatedAt
	gotActive := habit.Active

	if gotName != wantName {
		t.Errorf("got %v want %v", gotName, wantName)
	}

	_, err := time.Parse("2006-01-02", gotCreatedAt)
	if err != nil {
		t.Errorf("got %v, error: %v", gotCreatedAt, err)
	}

	if gotActive != true {
		t.Errorf("got %v want %v", gotActive, true)
	}

}

func TestGetActiveHabits(t *testing.T) {
	db := setup(t)
	g := Database{DB: db}

	habits := g.GetActiveHabits()

	if len(habits) != 2 {
		t.Errorf("got slice of length %d, want 2 elements", len(habits))
	}

}

func TestGetAllHabits(t *testing.T) {
	db := setup(t)
	g := Database{DB: db}

	habits := g.GetAllHabits()
	if len(habits) != 3 {
		t.Errorf("got slice of length %d, want 3 elements", len(habits))
	}
}

func currentDate() string {
	return time.Now().Format("2006-01-02")
}

func setup(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		log.Fatalf("unable to open in-memory SQLite DB: %v", err)
	}

	db.AutoMigrate(&Habit{})

	seedHabits(db)
	t.Cleanup(func() {
		db.Migrator().DropTable(&Habit{})
	})
	return db
}

func seedHabits(db *gorm.DB) {

	habits := []Habit{
		{Name: "cook", CreatedAt: currentDate(), Active: true},
		{Name: "read", CreatedAt: currentDate(), Active: true},
		{Name: "clean", CreatedAt: currentDate(), Active: false},
	}
	for _, i := range habits {
		db.Create(&i)
	}

}
