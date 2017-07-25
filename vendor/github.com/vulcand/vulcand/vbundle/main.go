package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	log "github.com/Sirupsen/logrus"
	"github.com/codegangsta/cli"
)

func main() {
	app := cli.NewApp()
	app.Name = "vulcanbundle"
	app.Usage = "Command line interface to compile plugins into vulcan binary"
	app.Commands = []cli.Command{
		{
			Name:   "init",
			Usage:  "Init bundle",
			Action: initBundle,
			Flags: []cli.Flag{
				cli.StringSliceFlag{
					Name:  "middleware, m",
					Value: &cli.StringSlice{},
					Usage: "Path to repo and revision, e.g. github.com/vulcand/vulcand-plugins/auth",
				},
			},
		},
	}
	err := app.Run(os.Args)
	if err != nil {
		log.Errorf("Error: %s\n", err)
	}
}

func initBundle(c *cli.Context) {
	b, err := NewBundler(c.StringSlice("middleware"))
	if err != nil {
		log.Errorf("Failed to bundle middlewares: %s", err)
		return
	}
	if err := b.bundle(); err != nil {
		log.Errorf("Failed to bundle middlewares: %s", err)
	} else {
		log.Infof("SUCCESS: bundle vulcand and vctl completed")
	}
}

type Bundler struct {
	bundleDir   string
	middlewares []string
}

func NewBundler(middlewares []string) (*Bundler, error) {
	return &Bundler{middlewares: middlewares}, nil
}

func (b *Bundler) bundle() error {
	if err := b.writeTemplates(); err != nil {
		return err
	}
	return nil
}

func (b *Bundler) writeTemplates() error {
	vulcandPath := "."
	packagePath, err := getPackagePath(vulcandPath)
	if err != nil {
		return err
	}

	context := struct {
		Packages    []Package
		PackagePath string
	}{
		Packages:    appendPackages(builtinPackages(), b.middlewares),
		PackagePath: packagePath,
	}

	if err := writeTemplate(
		filepath.Join(vulcandPath, "main.go"), mainTemplate, context); err != nil {
		return err
	}
	if err := writeTemplate(
		filepath.Join(vulcandPath, "registry", "registry.go"), registryTemplate, context); err != nil {
		return err
	}

	if err := writeTemplate(
		filepath.Join(vulcandPath, "vctl", "main.go"), vulcanctlTemplate, context); err != nil {
		return err
	}
	return nil
}

type Package string

func (p Package) Name() string {
	values := strings.Split(string(p), "/")
	return values[len(values)-1]
}

func builtinPackages() []Package {
	return []Package{
		"github.com/vulcand/vulcand/plugin/connlimit",
		"github.com/vulcand/vulcand/plugin/ratelimit",
		"github.com/vulcand/vulcand/plugin/rewrite",
		"github.com/vulcand/vulcand/plugin/cbreaker",
		"github.com/vulcand/vulcand/plugin/trace",
	}
}

func getPackagePath(dir string) (string, error) {
	path, err := filepath.Abs(dir)
	if err != nil {
		return "", err
	}
	out := strings.Split(path, "src/")
	if len(out) != 2 {
		return "", fmt.Errorf("failed to locate package path (missing top level src folder)")
	}
	return out[1], nil
}

func appendPackages(in []Package, a []string) []Package {
	for _, p := range a {
		in = append(in, Package(p))
	}
	return in
}

func writeTemplate(filename, contents string, data interface{}) error {
	if err := os.MkdirAll(filepath.Dir(filename), 0755); err != nil {
		return err
	}

	t, err := template.New(filename).Parse(contents)
	if err != nil {
		return err
	}
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()
	return t.Execute(file, data)
}
