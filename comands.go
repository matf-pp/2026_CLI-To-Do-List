package main

import (
	"fmt"
	"os"
	"slices"
	"strconv"
	"strings"
	"time"
)

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
	UpdateTodo(idx)
	fmt.Printf("%s[+] Updated:%s #%d — %s = \"%s\"\n", Green, Reset, id, field, value)
}

func MarkDone(id int) {
	idx := FindIndex(id)
	if idx == -1 {
		fmt.Fprintln(os.Stderr, "Error: Invalid todo id!")
		os.Exit(1)
	}
	todos[idx].Completed = true
	UpdateTodo(idx)
	fmt.Printf("%s[+] Done:%s %s\n", Green, Reset, todos[idx].Title)
}

func UnmarkDone(id int) {
	idx := FindIndex(id)
	if idx == -1 {
		fmt.Fprintln(os.Stderr, "Error: Invalid todo id!")
		os.Exit(1)
	}
	todos[idx].Completed = false
	UpdateTodo(idx)
	fmt.Printf("%s[-] Pending:%s %s\n", Yellow, Reset, todos[idx].Title)
}

func DeleteTodo(id int) {
	idx := FindIndex(id)
	if idx == -1 {
		fmt.Fprintln(os.Stderr, "Error: Invalid todo id!")
		os.Exit(1)
	}
	title := todos[idx].Title
	RemoveTodo(idx)
	fmt.Printf("%s[-] Deleted:%s %s\n", Red, Reset, title)
}
