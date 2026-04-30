package main

import (
	"encoding/json"
	"fmt"
	"os"
	"slices"
	"strconv"
	"strings"
	"time"
)

type Todo struct {
	ID        int    `json:"id"`
	Title     string `json:"title"`
	Category  string `json:"category"`
	Priority  int    `json:"priority"`
	DueDate   string `json:"due_date"`
	Completed bool   `json:"completed"`
}

const FileName = "data.json"

var Reset = "\033[0m"
var Red = "\033[31m"
var Green = "\033[32m"
var Yellow = "\033[33m"
var Blue = "\033[34m"
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
		// First run — start with empty list
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
	for _, todo := range todos {
		if todo.ID > id {
			id = todo.ID
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

func PriorityLabel(p int) string {
	switch p {
	case 1:
		return Green + "Low   " + Reset
	case 2:
		return Yellow + "Medium" + Reset
	case 3:
		return Red + "High  " + Reset
	default:
		return Dim + "None  " + Reset
	}
}

func StatusLabel(completed bool) string {
	if completed {
		return Green + "✓ Done   " + Reset
	}
	return Red + "✗ Pending" + Reset
}

func DueDateLabel(due string) string {
	if due == "" {
		return Dim + "—          " + Reset
	}
	t, err := time.Parse("2006-01-02", due)
	if err != nil {
		return due
	}
	today := time.Now().Truncate(24 * time.Hour)
	if t.Before(today) {
		return Red + due + " !" + Reset
	}
	if t.Equal(today) {
		return Yellow + due + " ★" + Reset
	}
	return Cyan + due + "  " + Reset
}

// Pad or truncate string to exact width
func pad(s string, width int) string {
	// Strip ANSI codes for length calculation
	visible := stripANSI(s)
	diff := width - len(visible)
	if diff <= 0 {
		return s[:width]
	}
	return s + strings.Repeat(" ", diff)
}

func stripANSI(s string) string {
	result := strings.Builder{}
	skip := false
	for _, r := range s {
		if r == '\033' {
			skip = true
		}
		if !skip {
			result.WriteRune(r)
		}
		if skip && r == 'm' {
			skip = false
		}
	}
	return result.String()
}

func PrintTable(list []Todo) {
	if len(list) == 0 {
		fmt.Println(Dim + "  No todos found." + Reset)
		return
	}

	// Column widths
	wID := 4
	wTitle := 22
	wCategory := 12
	wPriority := 8
	wDue := 13
	wStatus := 11

	sep := func() {
		fmt.Printf("├%s┼%s┼%s┼%s┼%s┼%s┤\n",
			strings.Repeat("─", wID+2),
			strings.Repeat("─", wTitle+2),
			strings.Repeat("─", wCategory+2),
			strings.Repeat("─", wPriority+2),
			strings.Repeat("─", wDue+2),
			strings.Repeat("─", wStatus+2),
		)
	}

	// Top border
	fmt.Printf("┌%s┬%s┬%s┬%s┬%s┬%s┐\n",
		strings.Repeat("─", wID+2),
		strings.Repeat("─", wTitle+2),
		strings.Repeat("─", wCategory+2),
		strings.Repeat("─", wPriority+2),
		strings.Repeat("─", wDue+2),
		strings.Repeat("─", wStatus+2),
	)

	// Header
	fmt.Printf("│ %s │ %s │ %s │ %s │ %s │ %s │\n",
		Bold+pad("ID", wID)+Reset,
		Bold+pad("Title", wTitle)+Reset,
		Bold+pad("Category", wCategory)+Reset,
		Bold+pad("Priority", wPriority)+Reset,
		Bold+pad("Due Date", wDue)+Reset,
		Bold+pad("Status", wStatus)+Reset,
	)

	sep()

	for _, t := range list {
		title := t.Title
		if len(title) > wTitle {
			title = title[:wTitle-1] + "…"
		}
		cat := t.Category
		if len(cat) > wCategory {
			cat = cat[:wCategory-1] + "…"
		}

		fmt.Printf("│ %s │ %s │ %s │ %s │ %s │ %s │\n",
			pad(strconv.Itoa(t.ID), wID),
			pad(title, wTitle),
			pad(cat, wCategory),
			pad(stripANSI(PriorityLabel(t.Priority)), wPriority)[:0]+PriorityLabel(t.Priority)+strings.Repeat(" ", max(0, wPriority-len(stripANSI(PriorityLabel(t.Priority))))),
			DueDateLabel(t.DueDate),
			StatusLabel(t.Completed),
		)
	}

	// Bottom border
	fmt.Printf("└%s┴%s┴%s┴%s┴%s┴%s┘\n",
		strings.Repeat("─", wID+2),
		strings.Repeat("─", wTitle+2),
		strings.Repeat("─", wCategory+2),
		strings.Repeat("─", wPriority+2),
		strings.Repeat("─", wDue+2),
		strings.Repeat("─", wStatus+2),
	)
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func AddTodo(title, category string, priority int, dueDate string) {
	id := GetNextID()
	todo := Todo{
		ID:        id,
		Title:     title,
		Category:  category,
		Priority:  priority,
		DueDate:   dueDate,
		Completed: false,
	}
	todos = append(todos, todo)
	saveToDos()
	fmt.Printf("%s✓ Added:%s %s (ID: %d)\n", Green, Reset, title, id)
}

func ListTodos(filterCategory, filterStatus string) {
	list := []Todo{}
	for _, t := range todos {
		if filterCategory != "" && !strings.EqualFold(t.Category, filterCategory) {
			continue
		}
		if filterStatus == "done" && !t.Completed {
			continue
		}
		if filterStatus == "pending" && t.Completed {
			continue
		}
		list = append(list, t)
	}

	// Sort by priority descending (3 = High first), then by ID
	slices.SortFunc(list, func(a, b Todo) int {
		if b.Priority != a.Priority {
			return b.Priority - a.Priority
		}
		return a.ID - b.ID
	})

	total := len(list)
	fmt.Printf("\n%s Todos%s", Bold, Reset)
	if filterCategory != "" {
		fmt.Printf(" — category: %s%s%s", Cyan, filterCategory, Reset)
	}
	if filterStatus != "" {
		fmt.Printf(" — status: %s%s%s", Cyan, filterStatus, Reset)
	}
	fmt.Printf(" (%d)\n\n", total)

	PrintTable(list)
	fmt.Println()
}

func EditTodo(id int, field, value string) {
	idx := FindIndex(id)
	if idx == -1 {
		fmt.Fprintln(os.Stderr, "Error: Invalid todo id!")
		os.Exit(1)
	}

	switch strings.ToLower(field) {
	case "title":
		todos[idx].Title = value
	case "category":
		todos[idx].Category = value
	case "priority":
		p, err := strconv.Atoi(value)
		if err != nil || p < 1 || p > 3 {
			fmt.Fprintln(os.Stderr, "Error: Priority must be 1 (low), 2 (medium), or 3 (high)")
			os.Exit(1)
		}
		todos[idx].Priority = p
	case "due":
		if value != "" {
			_, err := time.Parse("2006-01-02", value)
			if err != nil {
				fmt.Fprintln(os.Stderr, "Error: Due date must be in YYYY-MM-DD format")
				os.Exit(1)
			}
		}
		todos[idx].DueDate = value
	default:
		fmt.Fprintf(os.Stderr, "Error: Unknown field '%s'. Use: title, category, priority, due\n", field)
		os.Exit(1)
	}

	saveToDos()
	fmt.Printf("%s✓ Updated:%s todo #%d — %s set to \"%s\"\n", Green, Reset, id, field, value)
}

func MarkDone(id int) {
	idx := FindIndex(id)
	if idx == -1 {
		fmt.Fprintln(os.Stderr, "Error: Invalid todo id!")
		os.Exit(1)
	}
	todos[idx].Completed = true
	saveToDos()
	fmt.Printf("%s✓ Marked done:%s %s\n", Green, Reset, todos[idx].Title)
}

func UnmarkDone(id int) {
	idx := FindIndex(id)
	if idx == -1 {
		fmt.Fprintln(os.Stderr, "Error: Invalid todo id!")
		os.Exit(1)
	}
	todos[idx].Completed = false
	saveToDos()
	fmt.Printf("%s✓ Marked pending:%s %s\n", Yellow, Reset, todos[idx].Title)
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

func PrintStats() {
	total := len(todos)
	if total == 0 {
		fmt.Println(Dim + "No todos yet." + Reset)
		return
	}

	done := 0
	overdue := 0
	dueToday := 0
	high := 0
	today := time.Now().Truncate(24 * time.Hour)

	categories := map[string]int{}

	for _, t := range todos {
		if t.Completed {
			done++
		}
		if t.Priority == 3 {
			high++
		}
		if t.DueDate != "" {
			d, err := time.Parse("2006-01-02", t.DueDate)
			if err == nil {
				if !t.Completed && d.Before(today) {
					overdue++
				}
				if d.Equal(today) {
					dueToday++
				}
			}
		}
		categories[t.Category]++
	}

	pending := total - done
	pct := 0
	if total > 0 {
		pct = done * 100 / total
	}

	barWidth := 30
	filled := barWidth * pct / 100
	bar := Green + strings.Repeat("█", filled) + Reset + Dim + strings.Repeat("░", barWidth-filled) + Reset

	fmt.Printf("\n%s═══ Statistics ══════════════════════%s\n\n", Bold, Reset)
	fmt.Printf("  Total     : %s%d%s\n", Bold, total, Reset)
	fmt.Printf("  Done      : %s%d%s\n", Green, done, Reset)
	fmt.Printf("  Pending   : %s%d%s\n", Yellow, pending, Reset)
	fmt.Printf("  Progress  : [%s] %d%%\n\n", bar, pct)

	if overdue > 0 {
		fmt.Printf("  %sOverdue   : %d%s\n", Red, overdue, Reset)
	}
	if dueToday > 0 {
		fmt.Printf("  %sDue today : %d%s\n", Yellow, dueToday, Reset)
	}
	if high > 0 {
		fmt.Printf("  %sHigh prio : %d%s\n", Red, high, Reset)
	}

	fmt.Printf("\n  %sBy category:%s\n", Bold, Reset)
	for cat, count := range categories {
		fmt.Printf("    %-15s %d\n", cat, count)
	}
	fmt.Println()
}

func PrintUsage() {
	fmt.Printf(`
%sUsage:%s
  add <title> <category> [priority 1-3] [due YYYY-MM-DD]
  list [--category <cat>] [--status done|pending]
  edit <id> <field> <value>        field: title, category, priority, due
  delete <id>
  mark <id>
  unmark <id>
  stats

%sExamples:%s
  todo add "Buy milk" personal 1 2025-06-01
  todo list --category study
  todo list --status pending
  todo edit 3 priority 3
  todo edit 3 due 2025-06-15
  todo stats

%sPriority:%s  1 = Low   2 = Medium   3 = High
`, Bold, Reset, Bold, Reset, Bold, Reset)
}

func main() {
	if len(os.Args) < 2 {
		PrintUsage()
		os.Exit(1)
	}

	LoadTodos()
	command := os.Args[1]

	switch command {
	case "add":
		if len(os.Args) < 4 {
			fmt.Fprintln(os.Stderr, "Usage: add <title> <category> [priority 1-3] [due YYYY-MM-DD]")
			os.Exit(1)
		}
		title := os.Args[2]
		category := os.Args[3]
		priority := 0
		dueDate := ""

		// Parse optional args: priority and due date (order-independent)
		for i := 4; i < len(os.Args); i++ {
			arg := os.Args[i]
			if p, err := strconv.Atoi(arg); err == nil && p >= 1 && p <= 3 {
				priority = p
			} else if _, err := time.Parse("2006-01-02", arg); err == nil {
				dueDate = arg
			}
		}
		AddTodo(title, category, priority, dueDate)

	case "list":
		filterCat := ""
		filterStatus := ""
		args := os.Args[2:]
		for i := 0; i < len(args); i++ {
			switch args[i] {
			case "--category":
				if i+1 < len(args) {
					filterCat = args[i+1]
					i++
				}
			case "--status":
				if i+1 < len(args) {
					filterStatus = args[i+1]
					i++
				}
			}
		}
		ListTodos(filterCat, filterStatus)

	case "edit":
		if len(os.Args) < 5 {
			fmt.Fprintln(os.Stderr, "Usage: edit <id> <field> <value>")
			os.Exit(1)
		}
		id, err := strconv.Atoi(os.Args[2])
		if err != nil {
			fmt.Fprintln(os.Stderr, "Error: Invalid todo id!")
			os.Exit(1)
		}
		EditTodo(id, os.Args[3], os.Args[4])

	case "delete":
		if len(os.Args) < 3 {
			fmt.Fprintln(os.Stderr, "Error: Missing todo id!")
			os.Exit(1)
		}
		id, err := strconv.Atoi(os.Args[2])
		if err != nil {
			fmt.Fprintln(os.Stderr, "Error: Invalid todo id!")
			os.Exit(1)
		}
		DeleteTodo(id)

	case "mark":
		if len(os.Args) < 3 {
			fmt.Fprintln(os.Stderr, "Error: Missing todo id!")
			os.Exit(1)
		}
		id, err := strconv.Atoi(os.Args[2])
		if err != nil {
			fmt.Fprintln(os.Stderr, "Error: Invalid todo id!")
			os.Exit(1)
		}
		MarkDone(id)

	case "unmark":
		if len(os.Args) < 3 {
			fmt.Fprintln(os.Stderr, "Error: Missing todo id!")
			os.Exit(1)
		}
		id, err := strconv.Atoi(os.Args[2])
		if err != nil {
			fmt.Fprintln(os.Stderr, "Error: Invalid todo id!")
			os.Exit(1)
		}
		UnmarkDone(id)

	case "stats":
		PrintStats()

	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n", command)
		PrintUsage()
		os.Exit(1)
	}
}
