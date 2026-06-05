package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

var Reset = "\033[0m"
var Red = "\033[31m"
var Green = "\033[32m"
var Yellow = "\033[33m"
var Cyan = "\033[36m"
var Bold = "\033[1m"
var Dim = "\033[2m"

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
