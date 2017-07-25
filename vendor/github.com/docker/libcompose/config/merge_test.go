package config

import (
	"io/ioutil"
	"testing"
)

type NullLookup struct {
}

func (n *NullLookup) Lookup(file, relativeTo string) ([]byte, string, error) {
	return nil, "", nil
}

func (n *NullLookup) ResolvePath(path, inFile string) string {
	return ""
}

type FileLookup struct {
}

func (f *FileLookup) Lookup(file, relativeTo string) ([]byte, string, error) {
	bytes, err := ioutil.ReadFile(file)
	return bytes, file, err
}

func (f *FileLookup) ResolvePath(path, inFile string) string {
	return ""
}

func TestExtendsInheritImage(t *testing.T) {
	_, configV1, _, _, err := Merge(NewServiceConfigs(), nil, &NullLookup{}, "", []byte(`
parent:
  image: foo
child:
  extends:
    service: parent
`), nil)
	if err != nil {
		t.Fatal(err)
	}

	_, configV2, _, _, err := Merge(NewServiceConfigs(), nil, &NullLookup{}, "", []byte(`
version: '2'
services:
  parent:
    image: foo
  child:
    extends:
      service: parent
`), nil)
	if err != nil {
		t.Fatal(err)
	}

	for _, config := range []map[string]*ServiceConfig{configV1, configV2} {
		parent := config["parent"]
		child := config["child"]

		if parent.Image != "foo" {
			t.Fatal("Invalid parent image", parent.Image)
		}

		t.Logf("%#v", config["child"])
		if child.Build.Context != "" {
			t.Fatalf("Invalid build %#v", child.Build)
		}

		if child.Image != "foo" {
			t.Fatal("Invalid child image", child.Image)
		}
	}
}

func TestExtendsInheritBuild(t *testing.T) {
	_, configV1, _, _, err := Merge(NewServiceConfigs(), nil, &NullLookup{}, "", []byte(`
parent:
  build: .
child:
  extends:
    service: parent
`), nil)
	if err != nil {
		t.Fatal(err)
	}

	_, configV2, _, _, err := Merge(NewServiceConfigs(), nil, &NullLookup{}, "", []byte(`
version: '2'
services:
  parent:
    build:
      context: .
  child:
    extends:
      service: parent
`), nil)
	if err != nil {
		t.Fatal(err)
	}

	for _, config := range []map[string]*ServiceConfig{configV1, configV2} {
		parent := config["parent"]
		child := config["child"]

		if parent.Build.Context != "." {
			t.Fatal("Invalid build", parent.Build)
		}

		if child.Build.Context != "." {
			t.Fatal("Invalid build", child.Build)
		}

		if child.Image != "" {
			t.Fatal("Invalid image", child.Image)
		}
	}
}

func TestExtendBuildOverImageV1(t *testing.T) {
	_, config, _, _, err := Merge(NewServiceConfigs(), nil, &NullLookup{}, "", []byte(`
parent:
  image: foo
child:
  build: .
  extends:
    service: parent
`), nil)
	if err != nil {
		t.Fatal(err)
	}

	parent := config["parent"]
	child := config["child"]

	if parent.Image != "foo" {
		t.Fatal("Invalid image", parent.Image)
	}

	if child.Build.Context != "." {
		t.Fatal("Invalid build", child.Build)
	}

	if child.Image != "" {
		t.Fatal("Invalid image", child.Image)
	}
}

func TestExtendBuildOverImageV2(t *testing.T) {
	_, config, _, _, err := Merge(NewServiceConfigs(), nil, &NullLookup{}, "", []byte(`
version: '2'
services:
  parent:
    image: foo
  child:
    build:
      context: .
    extends:
      service: parent
`), nil)
	if err != nil {
		t.Fatal(err)
	}

	parent := config["parent"]
	child := config["child"]

	if parent.Image != "foo" {
		t.Fatal("Invalid image", parent.Image)
	}

	if child.Build.Context != "." {
		t.Fatal("Invalid build", child.Build)
	}

	if child.Image != "foo" {
		t.Fatal("Invalid image", child.Image)
	}
}

func TestExtendImageOverBuildV1(t *testing.T) {
	_, config, _, _, err := Merge(NewServiceConfigs(), nil, &NullLookup{}, "", []byte(`
parent:
  build: .
child:
  image: foo
  extends:
    service: parent
`), nil)
	if err != nil {
		t.Fatal(err)
	}

	parent := config["parent"]
	child := config["child"]

	if parent.Image != "" {
		t.Fatal("Invalid image", parent.Image)
	}

	if parent.Build.Context != "." {
		t.Fatal("Invalid build", parent.Build)
	}

	if child.Build.Context != "" {
		t.Fatal("Invalid build", child.Build)
	}

	if child.Image != "foo" {
		t.Fatal("Invalid image", child.Image)
	}

}

func TestExtendImageOverBuildV2(t *testing.T) {
	_, config, _, _, err := Merge(NewServiceConfigs(), nil, &NullLookup{}, "", []byte(`
version: '2'
services:
  parent:
    build:
      context: .
  child:
    image: foo
    extends:
      service: parent
`), nil)
	if err != nil {
		t.Fatal(err)
	}

	parent := config["parent"]
	child := config["child"]

	if parent.Image != "" {
		t.Fatal("Invalid image", parent.Image)
	}

	if parent.Build.Context != "." {
		t.Fatal("Invalid build", parent.Build)
	}

	if child.Build.Context != "." {
		t.Fatal("Invalid build", child.Build)
	}

	if child.Image != "foo" {
		t.Fatal("Invalid image", child.Image)
	}
}

func TestMergesEnvFile(t *testing.T) {
	_, configV1, _, _, err := Merge(NewServiceConfigs(), nil, &FileLookup{}, "", []byte(`
test:
  image: foo
  env_file:
    - testdata/.env
`), nil)
	if err != nil {
		t.Fatal(err)
	}

	_, configV2, _, _, err := Merge(NewServiceConfigs(), nil, &FileLookup{}, "", []byte(`
version: '2'
services:
  test:
    image: foo
    env_file:
      - testdata/.env
`), nil)
	if err != nil {
		t.Fatal(err)
	}

	for _, config := range []map[string]*ServiceConfig{configV1, configV2} {
		test := config["test"]

		if len(test.Environment) != 2 {
			t.Fatal("env_file is not merged", test.Environment)
		}

		for _, environment := range test.Environment {
			if (environment != "FOO=foo") && (environment != "BAR=bar") {
				t.Fatal("Empty line and comment should be excluded", environment)
			}
		}
	}
}

func TestRestartNo(t *testing.T) {
	_, configV1, _, _, err := Merge(NewServiceConfigs(), nil, &NullLookup{}, "", []byte(`
test:
  restart: "no"
  image: foo
`), nil)
	if err != nil {
		t.Fatal(err)
	}

	_, configV2, _, _, err := Merge(NewServiceConfigs(), nil, &NullLookup{}, "", []byte(`
version: '2'
services:
  test:
    restart: "no"
    image: foo
`), nil)
	if err != nil {
		t.Fatal(err)
	}

	for _, config := range []map[string]*ServiceConfig{configV1, configV2} {
		test := config["test"]

		if test.Restart != "no" {
			t.Fatal("Invalid restart policy", test.Restart)
		}
	}
}

func TestRestartAlways(t *testing.T) {
	_, configV1, _, _, err := Merge(NewServiceConfigs(), nil, &NullLookup{}, "", []byte(`
test:
  restart: always
  image: foo
`), nil)
	if err != nil {
		t.Fatal(err)
	}

	_, configV2, _, _, err := Merge(NewServiceConfigs(), nil, &NullLookup{}, "", []byte(`
version: '2'
services:
  test:
    restart: always
    image: foo
`), nil)
	if err != nil {
		t.Fatal(err)
	}

	for _, config := range []map[string]*ServiceConfig{configV1, configV2} {
		test := config["test"]

		if test.Restart != "always" {
			t.Fatal("Invalid restart policy", test.Restart)
		}
	}
}

func TestIsValidRemote(t *testing.T) {
	gitUrls := []string{
		"git://github.com/docker/docker",
		"git@github.com:docker/docker.git",
		"git@bitbucket.org:atlassianlabs/atlassian-docker.git",
		"https://github.com/docker/docker.git",
		"http://github.com/docker/docker.git",
		"http://github.com/docker/docker.git#branch",
		"http://github.com/docker/docker.git#:dir",
	}
	incompleteGitUrls := []string{
		"github.com/docker/docker",
	}
	invalidGitUrls := []string{
		"http://github.com/docker/docker.git:#branch",
	}
	for _, url := range gitUrls {
		if !IsValidRemote(url) {
			t.Fatalf("%q should have been a valid remote", url)
		}
	}
	for _, url := range incompleteGitUrls {
		if !IsValidRemote(url) {
			t.Fatalf("%q should have been a valid remote", url)
		}
	}
	for _, url := range invalidGitUrls {
		if !IsValidRemote(url) {
			t.Fatalf("%q should have been a valid remote", url)
		}
	}
}

func preprocess(services RawServiceMap) (RawServiceMap, error) {
	for name := range services {
		services[name]["image"] = "foo2"
	}
	return services, nil
}

func postprocess(services map[string]*ServiceConfig) (map[string]*ServiceConfig, error) {
	for name := range services {
		services[name].ContainerName = "cname"
	}
	return services, nil
}

func TestParseOptions(t *testing.T) {
	parseOptions := ParseOptions{
		Interpolate: false,
		Validate:    false,
		Preprocess:  preprocess,
		Postprocess: postprocess,
	}

	_, configV1, _, _, err := Merge(NewServiceConfigs(), nil, &NullLookup{}, "", []byte(`
test:
  image: foo
  labels:
    x: $X
test2:
  invalid_key: true
`), &parseOptions)
	if err != nil {
		t.Fatal(err)
	}

	_, configV2, _, _, err := Merge(NewServiceConfigs(), nil, &NullLookup{}, "", []byte(`
version: '2'
services:
  test:
    image: foo
    labels:
      x: $X
  test2:
    invalid_key: true
`), &parseOptions)
	if err != nil {
		t.Fatal(err)
	}

	for _, config := range []map[string]*ServiceConfig{configV1, configV2} {
		test := config["test"]

		if test.Image != "foo2" {
			t.Fatal("Preprocess failed to change image", test.Image)
		}
		if test.ContainerName != "cname" {
			t.Fatal("Postprocess failed to change container name", test.ContainerName)
		}
		if test.Labels["x"] != "$X" {
			t.Fatal("Failed to disable interpolation")
		}
	}
}
