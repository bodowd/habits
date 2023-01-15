package data

import (
	"log"
	"testing"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestCreateHabit(t *testing.T) {
	db := setup(t)
	g := Database{DB: db}

	t.Run("creates a new habit", func(t *testing.T) {
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
	})

	t.Run("does not create a duplicate habit", func(t *testing.T) {
		_, err := g.CreateHabit("eat")
		if err == nil {
			t.Errorf("Expected duplicate error")
		}
	})

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
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"),
		&gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
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
