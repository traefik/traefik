package plugins

const (
	runtimeYaegi = "yaegi"
	runtimeWasm  = "wasm"
)

const (
	typeMiddleware = "middleware"
	typeProvider   = "provider"
)

// Descriptor The static part of a plugin configuration.
type Descriptor struct {
	// ModuleName (required)
	ModuleName string `description:"plugin's module name." json:"moduleName,omitempty" toml:"moduleName,omitempty" yaml:"moduleName,omitempty" export:"true"`

	// Version (required)
	Version string `description:"plugin's version." json:"version,omitempty" toml:"version,omitempty" yaml:"version,omitempty" export:"true"`
}

// LocalDescriptor The static part of a local plugin configuration.
type LocalDescriptor struct {
	// ModuleName (required)
	ModuleName string `description:"plugin's module name." json:"moduleName,omitempty" toml:"moduleName,omitempty" yaml:"moduleName,omitempty" export:"true"`
}

// Manifest The plugin manifest.
type Manifest struct {
	DisplayName   string                 `yaml:"displayName"`
	Type          string                 `yaml:"type"`
	Runtime       string                 `yaml:"runtime"`
	WasmPath      string                 `yaml:"wasmPath"`
	Import        string                 `yaml:"import"`
	BasePkg       string                 `yaml:"basePkg"`
	Compatibility string                 `yaml:"compatibility"`
	Summary       string                 `yaml:"summary"`
	TestData      map[string]interface{} `yaml:"testData"`
}

// IsYaegiPlugin returns true if the plugin is a Yaegi plugin.
func (m *Manifest) IsYaegiPlugin() bool {
	// defaults always Yaegi to have backwards compatibility to plugins without runtime
	return m.Runtime == runtimeYaegi || m.Runtime == ""
}
