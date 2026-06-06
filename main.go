
package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

    tea "github.com/charmbracelet/bubbletea"
    "github.com/kavinbharathii/quest/db"
    "github.com/kavinbharathii/quest/history"
    "github.com/kavinbharathii/quest/index"
    "github.com/kavinbharathii/quest/tui"
)

func main() {
	dbPath := questDBPath()

	if len(os.Args) < 2 {
		runUI(dbPath)
		return
	}

	switch os.Args[1] {
	case "sync":
		runSync(dbPath)

	case "add":
		if len(os.Args) < 3 {
			fmt.Fprintln(os.Stderr, "usaeg: quest add <command>")
			os.Exit(1)
		}
		runAdd(dbPath, strings.Join(os.Args[2:], " "))

	case "ui":
		runUI(dbPath)

	case "help":
		printUsage()

	default:
		runSearch(dbPath, strings.Join(os.Args[1:], " "))
	}
}

func questDBPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		panic(err)
	}

	dir := filepath.Join(home, ".local", "share", "quest")
	if err := os.MkdirAll(dir, 0755); err != nil {
		panic(err)
	}

	return filepath.Join(dir, "quest.db")
}

func runSync (dbPath string) {
	home, _ := os.UserHomeDir()
	histPath := filepath.Join(home, ".bash_history")

	fmt.Printf("reading %s...\n", histPath)
	cmds, err := history.ParseBash(histPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error reading history: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("found %d commands\n", len(cmds))

	database, err := db.Open(dbPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error opening db: %v\n", err)
		os.Exit(1)
	}
	defer database.Close()

	for _, cmd := range cmds {
		tokens := index.TokenizeToString(cmd)
		if err := database.Upsert(cmd, tokens); err != nil {
			fmt.Fprintf(os.Stderr, "warn: %v\n", err)
		}
	}

    fmt.Printf("indexed %d commands into %s\n", len(cmds), dbPath)
    fmt.Println("done. run `quest <query>` to search.")
}

func runAdd(dbPath, cmd string) {
	database, err := db.Open(dbPath)
	if err != nil {
		os.Exit(1)
	}
	defer database.Close()

	tokens := index.TokenizeToString(cmd)
	_ = database.Upsert(cmd, tokens)
}

func runSearch (dbPath, query string) {
	database, err := db.Open(dbPath)
	if err != nil {
        fmt.Fprintf(os.Stderr, "error opening db: %v\n", err)
        os.Exit(1)
	}
	defer database.Close()

	cmds, err := database.All()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error loading commands: %v\n", err)
		os.Exit(1)
	}

	bm25 := index.Build(cmds)
	results := bm25.Search(query, 10)

	if len(results) == 0 {
        fmt.Fprintln(os.Stderr, "no results found")
        os.Exit(1)
	}

	fmt.Println(results[0].Command)
}

func runUI (dbPath string) {
    database, err := db.Open(dbPath)
    if err != nil {
        fmt.Fprintf(os.Stderr, "error opening db: %v\n", err)
        os.Exit(1)
    }
    defer database.Close()

    cmds, err := database.All()
    if err != nil {
        fmt.Fprintf(os.Stderr, "error loading commands: %v\n", err)
        os.Exit(1)
    }

    bm25 := index.Build(cmds)
    m := tui.New(bm25)

	p := tea.NewProgram(m, tea.WithAltScreen())
	finalModel, err := p.Run()

    if err != nil {
        fmt.Fprintf(os.Stderr, "error running TUI: %v\n", err)
        os.Exit(1)
    }

	chosen := finalModel.(tui.Model).Chosen()
	if chosen != "" {
		fmt.Println(chosen)
	}
}

func printUsage() {
    fmt.Println(`quest — fuzzy shell history search

usage:
  quest sync         index your bash history
  quest ui           open interactive TUI
  quest <query>      search and print top result

shell setup (add to ~/.bashrc):
  q() {
      cmd=$(quest "$@")
      if [ -n "$cmd" ]; then
          echo "running: $cmd"
          history -s "$cmd"
          eval "$cmd"
      fi
  }`)
}
