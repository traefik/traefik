package staert

import (
	"encoding/json"
	"errors"
	"github.com/containous/flaeg"
	"github.com/docker/libkv/store"
	"github.com/mitchellh/mapstructure"
	"reflect"
	"strings"
	"testing"
	"time"
)

func TestGenerateMapstructureBasic(t *testing.T) {
	moke := []*store.KVPair{
		&store.KVPair{
			Key:   "test/addr",
			Value: []byte("foo"),
		},
		&store.KVPair{
			Key:   "test/child/data",
			Value: []byte("bar"),
		},
	}
	prefix := "test"

	output, err := generateMapstructure(moke, prefix)
	if err != nil {
		t.Fatalf("Error :%s", err)
	}
	//check
	check := map[string]interface{}{
		"addr": "foo",
		"child": map[string]interface{}{
			"data": "bar",
		},
	}
	if !reflect.DeepEqual(check, output) {
		t.Fatalf("Expected %+v\nGot %+v", check, output)
	}
}

func TestGenerateMapstructureTrivialMap(t *testing.T) {
	moke := []*store.KVPair{
		&store.KVPair{
			Key:   "test/vfoo",
			Value: []byte("foo"),
		},
		&store.KVPair{
			Key:   "test/vother/foo",
			Value: []byte("foo"),
		},
		&store.KVPair{
			Key:   "test/vother/bar",
			Value: []byte("bar"),
		},
	}
	prefix := "test"

	output, err := generateMapstructure(moke, prefix)
	if err != nil {
		t.Fatalf("Error :%s", err)
	}
	//check
	check := map[string]interface{}{
		"vfoo": "foo",
		"vother": map[string]interface{}{
			"foo": "foo",
			"bar": "bar",
		},
	}
	if !reflect.DeepEqual(check, output) {
		t.Fatalf("Expected %#v\nGot %#v", check, output)
	}
}

func TestGenerateMapstructureTrivialSlice(t *testing.T) {
	moke := []*store.KVPair{
		&store.KVPair{
			Key:   "test/vfoo",
			Value: []byte("foo"),
		},
		&store.KVPair{
			Key:   "test/vother/0",
			Value: []byte("foo"),
		},
		&store.KVPair{
			Key:   "test/vother/1",
			Value: []byte("bar1"),
		},
		&store.KVPair{
			Key:   "test/vother/2",
			Value: []byte("bar2"),
		},
	}
	prefix := "test"

	output, err := generateMapstructure(moke, prefix)
	if err != nil {
		t.Fatalf("Error :%s", err)
	}
	//check
	check := map[string]interface{}{
		"vfoo": "foo",
		"vother": map[string]interface{}{
			"0": "foo",
			"1": "bar1",
			"2": "bar2",
		},
	}
	if !reflect.DeepEqual(check, output) {
		t.Fatalf("Expected %#v\nGot %#v", check, output)
	}
}

func TestGenerateMapstructureNotTrivialSlice(t *testing.T) {
	moke := []*store.KVPair{
		&store.KVPair{
			Key:   "test/vfoo",
			Value: []byte("foo"),
		},
		&store.KVPair{
			Key:   "test/vother/0/foo1",
			Value: []byte("bar"),
		},
		&store.KVPair{
			Key:   "test/vother/0/foo2",
			Value: []byte("bar"),
		},
		&store.KVPair{
			Key:   "test/vother/1/bar1",
			Value: []byte("foo"),
		},
		&store.KVPair{
			Key:   "test/vother/1/bar2",
			Value: []byte("foo"),
		},
	}
	prefix := "test"

	output, err := generateMapstructure(moke, prefix)
	if err != nil {
		t.Fatalf("Error :%s", err)
	}
	//check
	check := map[string]interface{}{
		"vfoo": "foo",
		"vother": map[string]interface{}{
			"0": map[string]interface{}{
				"foo1": "bar",
				"foo2": "bar",
			},
			"1": map[string]interface{}{
				"bar1": "foo",
				"bar2": "foo",
			},
		},
	}
	if !reflect.DeepEqual(check, output) {
		t.Fatalf("Expected %#v\nGot %#v", check, output)
	}
}

func TestDecodeHookSlice(t *testing.T) {
	data := map[string]interface{}{
		"10": map[string]interface{}{
			"bar1": "bar1",
			"bar2": "bar2",
		},
		"2": map[string]interface{}{
			"bar1": "foo1",
			"bar2": "foo2",
		},
	}
	output, err := decodeHook(reflect.TypeOf(data), reflect.TypeOf([]string{}), data)
	if err != nil {
		t.Fatalf("Error : %s", err)
	}

	check := []interface{}{
		map[string]interface{}{
			"bar1": "foo1",
			"bar2": "foo2",
		},
		map[string]interface{}{
			"bar1": "bar1",
			"bar2": "bar2",
		},
	}
	if !reflect.DeepEqual(check, output) {
		t.Fatalf("Expected %#v\nGot %#v", check, output)
	}

}

type BasicStruct struct {
	Bar1 string
	Bar2 string
}
type SliceStruct []BasicStruct
type Test struct {
	Vfoo   string
	Vother SliceStruct
}

func TestIntegrationMapstructureWithDecodeHook(t *testing.T) {
	input := map[string]interface{}{
		"vfoo": "foo",
		"vother": map[string]interface{}{
			"10": map[string]interface{}{
				"bar1": "bar1",
				"bar2": "bar2",
			},
			"2": map[string]interface{}{
				"bar1": "foo1",
				"bar2": "foo2",
			},
		},
	}
	var config Test

	//test
	configDecoder := &mapstructure.DecoderConfig{
		Metadata:         nil,
		Result:           &config,
		WeaklyTypedInput: true,
		DecodeHook:       decodeHook,
	}
	decoder, err := mapstructure.NewDecoder(configDecoder)
	if err != nil {
		t.Fatalf("got an err: %s", err.Error())
	}
	if err := decoder.Decode(input); err != nil {
		t.Fatalf("got an err: %s", err.Error())
	}

	//check
	check := Test{
		Vfoo: "foo",
		Vother: SliceStruct{
			BasicStruct{
				Bar1: "foo1",
				Bar2: "foo2",
			},
			BasicStruct{
				Bar1: "bar1",
				Bar2: "bar2",
			},
		},
	}

	if !reflect.DeepEqual(check, config) {
		t.Fatalf("Expected %#v\nGot %#v", check, config)
	}
}

// Extremely limited mock store so we can test initialization
type Mock struct {
	Error           bool
	KVPairs         []*store.KVPair
	WatchTreeMethod func() <-chan []*store.KVPair
}

func (s *Mock) Put(key string, value []byte, opts *store.WriteOptions) error {
	s.KVPairs = append(s.KVPairs, &store.KVPair{key, value, 0})
	return nil
}

func (s *Mock) Get(key string) (*store.KVPair, error) {
	if s.Error {
		return nil, errors.New("Error")
	}
	for _, kvPair := range s.KVPairs {
		if kvPair.Key == key {
			return kvPair, nil
		}
	}
	return nil, nil
}

func (s *Mock) Delete(key string) error {
	return errors.New("Delete not supported")
}

// Exists mock
func (s *Mock) Exists(key string) (bool, error) {
	return false, errors.New("Exists not supported")
}

// Watch mock
func (s *Mock) Watch(key string, stopCh <-chan struct{}) (<-chan *store.KVPair, error) {
	return nil, errors.New("Watch not supported")
}

// WatchTree mock
func (s *Mock) WatchTree(prefix string, stopCh <-chan struct{}) (<-chan []*store.KVPair, error) {
	return s.WatchTreeMethod(), nil
}

// NewLock mock
func (s *Mock) NewLock(key string, options *store.LockOptions) (store.Locker, error) {
	return nil, errors.New("NewLock not supported")
}

// List mock
func (s *Mock) List(prefix string) ([]*store.KVPair, error) {
	if s.Error {
		return nil, errors.New("Error")
	}
	kv := []*store.KVPair{}
	for _, kvPair := range s.KVPairs {
		if strings.HasPrefix(kvPair.Key, prefix+"/") {
			if secondSlashIndex := strings.IndexRune(kvPair.Key[len(prefix)+1:], '/'); secondSlashIndex == -1 {
				kv = append(kv, kvPair)
			} else {
				dir := &store.KVPair{
					Key: kvPair.Key[:secondSlashIndex+len(prefix)+1],
				}
				kv = append(kv, dir)
			}
		}
	}
	return kv, nil
}

// DeleteTree mock
func (s *Mock) DeleteTree(prefix string) error {
	return errors.New("DeleteTree not supported")
}

// AtomicPut mock
func (s *Mock) AtomicPut(key string, value []byte, previous *store.KVPair, opts *store.WriteOptions) (bool, *store.KVPair, error) {
	return false, nil, errors.New("AtomicPut not supported")
}

// AtomicDelete mock
func (s *Mock) AtomicDelete(key string, previous *store.KVPair) (bool, error) {
	return false, errors.New("AtomicDelete not supported")
}

// Close mock
func (s *Mock) Close() {
	return
}

func TestKvSourceEmpty(t *testing.T) {
	//Init
	config := &StructPtr{
		PtrStruct1: &Struct1{
			S1Int:    1,
			S1String: "S1StringInitConfig",
		},
		DurationField: time.Second,
	}

	//Test
	rootCmd := &flaeg.Command{
		Name:                  "test",
		Description:           "description test",
		Config:                config,
		DefaultPointersConfig: config,
		Run: func() error { return nil },
	}
	s := NewStaert(rootCmd)
	kv := &KvSource{
		&Mock{
			KVPairs: []*store.KVPair{},
		},
		"test/",
	}
	s.AddSource(kv)

	_, err := s.LoadConfig()
	if err != nil {
		t.Fatalf("Error %s", err)
	}

	//Check
	check := &StructPtr{
		PtrStruct1: &Struct1{
			S1Int:    1,
			S1String: "S1StringInitConfig",
		},
		DurationField: time.Second,
	}

	if !reflect.DeepEqual(check, rootCmd.Config) {
		t.Fatalf("\nexpected\t: %+v\ngot\t\t\t: %+v\n", check, rootCmd.Config)
	}
}

func TestGenerateMapstructureTrivial(t *testing.T) {
	input := []*store.KVPair{
		{
			Key:   "test/ptrstruct1/s1int",
			Value: []byte("28"),
		},
		{
			Key:   "test/durationfield",
			Value: []byte("28"),
		},
	}
	prefix := "test"
	output, err := generateMapstructure(input, prefix)
	if err != nil {
		t.Fatalf("Error :%s", err)
	}
	//check
	check := map[string]interface{}{
		"durationfield": "28",
		"ptrstruct1": map[string]interface{}{
			"s1int": "28",
		},
	}
	if !reflect.DeepEqual(check, output) {
		printResult, err := json.Marshal(output)
		if err != nil {
			t.Fatalf("error: %s", err)
		}
		printCheck, err := json.Marshal(check)
		if err != nil {
			t.Fatalf("error: %s", err)
		}
		t.Fatalf("\nexpected\t: %s\ngot\t\t\t: %s\n", printCheck, printResult)
	}
}

func TestIntegrationMapstructureWithDecodeHookPointer(t *testing.T) {
	mapstruct := map[string]interface{}{
		"durationfield": "28",
		"ptrstruct1": map[string]interface{}{
			"s1int": "28",
		},
	}
	config := StructPtr{}

	//test
	configDecoder := &mapstructure.DecoderConfig{
		Metadata:         nil,
		Result:           &config,
		WeaklyTypedInput: true,
		DecodeHook:       decodeHook,
		//TODO : ZeroFields:       false, doesn't work

	}
	decoder, err := mapstructure.NewDecoder(configDecoder)
	if err != nil {
		t.Fatalf("got an err: %s", err.Error())
	}
	if err := decoder.Decode(mapstruct); err != nil {
		t.Fatalf("got an err: %s", err.Error())
	}

	//check
	check := StructPtr{
		PtrStruct1: &Struct1{
			S1Int: 28,
		},
		DurationField: time.Nanosecond * 28,
	}

	if !reflect.DeepEqual(check, config) {
		printResult, err := json.Marshal(config)
		if err != nil {
			t.Fatalf("error: %s", err)
		}
		printCheck, err := json.Marshal(check)
		if err != nil {
			t.Fatalf("error: %s", err)
		}
		t.Fatalf("\nexpected\t: %s\ngot\t\t\t: %s\n", printCheck, printResult)
	}
}
func TestIntegrationMapstructureInitedPtrReset(t *testing.T) {
	mapstruct := map[string]interface{}{
		// "durationfield": "28",
		"ptrstruct1": map[string]interface{}{
			"s1int": "24",
		},
	}
	config := StructPtr{
		PtrStruct1: &Struct1{
			S1Int:    1,
			S1String: "S1StringInitConfig",
		},
		DurationField: time.Nanosecond * 28,
	}

	//test
	configDecoder := &mapstructure.DecoderConfig{
		Metadata:         nil,
		Result:           &config,
		WeaklyTypedInput: true,
		DecodeHook:       decodeHook,
	}
	decoder, err := mapstructure.NewDecoder(configDecoder)
	if err != nil {
		t.Fatalf("got an err: %s", err.Error())
	}
	if err := decoder.Decode(mapstruct); err != nil {
		t.Fatalf("got an err: %s", err.Error())
	}

	//check
	check := StructPtr{
		PtrStruct1: &Struct1{
			S1Int: 24,
		},
		DurationField: time.Nanosecond * 28,
	}

	if !reflect.DeepEqual(check, config) {
		printResult, err := json.Marshal(config)
		if err != nil {
			t.Fatalf("error: %s", err)
		}
		printCheck, err := json.Marshal(check)
		if err != nil {
			t.Fatalf("error: %s", err)
		}
		t.Fatalf("\nexpected\t: %s\ngot\t\t\t: %s\n", printCheck, printResult)
	}
}

func TestParseKvSourceTrivial(t *testing.T) {
	//Init
	config := StructPtr{}

	//Test
	rootCmd := &flaeg.Command{
		Name:                  "test",
		Description:           "description test",
		Config:                &config,
		DefaultPointersConfig: &config,
		Run: func() error { return nil },
	}
	kv := &KvSource{
		&Mock{
			KVPairs: []*store.KVPair{
				{
					Key:   "test/ptrstruct1/s1int",
					Value: []byte("28"),
				},
				{
					Key:   "test/durationfield",
					Value: []byte("28"),
				},
			},
		},
		"test",
	}
	if _, err := kv.Parse(rootCmd); err != nil {
		t.Fatalf("Error %s", err)
	}

	//Check
	check := &StructPtr{
		PtrStruct1: &Struct1{
			S1Int: 28,
		},
		DurationField: time.Nanosecond * 28,
	}

	if !reflect.DeepEqual(check, rootCmd.Config) {
		printResult, err := json.Marshal(rootCmd.Config)
		if err != nil {
			t.Fatalf("error: %s", err)
		}
		printCheck, err := json.Marshal(check)
		if err != nil {
			t.Fatalf("error: %s", err)
		}
		t.Fatalf("\nexpected\t: %s\ngot\t\t\t: %s\n", printCheck, printResult)
	}
}

func TestLoadConfigKvSourceNestedPtrsNil(t *testing.T) {
	//Init
	config := &StructPtr{}

	//Test
	kv := &KvSource{
		&Mock{
			KVPairs: []*store.KVPair{
				{
					Key:   "prefix/ptrstruct1/s1int",
					Value: []byte("1"),
				},
				{
					Key:   "prefix/ptrstruct1/s1string",
					Value: []byte("S1StringInitConfig"),
				},
				{
					Key:   "prefix/ptrstruct1/s1bool",
					Value: []byte("false"),
				},
				{
					Key:   "prefix/ptrstruct1/s1ptrstruct3/s3float64",
					Value: []byte("0"),
				},
				{
					Key:   "prefix/durationfield",
					Value: []byte("21000000000"),
				},
			},
		},
		"prefix",
	}
	if err := kv.LoadConfig(config); err != nil {
		t.Fatalf("Error %s", err)
	}

	//Check
	check := &StructPtr{
		PtrStruct1: &Struct1{
			S1Int:        1,
			S1String:     "S1StringInitConfig",
			S1PtrStruct3: &Struct3{},
		},
		DurationField: 21 * time.Second,
	}

	if !reflect.DeepEqual(check, config) {
		printResult, err := json.Marshal(config)
		if err != nil {
			t.Fatalf("error: %s", err)
		}
		printCheck, err := json.Marshal(check)
		if err != nil {
			t.Fatalf("error: %s", err)
		}
		t.Fatalf("\nexpected\t: %s\ngot\t\t\t: %s\n", printCheck, printResult)
	}
}

func TestParseKvSourceNestedPtrsNil(t *testing.T) {
	//Init
	config := StructPtr{}

	//Test
	rootCmd := &flaeg.Command{
		Name:                  "test",
		Description:           "description test",
		Config:                &config,
		DefaultPointersConfig: &config,
		Run: func() error { return nil },
	}
	kv := &KvSource{
		&Mock{
			KVPairs: []*store.KVPair{
				{
					Key:   "prefix/ptrstruct1/s1int",
					Value: []byte("1"),
				},
				{
					Key:   "prefix/ptrstruct1/s1string",
					Value: []byte("S1StringInitConfig"),
				},
				{
					Key:   "prefix/ptrstruct1/s1bool",
					Value: []byte("false"),
				},
				{
					Key:   "prefix/ptrstruct1/s1ptrstruct3/s3float64",
					Value: []byte("0"),
				},
				{
					Key:   "prefix/durationfield",
					Value: []byte("21000000000"),
				},
			},
		},
		"prefix",
	}
	if _, err := kv.Parse(rootCmd); err != nil {
		t.Fatalf("Error %s", err)
	}

	//Check
	check := &StructPtr{
		PtrStruct1: &Struct1{
			S1Int:        1,
			S1String:     "S1StringInitConfig",
			S1PtrStruct3: &Struct3{},
		},
		DurationField: 21 * time.Second,
	}

	if !reflect.DeepEqual(check, rootCmd.Config) {
		printResult, err := json.Marshal(rootCmd.Config)
		if err != nil {
			t.Fatalf("error: %s", err)
		}
		printCheck, err := json.Marshal(check)
		if err != nil {
			t.Fatalf("error: %s", err)
		}
		t.Fatalf("\nexpected\t: %s\ngot\t\t\t: %s\n", printCheck, printResult)
	}
}

func TestParseKvSourceMap(t *testing.T) {
	//Init
	config := &struct {
		Vmap map[string]int
	}{}

	//Test
	rootCmd := &flaeg.Command{
		Name:                  "test",
		Description:           "description test",
		Config:                config,
		DefaultPointersConfig: config,
		Run: func() error { return nil },
	}
	kv := &KvSource{
		&Mock{
			KVPairs: []*store.KVPair{
				{
					Key:   "prefix/vmap/toto",
					Value: []byte("1"),
				},
				{
					Key:   "prefix/vmap/tata",
					Value: []byte("2"),
				},
				{
					Key:   "prefix/vmap/titi",
					Value: []byte("3"),
				},
			},
		},
		"prefix",
	}
	if _, err := kv.Parse(rootCmd); err != nil {
		t.Fatalf("Error %s", err)
	}

	//Check
	check := &struct {
		Vmap map[string]int
	}{
		Vmap: map[string]int{
			"toto": 1,
			"tata": 2,
			"titi": 3,
		},
	}

	if !reflect.DeepEqual(check, rootCmd.Config) {
		t.Fatalf("\nexpected\t: %#v\ngot\t\t\t: %#v\n", check, rootCmd.Config)
	}
}

// TestCollateKvPairs
func TestCollateKvPairsBasic(t *testing.T) {
	//init
	config := &struct {
		Vstring string
		Vint    int
		Vuint   uint
		Vbool   bool
		Vfloat  float64
		Vextra  string
		vsilent bool
		Vdata   interface{}
	}{
		Vstring: "tata",
		Vint:    -15,
		Vuint:   51,
		Vbool:   true,
		Vfloat:  1.5,
		Vextra:  "toto",
		vsilent: true, //Unexported : must not be in the map
		Vdata:   42,
	}
	//test
	kv := map[string]string{}
	if err := collateKvRecursive(reflect.ValueOf(config), kv, "prefix"); err != nil {
		t.Fatalf("Error : %s", err)
	}
	//check
	check := map[string]string{
		"prefix/vbool":   "true",
		"prefix/vfloat":  "1.5",
		"prefix/vextra":  "toto",
		"prefix/vdata":   "42",
		"prefix/vstring": "tata",
		"prefix/vint":    "-15",
		"prefix/vuint":   "51",
	}
	if !reflect.DeepEqual(kv, check) {
		t.Fatalf("Expected %s\nGot %s", check, kv)
	}
}

func TestCollateKvPairsNestedPointers(t *testing.T) {
	//init
	config := &StructPtr{
		PtrStruct1: &Struct1{
			S1Int:        1,
			S1String:     "S1StringInitConfig",
			S1PtrStruct3: &Struct3{},
		},
		DurationField: 21 * time.Second,
	}
	//test
	kv := map[string]string{}
	if err := collateKvRecursive(reflect.ValueOf(config), kv, "prefix"); err != nil {
		t.Fatalf("Error : %s", err)
	}
	//check
	check := map[string]string{
		"prefix/ptrstruct1/s1int":                  "1",
		"prefix/ptrstruct1/s1string":               "S1StringInitConfig",
		"prefix/ptrstruct1/s1bool":                 "false",
		"prefix/ptrstruct1/s1ptrstruct3/":          "",
		"prefix/ptrstruct1/s1ptrstruct3/s3float64": "0",
		"prefix/durationfield":                     "21000000000",
	}
	if !reflect.DeepEqual(kv, check) {
		t.Fatalf("Expected %s\nGot %s", check, kv)
	}
}

func TestCollateKvPairsMapStringString(t *testing.T) {
	//init
	config := &struct {
		Vfoo   string
		Vother map[string]string
	}{
		Vfoo: "toto",
		Vother: map[string]string{
			"k1": "v1",
			"k2": "v2",
		},
	}
	//test
	kv := map[string]string{}
	if err := collateKvRecursive(reflect.ValueOf(config), kv, "prefix"); err != nil {
		t.Fatalf("Error : %s", err)
	}
	//check
	check := map[string]string{
		"prefix/vother/k1": "v1",
		"prefix/vother/k2": "v2",
		"prefix/vfoo":      "toto",
	}
	if !reflect.DeepEqual(kv, check) {
		t.Fatalf("Expected %s\nGot %s", check, kv)
	}
}
func TestCollateKvPairsMapIntString(t *testing.T) {
	//init
	config := &struct {
		Vfoo   string
		Vother map[int]string
	}{
		Vfoo: "toto",
		Vother: map[int]string{
			51: "v1",
			15: "v2",
		},
	}
	//test
	kv := map[string]string{}
	if err := collateKvRecursive(reflect.ValueOf(config), kv, "prefix"); err != nil {
		t.Fatalf("Error : %s", err)
	}
	//check
	check := map[string]string{
		"prefix/vother/51": "v1",
		"prefix/vother/15": "v2",
		"prefix/vfoo":      "toto",
	}
	if !reflect.DeepEqual(kv, check) {
		t.Fatalf("Expected %s\nGot %s", check, kv)
	}
}
func TestCollateKvPairsMapStringStruct(t *testing.T) {
	//init
	config := &struct {
		Vfoo   string
		Vother map[string]Struct1
	}{
		Vfoo: "toto",
		Vother: map[string]Struct1{
			"k1": Struct1{
				S1Bool:       true,
				S1Int:        51,
				S1PtrStruct3: nil,
			},
			"k2": Struct1{
				S1String: "tata",
			},
		},
	}
	//test
	kv := map[string]string{}
	if err := collateKvRecursive(reflect.ValueOf(config), kv, "prefix"); err != nil {
		t.Fatalf("Error : %s", err)
	}
	//check
	check := map[string]string{
		"prefix/vother/k1/s1bool":   "true",
		"prefix/vother/k1/s1int":    "51",
		"prefix/vother/k1/s1string": "",
		"prefix/vother/k2/s1bool":   "false",
		"prefix/vother/k2/s1int":    "0",
		"prefix/vother/k2/s1string": "tata",
		"prefix/vfoo":               "toto",
	}
	if !reflect.DeepEqual(kv, check) {
		t.Fatalf("Expected %s\nGot %s", check, kv)
	}
}
func TestCollateKvPairsMapStructStructSouldFail(t *testing.T) {
	//init
	config := &struct {
		Vfoo   string
		Vother map[Struct1]Struct1
	}{
		Vfoo: "toto",
		Vother: map[Struct1]Struct1{
			Struct1{
				S1Bool: true,
				S1Int:  1,
			}: Struct1{
				S1Int: 11,
			},
			Struct1{
				S1Bool: true,
				S1Int:  2,
			}: Struct1{
				S1Int: 22,
			},
		},
	}
	//test
	kv := map[string]string{}
	err := collateKvRecursive(reflect.ValueOf(config), kv, "prefix")
	if err == nil || !strings.Contains(err.Error(), "Struct as key not supported") {
		t.Fatalf("Expected error Struct as key not supported\ngot: %s", err)
	}
}

func TestCollateKvPairsSliceInt(t *testing.T) {
	//init
	config := &struct {
		Vfoo   string
		Vother []int
	}{
		Vfoo:   "toto",
		Vother: []int{51, 15},
	}
	//test
	kv := map[string]string{}
	if err := collateKvRecursive(reflect.ValueOf(config), kv, "prefix"); err != nil {
		t.Fatalf("Error : %s", err)
	}
	//check
	check := map[string]string{
		"prefix/vother/0": "51",
		"prefix/vother/1": "15",
		"prefix/vfoo":     "toto",
	}
	if !reflect.DeepEqual(kv, check) {
		t.Fatalf("Expected %s\nGot %s", check, kv)
	}
}

func TestCollateKvPairsSlicePtrOnStruct(t *testing.T) {
	//init
	config := &struct {
		Vfoo   string
		Vother []*BasicStruct
	}{
		Vfoo: "toto",
		Vother: []*BasicStruct{
			&BasicStruct{},
			&BasicStruct{
				Bar1: "tata",
				Bar2: "titi",
			},
		},
	}
	//test
	kv := map[string]string{}
	if err := collateKvRecursive(reflect.ValueOf(config), kv, "prefix"); err != nil {
		t.Fatalf("Error : %s", err)
	}
	//check
	check := map[string]string{
		"prefix/vother/0/":     "",
		"prefix/vother/0/bar1": "",
		"prefix/vother/0/bar2": "",
		"prefix/vother/1/":     "",
		"prefix/vother/1/bar1": "tata",
		"prefix/vother/1/bar2": "titi",
		"prefix/vfoo":          "toto",
	}
	if !reflect.DeepEqual(kv, check) {
		t.Fatalf("Expected %s\nGot %s", check, kv)
	}
}

func TestCollateKvPairsEmbedded(t *testing.T) {
	//init
	config := &struct {
		BasicStruct
		Vfoo string
	}{
		BasicStruct: BasicStruct{
			Bar1: "tata",
			Bar2: "titi",
		},
		Vfoo: "toto",
	}
	//test
	kv := map[string]string{}
	if err := collateKvRecursive(reflect.ValueOf(config), kv, "prefix"); err != nil {
		t.Fatalf("Error : %s", err)
	}
	//check
	check := map[string]string{
		"prefix/basicstruct/bar1": "tata",
		"prefix/basicstruct/bar2": "titi",
		"prefix/vfoo":             "toto",
	}
	if !reflect.DeepEqual(kv, check) {
		t.Fatalf("Expected %s\nGot %s", check, kv)
	}
}

func TestCollateKvPairsEmbeddedSquash(t *testing.T) {
	//init
	config := &struct {
		BasicStruct `mapstructure:",squash"`
		Vfoo        string
	}{
		BasicStruct: BasicStruct{
			Bar1: "tata",
			Bar2: "titi",
		},
		Vfoo: "toto",
	}
	//test
	kv := map[string]string{}
	if err := collateKvRecursive(reflect.ValueOf(config), kv, "prefix"); err != nil {
		t.Fatalf("Error : %s", err)
	}
	//check
	check := map[string]string{
		"prefix/bar1": "tata",
		"prefix/bar2": "titi",
		"prefix/vfoo": "toto",
	}
	if !reflect.DeepEqual(kv, check) {
		t.Fatalf("Expected %s\nGot %s", check, kv)
	}
}

func TestCollateKvPairsNotSupportedKindSouldFail(t *testing.T) {
	//init
	config := &struct {
		Vchan chan int
	}{
		Vchan: make(chan int),
	}
	//test
	kv := map[string]string{}
	err := collateKvRecursive(reflect.ValueOf(config), kv, "prefix")
	if err == nil || !strings.Contains(err.Error(), "Kind chan not supported") {
		t.Fatalf("Expected error : Kind chan not supported\nGot : %s", err)
	}
}

func TestStoreConfigEmbeddedSquash(t *testing.T) {
	//init
	config := &struct {
		BasicStruct `mapstructure:",squash"`
		Vfoo        string
	}{
		BasicStruct: BasicStruct{
			Bar1: "tata",
			Bar2: "titi",
		},
		Vfoo: "toto",
	}
	kv := &KvSource{
		&Mock{},
		"prefix",
	}
	//test
	if err := kv.StoreConfig(config); err != nil {
		t.Fatalf("Error : %s", err)
	}

	//check
	checkMap := map[string]string{
		"prefix/bar1": "tata",
		"prefix/bar2": "titi",
		"prefix/vfoo": "toto",
	}
	result := map[string][]byte{}
	err := kv.ListRecursive("prefix", result)
	if err != nil {
		t.Fatalf("Error : %s", err)
	}
	if len(result) != len(checkMap) {
		t.Fatalf("length of kv.List is not %d", len(checkMap))
	}
	for k, v := range result {
		if string(v) != checkMap[k] {
			t.Fatalf("Key %s\nExpected value %s, got %s", k, v, checkMap[k])
		}

	}

}

func TestCollateKvPairsUnexported(t *testing.T) {
	config := &struct {
		Vstring string
		vsilent string
	}{
		Vstring: "mustBeInTheMap",
		vsilent: "mustNotBeInTheMap",
	}
	//test
	kv := map[string]string{}
	if err := collateKvRecursive(reflect.ValueOf(config), kv, "prefix"); err != nil {
		t.Fatalf("Error : %s", err)
	}
	//check
	if _, ok := kv["prefix/vsilent"]; ok {
		t.Fatalf("Exported field should not be in the map : %s", kv)
	}

	check := map[string]string{
		"prefix/vstring": "mustBeInTheMap",
	}
	if !reflect.DeepEqual(kv, check) {
		t.Fatalf("Expected %s\nGot %s", check, kv)
	}

}

func TestCollateKvPairsShortNameUnexported(t *testing.T) {
	config := &struct {
		E string
		u string
	}{
		E: "mustBeInTheMap",
		u: "mustNotBeInTheMap",
	}
	//test
	kv := map[string]string{}
	if err := collateKvRecursive(reflect.ValueOf(config), kv, "prefix"); err != nil {
		t.Fatalf("Error : %s", err)
	}
	//check
	if _, ok := kv["prefix/u"]; ok {
		t.Fatalf("Exported field should not be in the map : %s", kv)
	}

	check := map[string]string{
		"prefix/e": "mustBeInTheMap",
	}
	if !reflect.DeepEqual(kv, check) {
		t.Fatalf("Expected %s\nGot %s", check, kv)
	}
}

func TestListRecursive5Levels(t *testing.T) {
	kv := &KvSource{
		&Mock{
			KVPairs: []*store.KVPair{
				{
					Key:   "prefix/l1",
					Value: []byte("level1"),
				},
				{
					Key:   "prefix/d1/l1",
					Value: []byte("level2"),
				},
				{
					Key:   "prefix/d1/l2",
					Value: []byte("level2"),
				},
				{
					Key:   "prefix/d2/d1/l1",
					Value: []byte("level3"),
				},
				{
					Key:   "prefix/d3/d2/d1/d1/d1",
					Value: []byte("level5"),
				},
			},
		},
		"prefix",
	}
	pairs := map[string][]byte{}
	err := kv.ListRecursive(kv.Prefix, pairs)
	if err != nil {
		t.Fatalf("Error : %s", err)
	}

	//check
	check := map[string][]byte{
		"prefix/l1":             []byte("level1"),
		"prefix/d1/l1":          []byte("level2"),
		"prefix/d1/l2":          []byte("level2"),
		"prefix/d2/d1/l1":       []byte("level3"),
		"prefix/d3/d2/d1/d1/d1": []byte("level5"),
	}
	if len(pairs) != len(check) {
		t.Fatalf("Expected length %d, got %d", len(check), len(pairs))
	}
	for k, v := range pairs {
		if !reflect.DeepEqual(v, check[k]) {
			t.Fatalf("Key %s\nExpected %s\nGot %s", k, check[k], v)
		}
	}
}

func TestListRecursiveEmpty(t *testing.T) {
	kv := &KvSource{
		&Mock{
			KVPairs: []*store.KVPair{},
		},
		"prefix",
	}
	pairs := map[string][]byte{}
	err := kv.ListRecursive(kv.Prefix, pairs)
	if err != nil {
		t.Fatalf("Error : %s", err)
	}

	//check
	check := map[string][]byte{}
	if len(pairs) != len(check) {
		t.Fatalf("Expected length %d, got %d", len(check), len(pairs))
	}
}

func TestConvertPairs5Levels(t *testing.T) {
	input := map[string][]byte{
		"prefix/l1":             []byte("level1"),
		"prefix/d1/l1":          []byte("level2"),
		"prefix/d1/l2":          []byte("level2"),
		"prefix/d2/d1/l1":       []byte("level3"),
		"prefix/d3/d2/d1/d1/d1": []byte("level5"),
	}
	//test
	output := convertPairs(input)

	//check
	check := map[string][]byte{
		"prefix/l1":             []byte("level1"),
		"prefix/d1/l1":          []byte("level2"),
		"prefix/d1/l2":          []byte("level2"),
		"prefix/d2/d1/l1":       []byte("level3"),
		"prefix/d3/d2/d1/d1/d1": []byte("level5"),
	}

	if len(output) != len(check) {
		t.Fatalf("Expected length %d, got %d", len(check), len(output))
	}
	for _, p := range output {
		if !reflect.DeepEqual(p.Value, check[p.Key]) {
			t.Fatalf("Key : %s\nExpected %s\nGot %s", p.Key, check[p.Key], p.Value)
		}
	}
}

func TestCollateKvPairsBase64(t *testing.T) {
	config := &struct {
		Base64Bytes []byte
	}{
		Base64Bytes: []byte("Testing automatic base64 if byte array"),
	}
	//test
	kv := map[string]string{}
	if err := collateKvRecursive(reflect.ValueOf(config), kv, "prefix"); err != nil {
		t.Fatalf("Error : %s", err)
	}
	//check

	check := map[string]string{
		"prefix/base64bytes": "VGVzdGluZyBhdXRvbWF0aWMgYmFzZTY0IGlmIGJ5dGUgYXJyYXk=",
	}
	if !reflect.DeepEqual(kv, check) {
		t.Fatalf("Expected %s\nGot %s", check, kv)
	}
}

type TestBase64Struct struct {
	Base64Bytes []byte
}

func TestParseKvSourceBase64(t *testing.T) {
	//Init
	config := TestBase64Struct{}

	//Test
	rootCmd := &flaeg.Command{
		Name:                  "test",
		Description:           "description test",
		Config:                &config,
		DefaultPointersConfig: &config,
		Run: func() error { return nil },
	}
	kv := &KvSource{
		&Mock{
			KVPairs: []*store.KVPair{
				{
					Key:   "test/base64bytes",
					Value: []byte("VGVzdGluZyBhdXRvbWF0aWMgYmFzZTY0IGlmIGJ5dGUgYXJyYXk="),
				},
			},
		},
		"test",
	}
	if _, err := kv.Parse(rootCmd); err != nil {
		t.Fatalf("Error %s", err)
	}

	//Check
	check := &TestBase64Struct{
		Base64Bytes: []byte("Testing automatic base64 if byte array"),
	}

	if !reflect.DeepEqual(check, rootCmd.Config) {
		printResult, err := json.Marshal(rootCmd.Config)
		if err != nil {
			t.Fatalf("error: %s", err)
		}
		printCheck, err := json.Marshal(check)
		if err != nil {
			t.Fatalf("error: %s", err)
		}
		t.Fatalf("\nexpected\t: %s\ngot\t\t\t: %s\n", printCheck, printResult)
	}
}

type CustomStruct struct {
	Bar1 string
	Bar2 string
}

// UnmarshalText define how unmarshal in TOML parsing
func (c *CustomStruct) UnmarshalText(text []byte) error {
	res := strings.Split(string(text), ",")
	c.Bar1 = res[0]
	c.Bar2 = res[1]
	return nil
}

// MarshalText encodes the receiver into UTF-8-encoded text and returns the result.
func (c *CustomStruct) MarshalText() (text []byte, err error) {
	return []byte(c.Bar1 + "," + c.Bar2), nil
}

func TestCollateCustomMarshaller(t *testing.T) {
	config := &CustomStruct{
		Bar1: "Bar1",
		Bar2: "Bar2",
	}
	//test
	kv := map[string]string{}
	if err := collateKvRecursive(reflect.ValueOf(config), kv, "prefix"); err != nil {
		t.Fatalf("Error : %s", err)
	}
	//check

	check := map[string]string{
		"prefix": "Bar1,Bar2",
	}
	if !reflect.DeepEqual(kv, check) {
		t.Fatalf("Expected %s\nGot %s", check, kv)
	}
}

func TestDecodeHookCustomMarshaller(t *testing.T) {
	data := &CustomStruct{
		Bar1: "Bar1",
		Bar2: "Bar2",
	}
	output, err := decodeHook(reflect.TypeOf([]string{}), reflect.TypeOf(data), "Bar1,Bar2")
	if err != nil {
		t.Fatalf("Error : %s", err)
	}

	if !reflect.DeepEqual(data, output) {
		t.Fatalf("Expected %#v\nGot %#v", data, output)
	}

}
