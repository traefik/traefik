package provider

import (
	"context"
	"maps"
	"reflect"
	"slices"
	"strings"

	"github.com/huandu/xstrings"
	"github.com/rs/zerolog/log"
	"github.com/traefik/traefik/v3/pkg/config/dynamic"
	"github.com/traefik/traefik/v3/pkg/observability/logs"
	"github.com/traefik/traefik/v3/pkg/tls"
)

var resourceLogFields = map[reflect.Type]string{
	reflect.TypeOf(dynamic.Router{}):              logs.RouterName,
	reflect.TypeOf(dynamic.Service{}):             logs.ServiceName,
	reflect.TypeOf(dynamic.Middleware{}):          logs.MiddlewareName,
	reflect.TypeOf(dynamic.ServersTransport{}):    logs.ServersTransportName,
	reflect.TypeOf(dynamic.TCPRouter{}):           logs.RouterName,
	reflect.TypeOf(dynamic.TCPService{}):          logs.ServiceName,
	reflect.TypeOf(dynamic.TCPMiddleware{}):       logs.MiddlewareName,
	reflect.TypeOf(dynamic.TCPServersTransport{}): logs.ServersTransportName,
	reflect.TypeOf(dynamic.UDPRouter{}):           logs.RouterName,
	reflect.TypeOf(dynamic.UDPService{}):          logs.ServiceName,
}

// Merge merges multiple configurations.
func Merge(ctx context.Context, configurations map[string]*dynamic.Configuration) *dynamic.Configuration {
	merged := &dynamic.Configuration{
		HTTP: &dynamic.HTTPConfiguration{
			Routers:           make(map[string]*dynamic.Router),
			Middlewares:       make(map[string]*dynamic.Middleware),
			Services:          make(map[string]*dynamic.Service),
			ServersTransports: make(map[string]*dynamic.ServersTransport),
		},
		TCP: &dynamic.TCPConfiguration{
			Routers:           make(map[string]*dynamic.TCPRouter),
			Services:          make(map[string]*dynamic.TCPService),
			Middlewares:       make(map[string]*dynamic.TCPMiddleware),
			ServersTransports: make(map[string]*dynamic.TCPServersTransport),
		},
		UDP: &dynamic.UDPConfiguration{
			Routers:  make(map[string]*dynamic.UDPRouter),
			Services: make(map[string]*dynamic.UDPService),
		},
		TLS: &dynamic.TLSConfiguration{
			Stores: make(map[string]tls.Store),
		},
	}

	tracker := newMergeTracker()

	providerNames := slices.Sorted(maps.Keys(configurations))
	for _, providerName := range providerNames {
		conf := configurations[providerName]

		if conf.HTTP != nil {
			mergeResourceMaps(reflect.ValueOf(merged.HTTP).Elem(), reflect.ValueOf(conf.HTTP).Elem(), providerName, tracker)
		}
		if conf.TCP != nil {
			mergeResourceMaps(reflect.ValueOf(merged.TCP).Elem(), reflect.ValueOf(conf.TCP).Elem(), providerName, tracker)
		}
		if conf.UDP != nil {
			mergeResourceMaps(reflect.ValueOf(merged.UDP).Elem(), reflect.ValueOf(conf.UDP).Elem(), providerName, tracker)
		}
		if conf.TLS != nil {
			mergeResourceMaps(reflect.ValueOf(merged.TLS).Elem(), reflect.ValueOf(conf.TLS).Elem(), providerName, tracker)
		}
	}

	deleteConflicts(ctx, tracker, resourceLogFields)

	return merged
}

// mergeResourceMaps merges all the resource maps defined in the provided struct.
// Conflicts are recorded in the given merge tracker.
func mergeResourceMaps(dst, src reflect.Value, providerName string, tracker *mergeTracker) {
	dstType := dst.Type()

	for i := 0; i < dstType.NumField(); i++ {
		field := dstType.Field(i)
		if !field.IsExported() {
			continue
		}

		dstField := dst.Field(i)
		srcField := src.Field(i)

		// Merge the resource maps of embedded structs.
		if field.Anonymous {
			mergeResourceMaps(dstField, srcField, providerName, tracker)
			continue
		}

		if dstField.Kind() == reflect.Map {
			mergeResourceMap(dstField, srcField, providerName, tracker)
		}
	}
}

// mergeResourceMap merges a resource map src into dst.
// New keys from src are added to dst.
// Duplicate keys are merged if the resource type implements a Merge method, otherwise
// the values must be identical. Conflicts are recorded in the given merge tracker.
func mergeResourceMap(dst, src reflect.Value, providerName string, tracker *mergeTracker) {
	if src.IsNil() {
		return
	}

	if dst.IsNil() {
		dst.Set(reflect.MakeMap(dst.Type()))
	}

	for _, key := range src.MapKeys() {
		keyStr := key.String()
		tracker.recordOrigin(dst, keyStr, providerName)

		srcValue := src.MapIndex(key)
		dstValue := dst.MapIndex(key)

		// Key doesn't exist in dst, add it.
		if !dstValue.IsValid() {
			dst.SetMapIndex(key, srcValue)
			continue
		}

		// Key exists, need to merge or detect conflict.
		if !tryMerge(dstValue, srcValue) {
			tracker.markForDeletion(dst, keyStr, dst.Type().Elem())
		}
	}
}

// tryMerge attempts to merge two resources.
// Returns true if the merge succeeds, false if values conflict.
func tryMerge(dst, src reflect.Value) bool {
	var dstActual, srcActual reflect.Value

	if dst.Kind() == reflect.Ptr {
		if dst.IsNil() || src.IsNil() {
			return reflect.DeepEqual(dst.Interface(), src.Interface())
		}

		dstActual = dst.Elem()
		srcActual = src.Elem()
	} else {
		dstActual = dst
		srcActual = src
	}

	// Check if the struct has the method `func (* T) Merge(other T) bool`.
	// We use reflection to detect this method because Go's type system doesn't allow type assertions
	// on generic interfaces (Mergeable[T]) practically.
	if dst.Kind() != reflect.Ptr {
		return reflect.DeepEqual(dst.Interface(), src.Interface())
	}

	mergeMethod := dst.MethodByName("Merge")
	if mergeMethod.IsValid() {
		methodType := mergeMethod.Type()
		if methodType.NumIn() == 1 && methodType.NumOut() == 1 && methodType.Out(0).Kind() == reflect.Bool {
			// Make sure the parameter type matches the type holding the method.
			if methodType.In(0).AssignableTo(src.Type()) {
				results := mergeMethod.Call([]reflect.Value{src})
				return results[0].Bool()
			}
		}
	}

	// When Merge is not implemented, merge is not allowed; the values must be the same.
	return reflect.DeepEqual(dstActual.Interface(), srcActual.Interface())
}

// deleteConflicts removes conflicting items and logs errors.
func deleteConflicts(ctx context.Context, tracker *mergeTracker, resourceLogFields map[reflect.Type]string) {
	logger := log.Ctx(ctx)

	for ck, info := range tracker.toDelete {
		origins := tracker.origins[ck]
		resourceType := typeName(info.elemType)
		keyField, ok := resourceLogFields[info.elemType]
		if !ok {
			keyField = xstrings.ToCamelCase(resourceType) + "Name"
		}

		typeWords := strings.ReplaceAll(xstrings.ToKebabCase(resourceType), "-", " ")

		logger.Error().
			Str(keyField, ck.key).
			Interface("configuration", origins).
			Msgf("%s defined multiple times with different configurations", xstrings.FirstRuneToUpper(typeWords))

		mapKey := reflect.ValueOf(ck.key)
		info.mapRef.SetMapIndex(mapKey, reflect.Value{})
	}
}

// typeName retrieves the type name.
func typeName(elemType reflect.Type) string {
	if elemType.Kind() == reflect.Ptr {
		elemType = elemType.Elem()
	}

	return elemType.Name()
}

// mergeTracker tracks item origins and items marked for deletion during merge.
type mergeTracker struct {
	toDelete map[conflictKey]conflictInfo
	origins  map[conflictKey][]string
}

// conflictKey uniquely identifies an entry in a map.
type conflictKey struct {
	mapPtr uintptr
	key    string
}

// conflictInfo stores information about a merge conflict.
type conflictInfo struct {
	mapRef   reflect.Value // The map to delete from.
	elemType reflect.Type
}

func newMergeTracker() *mergeTracker {
	return &mergeTracker{
		toDelete: make(map[conflictKey]conflictInfo),
		origins:  make(map[conflictKey][]string),
	}
}

func (t *mergeTracker) recordOrigin(mapRef reflect.Value, key, providerName string) {
	ck := conflictKey{mapPtr: mapRef.Pointer(), key: key}
	t.origins[ck] = append(t.origins[ck], providerName)
}

func (t *mergeTracker) markForDeletion(mapRef reflect.Value, key string, elemType reflect.Type) {
	ck := conflictKey{mapPtr: mapRef.Pointer(), key: key}
	t.toDelete[ck] = conflictInfo{
		mapRef:   mapRef,
		elemType: elemType,
	}
}
