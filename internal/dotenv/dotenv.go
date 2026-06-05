package dotenv

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// Upsert writes key=value pairs into file, updating existing keys in place
// and appending new ones. Existing lines that don't match are preserved.
func Upsert(file string, pairs map[string]string) error {
	lines := readLines(file)

	updated := make(map[string]bool, len(pairs))
	result := make([]string, 0, len(lines)+len(pairs))

	for _, line := range lines {
		key, _, found := strings.Cut(line, "=")
		key = strings.TrimSpace(key)
		if found && pairs[key] != "" {
			result = append(result, key+"="+pairs[key])
			updated[key] = true
		} else {
			result = append(result, line)
		}
	}

	for k, v := range pairs {
		if !updated[k] {
			result = append(result, k+"="+v)
		}
	}

	f, err := os.Create(file)
	if err != nil {
		return fmt.Errorf("writing %s: %w", file, err)
	}
	defer f.Close()

	for _, line := range result {
		fmt.Fprintln(f, line)
	}
	return nil
}

// Load reads key=value pairs from file without modifying the process environment.
// Lines starting with # and empty lines are skipped.
func Load(file string) map[string]string {
	out := map[string]string{}
	for _, line := range readLines(file) {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		k, v, found := strings.Cut(line, "=")
		if found {
			out[strings.TrimSpace(k)] = strings.TrimSpace(v)
		}
	}
	return out
}

func readLines(file string) []string {
	f, err := os.Open(file)
	if err != nil {
		return nil
	}
	defer f.Close()

	var lines []string
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	return lines
}
