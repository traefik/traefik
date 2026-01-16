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
	reflect.TypeFor[dynamic.Router]():              logs.RouterName,
	reflect.TypeFor[dynamic.Service]():             logs.ServiceName,
	reflect.TypeFor[dynamic.Middleware]():          logs.MiddlewareName,
	reflect.TypeFor[dynamic.ServersTransport]():    logs.ServersTransportName,
	reflect.TypeFor[dynamic.TCPRouter]():           logs.RouterName,
	reflect.TypeFor[dynamic.TCPService]():          logs.ServiceName,
	reflect.TypeFor[dynamic.TCPMiddleware]():       logs.MiddlewareName,
	reflect.TypeFor[dynamic.TCPServersTransport](): logs.ServersTransportName,
	reflect.TypeFor[dynamic.UDPRouter]():           logs.RouterName,
	reflect.TypeFor[dynamic.UDPService]():          logs.ServiceName,
}

// ResourceStrategy defines how the merge should handle resources.
type ResourceStrategy int

const (
	// ResourceStrategyMerge tries to call the Merge method on the resource.
	ResourceStrategyMerge ResourceStrategy = iota
	// ResourceStrategySkipDuplicates skips duplicate resources.
	ResourceStrategySkipDuplicates
)

// Merge merges multiple configurations.
func Merge(ctx context.Context, configurations map[string]*dynamic.Configuration, strategy ResourceStrategy) *dynamic.Configuration {
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

	origins := slices.Sorted(maps.Keys(configurations))
	for _, origin := range origins {
		conf := configurations[origin]

		if conf.HTTP != nil {
			mergeResourceMaps(ctx, reflect.ValueOf(merged.HTTP).Elem(), reflect.ValueOf(conf.HTTP).Elem(), origin, tracker, strategy, resourceLogFields)
		}
		if conf.TCP != nil {
			mergeResourceMaps(ctx, reflect.ValueOf(merged.TCP).Elem(), reflect.ValueOf(conf.TCP).Elem(), origin, tracker, strategy, resourceLogFields)
		}
		if conf.UDP != nil {
			mergeResourceMaps(ctx, reflect.ValueOf(merged.UDP).Elem(), reflect.ValueOf(conf.UDP).Elem(), origin, tracker, strategy, resourceLogFields)
		}
		if conf.TLS != nil {
			mergeResourceMaps(ctx, reflect.ValueOf(merged.TLS).Elem(), reflect.ValueOf(conf.TLS).Elem(), origin, tracker, strategy, resourceLogFields)

			merged.TLS.Certificates = mergeCertificates(ctx, merged.TLS.Certificates, conf.TLS.Certificates, origin, strategy)
		}
	}

	deleteConflicts(ctx, tracker, resourceLogFields)

	return merged
}

// mergeResourceMaps merges all the resource maps defined in the provided struct.
// Conflicts are recorded in the given merge tracker.
func mergeResourceMaps(ctx context.Context, dst, src reflect.Value, origin string, tracker *mergeTracker, strategy ResourceStrategy, resourceLogFields map[reflect.Type]string) {
	dstType := dst.Type()

	for i := range dstType.NumField() {
		field := dstType.Field(i)
		if !field.IsExported() {
			continue
		}

		dstField := dst.Field(i)
		srcField := src.Field(i)

		// Merge the resource maps of embedded structs.
		if field.Anonymous {
			mergeResourceMaps(ctx, dstField, srcField, origin, tracker, strategy, resourceLogFields)
			continue
		}

		if dstField.Kind() == reflect.Map {
			mergeResourceMap(ctx, dstField, srcField, origin, tracker, strategy, resourceLogFields)
		}
	}
}

// mergeResourceMap merges a resource map src into dst.
// New keys from src are added to dst.
// Duplicate keys are merged if the resource type implements a Merge method, otherwise
// the values must be identical. Conflicts are recorded in the given merge tracker.
func mergeResourceMap(ctx context.Context, dst, src reflect.Value, origin string, tracker *mergeTracker, strategy ResourceStrategy, resourceLogFields map[reflect.Type]string) {
	if src.IsNil() {
		return
	}

	if dst.IsNil() {
		dst.Set(reflect.MakeMap(dst.Type()))
	}

	for _, resourceKey := range src.MapKeys() {
		resourceKeyStr := resourceKey.String()
		tracker.recordOrigin(dst, resourceKeyStr, origin)

		srcValue := src.MapIndex(resourceKey)
		dstValue := dst.MapIndex(resourceKey)

		// Key doesn't exist in dst, add it.
		if !dstValue.IsValid() {
			dst.SetMapIndex(resourceKey, srcValue)
			continue
		}

		// Key exists, need to merge or detect conflict.
		switch strategy {
		case ResourceStrategyMerge:
			if !tryMerge(dstValue, srcValue) {
				tracker.markForDeletion(dst, resourceKeyStr, dst.Type().Elem())
			}
		case ResourceStrategySkipDuplicates:
			logSkippedDuplicate(ctx, dst.Type().Elem(), resourceKeyStr, origin, resourceLogFields)
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
		resourceNameField, resourceTypeWords := resourceLogMeta(info.resourceType, resourceLogFields)
		logger.Error().
			Str(resourceNameField, ck.resourceKey).
			Interface("configuration", tracker.origins[ck]).
			Msgf("%s defined multiple times with different configurations", xstrings.FirstRuneToUpper(resourceTypeWords))

		info.resourceMap.SetMapIndex(reflect.ValueOf(ck.resourceKey), reflect.Value{})
	}
}

// mergeCertificates merges multiple certificates.
func mergeCertificates(ctx context.Context, certificates []*tls.CertAndStores, newCertificates []*tls.CertAndStores, origin string, strategy ResourceStrategy) []*tls.CertAndStores {
	for _, certificate := range newCertificates {
		var found bool
		for _, existingCertificate := range certificates {
			if existingCertificate.Certificate == certificate.Certificate {
				found = true

				switch strategy {
				case ResourceStrategyMerge:
					existingCertificate.Stores = mergeStores(existingCertificate.Stores, certificate.Stores)
				case ResourceStrategySkipDuplicates:
					log.Ctx(ctx).Warn().
						Str("origin", origin).
						Msgf("TLS certificate %v already configured, skipping", certificate.Certificate)
				}

				break
			}
		}

		if !found {
			certificates = append(certificates, certificate)
		}
	}

	return certificates
}

// mergeStores merges two store slices, deduplicating entries while. Order is preserved.
func mergeStores(existing, other []string) []string {
	seen := make(map[string]struct{}, len(existing))
	for _, s := range existing {
		seen[s] = struct{}{}
	}

	for _, s := range other {
		if _, ok := seen[s]; !ok {
			existing = append(existing, s)
			seen[s] = struct{}{}
		}
	}

	return existing
}

// logSkippedDuplicate logs a warning when a duplicate resource is skipped.
func logSkippedDuplicate(ctx context.Context, resourceType reflect.Type, resourceKey, origin string, resourceLogFields map[reflect.Type]string) {
	resourceNameField, resourceTypeWords := resourceLogMeta(resourceType, resourceLogFields)

	log.Ctx(ctx).Warn().
		Str("origin", origin).
		Str(resourceNameField, resourceKey).
		Msgf("%s already configured, skipping", xstrings.FirstRuneToUpper(resourceTypeWords))
}

// resourceLogMeta returns the log field name and human-readable type description for the given resource element type.
func resourceLogMeta(resourceType reflect.Type, resourceLogFields map[reflect.Type]string) (resourceNameField, resourceTypeWords string) {
	if resourceType.Kind() == reflect.Ptr {
		resourceType = resourceType.Elem()
	}

	resourceTypeName := resourceType.Name()

	resourceNameField, ok := resourceLogFields[resourceType]
	if !ok {
		resourceNameField = xstrings.ToCamelCase(resourceTypeName) + "Name"
	}

	resourceTypeWords = strings.ReplaceAll(xstrings.ToKebabCase(resourceTypeName), "-", " ")

	return resourceNameField, resourceTypeWords
}

// mergeTracker tracks item origins and items marked for deletion during merge.
type mergeTracker struct {
	toDelete map[conflictKey]conflictInfo
	origins  map[conflictKey][]string
}

// conflictKey uniquely identifies an entry in a map.
type conflictKey struct {
	mapPtr      uintptr
	resourceKey string
}

// conflictInfo stores information about a merge conflict.
type conflictInfo struct {
	resourceMap  reflect.Value // The map to delete from.
	resourceType reflect.Type
}

func newMergeTracker() *mergeTracker {
	return &mergeTracker{
		toDelete: make(map[conflictKey]conflictInfo),
		origins:  make(map[conflictKey][]string),
	}
}

func (t *mergeTracker) recordOrigin(resourceMap reflect.Value, resourceKey, origin string) {
	ck := conflictKey{mapPtr: resourceMap.Pointer(), resourceKey: resourceKey}
	t.origins[ck] = append(t.origins[ck], origin)
}

func (t *mergeTracker) markForDeletion(resourceMap reflect.Value, resourceKey string, resourceType reflect.Type) {
	ck := conflictKey{mapPtr: resourceMap.Pointer(), resourceKey: resourceKey}
	t.toDelete[ck] = conflictInfo{
		resourceMap:  resourceMap,
		resourceType: resourceType,
	}
}
