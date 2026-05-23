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

// ── Display helpers ───────────────────────────

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
		return Green + "[+] Done   " + Reset
	}
	return Red + "[-] Pending" + Reset
}

func DueDateLabel(due string) string {
	if due == "" {
		return Dim + "-           " + Reset
	}
	t, err := time.Parse("2006-01-02", due)
	if err != nil {
		return due + "  "
	}
	today := time.Now().Truncate(24 * time.Hour)
	if t.Before(today) {
		return Red + due + " !" + Reset
	}
	if t.Equal(today) {
		return Yellow + due + " *" + Reset
	}
	return Cyan + due + "  " + Reset
}

// ── Table ─────────────────────────────────────

func PrintTable(list []Todo) {
	if len(list) == 0 {
		fmt.Println(Dim + "  No todos found." + Reset)
		return
	}

	wID, wTitle, wDesc, wCat, wPri, wDue, wStatus := 4, 20, 18, 11, 8, 12, 11

	border := func(l, m, r, h string) {
		cols := []int{wID, wTitle, wDesc, wCat, wPri, wDue, wStatus}
		fmt.Print(l)
		for i, w := range cols {
			fmt.Print(strings.Repeat(h, w+2))
			if i < len(cols)-1 {
				fmt.Print(m)
			}
		}
		fmt.Println(r)
	}

	padColor := func(s string, width int) string {
		diff := width - len([]rune(stripANSI(s)))
		if diff <= 0 {
			return s
		}
		return s + strings.Repeat(" ", diff)
	}

	row := func(id, title, desc, cat, pri, due, status string) {
		fmt.Printf("│ %s │ %s │ %s │ %s │ %s │ %s │ %s │\n",
			padColor(id, wID),
			padColor(title, wTitle),
			padColor(desc, wDesc),
			padColor(cat, wCat),
			padColor(pri, wPri),
			padColor(due, wDue),
			padColor(status, wStatus),
		)
	}

	border("┌", "┬", "┐", "─")
	row(Bold+"ID"+Reset, Bold+"Title"+Reset, Bold+"Description"+Reset,
		Bold+"Category"+Reset, Bold+"Priority"+Reset, Bold+"Due Date"+Reset, Bold+"Status"+Reset)
	border("├", "┼", "┤", "─")

	for _, t := range list {
		title := t.Title
		desc := t.Description
		cat := t.Category

		// # indicator counts as 2 chars — shrink title to fit
		titleDisplay := title
		if t.Description != "" {
			maxT := wTitle - 2
			if len([]rune(title)) > maxT {
				title = string([]rune(title)[:maxT-1]) + "…"
			}
			titleDisplay = title + Cyan + " #" + Reset
		} else if len([]rune(title)) > wTitle {
			title = string([]rune(title)[:wTitle-1]) + "…"
			titleDisplay = title
		}

		if desc == "" {
			desc = Dim + "-" + Reset
		} else if len([]rune(desc)) > wDesc {
			desc = string([]rune(desc)[:wDesc-1]) + "…"
		}

		if len([]rune(cat)) > wCat {
			cat = string([]rune(cat)[:wCat-1]) + "…"
		}

		row(strconv.Itoa(t.ID), titleDisplay, desc, cat,
			PriorityLabel(t.Priority), DueDateLabel(t.DueDate), StatusLabel(t.Completed))
	}

	border("└", "┴", "┘", "─")
}

// ── Commands ──────────────────────────────────

func ListTodos(filterCat, filterStatus string) {
	list := []Todo{}
	for _, t := range todos {
		if filterCat != "" && !strings.EqualFold(t.Category, filterCat) {
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

	slices.SortFunc(list, func(a, b Todo) int {
		if b.Priority != a.Priority {
			return b.Priority - a.Priority
		}
		return a.ID - b.ID
	})

	fmt.Printf("\n%sTodos%s", Bold, Reset)
	if filterCat != "" {
		fmt.Printf(" — category: %s%s%s", Cyan, filterCat, Reset)
	}
	if filterStatus != "" {
		fmt.Printf(" — status: %s%s%s", Cyan, filterStatus, Reset)
	}
	fmt.Printf(" (%d)\n\n", len(list))
	PrintTable(list)
	fmt.Printf("\n  %s#%s  = has description  ->  show_description <id>\n\n", Cyan, Reset)
}

func ShowDescription(id int) {
	idx := FindIndex(id)
	if idx == -1 {
		fmt.Fprintln(os.Stderr, "Error: Invalid todo id!")
		os.Exit(1)
	}
	t := todos[idx]

	w := 54
	bline := func(content string) {
		p := w - 2 - len([]rune(stripANSI(content)))
		if p < 0 {
			p = 0
		}
		fmt.Printf("│ %s%s │\n", content, strings.Repeat(" ", p))
	}
	sep := func() { fmt.Printf("├%s┤\n", strings.Repeat("─", w-2)) }

	fmt.Printf("┌%s┐\n", strings.Repeat("─", w-2))
	bline(Bold + fmt.Sprintf("  #%d — %s", t.ID, t.Title) + Reset)
	sep()

	dueStr := t.DueDate
	if dueStr == "" {
		dueStr = "-"
	}
	bline(Dim + "Category " + Reset + ": " + t.Category)
	bline(Dim + "Priority " + Reset + ": " + stripANSI(PriorityLabel(t.Priority)))
	bline(Dim + "Due date " + Reset + ": " + dueStr)
	bline(Dim + "Status   " + Reset + ": " + stripANSI(StatusLabel(t.Completed)))
	sep()

	if t.Description == "" {
		bline(Dim + "  (no description)" + Reset)
	} else {
		words := strings.Fields(t.Description)
		cur := ""
		for _, word := range words {
			if len(cur)+len(word)+1 > w-4 {
				bline("  " + cur)
				cur = word
			} else {
				if cur == "" {
					cur = word
				} else {
					cur += " " + word
				}
			}
		}
		if cur != "" {
			bline("  " + cur)
		}
	}

	fmt.Printf("└%s┘\n", strings.Repeat("─", w-2))
}

func PrintStats() {
	total := len(todos)
	if total == 0 {
		fmt.Println(Dim + "No todos yet." + Reset)
		return
	}
	done, overdue, dueToday, high := 0, 0, 0, 0
	today := time.Now().Truncate(24 * time.Hour)
	cats := map[string]int{}

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
		cats[t.Category]++
	}

	pending := total - done
	pct := done * 100 / total
	bw := 30
	filled := bw * pct / 100
	bar := Green + strings.Repeat("█", filled) + Reset + Dim + strings.Repeat("░", bw-filled) + Reset

	fmt.Printf("\n%s=== Statistics =====================================%s\n\n", Bold, Reset)
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
	for cat, count := range cats {
		fmt.Printf("    %-15s %d\n", cat, count)
	}
	fmt.Println()
}

func MarkDone(id int) {
	idx := FindIndex(id)
	if idx == -1 {
		fmt.Fprintln(os.Stderr, "Error: Invalid todo id!")
		os.Exit(1)
	}
	todos[idx].Completed = true
	saveToDos()
	fmt.Printf("%s[+] Done:%s %s\n", Green, Reset, todos[idx].Title)
}

func UnmarkDone(id int) {
	idx := FindIndex(id)
	if idx == -1 {
		fmt.Fprintln(os.Stderr, "Error: Invalid todo id!")
		os.Exit(1)
	}
	todos[idx].Completed = false
	saveToDos()
	fmt.Printf("%s[-] Pending:%s %s\n", Yellow, Reset, todos[idx].Title)
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
	fmt.Printf("%s[-] Deleted:%s %s\n", Red, Reset, title)
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
	case "list":
		filterCat, filterStatus := "", ""
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

	case "show_description":
		if len(os.Args) < 3 {
			fmt.Fprintln(os.Stderr, "Usage: show_description <id>")
			os.Exit(1)
		}
		id, err := strconv.Atoi(os.Args[2])
		if err != nil {
			fmt.Fprintln(os.Stderr, "Error: Invalid id!")
			os.Exit(1)
		}
		ShowDescription(id)

	case "stats":
		PrintStats()

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
