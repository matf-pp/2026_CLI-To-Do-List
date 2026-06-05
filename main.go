package main

import (
	"fmt"
	"os"
	"strconv"
)

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

	InitDB()
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
