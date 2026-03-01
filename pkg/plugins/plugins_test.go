package plugins

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_checkRemotePluginsConfiguration(t *testing.T) {
	testCases := []struct {
		name    string
		plugins map[string]Descriptor
		wantErr bool
	}{
		{
			name:    "nil plugins configuration returns no error",
			plugins: nil,
			wantErr: false,
		},
		{
			name: "malformed module name returns error",
			plugins: map[string]Descriptor{
				"plugin1": {ModuleName: "invalid/module/name", Version: "v1.0.0"},
			},
			wantErr: true,
		},
		{
			name: "malformed module name with path traversal returns error",
			plugins: map[string]Descriptor{
				"plugin1": {ModuleName: "github.com/module/../name", Version: "v1.0.0"},
			},
			wantErr: true,
		},
		{
			name: "malformed module name with encoded path traversal returns error",
			plugins: map[string]Descriptor{
				"plugin1": {ModuleName: "github.com/module%2F%2E%2E%2Fname", Version: "v1.0.0"},
			},
			wantErr: true,
		},
		{
			name: "malformed module name returns error",
			plugins: map[string]Descriptor{
				"plugin1": {ModuleName: "invalid/module/name", Version: "v1.0.0"},
			},
			wantErr: true,
		},
		{
			name: "missing plugin version returns error",
			plugins: map[string]Descriptor{
				"plugin1": {ModuleName: "github.com/module/name", Version: ""},
			},
			wantErr: true,
		},
		{
			name: "duplicate plugin module name returns error",
			plugins: map[string]Descriptor{
				"plugin1": {ModuleName: "github.com/module/name", Version: "v1.0.0"},
				"plugin2": {ModuleName: "github.com/module/name", Version: "v1.1.0"},
			},
			wantErr: true,
		},
		{
			name: "valid plugins configuration returns no error",
			plugins: map[string]Descriptor{
				"plugin1": {ModuleName: "github.com/module/name1", Version: "v1.0.0"},
				"plugin2": {ModuleName: "github.com/module/name2", Version: "v1.1.0"},
			},
			wantErr: false,
		},
	}

	for _, test := range testCases {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			err := checkRemotePluginsConfiguration(test.plugins)
			if test.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
