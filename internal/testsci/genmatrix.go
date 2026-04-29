// genmatrix lists Go test function names under the configured roots, optionally
// reads timing data from a JUnit XML, then bin-packs the tests into N groups
// (greedy first-fit-decreasing) and prints the resulting matrix as JSON for
// GitHub Actions to consume via fromJson(). Listing is done by parsing the
// `_test.go` files with go/ast — no compilation, so it runs in a couple of
// seconds even on a cold cache.
package main

import (
	"encoding/json"
	"encoding/xml"
	"flag"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

const defaultDuration = 1.0

type group struct {
	ID    int    `json:"id"`
	Regex string `json:"regex"`
}

type junitTestCase struct {
	Name string  `xml:"name,attr"`
	Time float64 `xml:"time,attr"`
}

type junitSuite struct {
	TestCases []junitTestCase `xml:"testcase"`
}

type junitSummary struct {
	XMLName xml.Name     `xml:"testsuites"`
	Suites  []junitSuite `xml:"testsuite"`
}

func main() {
	var (
		groupCount = flag.Int("groups", 6, "number of shards")
		junitPath  = flag.String("junit", "", "JUnit XML with previous-run timings")
		rootsArg   = flag.String("roots", "./cmd,./pkg", "comma-separated roots to walk for `_test.go` files")
	)
	flag.Parse()

	timings := loadTimings(*junitPath)
	tests := listTestFunctions(strings.Split(*rootsArg, ","))

	matrix := binPack(tests, timings, *groupCount)

	out, err := json.Marshal(matrix)
	if err != nil {
		fmt.Fprintf(os.Stderr, "marshal matrix: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("matrix=%s\n", string(out))
}

// loadTimings parses a JUnit XML produced by gotestsum and returns a map of
// top-level test name → cumulative wall time across all packages it appears in.
// Sub-test entries (those containing a `/` in their name) are skipped because
// their time is already rolled into the parent entry.
func loadTimings(path string) map[string]float64 {
	if path == "" {
		return nil
	}
	data, err := os.ReadFile(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "warning: cannot read %s: %v — falling back to count-based split\n", path, err)
		return nil
	}
	var summary junitSummary
	if err := xml.Unmarshal(data, &summary); err != nil {
		fmt.Fprintf(os.Stderr, "warning: cannot parse %s: %v — falling back to count-based split\n", path, err)
		return nil
	}
	timings := map[string]float64{}
	for _, suite := range summary.Suites {
		for _, tc := range suite.TestCases {
			if strings.Contains(tc.Name, "/") {
				continue
			}
			// Skip zero-time entries: they typically come from a canceled or
			// skipped previous run and would mislead the bin-packer into
			// treating those tests as instantaneous.
			if tc.Time == 0 {
				continue
			}
			timings[tc.Name] += tc.Time
		}
	}
	return timings
}

// listTestFunctions walks the given roots and returns the unique set of
// top-level `Test*` function names declared in `*_test.go` files. TestMain is
// excluded because Go runs it implicitly.
func listTestFunctions(roots []string) []string {
	seen := map[string]struct{}{}
	for _, root := range roots {
		root = strings.TrimSpace(root)
		if root == "" {
			continue
		}
		_ = filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				// Skip unreadable entries but continue the walk.
				return nil //nolint:nilerr // intentional: we want partial enumeration on transient I/O errors.
			}
			if d.IsDir() || !strings.HasSuffix(path, "_test.go") {
				return nil
			}
			f, perr := parser.ParseFile(token.NewFileSet(), path, nil, 0)
			if perr != nil {
				// Skip files that don't parse (e.g., build-tag-gated stubs); the
				// rest of the tree still needs to be enumerated.
				return nil //nolint:nilerr // intentional: a single unparsable file shouldn't kill the listing.
			}
			for _, decl := range f.Decls {
				fn, ok := decl.(*ast.FuncDecl)
				if !ok || fn.Recv != nil {
					continue
				}
				name := fn.Name.Name
				if !strings.HasPrefix(name, "Test") || name == "TestMain" {
					continue
				}
				seen[name] = struct{}{}
			}
			return nil
		})
	}
	tests := make([]string, 0, len(seen))
	for name := range seen {
		tests = append(tests, name)
	}
	sort.Strings(tests)
	return tests
}

// binPack distributes tests into groupCount bins using greedy first-fit
// decreasing on the per-test timing. Tests without timing data get the
// average timing of those with data (or `defaultDuration` if none).
func binPack(tests []string, timings map[string]float64, groupCount int) []group {
	avg := defaultDuration
	if len(timings) > 0 {
		var sum float64
		var n int
		for _, t := range tests {
			if v, ok := timings[t]; ok {
				sum += v
				n++
			}
		}
		if n > 0 {
			avg = sum / float64(n)
		}
	}

	type weighted struct {
		name string
		time float64
	}
	w := make([]weighted, len(tests))
	for i, name := range tests {
		t, ok := timings[name]
		if !ok {
			t = avg
		}
		w[i] = weighted{name: name, time: t}
	}
	sort.Slice(w, func(i, j int) bool {
		if w[i].time != w[j].time {
			return w[i].time > w[j].time
		}
		return w[i].name < w[j].name
	})

	bins := make([][]string, groupCount)
	totals := make([]float64, groupCount)
	for _, t := range w {
		minIdx := 0
		for i := 1; i < groupCount; i++ {
			if totals[i] < totals[minIdx] {
				minIdx = i
			}
		}
		bins[minIdx] = append(bins[minIdx], t.name)
		totals[minIdx] += t.time
	}

	matrix := make([]group, 0, groupCount)
	for i, b := range bins {
		if len(b) == 0 {
			continue
		}
		sort.Strings(b)
		matrix = append(matrix, group{
			ID:    i,
			Regex: "^(?:" + strings.Join(b, "|") + ")$",
		})
		fmt.Fprintf(os.Stderr, "bin %d: %d tests, ~%.1fs estimated\n", i, len(b), totals[i])
	}
	return matrix
}
