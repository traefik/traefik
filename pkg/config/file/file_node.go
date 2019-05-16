package file

import (
	"fmt"
	"io/ioutil"
	"path/filepath"

	"github.com/BurntSushi/toml"
	"github.com/containous/traefik/pkg/config/parser"
	"gopkg.in/yaml.v2"
)

func decodeFileToNode(filePath string, excludes ...string) (*parser.Node, error) {
	content, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	data := make(map[string]interface{})

	switch filepath.Ext(filePath) {
	case ".toml":
		err = toml.Unmarshal(content, &data)
		if err != nil {
			return nil, err
		}

	case ".yml", ".yaml":
		var err error
		err = yaml.Unmarshal(content, data)
		if err != nil {
			return nil, err
		}

		return decodeRawToNode(data, excludes...)

	default:
		return nil, fmt.Errorf("unsupported file extension: %s", filePath)
	}

	return decodeRawToNode(data, excludes...)
}
