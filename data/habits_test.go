package data

import (
	"log"
	"strings"
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

		assertDate(t, gotCreatedAt)

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

	want := 4
	if len(habits) != want {
		t.Errorf("got slice of length %d, want %d elements", len(habits), want)
	}

}

func TestGetInactiveHabits(t *testing.T) {
	db := setup(t)
	g := Database{DB: db}

	habits := g.GetInactiveHabits()
	want := 1
	if len(habits) != want {
		t.Errorf("got slice of length %d, want %d", len(habits), want)
	}
}

func TestGetAllHabits(t *testing.T) {
	db := setup(t)
	g := Database{DB: db}

	habits := g.GetAllHabits()
	want := 5
	if len(habits) != want {
		t.Errorf("got slice of length %d, want %d elements", len(habits), want)
	}
}

func TestRecordCompletion(t *testing.T) {
	db := setup(t)
	g := Database{DB: db}

	t.Run("adds to streak if a completion was recorded yesterday", func(t *testing.T) {
		// id 1 is cook
		got, _ := g.RecordCompletion("cook")

		assertDate(t, got.RecordedAt)

		wantStreak := 4
		if got.Streak != wantStreak {
			t.Errorf("got %d want %d", got.Streak, wantStreak)
		}

		if got.HabitID != 1 {
			t.Errorf("got id %d want %d", got.HabitID, 1)
		}
	})

	t.Run("streak starts over if a record from yesterday was not found", func(t *testing.T) {
		got, _ := g.RecordCompletion("read")

		assertDate(t, got.RecordedAt)

		wantStreak := 1
		if got.Streak != wantStreak {
			t.Errorf("got %d want %d", got.Streak, wantStreak)
		}

		if got.HabitID != 2 {
			t.Errorf("got id %d want %d", got.HabitID, 2)
		}
	})

	t.Run("does not record completion for inactive habits", func(t *testing.T) {
		_, err := g.RecordCompletion("clean")
		assertRecordNotFound(t, err)
	})

	t.Run("only records one completion per day", func(t *testing.T) {
		_, err := g.RecordCompletion("play guitar")
		if err == nil {
			t.Errorf("Expected already recorded error. Got %v", err.Error())
		}
	})

}

func TestGetHabitByName(t *testing.T) {
	db := setup(t)
	g := Database{DB: db}

	t.Run("gets habit by name", func(t *testing.T) {
		want := "cook"
		h, _ := g.getHabitByName(want)

		if h.Name != want {
			t.Errorf("got %v name want %v", h.Name, want)
		}

	})

	t.Run("returns record not found if cannot find habit", func(t *testing.T) {
		_, err := g.getHabitByName("NOT EXISTING")
		assertRecordNotFound(t, err)
	})
}

func TestArchiveHabit(t *testing.T) {
	db := setup(t)
	g := Database{DB: db}

	t.Run("archives an active habit", func(t *testing.T) {
		err := g.ArchiveHabit("cook")
		didNotExpectError(t, err)

		habit, err := g.getHabitByName("cook")
		didNotExpectError(t, err)

		if habit.Active {
			t.Errorf("got %v expected false", habit.Active)
		}
	})

	t.Run("cannot archive an inactive habit", func(t *testing.T) {
		g.ArchiveHabit("clean")
		habit, _ := g.getHabitByName("clean")

		if habit.Active {
			t.Errorf("got %v expected false", habit.Active)
		}
	})
}

func TestRestoreHabit(t *testing.T) {
	db := setup(t)
	g := Database{DB: db}

	t.Run("restores an inactive habit", func(t *testing.T) {
		err := g.RestoreHabit("clean")
		didNotExpectError(t, err)
	})

	t.Run("does not do anything to an active habit", func(t *testing.T) {
		g.RestoreHabit("cook")
		habit, _ := g.getHabitByName("cook")
		if !habit.Active {
			t.Errorf("got %v expected true", habit.Active)
		}

	})
}

func TestGetActiveHabitsAndCompletions(t *testing.T) {
	db := setup(t)
	g := Database{DB: db}

	t.Run("gets active habits and their completions", func(t *testing.T) {
		type Result struct {
			name       string
			recordedAt string
		}

		year, month, _ := time.Now().Date()

		habitsAndCompletions := g.GetActiveHabitsAndCompletions(int(month), year)

		result := []Result{}
		resultMap := map[string]bool{}
		for _, h := range habitsAndCompletions {
			_, ok := resultMap[h.Name]
			if !ok {
				resultMap[h.Name] = true
			}
			if h.Habit.Name == "read" {
				result = append(result, Result{name: h.Name, recordedAt: h.Completion.RecordedAt})
			}
		}

		if len(result) != 3 {
			t.Errorf("expected 3 recordings for read, got %d", len(result))
		}

		if len(resultMap) != 4 {
			t.Errorf("expected 4 habits, got %d", len(resultMap))
		}

	})

	t.Run("returns empty slice if nothing there", func(t *testing.T) {
		month := time.Now().AddDate(0, 1, 0).Month().String()[0:3]
		year := time.Now().AddDate(1, 0, 0).Year()

		h := g.GetActiveHabitsAndCompletions(MonthToIntMap[month], year)
		if len(h) != 0 {
			t.Errorf("expected slice of length 0 but got %v", h)
		}
	})
}

func TestGetAvailableYears(t *testing.T) {
	db := setup(t)
	g := Database{DB: db}

	years := g.GetAvailableYears()
	if len(years) != 3 {
		t.Errorf("expected 3 distinct years, got %d", len(years))
	}
}

func setup(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"),
		&gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		log.Fatalf("unable to open in-memory SQLite DB: %v", err)
	}

	db.AutoMigrate(&Habit{}, &Completion{})

	seedHabits(db)
	t.Cleanup(func() {
		db.Migrator().DropTable(&Habit{}, &Completion{})
	})
	return db
}

func seedHabits(db *gorm.DB) {

	habits := []Habit{
		{Name: "cook", CreatedAt: currentDate(), Active: true},
		{Name: "read", CreatedAt: currentDate(), Active: true},
		{Name: "clean", CreatedAt: currentDate(), Active: false},
		{Name: "garden", CreatedAt: currentDate(), Active: true},
		{Name: "play guitar", CreatedAt: currentDate(), Active: true},
	}

	records := []Completion{
		{RecordedAt: yesterdaysDate(), Streak: 3, HabitID: 1},
		{RecordedAt: time.Now().AddDate(0, 0, -2).Format("2006-01-02"),
			Streak:  4,
			HabitID: 2},
		{RecordedAt: time.Now().AddDate(0, 0, -3).Format("2006-01-02"),
			Streak:  3,
			HabitID: 2},
		{RecordedAt: time.Now().AddDate(0, 0, -4).Format("2006-01-02"),
			Streak:  2,
			HabitID: 2},
		{RecordedAt: yesterdaysDate(), Streak: 510, HabitID: 4},
		{RecordedAt: currentDate(), Streak: 1, HabitID: 5},
		{RecordedAt: time.Now().AddDate(-2, 0, 0).Format("2006-01-02"), Streak: 20, HabitID: 5},
		{RecordedAt: time.Now().AddDate(-10, 0, 0).Format("2006-01-02"), Streak: 20, HabitID: 5},
	}
	for _, i := range habits {
		db.Create(&i)
	}

	for _, i := range records {
		db.Create(&i)
	}

}

func assertDate(t *testing.T, got string) {
	t.Helper()
	_, err := time.Parse("2006-01-02", got)
	if err != nil {
		t.Errorf("got %v, error: %v", got, err)
	}

}

func assertRecordNotFound(t *testing.T, err error) {
	t.Helper()
	if !strings.Contains("record not found", err.Error()) {
		t.Errorf("Expected record not found error")
	}

}

func didNotExpectError(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Errorf("did not expect error %v", err.Error())
	}

}
