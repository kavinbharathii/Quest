
package history

import (
	"bufio"
	"os"
	"strings"
)

func ParseBash (path string) ([]string, error) {
	f, err := os.Open(path)

	if err != nil {
		return nil, err 
	}

	defer f.Close()

	seen := make(map[string]bool)
	var cmds []string

	scanner := bufio.NewScanner(f)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		if line == "" {
			continue
		}

		if strings.HasPrefix(line, "#") {
			continue
		}

		if len(line) < 2 {
			continue
		}

		if isBoring(line) {
			continue
		}

		if !seen[line] {
			seen[line] = true
			cmds = append(cmds, line)
		}
	}

	return cmds, scanner.Err()
}


func isBoring(cmd string) bool {
	boring := []string{
        "ls", "ll", "la", "l", "pwd", "clear", "exit", "cd", "history",
        "q", "qq", "z", "zz", "bg", "fg", "jobs",
	}

	trimmed := strings.Fields(cmd)[0]

	for _, b := range boring {
		if trimmed == b {
			return true
		}
	}

	return false
}


