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

	outputPath := path.Join(os.TempDir(), fmt.Sprintf(".goreleaser_%s.yml", goos))

	output, err := os.Create(outputPath)
	if err != nil {
		log.Fatal(err)
	}

	err = tmpl.Execute(output, map[string]string{"GOOS": goos})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Print(outputPath)
}
