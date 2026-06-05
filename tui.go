package main

import (
	"fmt"
	"os"
	"strings"
	"time"

	"golang.org/x/term"
)

const boxW = 54

type formState struct {
	title       string
	description string
	category    string
	priority    int
	dueDate     string
	field       int
}

func clearScreen()            { fmt.Print("\033[H\033[2J") }
func moveCursor(row, col int) { fmt.Printf("\033[%d;%dH", row, col) }
func hideCursor()             { fmt.Print("\033[?25l") }
func showCursor()             { fmt.Print("\033[?25h") }

func readKey(fd int) []byte {
	buf := make([]byte, 6)
	n, _ := os.NewFile(uintptr(fd), "stdin").Read(buf)
	return buf[:n]
}

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

func drawAddBox(s formState, errMsg string) {
	clearScreen()

	bline := func(content string) {
		visible := stripANSI(content)
		p := boxW - 2 - len([]rune(visible))
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
			val = Dim + "-" + Reset
		} else if len([]rune(val)) > 24 {
			val = string([]rune(val)[:23]) + "…"
		}
		bline(cur + Dim + f.label + Reset + " : " + val)
	}

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

	dueCur := "  "
	if s.field == 4 {
		dueCur = Cyan + "▶ " + Reset
	}
	dueVal := s.dueDate
	if dueVal == "" {
		dueVal = Dim + "-" + Reset
	}
	bline(dueCur + Dim + "Due date   " + Reset + " : " + dueVal)

	sep()

	var confirmStr string
	switch s.field {
	case 5:
		confirmStr = Green + Bold + "[ Confirm ]" + Reset + "   " + Dim + "[ Cancel ]" + Reset
	case 6:
		confirmStr = Dim + "[ Confirm ]" + Reset + "   " + Red + Bold + "[ Cancel ]" + Reset
	default:
		confirmStr = Dim + "[ Confirm ]   [ Cancel ]" + Reset
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

		if len(key) == 3 && key[0] == 27 && key[1] == 91 {
			switch key[2] {
			case 65:
				if s.field > 0 {
					s.field--
				}
			case 66:
				if s.field < 6 {
					s.field++
				}
			case 67:
				if s.field == 3 && s.priority < 3 {
					s.priority++
				}
			case 68:
				if s.field == 3 && s.priority > 0 {
					s.priority--
				}
			}
			continue
		}

		if key[0] == 9 {
			s.field = (s.field + 1) % 7
			continue
		}

		if key[0] == 27 && len(key) == 1 {
			return Todo{}, false
		}

		if key[0] == 13 || key[0] == 10 {
			switch s.field {
			case 0, 1, 2, 4:
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
							errMsg = "Date must be YYYY-MM-DD  e.g. 2026-06-01"
						} else {
							s.dueDate = val
						}
					}
				}
				oldState, _ = term.MakeRaw(fd)

			case 3:
				s.priority = (s.priority + 1) % 4

			case 5:
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

			case 6:
				return Todo{}, false
			}
		}
	}
}

func AddTodoInteractive() {
	t, ok := runAddForm()
	clearScreen()
	if !ok {
		fmt.Println(Dim + "Cancelled." + Reset)
		return
	}
	SaveTodo(t)
	fmt.Printf("%s[+] Added:%s \"%s\" (ID: %d)\n", Green, Reset, t.Title, t.ID)
}
