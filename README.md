# 2026_CLI-To-Do-List

[![Codacy Badge](https://app.codacy.com/project/badge/Grade/218997beba334ff4ae5b8d8a05ea6dc6)](https://app.codacy.com/gh/matf-pp/2026_CLI-To-Do-List/dashboard?utm_source=gh&utm_medium=referral&utm_content=&utm_campaign=Badge_grade)

CLI To-Do List (Go) Jednostavna komandno-linijska aplikacija u programskom jeziku Go za upravljanje dnevnim zadacima

## Requirements

- Go 1.21+

## Installation

```bash
git clone <repo-url>
cd <repo-dir>
go build -o todo .
```

## Usage

```
./todo <command> [options]
```

### Commands

| Command | Description |
|---|---|
| `add` | Open interactive TUI form to add a new task |
| `list` | List all tasks, sorted by priority |
| `show_description <id>` | Show full details for a task |
| `edit <id> <field> <value>` | Edit a field on an existing task |
| `mark <id>` | Mark a task as done |
| `unmark <id>` | Mark a task as pending |
| `delete <id>` | Delete a task |
| `stats` | Show statistics and progress |

### List filters

```bash
./todo list --category faks
./todo list --status done
./todo list --status pending
./todo list --category faks --status pending
```

### Editable fields

```bash
./todo edit <id> title      "New title"
./todo edit <id> description "New description"
./todo edit <id> category   "work"
./todo edit <id> priority   3        # 1=low, 2=medium, 3=high
./todo edit <id> due        2026-07-01
```

## Add form controls

When running `add`, a TUI form opens in the terminal:

| Key | Action |
|---|---|
| `Tab` / `↑` `↓` | Navigate between fields |
| `Enter` | Edit the selected text field |
| `←` `→` | Change priority (on the priority row) |
| `Escape` | Cancel and exit |

Title and Category are required. Due date must be in `YYYY-MM-DD` format.

## Data

Tasks are saved to `data.json` in the current directory. The file is created automatically on first use. You can edit it manually if needed — each task has the following fields:

```json
{
  "id": 1,
  "title": "Task title",
  "description": "",
  "category": "work",
  "priority": 3,
  "due_date": "2026-06-01",
  "completed": false
}
```

Priority values: `0` = none, `1` = low, `2` = medium, `3` = high.

