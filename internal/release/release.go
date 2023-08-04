package main

import (
	"fmt"
	"log"
	"os"
	"path"
	"text/template"
)

func main() {
	t := template.New(".goreleaser.yml.tmpl").Delims("[[", "]]")
	tmpl := template.Must(t.ParseFiles("./.goreleaser.yml.tmpl"))

	goos := os.Args[1]

	f := path.Join(os.TempDir(), fmt.Sprintf(".goreleaser_%s.yml", goos))
	file, err := os.Create(f)
	if err != nil {
		log.Fatal(err)
	}

	err = tmpl.Execute(file, map[string]string{"GOOS": goos})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Print(f)
}
