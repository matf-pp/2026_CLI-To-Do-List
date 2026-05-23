package main

import (
	"encoding/json"
	"fmt"
	"os"
	"slices"
	"strconv"
	"strings"
)

type Todo struct {
	ID          int    `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Category    string `json:"category"`
	Priority    int    `json:"priority"`
	DueDate     string `json:"due_date"`
	Completed   bool   `json:"completed"`
}

const FileName = "data.json"

var Reset = "\033[0m"
var Red = "\033[31m"
var Green = "\033[32m"
var Yellow = "\033[33m"
var Cyan = "\033[36m"
var Bold = "\033[1m"
var Dim = "\033[2m"

var todos []Todo

func saveToDos() {
	file, err := os.Create(FileName)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error: Could not open file to write")
		os.Exit(1)
	}
	defer file.Close()
	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(todos); err != nil {
		fmt.Fprintln(os.Stderr, "Error: Could not encode todos!")
		os.Exit(1)
	}
}

func LoadTodos() {
	file, err := os.Open(FileName)
	if err != nil {
		todos = []Todo{}
		return
	}
	defer file.Close()
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&todos); err != nil {
		fmt.Fprintln(os.Stderr, "Error: Could not decode todos!")
		os.Exit(1)
	}
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

func stripANSI(s string) string {
	b := strings.Builder{}
	skip := false
	for _, r := range s {
		if r == '\033' {
			skip = true
		}
		if !skip {
			b.WriteRune(r)
		}
		if skip && r == 'm' {
			skip = false
		}
	}
	return b.String()
}

func pad(s string, width int) string {
	diff := width - len([]rune(stripANSI(s)))
	if diff <= 0 {
		return s
	}
	return s + strings.Repeat(" ", diff)
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func MarkDone(id int) {
	idx := FindIndex(id)
	if idx == -1 {
		fmt.Fprintln(os.Stderr, "Error: Invalid todo id!")
		os.Exit(1)
	}
	todos[idx].Completed = true
	saveToDos()
	fmt.Printf("%s✓ Done:%s %s\n", Green, Reset, todos[idx].Title)
}

func UnmarkDone(id int) {
	idx := FindIndex(id)
	if idx == -1 {
		fmt.Fprintln(os.Stderr, "Error: Invalid todo id!")
		os.Exit(1)
	}
	todos[idx].Completed = false
	saveToDos()
	fmt.Printf("%s✓ Pending:%s %s\n", Yellow, Reset, todos[idx].Title)
}

func DeleteTodo(id int) {
	idx := FindIndex(id)
	if idx == -1 {
		fmt.Fprintln(os.Stderr, "Error: Invalid todo id!")
		os.Exit(1)
	}
	title := todos[idx].Title
	todos = slices.Delete(todos, idx, idx+1)
	saveToDos()
	fmt.Printf("%s✓ Deleted:%s %s\n", Red, Reset, title)
}

func PrintUsage() {
	fmt.Printf(`
%sUsage:%s
  add                              -- Interactive TUI form
  list [--category <c>] [--status done|pending]
  show_description <id>
  edit <id> <field> <value>        -- fields: title, description, category, priority, due
  delete <id>
  mark <id>   /   unmark <id>
  stats

`, Bold, Reset)
}

func main() {
	if len(os.Args) < 2 {
		PrintUsage()
		os.Exit(1)
	}

	LoadTodos()

	switch os.Args[1] {
	case "delete":
		if len(os.Args) < 3 {
			fmt.Fprintln(os.Stderr, "Error: Missing id!")
			os.Exit(1)
		}
		id, err := strconv.Atoi(os.Args[2])
		if err != nil {
			fmt.Fprintln(os.Stderr, "Error: Invalid id!")
			os.Exit(1)
		}
		DeleteTodo(id)

	case "mark":
		if len(os.Args) < 3 {
			fmt.Fprintln(os.Stderr, "Error: Missing id!")
			os.Exit(1)
		}
		id, err := strconv.Atoi(os.Args[2])
		if err != nil {
			fmt.Fprintln(os.Stderr, "Error: Invalid id!")
			os.Exit(1)
		}
		MarkDone(id)

	case "unmark":
		if len(os.Args) < 3 {
			fmt.Fprintln(os.Stderr, "Error: Missing id!")
			os.Exit(1)
		}
		id, err := strconv.Atoi(os.Args[2])
		if err != nil {
			fmt.Fprintln(os.Stderr, "Error: Invalid id!")
			os.Exit(1)
		}
		UnmarkDone(id)

	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n", os.Args[1])
		PrintUsage()
		os.Exit(1)
	}
}
