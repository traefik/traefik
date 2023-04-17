package dynamic

import (
	"encoding/json"
	"fmt"
	"reflect"
)

// +k8s:deepcopy-gen=false

// PluginConf holds the plugin configuration.
type PluginConf map[string]any

// DeepCopyInto is a deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *PluginConf) DeepCopyInto(out *PluginConf) {
	if in == nil {
		*out = nil
	} else {
		*out = deepCopyJSON(*in)
	}
}

// DeepCopy is a deepcopy function, copying the receiver, creating a new PluginConf.
func (in *PluginConf) DeepCopy() *PluginConf {
	if in == nil {
		return nil
	}
	out := new(PluginConf)
	in.DeepCopyInto(out)
	return out
}

// inspired by https://github.com/kubernetes/apimachinery/blob/53ecdf01b997ca93c7db7615dfe7b27ad8391983/pkg/runtime/converter.go#L607
func deepCopyJSON(x map[string]any) map[string]any {
	return deepCopyJSONValue(x).(map[string]any)
}

func deepCopyJSONValue(x any) any {
	switch x := x.(type) {
	case map[string]any:
		if x == nil {
			// Typed nil - an any that contains a type map[string]any with a value of nil
			return x
		}
		clone := make(map[string]any, len(x))
		for k, v := range x {
			clone[k] = deepCopyJSONValue(v)
		}
		return clone
	case []any:
		if x == nil {
			// Typed nil - an any that contains a type []any with a value of nil
			return x
		}
		clone := make([]any, len(x))
		for i, v := range x {
			clone[i] = deepCopyJSONValue(v)
		}
		return clone
	case string, int64, bool, float64, nil, json.Number:
		return x
	default:
		v := reflect.ValueOf(x)

		if v.NumMethod() == 0 {
			panic(fmt.Errorf("cannot deep copy %T", x))
		}

		method := v.MethodByName("DeepCopy")
		if method.Kind() == reflect.Invalid {
			panic(fmt.Errorf("cannot deep copy %T", x))
		}

		call := method.Call(nil)
		return call[0].Interface()
	}
}
