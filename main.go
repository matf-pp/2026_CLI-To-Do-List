package main

import (
	"encoding/json"
	"fmt"
	"os"
	"slices"
	"strconv"
	"strings"
	"time"

	"golang.org/x/term"
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

var Reset  = "\033[0m"
var Red    = "\033[31m"
var Green  = "\033[32m"
var Yellow = "\033[33m"
var Cyan   = "\033[36m"
var Bold   = "\033[1m"
var Dim    = "\033[2m"

var todos []Todo

// ─────────────────────────────────────────────
//  File I/O
// ─────────────────────────────────────────────

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

// ─────────────────────────────────────────────
//  Terminal helpers
// ─────────────────────────────────────────────

func clearScreen()           { fmt.Print("\033[H\033[2J") }
func moveCursor(row, col int) { fmt.Printf("\033[%d;%dH", row, col) }
func hideCursor()            { fmt.Print("\033[?25l") }
func showCursor()            { fmt.Print("\033[?25h") }

func readKey(fd int) []byte {
	buf := make([]byte, 6)
	n, _ := os.NewFile(uintptr(fd), "stdin").Read(buf)
	return buf[:n]
}

// readLine reads a line of text in raw mode with echo + backspace.
func readLine(fd int, maxLen int) string {
	input := []rune{}
	raw := os.NewFile(uintptr(fd), "stdin")
	buf := make([]byte, 4)
	for {
		n, _ := raw.Read(buf)
		if n == 0 {
			continue
		}
		b := buf[:n]
		switch {
		case b[0] == 13 || b[0] == 10:
			return string(input)
		case b[0] == 127 || b[0] == 8:
			if len(input) > 0 {
				input = input[:len(input)-1]
				fmt.Print("\b \b")
			}
		case b[0] >= 32 && len(input) < maxLen:
			input = append(input, rune(b[0]))
			fmt.Printf("%c", rune(b[0]))
		}
	}
}

// ─────────────────────────────────────────────
//  Interactive TUI — Add Todo
// ─────────────────────────────────────────────

const boxW = 54

type formState struct {
	title       string
	description string
	category    string
	priority    int // 0=none 1=low 2=medium 3=high
	dueDate     string
	field       int // 0=title 1=desc 2=category 3=priority 4=due 5=confirm 6=cancel
}

func drawAddBox(s formState, errMsg string) {
	clearScreen()

	bline := func(content string) {
		visible := stripANSI(content)
		p := boxW - 2 - len(visible)
		if p < 0 {
			p = 0
		}
		fmt.Printf("│ %s%s │\n", content, strings.Repeat(" ", p))
	}
	sep := func() { fmt.Printf("├%s┤\n", strings.Repeat("─", boxW-2)) }

	fmt.Printf("┌%s┐\n", strings.Repeat("─", boxW-2))
	bline(Bold + "  Add New Todo" + Reset)
	sep()

	type fieldDef struct {
		label string
		value string
		idx   int
	}
	textFields := []fieldDef{
		{"Title      ", s.title, 0},
		{"Description", s.description, 1},
		{"Category   ", s.category, 2},
	}
	for _, f := range textFields {
		cur := "  "
		if s.field == f.idx {
			cur = Cyan + "▶ " + Reset
		}
		val := f.value
		if val == "" {
			val = Dim + "—" + Reset
		} else if len(val) > 24 {
			val = val[:23] + "…"
		}
		bline(cur + Dim + f.label + Reset + " : " + val)
	}

	// Priority row
	priCur := "  "
	if s.field == 3 {
		priCur = Cyan + "▶ " + Reset
	}
	prios := []string{"None", "Low", "Medium", "High"}
	priColors := []string{Dim, Green, Yellow, Red}
	priStr := ""
	for i, p := range prios {
		if i == s.priority {
			priStr += priColors[i] + Bold + "[" + p + "]" + Reset + " "
		} else {
			priStr += Dim + p + " " + Reset
		}
	}
	bline(priCur + Dim + "Priority   " + Reset + " : " + priStr)

	// Due date row
	dueCur := "  "
	if s.field == 4 {
		dueCur = Cyan + "▶ " + Reset
	}
	dueVal := s.dueDate
	if dueVal == "" {
		dueVal = Dim + "—" + Reset
	}
	bline(dueCur + Dim + "Due date   " + Reset + " : " + dueVal)

	sep()

	// Confirm / Cancel
	var confirmStr string
	switch s.field {
	case 5:
		confirmStr = Green + Bold + "[ ✓ Confirm ]" + Reset + "   " + Dim + "[ ✗ Cancel ]" + Reset
	case 6:
		confirmStr = Dim + "[ ✓ Confirm ]" + Reset + "   " + Red + Bold + "[ ✗ Cancel ]" + Reset
	default:
		confirmStr = Dim + "[ ✓ Confirm ]   [ ✗ Cancel ]" + Reset
	}
	bline("  " + confirmStr)
	fmt.Printf("└%s┘\n", strings.Repeat("─", boxW-2))

	if errMsg != "" {
		fmt.Printf("\n  %s%s%s\n", Red, errMsg, Reset)
	} else {
		fmt.Printf("\n  %sTab/↑↓ navigate  ·  Enter edit  ·  ←→ priority%s\n", Dim, Reset)
	}
}

func runAddForm() (Todo, bool) {
	fd := int(os.Stdin.Fd())
	oldState, err := term.MakeRaw(fd)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error: raw mode failed:", err)
		os.Exit(1)
	}
	defer func() {
		term.Restore(fd, oldState)
		showCursor()
		fmt.Println()
	}()

	hideCursor()
	s := formState{}
	errMsg := ""

	for {
		drawAddBox(s, errMsg)
		errMsg = ""
		key := readKey(fd)

		// Arrow keys
		if len(key) == 3 && key[0] == 27 && key[1] == 91 {
			switch key[2] {
			case 65: // Up
				if s.field > 0 {
					s.field--
				}
			case 66: // Down
				if s.field < 6 {
					s.field++
				}
			case 67: // Right — priority +1
				if s.field == 3 && s.priority < 3 {
					s.priority++
				}
			case 68: // Left — priority -1
				if s.field == 3 && s.priority > 0 {
					s.priority--
				}
			}
			continue
		}

		// Tab — move forward
		if key[0] == 9 {
			s.field = (s.field + 1) % 7
			continue
		}

		// Escape — cancel
		if key[0] == 27 && len(key) == 1 {
			return Todo{}, false
		}

		// Enter
		if key[0] == 13 || key[0] == 10 {
			switch s.field {
			case 0, 1, 2, 4: // Text input fields
				// Row numbers in the drawn box (1-indexed, including borders)
				rowMap := map[int]int{0: 4, 1: 5, 2: 6, 4: 8}
				labelMap := map[int]string{
					0: "Title      ",
					1: "Description",
					2: "Category   ",
					4: "Due date   ",
				}
				maxLen := 60
				if s.field == 1 {
					maxLen = 120
				}

				term.Restore(fd, oldState)
				showCursor()

				// Redraw and position cursor inline in the right row
				drawAddBox(s, "")
				moveCursor(rowMap[s.field], 1)
				fmt.Printf("│ %s%s%s : ",
					Cyan+"▶ "+Reset,
					Dim+labelMap[s.field]+Reset,
					"",
				)

				newRaw, _ := term.MakeRaw(fd)
				val := readLine(fd, maxLen)
				term.Restore(fd, newRaw)
				hideCursor()

				if val != "" {
					switch s.field {
					case 0:
						s.title = val
					case 1:
						s.description = val
					case 2:
						s.category = val
					case 4:
						if _, e := time.Parse("2006-01-02", val); e != nil {
							errMsg = "Date must be YYYY-MM-DD  e.g. 2025-06-01"
						} else {
							s.dueDate = val
						}
					}
				}
				oldState, _ = term.MakeRaw(fd)

			case 3: // Priority — Enter cycles
				s.priority = (s.priority + 1) % 4

			case 5: // Confirm
				if s.title == "" {
					errMsg = "Title cannot be empty!"
					continue
				}
				if s.category == "" {
					errMsg = "Category cannot be empty!"
					continue
				}
				t := Todo{
					ID:          GetNextID(),
					Title:       s.title,
					Description: s.description,
					Category:    s.category,
					Priority:    s.priority,
					DueDate:     s.dueDate,
					Completed:   false,
				}
				return t, true

			case 6: // Cancel
				return Todo{}, false
			}
		}
	}
}

// ─────────────────────────────────────────────
//  Helpers
// ─────────────────────────────────────────────

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
		return Dim + "-          " + Reset
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

// ─────────────────────────────────────────────
//  Table display
// ─────────────────────────────────────────────

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

	// padColor pads a string that may contain ANSI codes to exact visible width
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
		if len([]rune(title)) > wTitle {
			title = string([]rune(title)[:wTitle-1]) + "…"
		}

		desc := t.Description
		if desc == "" {
			desc = Dim + "-" + Reset
		} else if len([]rune(desc)) > wDesc {
			desc = string([]rune(desc)[:wDesc-1]) + "…"
		}

		cat := t.Category
		if len([]rune(cat)) > wCat {
			cat = string([]rune(cat)[:wCat-1]) + "…"
		}

		// # indicator if description exists — takes 2 chars, so truncate title accordingly
		titleDisplay := title
		if t.Description != "" {
			maxT := wTitle - 2
			if len([]rune(title)) > maxT {
				title = string([]rune(title)[:maxT-1]) + "…"
			}
			titleDisplay = title + Cyan + " #" + Reset
		}

		row(strconv.Itoa(t.ID), titleDisplay, desc, cat,
			PriorityLabel(t.Priority), DueDateLabel(t.DueDate), StatusLabel(t.Completed))
	}

	border("└", "┴", "┘", "─")
}

// ─────────────────────────────────────────────
//  Commands
// ─────────────────────────────────────────────

func AddTodoInteractive() {
	t, ok := runAddForm()
	clearScreen()
	if !ok {
		fmt.Println(Dim + "Cancelled." + Reset)
		return
	}
	todos = append(todos, t)
	saveToDos()
	fmt.Printf("%s✓ Added:%s \"%s\" (ID: %d)\n", Green, Reset, t.Title, t.ID)
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
		p := w - 2 - len(stripANSI(content))
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
		dueStr = "—"
	}
	bline(Dim + "Category " + Reset + ": " + t.Category)
	bline(Dim + "Priority " + Reset + ": " + stripANSI(PriorityLabel(t.Priority)))
	bline(Dim + "Due date " + Reset + ": " + dueStr)
	bline(Dim + "Status   " + Reset + ": " + stripANSI(StatusLabel(t.Completed)))
	sep()

	if t.Description == "" {
		bline(Dim + "  (no description)" + Reset)
	} else {
		// Word-wrap to fit box
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
	fmt.Printf("\n  %s#%s  = has description  →  show_description <id>\n\n", Cyan, Reset)
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
	case "description":
		todos[idx].Description = value
	case "category":
		todos[idx].Category = value
	case "priority":
		p, err := strconv.Atoi(value)
		if err != nil || p < 1 || p > 3 {
			fmt.Fprintln(os.Stderr, "Error: Priority must be 1, 2, or 3")
			os.Exit(1)
		}
		todos[idx].Priority = p
	case "due":
		if value != "" {
			if _, err := time.Parse("2006-01-02", value); err != nil {
				fmt.Fprintln(os.Stderr, "Error: Due date must be YYYY-MM-DD")
				os.Exit(1)
			}
		}
		todos[idx].DueDate = value
	default:
		fmt.Fprintf(os.Stderr, "Error: Unknown field '%s'. Use: title, description, category, priority, due\n", field)
		os.Exit(1)
	}
	saveToDos()
	fmt.Printf("%s✓ Updated:%s #%d — %s = \"%s\"\n", Green, Reset, id, field, value)
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
	for cat, count := range cats {
		fmt.Printf("    %-15s %d\n", cat, count)
	}
	fmt.Println()
}

// ─────────────────────────────────────────────
//  Usage & Main
// ─────────────────────────────────────────────

func PrintUsage() {
	fmt.Printf(`
%sUsage:%s
  add                              -- Interactive TUI form
  list [--category <c>] [--status done|pending]
  show_description <id>            -- Show full description
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

	case "add":
		AddTodoInteractive()

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

	case "edit":
		if len(os.Args) < 5 {
			fmt.Fprintln(os.Stderr, "Usage: edit <id> <field> <value>")
			os.Exit(1)
		}
		id, err := strconv.Atoi(os.Args[2])
		if err != nil {
			fmt.Fprintln(os.Stderr, "Error: Invalid id!")
			os.Exit(1)
		}
		EditTodo(id, os.Args[3], os.Args[4])

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

	case "stats":
		PrintStats()

	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n", os.Args[1])
		PrintUsage()
		os.Exit(1)
	}
}
