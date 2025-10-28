package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

var (
	// detect any existing <a ...> tag in the cell (case-insensitive).
	reExistingAnchor = regexp.MustCompile(`(?i)<\s*a\b[^>]*>.*?</\s*a\s*>`)
	// separator cell like --- or :---: (3+ dashes, optional leading/trailing colon).
	reSepCell = regexp.MustCompile(`^\s*:?-{3,}:?\s*$`)
	// markdown link [text](url) → text (used to strip link wrappers in id).
	reMdLink = regexp.MustCompile(`\[(.*?)\]\((.*?)\)`)
	// collapse multiple hyphens.
	reMultiHyphens = regexp.MustCompile(`-+`)
)

// splitTableRow splits a markdown table line on pipes, while keeping escaped pipes.
// parts[1] will be the first data cell for lines that start with '|'.
func splitTableRow(line string) []string {
	var parts []string
	var b strings.Builder
	escaped := false
	for _, r := range line {
		if escaped {
			b.WriteRune(r)
			escaped = false
			continue
		}
		if r == '\\' {
			escaped = true
			b.WriteRune(r)
			continue
		}
		if r == '|' {
			parts = append(parts, b.String())
			b.Reset()
			continue
		}
		b.WriteRune(r)
	}
	parts = append(parts, b.String())
	return parts
}

func isTableRow(line string) bool {
	s := strings.TrimSpace(line)
	if !strings.HasPrefix(s, "|") {
		return false
	}
	parts := splitTableRow(line)
	return len(parts) >= 3
}

func isSeparatorRow(line string) bool {
	if !isTableRow(line) {
		return false
	}
	parts := splitTableRow(line)
	// check all middle cells (skip first and last which are outside pipes)
	for i := 1; i < len(parts)-1; i++ {
		cell := strings.TrimSpace(parts[i])
		if cell == "" {
			continue
		}
		if !reSepCell.MatchString(cell) {
			return false
		}
	}
	return true
}

// Create ID from cell text, preserving letter case, removing <br /> and markdown decorations.
func makeID(text string) string {
	id := strings.TrimSpace(text)

	// remove BR tags (common in table cells)
	id = strings.ReplaceAll(id, "<br />", " ")
	id = strings.ReplaceAll(id, "<br/>", " ")
	id = strings.ReplaceAll(id, "<br>", " ")

	// remove the dots
	id = strings.ReplaceAll(id, ".", "-")

	// strip markdown link wrappers [text](url) -> text
	id = reMdLink.ReplaceAllString(id, "$1")

	// remove inline markdown characters
	id = strings.ReplaceAll(id, "`", "")
	id = strings.ReplaceAll(id, "*", "")
	id = strings.ReplaceAll(id, "~", "")

	// replace spaces/underscores with hyphen
	id = strings.ReplaceAll(id, " ", "-")
	id = strings.ReplaceAll(id, "_", "-")

	// keep only letters (both cases), digits and hyphens
	var clean []rune
	for _, r := range id {
		if (r >= 'a' && r <= 'z') ||
			(r >= 'A' && r <= 'Z') ||
			(r >= '0' && r <= '9') ||
			r == '-' || r == '.' {
			// keep dot as you requested (we won't replace it)
			clean = append(clean, r)
		}
	}
	id = string(clean)

	// collapse multiple hyphens and trim
	id = reMultiHyphens.ReplaceAllString(id, "-")
	id = strings.Trim(id, "-")
	if id == "" {
		id = "row"
	}
	return id
}

// Dedupe ID within a file: if id already seen, append -2, -3...
// Use "opt-" prefix to avoid conflicts with section headings.
func dedupeID(base string, seen map[string]int) string {
	if base == "" {
		base = "row"
	}

	// Add prefix to avoid conflicts with section headings.
	optID := "opt-" + base

	count, ok := seen[optID]
	if !ok {
		seen[optID] = 1
		return optID
	}

	seen[optID] = count + 1
	return fmt.Sprintf("%s-%d", optID, count+1)
}

// Clean existing anchors from cell content.
func cleanExistingAnchors(text string) string {
	return reExistingAnchor.ReplaceAllStringFunc(text, func(match string) string {
		// Extract content between <a> tags.
		start := strings.Index(match, ">")
		end := strings.LastIndex(match, "</")
		if start >= 0 && end > start {
			return strings.TrimSpace(match[start+1 : end])
		}
		return strings.TrimSpace(match)
	})
}

// Inject clickable link that is also the target (id + href on same element).
func injectClickableFirstCell(line string, seen map[string]int) string {
	parts := splitTableRow(line)
	// first data cell is index 1
	firstCellRaw := parts[1]

	// Clean any existing anchors first.
	firstCellRaw = cleanExistingAnchors(firstCellRaw)
	firstTrimmed := strings.TrimSpace(firstCellRaw)

	id := makeID(firstTrimmed)
	if id == "" {
		return strings.Join(parts, "|")
	}
	id = dedupeID(id, seen)

	// wrap the visible cell content in a link that is also the target
	// keep the cell inner HTML/text (firstCellRaw) as-is inside the <a>
	parts[1] = fmt.Sprintf(" <a id=\"%s\" href=\"#%s\" title=\"#%s\">%s</a> ", id, id, id, strings.TrimSpace(firstCellRaw))
	return strings.Join(parts, "|")
}

func processFile(path string) error {
	// read file
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	var lines []string
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		lines = append(lines, sc.Text())
	}
	if err := sc.Err(); err != nil {
		_ = f.Close()
		return err
	}
	_ = f.Close()

	inFence := false
	seen := make(map[string]int)
	out := make([]string, len(lines))

	for i, line := range lines {
		trim := strings.TrimSpace(line)

		// toggle code fence (``` or ~~~)
		if strings.HasPrefix(trim, "```") || strings.HasPrefix(trim, "~~~") {
			inFence = !inFence
			out[i] = line
			continue
		}
		if inFence {
			out[i] = line
			continue
		}

		// not a table row -> copy as-is
		if !isTableRow(line) {
			out[i] = line
			continue
		}

		// separator row -> copy as-is
		if isSeparatorRow(line) {
			out[i] = line
			continue
		}

		// detect header row (the row immediately before a separator) and skip it
		isHeader := false
		for j := i + 1; j < len(lines); j++ {
			if strings.TrimSpace(lines[j]) == "" {
				continue
			}
			if isSeparatorRow(lines[j]) {
				isHeader = true
			}
			break
		}
		if isHeader {
			out[i] = line
			continue
		}

		// otherwise inject clickable link in first cell
		out[i] = injectClickableFirstCell(line, seen)
	}

	// overwrite file in place
	wf, err := os.Create(path)
	if err != nil {
		return err
	}
	bw := bufio.NewWriter(wf)
	for _, l := range out {
		fmt.Fprintln(bw, l)
	}
	if err := bw.Flush(); err != nil {
		_ = wf.Close()
		return err
	}
	return wf.Close()
}

func genAnchors() {
	root := "./docs/content/reference/"
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && strings.HasSuffix(strings.ToLower(info.Name()), ".md") {
			if perr := processFile(path); perr != nil {
				fmt.Printf("⚠️ Error processing %s: %v\n", path, perr)
			} else {
				fmt.Printf("✅ Processed %s\n", path)
			}
		}
		return nil
	})
	if err != nil {
		log.Fatalf("walk error: %v", err)
	}
}
