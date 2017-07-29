package staert

import (
	"bytes"
	"fmt"
	"github.com/containous/flaeg"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"
)

//StructPtr : Struct with pointers
type StructPtr struct {
	PtrStruct1    *Struct1      `description:"Enable Struct1"`
	PtrStruct2    *Struct2      `description:"Enable Struct1"`
	DurationField time.Duration `description:"Duration Field"`
}

//Struct1 : Struct with pointer
type Struct1 struct {
	S1Int        int      `description:"Struct 1 Int"`
	S1String     string   `description:"Struct 1 String"`
	S1Bool       bool     `description:"Struct 1 Bool"`
	S1PtrStruct3 *Struct3 `description:"Enable Struct3"`
}

//Struct2 : trivial Struct
type Struct2 struct {
	S2Int64  int64  `description:"Struct 2 Int64"`
	S2String string `description:"Struct 2 String"`
	S2Bool   bool   `description:"Struct 2 Bool"`
}

//Struct3 : trivial Struct
type Struct3 struct {
	S3Float64 float64 `description:"Struct 3 float64"`
}

func TestFleagSourceNoArgs(t *testing.T) {
	//use buffer instead of stdout
	var b bytes.Buffer
	//Init
	config := &StructPtr{
		PtrStruct1: &Struct1{
			S1Int:    1,
			S1String: "S1StringInitConfig",
		},
		DurationField: time.Second,
	}
	defaultPointersConfig := &StructPtr{
		PtrStruct1: &Struct1{
			S1Int:    11,
			S1String: "S1StringDefaultPointersConfig",
			S1Bool:   true,
			S1PtrStruct3: &Struct3{
				S3Float64: 11.11,
			},
		},
		PtrStruct2: &Struct2{
			S2Int64:  22,
			S2String: "S2StringDefaultPointersConfig",
			S2Bool:   false,
		},
	}
	args := []string{}

	//Test
	rootCmd := &flaeg.Command{
		Name:                  "test",
		Description:           "description test",
		Config:                config,
		DefaultPointersConfig: defaultPointersConfig,
		Run: func() error {
			fmt.Fprintf(&b, "Run with config :\n%+v\n", config)
			return nil
		},
	}
	s := NewStaert(rootCmd)
	fs := flaeg.New(rootCmd, args)
	s.AddSource(fs)
	if err := s.parseConfigAllSources(rootCmd); err != nil {
		t.Errorf("Error %s", err.Error())
	}

	//Check
	check := &StructPtr{
		PtrStruct1: &Struct1{
			S1Int:    1,
			S1String: "S1StringInitConfig",
		},
		DurationField: time.Second,
	}

	if !reflect.DeepEqual(rootCmd.Config, check) {
		t.Errorf("\nexpected\t: %+v\ngot\t\t\t: %+v\n", check, rootCmd.Config)
	}
}

func TestFleagSourcePtrUnderPtrArgs(t *testing.T) {
	//use buffer instead of stdout
	var b bytes.Buffer
	//Init
	config := &StructPtr{
		PtrStruct1: &Struct1{
			S1Int:    1,
			S1String: "S1StringInitConfig",
		},
		DurationField: time.Second,
	}
	defaultPointersConfig := &StructPtr{
		PtrStruct1: &Struct1{
			S1Int:    11,
			S1String: "S1StringDefaultPointersConfig",
			S1Bool:   true,
			S1PtrStruct3: &Struct3{
				S3Float64: 11.11,
			},
		},
		PtrStruct2: &Struct2{
			S2Int64:  22,
			S2String: "S2StringDefaultPointersConfig",
			S2Bool:   false,
		},
	}
	args := []string{
		"--ptrstruct1.s1ptrstruct3",
	}

	//Test
	rootCmd := &flaeg.Command{
		Name:                  "test",
		Description:           "description test",
		Config:                config,
		DefaultPointersConfig: defaultPointersConfig,
		Run: func() error {
			fmt.Fprintf(&b, "Run with config :\n%+v\n", config)
			return nil
		},
	}
	s := NewStaert(rootCmd)
	fs := flaeg.New(rootCmd, args)
	s.AddSource(fs)
	if err := s.parseConfigAllSources(rootCmd); err != nil {
		t.Errorf("Error %s", err.Error())
	}
	//Check
	check := &StructPtr{
		PtrStruct1: &Struct1{
			S1Int:    1,
			S1String: "S1StringInitConfig",
			S1PtrStruct3: &Struct3{
				S3Float64: 11.11,
			},
		},
		DurationField: time.Second,
	}

	if !reflect.DeepEqual(rootCmd.Config, check) {
		result, ok := rootCmd.Config.(*StructPtr)
		if ok {
			fmt.Printf("\nexpected\t: %+v\ngot\t\t\t: %+v\n", check.PtrStruct1, result.PtrStruct1)
		}

		t.Errorf("\nexpected\t: %+v\ngot\t\t\t: %+v\n", check, rootCmd.Config)
	}
}

func TestFleagSourceFieldUnderPtrUnderPtrArgs(t *testing.T) {
	//use buffer instead of stdout
	var b bytes.Buffer
	//Init
	config := &StructPtr{
		PtrStruct1: &Struct1{
			S1Int:    1,
			S1String: "S1StringInitConfig",
		},
		DurationField: time.Second,
	}
	defaultPointersConfig := &StructPtr{
		PtrStruct1: &Struct1{
			S1Int:    11,
			S1String: "S1StringDefaultPointersConfig",
			S1Bool:   true,
			S1PtrStruct3: &Struct3{
				S3Float64: 11.11,
			},
		},
		PtrStruct2: &Struct2{
			S2Int64:  22,
			S2String: "S2StringDefaultPointersConfig",
			S2Bool:   false,
		},
	}
	args := []string{
		"--ptrstruct1.s1ptrstruct3.s3float64=55.55",
	}

	//Test
	rootCmd := &flaeg.Command{
		Name:                  "test",
		Description:           "description test",
		Config:                config,
		DefaultPointersConfig: defaultPointersConfig,
		Run: func() error {
			fmt.Fprintf(&b, "Run with config :\n%+v\n", config)
			return nil
		},
	}
	s := NewStaert(rootCmd)
	fs := flaeg.New(rootCmd, args)
	s.AddSource(fs)
	if err := s.parseConfigAllSources(rootCmd); err != nil {
		t.Errorf("Error %s", err.Error())
	}
	//Check
	check := &StructPtr{
		PtrStruct1: &Struct1{
			S1Int:    1,
			S1String: "S1StringInitConfig",
			S1PtrStruct3: &Struct3{
				S3Float64: 55.55,
			},
		},
		DurationField: time.Second,
	}

	if !reflect.DeepEqual(rootCmd.Config, check) {
		t.Errorf("\nexpected\t: %+v\ngot\t\t\t: %+v\n", check, rootCmd.Config)
	}
}

func TestTomlSourceNothing(t *testing.T) {
	//use buffer instead of stdout
	var b bytes.Buffer
	//Init
	config := &StructPtr{
		PtrStruct1: &Struct1{
			S1Int:    1,
			S1String: "S1StringInitConfig",
		},
		DurationField: time.Second,
	}
	defaultPointersConfig := &StructPtr{
		PtrStruct1: &Struct1{
			S1Int:    11,
			S1String: "S1StringDefaultPointersConfig",
			S1Bool:   true,
			S1PtrStruct3: &Struct3{
				S3Float64: 11.11,
			},
		},
		PtrStruct2: &Struct2{
			S2Int64:  22,
			S2String: "S2StringDefaultPointersConfig",
			S2Bool:   false,
		},
	}

	//Test
	rootCmd := &flaeg.Command{
		Name:                  "test",
		Description:           "description test",
		Config:                config,
		DefaultPointersConfig: defaultPointersConfig,
		Run: func() error {
			fmt.Fprintf(&b, "Run with config :\n%+v\n", config)
			return nil
		},
	}
	s := NewStaert(rootCmd)
	toml := NewTomlSource("nothing", []string{"./toml/", "/any/other/path"})
	s.AddSource(toml)

	if err := s.parseConfigAllSources(rootCmd); err != nil {
		t.Errorf("Error %s", err.Error())
	}
	//Check
	check := &StructPtr{
		PtrStruct1: &Struct1{
			S1Int:    1,
			S1String: "S1StringInitConfig",
		},
		DurationField: time.Second,
	}

	if !reflect.DeepEqual(rootCmd.Config, check) {
		t.Errorf("\nexpected\t: %+v\ngot\t\t\t: %+v\n", check, rootCmd.Config)
	}
}

func TestTomlSourceTrivial(t *testing.T) {
	//use buffer instead of stdout
	var b bytes.Buffer
	//Init
	config := &StructPtr{
		PtrStruct1: &Struct1{
			S1Int:    1,
			S1String: "S1StringInitConfig",
		},
		DurationField: time.Second,
	}
	defaultPointersConfig := &StructPtr{
		PtrStruct1: &Struct1{
			S1Int:    11,
			S1String: "S1StringDefaultPointersConfig",
			S1Bool:   true,
			S1PtrStruct3: &Struct3{
				S3Float64: 11.11,
			},
		},
		PtrStruct2: &Struct2{
			S2Int64:  22,
			S2String: "S2StringDefaultPointersConfig",
			S2Bool:   false,
		},
	}

	//Test
	rootCmd := &flaeg.Command{
		Name:                  "test",
		Description:           "description test",
		Config:                config,
		DefaultPointersConfig: defaultPointersConfig,
		Run: func() error {
			fmt.Fprintf(&b, "Run with config :\n%+v\n", config)
			return nil
		},
	}
	s := NewStaert(rootCmd)
	toml := NewTomlSource("trivial", []string{"./toml/", "/any/other/path"})
	s.AddSource(toml)

	if err := s.parseConfigAllSources(rootCmd); err != nil {
		t.Errorf("Error %s", err.Error())
	}
	//Check
	check := &StructPtr{
		PtrStruct1: &Struct1{
			S1Int:    28,
			S1String: "S1StringDefaultPointersConfig",
			S1Bool:   true,
		},
		DurationField: 28 * time.Nanosecond,
	}

	if !reflect.DeepEqual(rootCmd.Config, check) {
		t.Errorf("\nexpected\t: %+v\ngot\t\t\t: %+v\n", check, rootCmd.Config)
		t.Errorf("\nexpected\t: %+v\ngot\t\t\t: %+v\n", check.PtrStruct1, config.PtrStruct1)
	}
}

func TestTomlSourcePointer(t *testing.T) {
	//use buffer instead of stdout
	var b bytes.Buffer
	//Init
	config := &StructPtr{
		PtrStruct1: &Struct1{
			S1Int:    1,
			S1String: "S1StringInitConfig",
		},
		DurationField: time.Second,
	}
	defaultPointersConfig := &StructPtr{
		PtrStruct1: &Struct1{
			S1Int:    11,
			S1String: "S1StringDefaultPointersConfig",
			S1Bool:   true,
			S1PtrStruct3: &Struct3{
				S3Float64: 11.11,
			},
		},
		PtrStruct2: &Struct2{
			S2Int64:  22,
			S2String: "S2StringDefaultPointersConfig",
			S2Bool:   false,
		},
	}

	//Test
	rootCmd := &flaeg.Command{
		Name:                  "test",
		Description:           "description test",
		Config:                config,
		DefaultPointersConfig: defaultPointersConfig,
		Run: func() error {
			fmt.Fprintf(&b, "Run with config :\n%+v\n", config)
			return nil
		},
	}
	s := NewStaert(rootCmd)
	toml := NewTomlSource("pointer", []string{"./toml/", "/any/other/path"})
	s.AddSource(toml)

	if err := s.parseConfigAllSources(rootCmd); err != nil {
		t.Errorf("Error %s", err.Error())
	}
	//Check
	check := &StructPtr{
		PtrStruct1: &Struct1{
			S1Int:    1,
			S1String: "S1StringInitConfig",
		},
		PtrStruct2: &Struct2{
			S2Int64:  22,
			S2String: "S2StringDefaultPointersConfig",
			S2Bool:   false,
		},
		DurationField: time.Second,
	}

	if !reflect.DeepEqual(rootCmd.Config, check) {
		t.Errorf("\nexpected\t: %+v\ngot\t\t\t: %+v\n", check, rootCmd.Config)
	}
}

func TestTomlSourceFieldUnderPointer(t *testing.T) {
	//Init
	config := &StructPtr{
		PtrStruct1: &Struct1{
			S1Int:    1,
			S1String: "S1StringInitConfig",
		},
		DurationField: time.Second,
	}
	defaultPointersConfig := &StructPtr{
		PtrStruct1: &Struct1{
			S1Int:    11,
			S1String: "S1StringDefaultPointersConfig",
			S1Bool:   true,
			S1PtrStruct3: &Struct3{
				S3Float64: 11.11,
			},
		},
		PtrStruct2: &Struct2{
			S2Int64:  22,
			S2String: "S2StringDefaultPointersConfig",
			S2Bool:   false,
		},
	}

	//Test
	rootCmd := &flaeg.Command{
		Name:                  "test",
		Description:           "description test",
		Config:                config,
		DefaultPointersConfig: defaultPointersConfig,
		Run: func() error {
			fmt.Printf("Run with config :\n%+v\n", config)
			return nil
		},
	}
	s := NewStaert(rootCmd)
	toml := NewTomlSource("fieldUnderPointer", []string{"./toml/", "/any/other/path"})
	s.AddSource(toml)

	if err := s.parseConfigAllSources(rootCmd); err != nil {
		t.Errorf("Error %s", err.Error())
	}
	//Check
	check := &StructPtr{
		PtrStruct1: &Struct1{
			S1Int:    11,
			S1String: "S1StringDefaultPointersConfig",
			S1Bool:   true,
		},
		DurationField: 42 * time.Second,
	}

	if !reflect.DeepEqual(rootCmd.Config, check) {
		t.Errorf("\nexpected\t: %+v\ngot\t\t\t: %+v\n", check, config)
		t.Errorf("\nexpected\t: %+v\ngot\t\t\t: %+v\n", check.PtrStruct1, config.PtrStruct1)
	}
}

func TestTomlSourcePointerUnderPointer(t *testing.T) {
	//use buffer instead of stdout
	var b bytes.Buffer
	//Init
	config := &StructPtr{
		PtrStruct1: &Struct1{
			S1Int:    1,
			S1String: "S1StringInitConfig",
		},
		DurationField: time.Second,
	}
	defaultPointersConfig := &StructPtr{
		PtrStruct1: &Struct1{
			S1Int:    11,
			S1String: "S1StringDefaultPointersConfig",
			S1Bool:   true,
			S1PtrStruct3: &Struct3{
				S3Float64: 11.11,
			},
		},
		PtrStruct2: &Struct2{
			S2Int64:  22,
			S2String: "S2StringDefaultPointersConfig",
			S2Bool:   false,
		},
	}

	//Test
	rootCmd := &flaeg.Command{
		Name:                  "test",
		Description:           "description test",
		Config:                config,
		DefaultPointersConfig: defaultPointersConfig,
		Run: func() error {
			fmt.Fprintf(&b, "Run with config :\n%+v\n", config)
			return nil
		},
	}
	s := NewStaert(rootCmd)
	toml := NewTomlSource("pointerUnderPointer", []string{"./toml/", "/any/other/path"})
	s.AddSource(toml)

	if err := s.parseConfigAllSources(rootCmd); err != nil {
		t.Errorf("Error %s", err.Error())
	}
	//Check
	check := &StructPtr{
		PtrStruct1: &Struct1{
			S1Int:    11,
			S1String: "S1StringDefaultPointersConfig",
			S1Bool:   true,
			S1PtrStruct3: &Struct3{
				S3Float64: 11.11,
			},
		},
		DurationField: time.Second,
	}

	if !reflect.DeepEqual(rootCmd.Config, check) {
		t.Errorf("\nexpected\t: %+v\ngot\t\t\t: %+v\n", check, rootCmd.Config)
		t.Errorf("\nexpected\t: %+v\ngot\t\t\t: %+v\n", check.PtrStruct1, config.PtrStruct1)
		t.Errorf("\nexpected\t: %+v\ngot\t\t\t: %+v\n", check.PtrStruct1.S1PtrStruct3, config.PtrStruct1.S1PtrStruct3)
	}
}

func TestTomlSourceFieldUnderPointerUnderPointer(t *testing.T) {
	//use buffer instead of stdout
	var b bytes.Buffer
	//Init
	config := &StructPtr{
		PtrStruct1: &Struct1{
			S1Int:    1,
			S1String: "S1StringInitConfig",
		},
		DurationField: time.Second,
	}
	defaultPointersConfig := &StructPtr{
		PtrStruct1: &Struct1{
			S1Int:    11,
			S1String: "S1StringDefaultPointersConfig",
			S1Bool:   true,
			S1PtrStruct3: &Struct3{
				S3Float64: 11.11,
			},
		},
		PtrStruct2: &Struct2{
			S2Int64:  22,
			S2String: "S2StringDefaultPointersConfig",
			S2Bool:   false,
		},
	}

	//Test
	rootCmd := &flaeg.Command{
		Name:                  "test",
		Description:           "description test",
		Config:                config,
		DefaultPointersConfig: defaultPointersConfig,
		Run: func() error {
			fmt.Fprintf(&b, "Run with config :\n%+v\n", config)
			return nil
		},
	}
	s := NewStaert(rootCmd)
	toml := NewTomlSource("fieldUnderPtrUnderPtr", []string{"./toml/", "/any/other/path"})
	s.AddSource(toml)

	if err := s.parseConfigAllSources(rootCmd); err != nil {
		t.Errorf("Error %s", err.Error())
	}
	//Check
	check := &StructPtr{
		PtrStruct1: &Struct1{
			S1Int:    1,
			S1String: "S1StringInitConfig",
			S1PtrStruct3: &Struct3{
				S3Float64: 28.28,
			},
		},
		DurationField: time.Second,
	}

	if !reflect.DeepEqual(rootCmd.Config, check) {
		t.Errorf("\nexpected\t: %+v\ngot\t\t\t: %+v\n", check, rootCmd.Config)
	}
}

func TestMergeTomlNothingFlaegNoArgs(t *testing.T) {
	//use buffer instead of stdout
	var b bytes.Buffer
	//Init
	config := &StructPtr{
		PtrStruct1: &Struct1{
			S1Int:    1,
			S1String: "S1StringInitConfig",
		},
		DurationField: time.Second,
	}
	defaultPointersConfig := &StructPtr{
		PtrStruct1: &Struct1{
			S1Int:    11,
			S1String: "S1StringDefaultPointersConfig",
			S1Bool:   true,
			S1PtrStruct3: &Struct3{
				S3Float64: 11.11,
			},
		},
		PtrStruct2: &Struct2{
			S2Int64:  22,
			S2String: "S2StringDefaultPointersConfig",
			S2Bool:   false,
		},
	}

	args := []string{}

	//Test
	rootCmd := &flaeg.Command{
		Name:                  "test",
		Description:           "description test",
		Config:                config,
		DefaultPointersConfig: defaultPointersConfig,
		Run: func() error {
			fmt.Fprintf(&b, "Run with config :\n%+v\n", config)
			return nil
		},
	}
	s := NewStaert(rootCmd)
	toml := NewTomlSource("nothing", []string{"./toml/", "/any/other/path"})
	s.AddSource(toml)
	fs := flaeg.New(rootCmd, args)
	s.AddSource(fs)
	if err := s.parseConfigAllSources(rootCmd); err != nil {
		t.Errorf("Error %s", err.Error())
	}

	//Check
	check := &StructPtr{
		PtrStruct1: &Struct1{
			S1Int:    1,
			S1String: "S1StringInitConfig",
		},
		DurationField: time.Second,
	}

	if !reflect.DeepEqual(rootCmd.Config, check) {
		t.Errorf("\nexpected\t: %+v\ngot\t\t\t: %+v\n", check, rootCmd.Config)
	}

}

func TestMergeTomlFieldUnderPointerUnderPointerFlaegNoArgs(t *testing.T) {
	//use buffer instead of stdout
	var b bytes.Buffer
	//Init
	config := &StructPtr{
		PtrStruct1: &Struct1{
			S1Int:    1,
			S1String: "S1StringInitConfig",
		},
		DurationField: time.Second,
	}
	defaultPointersConfig := &StructPtr{
		PtrStruct1: &Struct1{
			S1Int:    11,
			S1String: "S1StringDefaultPointersConfig",
			S1Bool:   true,
			S1PtrStruct3: &Struct3{
				S3Float64: 11.11,
			},
		},
		PtrStruct2: &Struct2{
			S2Int64:  22,
			S2String: "S2StringDefaultPointersConfig",
			S2Bool:   false,
		},
	}

	args := []string{}

	//Test
	rootCmd := &flaeg.Command{
		Name:                  "test",
		Description:           "description test",
		Config:                config,
		DefaultPointersConfig: defaultPointersConfig,
		Run: func() error {
			fmt.Fprintf(&b, "Run with config :\n%+v\n", config)
			return nil
		},
	}
	s := NewStaert(rootCmd)
	toml := NewTomlSource("fieldUnderPtrUnderPtr", []string{"./toml/", "/any/other/path"})
	s.AddSource(toml)
	fs := flaeg.New(rootCmd, args)
	s.AddSource(fs)
	if err := s.parseConfigAllSources(rootCmd); err != nil {
		t.Errorf("Error %s", err.Error())
	}

	//Check
	check := &StructPtr{
		PtrStruct1: &Struct1{
			S1Int:    1,
			S1String: "S1StringInitConfig",
			S1PtrStruct3: &Struct3{
				S3Float64: 28.28,
			},
		},
		DurationField: time.Second,
	}

	if !reflect.DeepEqual(rootCmd.Config, check) {
		t.Errorf("\nexpected\t: %+v\ngot\t\t\t: %+v\n", check, rootCmd.Config)
	}

}

func TestMergeTomlTrivialFlaegOverwriteField(t *testing.T) {
	//use buffer instead of stdout
	var b bytes.Buffer
	//Init
	config := &StructPtr{
		PtrStruct1: &Struct1{
			S1Int:    1,
			S1String: "S1StringInitConfig",
		},
		DurationField: time.Second,
	}
	defaultPointersConfig := &StructPtr{
		PtrStruct1: &Struct1{
			S1Int:    11,
			S1String: "S1StringDefaultPointersConfig",
			S1Bool:   true,
			S1PtrStruct3: &Struct3{
				S3Float64: 11.11,
			},
		},
		PtrStruct2: &Struct2{
			S2Int64:  22,
			S2String: "S2StringDefaultPointersConfig",
			S2Bool:   false,
		},
	}

	args := []string{"--ptrstruct1.s1int=55"}

	//Test
	rootCmd := &flaeg.Command{
		Name:                  "test",
		Description:           "description test",
		Config:                config,
		DefaultPointersConfig: defaultPointersConfig,
		Run: func() error {
			fmt.Fprintf(&b, "Run with config :\n%+v\n", config)
			return nil
		},
	}
	s := NewStaert(rootCmd)
	toml := NewTomlSource("trivial", []string{"./toml/", "/any/other/path"})
	s.AddSource(toml)
	fs := flaeg.New(rootCmd, args)
	s.AddSource(fs)
	if err := s.parseConfigAllSources(rootCmd); err != nil {
		t.Errorf("Error %s", err.Error())
	}
	//Check
	check := &StructPtr{
		PtrStruct1: &Struct1{
			S1Int:    55,
			S1String: "S1StringDefaultPointersConfig",
			S1Bool:   true,
		},
		DurationField: 28 * time.Nanosecond,
	}

	if !reflect.DeepEqual(rootCmd.Config, check) {
		t.Errorf("\nexpected\t: %+v\ngot\t\t\t: %+v\n", check, rootCmd.Config)
		t.Errorf("\nexpected\t: %+v\ngot\t\t\t: %+v\n", check.PtrStruct1, config.PtrStruct1)

	}

}

func TestMergeTomlPointerUnderPointerFlaegManyArgs(t *testing.T) {
	//use buffer instead of stdout
	var b bytes.Buffer
	//Init
	config := &StructPtr{
		PtrStruct1: &Struct1{
			S1Int:    1,
			S1String: "S1StringInitConfig",
		},
		DurationField: time.Second,
	}
	defaultPointersConfig := &StructPtr{
		PtrStruct1: &Struct1{
			S1Int:    11,
			S1String: "S1StringDefaultPointersConfig",
			S1Bool:   true,
			S1PtrStruct3: &Struct3{
				S3Float64: 11.11,
			},
		},
		PtrStruct2: &Struct2{
			S2Int64:  22,
			S2String: "S2StringDefaultPointersConfig",
			S2Bool:   false,
		},
	}

	args := []string{
		"--ptrstruct1.s1int=55",
		"--durationfield=55s",
		"--ptrstruct2.s2string=S2StringFlaeg",
	}

	//Test
	rootCmd := &flaeg.Command{
		Name:                  "test",
		Description:           "description test",
		Config:                config,
		DefaultPointersConfig: defaultPointersConfig,
		Run: func() error {
			fmt.Fprintf(&b, "Run with config :\n%+v\n", config)
			return nil
		},
	}
	s := NewStaert(rootCmd)
	toml := NewTomlSource("pointerUnderPointer", []string{"./toml/", "/any/other/path"})
	s.AddSource(toml)
	fs := flaeg.New(rootCmd, args)
	s.AddSource(fs)
	if err := s.parseConfigAllSources(rootCmd); err != nil {
		t.Errorf("Error %s", err.Error())
	}

	//Check
	check := &StructPtr{
		PtrStruct1: &Struct1{
			S1Int:    55,
			S1String: "S1StringDefaultPointersConfig",
			S1Bool:   true,
			S1PtrStruct3: &Struct3{
				S3Float64: 11.11,
			},
		},
		PtrStruct2: &Struct2{
			S2Int64:  22,
			S2String: "S2StringFlaeg",
			S2Bool:   false,
		},
		DurationField: time.Second * 55,
	}
	if !reflect.DeepEqual(rootCmd.Config, check) {
		t.Errorf("\nexpected\t: %+v\ngot\t\t\t: %+v\n", check, rootCmd.Config)
	}

}

func TestMergeFlaegNoArgsTomlNothing(t *testing.T) {
	//use buffer instead of stdout
	var b bytes.Buffer
	//Init
	config := &StructPtr{
		PtrStruct1: &Struct1{
			S1Int:    1,
			S1String: "S1StringInitConfig",
		},
		DurationField: time.Second,
	}
	defaultPointersConfig := &StructPtr{
		PtrStruct1: &Struct1{
			S1Int:    11,
			S1String: "S1StringDefaultPointersConfig",
			S1Bool:   true,
			S1PtrStruct3: &Struct3{
				S3Float64: 11.11,
			},
		},
		PtrStruct2: &Struct2{
			S2Int64:  22,
			S2String: "S2StringDefaultPointersConfig",
			S2Bool:   false,
		},
	}

	args := []string{}

	//Test
	rootCmd := &flaeg.Command{
		Name:                  "test",
		Description:           "description test",
		Config:                config,
		DefaultPointersConfig: defaultPointersConfig,
		Run: func() error {
			fmt.Fprintf(&b, "Run with config :\n%+v\n", config)
			return nil
		},
	}
	s := NewStaert(rootCmd)
	fs := flaeg.New(rootCmd, args)
	s.AddSource(fs)
	toml := NewTomlSource("nothing", []string{"./toml/", "/any/other/path"})
	s.AddSource(toml)
	if err := s.parseConfigAllSources(rootCmd); err != nil {
		t.Errorf("Error %s", err.Error())
	}

	//Check
	check := &StructPtr{
		PtrStruct1: &Struct1{
			S1Int:    1,
			S1String: "S1StringInitConfig",
		},
		DurationField: time.Second,
	}

	if !reflect.DeepEqual(rootCmd.Config, check) {
		t.Errorf("\nexpected\t: %+v\ngot\t\t\t: %+v\n", check, rootCmd.Config)
	}

}

func TestMergeFlaegFieldUnderPointerUnderPointerTomlNothing(t *testing.T) {
	//use buffer instead of stdout
	var b bytes.Buffer
	//Init
	config := &StructPtr{
		PtrStruct1: &Struct1{
			S1Int:    1,
			S1String: "S1StringInitConfig",
		},
		DurationField: time.Second,
	}
	defaultPointersConfig := &StructPtr{
		PtrStruct1: &Struct1{
			S1Int:    11,
			S1String: "S1StringDefaultPointersConfig",
			S1Bool:   true,
			S1PtrStruct3: &Struct3{
				S3Float64: 11.11,
			},
		},
		PtrStruct2: &Struct2{
			S2Int64:  22,
			S2String: "S2StringDefaultPointersConfig",
			S2Bool:   false,
		},
	}

	args := []string{
		"--ptrstruct1.s1ptrstruct3.s3float64=55.55",
	}
	//Test
	rootCmd := &flaeg.Command{
		Name:                  "test",
		Description:           "description test",
		Config:                config,
		DefaultPointersConfig: defaultPointersConfig,
		Run: func() error {
			fmt.Fprintf(&b, "Run with config :\n%+v\n", config)
			return nil
		},
	}
	s := NewStaert(rootCmd)
	fs := flaeg.New(rootCmd, args)
	s.AddSource(fs)
	toml := NewTomlSource("nothing", []string{"./toml/", "/any/other/path"})
	s.AddSource(toml)
	if err := s.parseConfigAllSources(rootCmd); err != nil {
		t.Errorf("Error %s", err.Error())
	}

	//Check
	check := &StructPtr{
		PtrStruct1: &Struct1{
			S1Int:    1,
			S1String: "S1StringInitConfig",
			S1PtrStruct3: &Struct3{
				S3Float64: 55.55,
			},
		},
		DurationField: time.Second,
	}

	if !reflect.DeepEqual(rootCmd.Config, check) {
		t.Errorf("\nexpected\t: %+v\ngot\t\t\t: %+v\n", check, rootCmd.Config)
	}

}

func TestMergeFlaegManyArgsTomlOverwriteField(t *testing.T) {
	//use buffer instead of stdout
	var b bytes.Buffer
	//Init
	config := &StructPtr{
		PtrStruct1: &Struct1{
			S1Int:    1,
			S1String: "S1StringInitConfig",
		},
		DurationField: time.Second,
	}
	defaultPointersConfig := &StructPtr{
		PtrStruct1: &Struct1{
			S1Int:    11,
			S1String: "S1StringDefaultPointersConfig",
			S1Bool:   true,
			S1PtrStruct3: &Struct3{
				S3Float64: 11.11,
			},
		},
		PtrStruct2: &Struct2{
			S2Int64:  22,
			S2String: "S2StringDefaultPointersConfig",
			S2Bool:   false,
		},
	}

	args := []string{
		"--ptrstruct1.s1int=55",
		"--durationfield=55s",
		"--ptrstruct2.s2string=S2StringFlaeg",
	}
	//Test
	rootCmd := &flaeg.Command{
		Name:                  "test",
		Description:           "description test",
		Config:                config,
		DefaultPointersConfig: defaultPointersConfig,
		Run: func() error {
			fmt.Fprintf(&b, "Run with config :\n%+v\n", config)
			return nil
		},
	}
	s := NewStaert(rootCmd)
	fs := flaeg.New(rootCmd, args)
	s.AddSource(fs)
	toml := NewTomlSource("trivial", []string{"./toml/", "/any/other/path"})
	s.AddSource(toml)
	if err := s.parseConfigAllSources(rootCmd); err != nil {
		t.Errorf("Error %s", err.Error())
	}

	//Check
	check := &StructPtr{
		PtrStruct1: &Struct1{
			S1Int:    28,
			S1String: "S1StringDefaultPointersConfig",
			S1Bool:   true,
		},
		DurationField: time.Nanosecond * 28,
		PtrStruct2: &Struct2{
			S2Int64:  22,
			S2String: "S2StringFlaeg",
			S2Bool:   false,
		},
	}

	if !reflect.DeepEqual(rootCmd.Config, check) {
		t.Errorf("\nexpected\t: %+v\ngot\t\t\t: %+v\n", check, rootCmd.Config)
	}

}

func TestRunFleagFieldUnderPtrUnderPtr1Command(t *testing.T) {
	//use buffer instead of stdout
	var b bytes.Buffer
	//Init
	config := &StructPtr{
		PtrStruct1: &Struct1{
			S1Int:    1,
			S1String: "S1StringInitConfig",
		},
		DurationField: time.Second,
	}
	defaultPointersConfig := &StructPtr{
		PtrStruct1: &Struct1{
			S1Int:    11,
			S1String: "S1StringDefaultPointersConfig",
			S1Bool:   true,
			S1PtrStruct3: &Struct3{
				S3Float64: 11.11,
			},
		},
		PtrStruct2: &Struct2{
			S2Int64:  22,
			S2String: "S2StringDefaultPointersConfig",
			S2Bool:   false,
		},
	}
	args := []string{
		"--ptrstruct1.s1ptrstruct3.s3float64=55.55",
	}

	//Test
	rootCmd := &flaeg.Command{
		Name:                  "test",
		Description:           "description test",
		Config:                config,
		DefaultPointersConfig: defaultPointersConfig,
		Run: func() error {
			fmt.Fprintf(&b, "Run with config :\n%+v\n", config)
			check := &StructPtr{
				PtrStruct1: &Struct1{
					S1Int:    1,
					S1String: "S1StringInitConfig",
					S1PtrStruct3: &Struct3{
						S3Float64: 55.55,
					},
				},
				DurationField: time.Second,
			}

			if !reflect.DeepEqual(config, check) {
				return fmt.Errorf("\nexpected\t: %+v\ngot\t\t\t: %+v\n", check, config)
			}
			return nil
		},
	}
	s := NewStaert(rootCmd)
	fs := flaeg.New(rootCmd, args)
	s.AddSource(fs)
	_, err := s.LoadConfig()
	if err != nil {
		t.Errorf("Error %s", err.Error())
	}
	if err := s.Run(); err != nil {
		t.Errorf("Error %s", err.Error())
	}
	//check buffer
	checkB := `Run with config`
	if !strings.Contains(b.String(), checkB) {
		t.Errorf("Error output doesn't contain %s,\ngot: %s", checkB, &b)
	}
}

//Version Config
type VersionConfig struct {
	Version string `short:"v" description:"Version"`
}

func TestRunFleagFieldUnderPtrUnderPtr2Command(t *testing.T) {
	//use buffer instead of stdout
	var b bytes.Buffer
	//Init
	config := &StructPtr{
		PtrStruct1: &Struct1{
			S1Int:    1,
			S1String: "S1StringInitConfig",
		},
		DurationField: time.Second,
	}
	defaultPointersConfig := &StructPtr{
		PtrStruct1: &Struct1{
			S1Int:    11,
			S1String: "S1StringDefaultPointersConfig",
			S1Bool:   true,
			S1PtrStruct3: &Struct3{
				S3Float64: 11.11,
			},
		},
		PtrStruct2: &Struct2{
			S2Int64:  22,
			S2String: "S2StringDefaultPointersConfig",
			S2Bool:   false,
		},
	}
	//init version config
	versionConfig := &VersionConfig{"0.1"}

	args := []string{
		"--ptrstruct1.s1ptrstruct3.s3float64=55.55",
	}

	rootCmd := &flaeg.Command{
		Name:                  "test",
		Description:           "description test",
		Config:                config,
		DefaultPointersConfig: defaultPointersConfig,
		Run: func() error {
			fmt.Fprintf(&b, "Run with config :\n%+v\n", config)
			check := &StructPtr{
				PtrStruct1: &Struct1{
					S1Int:    1,
					S1String: "S1StringInitConfig",
					S1PtrStruct3: &Struct3{
						S3Float64: 55.55,
					},
				},
				DurationField: time.Second,
			}

			if !reflect.DeepEqual(config, check) {
				return fmt.Errorf("\nexpected\t: %+v\ngot\t\t\t: %+v\n", check, config)
			}
			return nil
		},
	}
	//vesion command
	versionCmd := &flaeg.Command{
		Name:        "version",
		Description: `Print version`,

		Config:                versionConfig,
		DefaultPointersConfig: versionConfig,
		//test in run
		Run: func() error {
			fmt.Fprintf(&b, "Version %s \n", versionConfig.Version)
			//CHECK
			if versionConfig.Version != "0.1" {
				return fmt.Errorf("expected 0.1 got %s", versionConfig.Version)
			}
			return nil

		},
	}
	//TEST
	s := NewStaert(rootCmd)
	fs := flaeg.New(rootCmd, args)
	fs.AddCommand(versionCmd)
	s.AddSource(fs)
	//check in command run func
	_, err := s.LoadConfig()
	if err != nil {
		t.Errorf("Error %s", err.Error())
	}
	if err := s.Run(); err != nil {
		t.Errorf("Error %s", err.Error())
	}
	//check buffer
	checkB := `Run with config`
	if !strings.Contains(b.String(), checkB) {
		t.Errorf("Error output doesn't contain %s,\ngot: %s", checkB, &b)
	}
}

func TestRunFleagVersion2CommandCallVersion(t *testing.T) {
	//use buffer instead of stdout
	var b bytes.Buffer
	//Init
	config := &StructPtr{
		PtrStruct1: &Struct1{
			S1Int:    1,
			S1String: "S1StringInitConfig",
		},
		DurationField: time.Second,
	}
	defaultPointersConfig := &StructPtr{
		PtrStruct1: &Struct1{
			S1Int:    11,
			S1String: "S1StringDefaultPointersConfig",
			S1Bool:   true,
			S1PtrStruct3: &Struct3{
				S3Float64: 11.11,
			},
		},
		PtrStruct2: &Struct2{
			S2Int64:  22,
			S2String: "S2StringDefaultPointersConfig",
			S2Bool:   false,
		},
	}
	//init version config
	versionConfig := &VersionConfig{"0.1"}

	args := []string{
		// "--toto",  //it now has effet
		"version", //call Command
		"-v2.2beta",
	}

	rootCmd := &flaeg.Command{
		Name:                  "test",
		Description:           "description test",
		Config:                config,
		DefaultPointersConfig: defaultPointersConfig,
		Run: func() error {

			fmt.Fprintf(&b, "Run with config :\n%+v\n", config)
			check := &StructPtr{
				PtrStruct1: &Struct1{
					S1Int:    1,
					S1String: "S1StringInitConfig",
					S1PtrStruct3: &Struct3{
						S3Float64: 55.55,
					},
				},
				DurationField: time.Second,
			}

			if !reflect.DeepEqual(config, check) {
				return fmt.Errorf("\nexpected\t: %+v\ngot\t\t\t: %+v\n", check, config)
			}
			return nil
		},
	}
	//vesion command
	versionCmd := &flaeg.Command{
		Name:        "version",
		Description: `Print version`,

		Config:                versionConfig,
		DefaultPointersConfig: versionConfig,
		//test in run
		Run: func() error {
			fmt.Fprintf(&b, "Version %s \n", versionConfig.Version)
			//CHECK
			if versionConfig.Version != "2.2beta" {
				return fmt.Errorf("expected 2.2beta got %s", versionConfig.Version)
			}
			return nil

		},
	}
	//TEST
	s := NewStaert(rootCmd)
	fs := flaeg.New(rootCmd, args)
	fs.AddCommand(versionCmd)
	s.AddSource(fs)
	//check in command run func
	_, err := s.LoadConfig()
	if err != nil {
		t.Errorf("Error %s", err.Error())
	}
	if err := s.Run(); err != nil {
		t.Errorf("Error %s", err.Error())
	}
	//check buffer
	checkB := `Version 2.2beta`
	if !strings.Contains(b.String(), checkB) {
		t.Errorf("Error output doesn't contain %s,\ngot: %s", checkB, &b)
	}
}

func TestRunMergeFlaegToml2CommmandCallRootCmd(t *testing.T) {
	//use buffer instead of stdout
	var b bytes.Buffer
	//Init
	config := &StructPtr{
		PtrStruct1: &Struct1{
			S1Int:    1,
			S1String: "S1StringInitConfig",
		},
		DurationField: time.Second,
	}
	defaultPointersConfig := &StructPtr{
		PtrStruct1: &Struct1{
			S1Int:    11,
			S1String: "S1StringDefaultPointersConfig",
			S1Bool:   true,
			S1PtrStruct3: &Struct3{
				S3Float64: 11.11,
			},
		},
		PtrStruct2: &Struct2{
			S2Int64:  22,
			S2String: "S2StringDefaultPointersConfig",
			S2Bool:   false,
		},
	}
	//init version config
	versionConfig := &VersionConfig{"0.1"}

	args := []string{
		"--ptrstruct1.s1int=55",
		"--durationfield=55s",
		"--ptrstruct2.s2string=S2StringFlaeg",
	}
	//Test
	rootCmd := &flaeg.Command{
		Name:                  "test",
		Description:           "description test",
		Config:                config,
		DefaultPointersConfig: defaultPointersConfig,
		Run: func() error {
			fmt.Fprintf(&b, "Run with config :\n%+v\n", config)
			//Check
			check := &StructPtr{
				PtrStruct1: &Struct1{
					S1Int:    28,
					S1String: "S1StringDefaultPointersConfig",
					S1Bool:   true,
				},
				DurationField: time.Nanosecond * 28,
				PtrStruct2: &Struct2{
					S2Int64:  22,
					S2String: "S2StringFlaeg",
					S2Bool:   false,
				},
			}

			if !reflect.DeepEqual(config, check) {
				return fmt.Errorf("\nexpected\t: %+v\ngot\t\t\t: %+v\n", check, config)
			}
			return nil
		},
	}
	//vesion command
	versionCmd := &flaeg.Command{
		Name:        "version",
		Description: `Print version`,

		Config:                versionConfig,
		DefaultPointersConfig: versionConfig,
		//test in run
		Run: func() error {
			fmt.Fprintf(&b, "Version %s \n", versionConfig.Version)
			//CHECK
			if versionConfig.Version != "0.1" {
				return fmt.Errorf("expected 0.1 got %s", versionConfig.Version)
			}
			return nil

		},
	}

	s := NewStaert(rootCmd)
	fs := flaeg.New(rootCmd, args)
	fs.AddCommand(versionCmd)
	s.AddSource(fs)
	toml := NewTomlSource("trivial", []string{"./toml/", "/any/other/path"})
	s.AddSource(toml)
	//check in command run func
	_, err := s.LoadConfig()
	if err != nil {
		t.Errorf("Error %s", err.Error())
	}
	if err := s.Run(); err != nil {
		t.Errorf("Error %s", err.Error())
	}
	//check buffer
	checkB := `Run with config :`
	if !strings.Contains(b.String(), checkB) {
		t.Errorf("Error output doesn't contain %s,\ngot: %s", checkB, &b)
	}

}

func TestTomlSourceErrorFileNotFound(t *testing.T) {
	//use buffer instead of stdout
	var b bytes.Buffer
	//Init
	config := &StructPtr{
		PtrStruct1: &Struct1{
			S1Int:    1,
			S1String: "S1StringInitConfig",
		},
		DurationField: time.Second,
	}
	defaultPointersConfig := &StructPtr{
		PtrStruct1: &Struct1{
			S1Int:    11,
			S1String: "S1StringDefaultPointersConfig",
			S1Bool:   true,
			S1PtrStruct3: &Struct3{
				S3Float64: 11.11,
			},
		},
		PtrStruct2: &Struct2{
			S2Int64:  22,
			S2String: "S2StringDefaultPointersConfig",
			S2Bool:   false,
		},
	}

	//Test
	rootCmd := &flaeg.Command{
		Name:                  "test",
		Description:           "description test",
		Config:                config,
		DefaultPointersConfig: defaultPointersConfig,
		Run: func() error {
			fmt.Fprintf(&b, "Run with config :\n%+v\n", config)
			return nil
		},
	}
	checkCmd := *rootCmd
	s := NewStaert(rootCmd)
	toml := NewTomlSource("nothing", []string{"../path", "/any/other/path"})
	s.AddSource(toml)

	//Check
	if err := s.parseConfigAllSources(rootCmd); err != nil {
		t.Errorf("No Error expected\nGot Error : %s", err)
	}
	if !reflect.DeepEqual(checkCmd.Config, rootCmd.Config) {
		t.Errorf("Expected %+v \nGot %+v", checkCmd.Config, rootCmd.Config)
	}
	if !reflect.DeepEqual(checkCmd.DefaultPointersConfig, rootCmd.DefaultPointersConfig) {
		t.Errorf("Expected %+v \nGot %+v", checkCmd.DefaultPointersConfig, rootCmd.DefaultPointersConfig)
	}

}

func TestPreprocessDir(t *testing.T) {
	thisPath, err := filepath.Abs(".")
	if err != nil {
		t.Errorf("Error: %s", err.Error())
	}
	checkMap := map[string]string{
		".":                   thisPath,
		"$HOME":               os.Getenv("HOME"),
		"dir1/dir2":           thisPath + "/dir1/dir2",
		"$HOME/dir1/dir2":     os.Getenv("HOME") + "/dir1/dir2",
		"/etc/test":           "/etc/test",
		"/etc/dir1/file1.ext": "/etc/dir1/file1.ext",
	}
	for in, check := range checkMap {
		out, err := preprocessDir(in)
		if err != nil {
			t.Errorf("Error: %s", err.Error())
		}
		if check != out {
			t.Errorf("input %s\nexpected %s\n got %s", in, check, out)
		}
	}

}

func TestFindFile(t *testing.T) {
	result := findFile("nothing", []string{"", "$HOME/test", "toml"})

	//check
	thisPath, err := filepath.Abs(".")
	if err != nil {
		t.Errorf("Error: %s", err.Error())
	}
	expected := thisPath + "/toml/nothing.toml"
	if result != expected {
		t.Errorf("Expected %s\ngot %s", expected, result)
	}
}

type SliceStr []string
type StructPtrCustom struct {
	PtrCustom *StructCustomParser `description:"Ptr on StructCustomParser"`
}
type StructCustomParser struct {
	CustomField SliceStr `description:"CustomField which requiers custom parser"`
}

func TestTomlMissingCustomParser(t *testing.T) {
	config := &StructPtrCustom{}
	defaultPointersConfig := &StructPtrCustom{&StructCustomParser{SliceStr{"str1", "str2"}}}
	command := &flaeg.Command{
		Name:                  "MissingCustomParser",
		Description:           "This is an example of description",
		Config:                config,
		DefaultPointersConfig: defaultPointersConfig,
		Run: func() error {
			// fmt.Printf("Run example with the config :\n%+v\n", config)
			//check
			check := &StructPtrCustom{&StructCustomParser{SliceStr{"str1", "str2"}}}
			if !reflect.DeepEqual(config, check) {
				return fmt.Errorf("Expected %+v\ngot %+v", check.PtrCustom, config.PtrCustom)
			}
			return nil
		},
	}
	s := NewStaert(command)
	toml := NewTomlSource("missingCustomParser", []string{"toml"})
	s.AddSource(toml)
	_, err := s.LoadConfig()
	if err != nil {
		t.Errorf("Error %s", err.Error())
	}
	if err := s.Run(); err != nil {
		t.Errorf("Error :%s", err)
	}

	//check
	check := &StructPtrCustom{&StructCustomParser{SliceStr{"str1", "str2"}}}
	if !reflect.DeepEqual(config, check) {
		t.Errorf("Expected %+v\ngot %+v", check.PtrCustom, config.PtrCustom)
	}
}
func TestFindFileSliceFileAndDirLastIf(t *testing.T) {

	//check
	thisPath, _ := filepath.Abs(".")
	check := thisPath + "/toml/trivial.toml"
	if result := findFile("trivial", []string{"./toml/", "/any/other/path"}); result != check {
		t.Errorf("Expected %s\nGot %s", check, result)
	}
}
func TestFindFileSliceFileAndDirFirstIf(t *testing.T) {
	inFilename := ""
	inDirNfile := []string{"$PWD/toml/nothing.toml"}
	//check
	thisPath, _ := filepath.Abs(".")
	check := thisPath + "/toml/nothing.toml"
	if result := findFile(inFilename, inDirNfile); result != check {
		t.Errorf("Expected %s\nGot %s", check, result)
	}
}
func TestRunWithoutLoadConfig(t *testing.T) {
	//use buffer instead of stdout
	var b bytes.Buffer
	//Init
	config := &StructPtr{
		PtrStruct1: &Struct1{
			S1Int:    1,
			S1String: "S1StringInitConfig",
		},
		DurationField: time.Second,
	}
	defaultPointersConfig := &StructPtr{
		PtrStruct1: &Struct1{
			S1Int:    11,
			S1String: "S1StringDefaultPointersConfig",
			S1Bool:   true,
			S1PtrStruct3: &Struct3{
				S3Float64: 11.11,
			},
		},
		PtrStruct2: &Struct2{
			S2Int64:  22,
			S2String: "S2StringDefaultPointersConfig",
			S2Bool:   false,
		},
	}

	args := []string{
		"--ptrstruct2",
	}

	//Test
	rootCmd := &flaeg.Command{
		Name:                  "test",
		Description:           "description test",
		Config:                config,
		DefaultPointersConfig: defaultPointersConfig,
		Run: func() error {
			fmt.Fprintf(&b, "Run with config :\n%+v\n", config)
			return nil
		},
	}
	s := NewStaert(rootCmd)
	toml := NewTomlSource("trivial", []string{"./toml/", "/any/other/path"})
	s.AddSource(toml)
	fs := flaeg.New(rootCmd, args)
	s.AddSource(fs)
	// s.LoadConfig() IS MISSING
	s.Run()
	check := &StructPtr{
		PtrStruct1: &Struct1{
			S1Int:    1,
			S1String: "S1StringInitConfig",
		},
		DurationField: time.Second,
	}
	if !reflect.DeepEqual(rootCmd.Config, check) {
		t.Errorf("\nexpected\t: %+v\ngot\t\t\t: %+v\n", check, rootCmd.Config)
	}
	//check buffer
	checkB := `Run with config`
	if !strings.Contains(b.String(), checkB) {
		t.Errorf("Error output doesn't contain %s,\ngot: %s", checkB, &b)
	}
}

func TestFlaegTomlSubCommandParseAllSources(t *testing.T) {
	//use buffer instead of stdout
	var b bytes.Buffer
	//init
	args := []string{
		"subcmd",
		"--Vstring=toto",
	}
	config := &struct {
		Vstring string `description:"string field"`
		Vint    int    `description:"int field"`
	}{
		Vstring: "tata",
		Vint:    -15,
	}
	rootCmd := &flaeg.Command{
		Name:                  "test",
		Description:           "description test",
		Config:                config,
		DefaultPointersConfig: config,
		Run: func() error {
			fmt.Fprintln(&b, "rootCmd")
			fmt.Fprintf(&b, "run with config : %+v\n", config)
			return nil
		},
	}
	subCmd := &flaeg.Command{
		Name:                  "subcmd",
		Description:           "description subcmd",
		Config:                config,
		DefaultPointersConfig: config,
		Run: func() error {
			fmt.Fprintln(&b, "subcmd")
			fmt.Fprintf(&b, "run with config : %+v\n", config)
			return nil
		},
		Metadata: map[string]string{
			"parseAllSources": "true",
		},
	}
	s := NewStaert(rootCmd)
	toml := NewTomlSource("subcmd", []string{"./toml/", "/any/other/path"})
	s.AddSource(toml)
	fs := flaeg.New(rootCmd, args)
	fs.AddCommand(subCmd)
	s.AddSource(fs)
	_, err := s.LoadConfig()
	if err != nil {
		t.Errorf("Error %s", err.Error())
	}
	if err = s.Run(); err != nil {
		t.Errorf("Error %s", err.Error())
	}

	//test
	if !strings.Contains(b.String(), "subcmd") ||
		!strings.Contains(b.String(), "Vstring:toto") ||
		!strings.Contains(b.String(), "Vint:777") {
		t.Errorf("expected: subcmd, Vstring = toto, Vint = 777\n got %s", b.String())
	}
}

func TestFlaegTomlSubCommandParseAllSourcesShouldError(t *testing.T) {
	//use buffer instead of stdout
	var b bytes.Buffer
	//init
	args := []string{
		"subcmd",
		"--Vstring=toto",
	}
	config := &struct {
		Vstring string `description:"string field"`
		Vint    int    `description:"int field"`
	}{
		Vstring: "tata",
		Vint:    -15,
	}

	config2 := &struct {
		Vstring int `description:"int field"` // TO check error
		Vint    int `description:"int field"`
	}{
		Vstring: -1,
		Vint:    -15,
	}

	rootCmd := &flaeg.Command{
		Name:                  "test",
		Description:           "description test",
		Config:                config,
		DefaultPointersConfig: config,
		Run: func() error {
			fmt.Fprintln(&b, "rootCmd")
			fmt.Fprintf(&b, "run with config : %+v\n", config)
			return nil
		},
	}
	subCmd := &flaeg.Command{
		Name:                  "subcmd",
		Description:           "description subcmd",
		Config:                config2,
		DefaultPointersConfig: config2,
		Run: func() error {
			fmt.Fprintln(&b, "subcmd")
			fmt.Fprintf(&b, "run with config : %+v\n", config)
			return nil
		},
		Metadata: map[string]string{
			"parseAllSources": "true",
		},
	}
	s := NewStaert(rootCmd)
	toml := NewTomlSource("subcmd", []string{"./toml/", "/any/other/path"})
	s.AddSource(toml)
	fs := flaeg.New(rootCmd, args)
	fs.AddCommand(subCmd)
	s.AddSource(fs)
	_, err := s.LoadConfig()
	errExp := "Config type doesn't match with root command config type."
	if err == nil || !strings.Contains(err.Error(), errExp) {
		t.Errorf("Experted error %s\n got : %s", errExp, err)
	}
}
