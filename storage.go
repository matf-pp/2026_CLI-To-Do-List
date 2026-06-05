package main

import (
	"fmt"
	"os"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

const DBFile = "todos.db"

var db *gorm.DB
var todos []Todo

func InitDB() {
	var err error
	db, err = gorm.Open(sqlite.Open(DBFile), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error: Could not open database!")
		os.Exit(1)
	}

	if err := db.AutoMigrate(&Todo{}); err != nil {
		fmt.Fprintln(os.Stderr, "Error: Could not migrate database!")
		os.Exit(1)
	}
}

func LoadTodos() {
	db.Order("id asc").Find(&todos)
}

func GetNextID() int {
	id := 0
	for _, t := range todos {
		if t.ID > id {
			id = t.ID
		}
	}
	return id + 1
}

func FindIndex(id int) int {
	for i, t := range todos {
		if t.ID == id {
			return i
		}
	}
	return -1
}

func SaveTodo(t Todo) {
	db.Create(&t)
	todos = append(todos, t)
}

func UpdateTodo(idx int) {
	db.Save(&todos[idx])
}

func RemoveTodo(idx int) {
	db.Delete(&todos[idx])
	todos = append(todos[:idx], todos[idx+1:]...)
}
