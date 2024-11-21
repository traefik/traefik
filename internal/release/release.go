package main

import (
	"fmt"
	"log"
	"os"
	"path"
	"strings"
	"text/template"
)

func main() {
	if len(os.Args) < 2 {
		log.Fatal("GOOS should be provided as a CLI argument")
	}

	goos := strings.TrimSpace(os.Args[1])
	if goos == "" {
		log.Fatal("GOOS should be provided as a CLI argument")
	}

	tmpl := template.Must(
		template.New(".goreleaser.yml.tmpl").
			Delims("[[", "]]").
			ParseFiles("./.goreleaser.yml.tmpl"),
	)

	goarch := ""
	outputFileName := fmt.Sprintf(".goreleaser_%s.yml", goos)
	if strings.Contains(goos, "-") {
		split := strings.Split(goos, "-")
		goos = split[0]
		goarch = split[1]
	}

	outputPath := path.Join(os.TempDir(), outputFileName)

	output, err := os.Create(outputPath)
	if err != nil {
		log.Fatal(err)
	}

	err = tmpl.Execute(output, map[string]string{"GOOS": goos, "GOARCH": goarch})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Print(outputPath)
}
