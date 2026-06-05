package main

type Todo struct {
	ID          int    `gorm:"primaryKey" json:"id"`
	Title       string `gorm:"not null"   json:"title"`
	Description string `json:"description"`
	Category    string `json:"category"`
	Priority    int    `json:"priority"`
	DueDate     string `json:"due_date"`
	Completed   bool   `json:"completed"`
}
