package config

import (
	"io/ioutil"
	"path/filepath"
	"strings"
	"testing"
)

func TestMergeOnValidFixtures(t *testing.T) {
	files, err := ioutil.ReadDir("testdata/")
	if err != nil {
		t.Fatal(err)
	}
	for _, file := range files {
		if file.IsDir() || !strings.HasSuffix(file.Name(), ".yml") {
			continue
		}
		data, err := ioutil.ReadFile(filepath.Join("testdata", file.Name()))
		if err != nil {
			t.Fatalf("error reading %q: %v", file.Name(), err)
		}
		_, _, _, _, err = Merge(NewServiceConfigs(), nil, nil, file.Name(), data, nil)
		if err != nil {
			t.Errorf("error loading %q: %v\n %v", file.Name(), string(data), err)
		}
	}
}
