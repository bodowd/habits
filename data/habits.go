package data

import (
	"fmt"
	"time"

	"gorm.io/gorm"
)

type Habit struct {
	gorm.Model
	Name      string
	CreatedAt string
	Active    bool
}

type Database struct {
	DB *gorm.DB
}

func NewHabit(name string) *Habit {
	return &Habit{
		Name:   name,
		Active: true,
	}
}

func (d *Database) CreateHabit(name string) (Habit, error) {
	hab := Habit{Name: name, CreatedAt: time.Now().Format("2006-01-02"), Active: true}
	if err := d.DB.Create(&hab).Error; err != nil {
		return hab, fmt.Errorf("Cannot create habit: %v", err)
	}
	return hab, nil
}

func (d *Database) GetActiveHabits() []Habit {
	var habits []Habit
	d.DB.Where("active = ?", true).Find(&habits)

	return habits
}

func (d *Database) GetAllHabits() []Habit {
	var habits []Habit
	d.DB.Find(&habits)
	return habits
}
