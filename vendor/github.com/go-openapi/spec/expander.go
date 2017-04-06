// Copyright 2015 go-swagger maintainers
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package spec

import (
	"encoding/json"
	"fmt"
	"log"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"reflect"
	"strings"
	"sync"

	"github.com/go-openapi/jsonpointer"
	"github.com/go-openapi/swag"
)

var (
	// Debug enables logging when SWAGGER_DEBUG env var is not empty
	Debug = os.Getenv("SWAGGER_DEBUG") != ""
)

// ExpandOptions provides options for expand.
type ExpandOptions struct {
	RelativeBase string
	SkipSchemas  bool
}

// ResolutionCache a cache for resolving urls
type ResolutionCache interface {
	Get(string) (interface{}, bool)
	Set(string, interface{})
}

type simpleCache struct {
	lock  sync.Mutex
	store map[string]interface{}
}

var resCache ResolutionCache

func init() {
	resCache = initResolutionCache()
}

func initResolutionCache() ResolutionCache {
	return &simpleCache{store: map[string]interface{}{
		"http://swagger.io/v2/schema.json":       MustLoadSwagger20Schema(),
		"http://json-schema.org/draft-04/schema": MustLoadJSONSchemaDraft04(),
	}}
}

func (s *simpleCache) Get(uri string) (interface{}, bool) {
	debugLog("getting %q from resolution cache", uri)
	s.lock.Lock()
	v, ok := s.store[uri]
	debugLog("got %q from resolution cache: %t", uri, ok)

	s.lock.Unlock()
	return v, ok
}

func (s *simpleCache) Set(uri string, data interface{}) {
	s.lock.Lock()
	s.store[uri] = data
	s.lock.Unlock()
}

// ResolveRefWithBase resolves a reference against a context root with preservation of base path
func ResolveRefWithBase(root interface{}, ref *Ref, opts *ExpandOptions) (*Schema, error) {
	resolver, err := defaultSchemaLoader(root, nil, opts, nil)
	if err != nil {
		return nil, err
	}

	result := new(Schema)
	if err := resolver.Resolve(ref, result); err != nil {
		return nil, err
	}
	return result, nil
}

// ResolveRef resolves a reference against a context root
func ResolveRef(root interface{}, ref *Ref) (*Schema, error) {
	return ResolveRefWithBase(root, ref, nil)
}

// ResolveParameter resolves a paramter reference against a context root
func ResolveParameter(root interface{}, ref Ref) (*Parameter, error) {
	return ResolveParameterWithBase(root, ref, nil)
}

// ResolveParameterWithBase resolves a paramter reference against a context root and base path
func ResolveParameterWithBase(root interface{}, ref Ref, opts *ExpandOptions) (*Parameter, error) {
	resolver, err := defaultSchemaLoader(root, nil, opts, nil)
	if err != nil {
		return nil, err
	}

	result := new(Parameter)
	if err := resolver.Resolve(&ref, result); err != nil {
		return nil, err
	}
	return result, nil
}

// ResolveResponse resolves response a reference against a context root
func ResolveResponse(root interface{}, ref Ref) (*Response, error) {
	return ResolveResponseWithBase(root, ref, nil)
}

// ResolveResponseWithBase resolves response a reference against a context root and base path
func ResolveResponseWithBase(root interface{}, ref Ref, opts *ExpandOptions) (*Response, error) {
	resolver, err := defaultSchemaLoader(root, nil, opts, nil)
	if err != nil {
		return nil, err
	}

	result := new(Response)
	if err := resolver.Resolve(&ref, result); err != nil {
		return nil, err
	}
	return result, nil
}

// ResolveItems resolves header and parameter items reference against a context root and base path
func ResolveItems(root interface{}, ref Ref, opts *ExpandOptions) (*Items, error) {
	resolver, err := defaultSchemaLoader(root, nil, opts, nil)
	if err != nil {
		return nil, err
	}

	result := new(Items)
	if err := resolver.Resolve(&ref, result); err != nil {
		return nil, err
	}
	return result, nil
}

// ResolvePathItem resolves response a path item against a context root and base path
func ResolvePathItem(root interface{}, ref Ref, opts *ExpandOptions) (*PathItem, error) {
	resolver, err := defaultSchemaLoader(root, nil, opts, nil)
	if err != nil {
		return nil, err
	}

	result := new(PathItem)
	if err := resolver.Resolve(&ref, result); err != nil {
		return nil, err
	}
	return result, nil
}

type schemaLoader struct {
	loadingRef  *Ref
	startingRef *Ref
	currentRef  *Ref
	root        interface{}
	options     *ExpandOptions
	cache       ResolutionCache
	loadDoc     func(string) (json.RawMessage, error)
}

var idPtr, _ = jsonpointer.New("/id")
var refPtr, _ = jsonpointer.New("/$ref")

// PathLoader function to use when loading remote refs
var PathLoader func(string) (json.RawMessage, error)

func init() {
	PathLoader = func(path string) (json.RawMessage, error) {
		data, err := swag.LoadFromFileOrHTTP(path)
		if err != nil {
			return nil, err
		}
		return json.RawMessage(data), nil
	}
}

func defaultSchemaLoader(
	root interface{}, ref *Ref,
	expandOptions *ExpandOptions, cache ResolutionCache) (*schemaLoader, error) {

	if cache == nil {
		cache = resCache
	}
	if expandOptions == nil {
		expandOptions = &ExpandOptions{}
	}

	var ptr *jsonpointer.Pointer
	if ref != nil {
		ptr = ref.GetPointer()
	}

	currentRef := nextRef(root, ref, ptr)

	return &schemaLoader{
		loadingRef:  ref,
		startingRef: ref,
		currentRef:  currentRef,
		root:        root,
		options:     expandOptions,
		cache:       cache,
		loadDoc: func(path string) (json.RawMessage, error) {
			debugLog("fetching document at %q", path)
			return PathLoader(path)
		},
	}, nil
}

func idFromNode(node interface{}) (*Ref, error) {
	if idValue, _, err := idPtr.Get(node); err == nil {
		if refStr, ok := idValue.(string); ok && refStr != "" {
			idRef, err := NewRef(refStr)
			if err != nil {
				return nil, err
			}
			return &idRef, nil
		}
	}
	return nil, nil
}

func nextRef(startingNode interface{}, startingRef *Ref, ptr *jsonpointer.Pointer) *Ref {
	if startingRef == nil {
		return nil
	}

	if ptr == nil {
		return startingRef
	}

	ret := startingRef
	var idRef *Ref
	node := startingNode

	for _, tok := range ptr.DecodedTokens() {
		node, _, _ = jsonpointer.GetForToken(node, tok)
		if node == nil {
			break
		}

		idRef, _ = idFromNode(node)
		if idRef != nil {
			nw, err := ret.Inherits(*idRef)
			if err != nil {
				break
			}
			ret = nw
		}

		refRef, _, _ := refPtr.Get(node)
		if refRef != nil {
			var rf Ref
			switch value := refRef.(type) {
			case string:
				rf, _ = NewRef(value)
			}
			nw, err := ret.Inherits(rf)
			if err != nil {
				break
			}
			nwURL := nw.GetURL()
			if nwURL.Scheme == "file" || (nwURL.Scheme == "" && nwURL.Host == "") {
				nwpt := filepath.ToSlash(nwURL.Path)
				if filepath.IsAbs(nwpt) {
					_, err := os.Stat(nwpt)
					if err != nil {
						nwURL.Path = filepath.Join(".", nwpt)
					}
				}
			}

			ret = nw
		}

	}

	return ret
}

func debugLog(msg string, args ...interface{}) {
	if Debug {
		log.Printf(msg, args...)
	}
}

func normalizeFileRef(ref *Ref, relativeBase string) *Ref {
	refURL := ref.GetURL()
	debugLog("normalizing %s against %s", ref.String(), relativeBase)
	if strings.HasPrefix(refURL.String(), "#") {
		return ref
	}

	if refURL.Scheme == "file" || (refURL.Scheme == "" && refURL.Host == "") {
		filePath := refURL.Path
		debugLog("normalizing file path: %s", filePath)

		if !filepath.IsAbs(filepath.FromSlash(filePath)) && len(relativeBase) != 0 {
			debugLog("joining %s with %s", relativeBase, filePath)
			if fi, err := os.Stat(filepath.FromSlash(relativeBase)); err == nil {
				if !fi.IsDir() {
					relativeBase = path.Dir(relativeBase)
				}
			}
			filePath = filepath.Join(filepath.FromSlash(relativeBase), filepath.FromSlash(filePath))
		}
		if !filepath.IsAbs(filepath.FromSlash(filePath)) {
			pwd, err := os.Getwd()
			if err == nil {
				debugLog("joining cwd %s with %s", pwd, filePath)
				filePath = filepath.Join(pwd, filePath)
			}
		}

		debugLog("cleaning %s", filePath)
		filePath = filepath.Clean(filePath)
		_, err := os.Stat(filepath.FromSlash(filePath))
		if err == nil {
			debugLog("rewriting url to scheme \"\" path %s", filePath)
			refURL.Scheme = ""
			refURL.Path = filepath.ToSlash(filePath)
			debugLog("new url with joined filepath: %s", refURL.String())
			*ref = MustCreateRef(refURL.String())
		}
	}

	return ref
}

func (r *schemaLoader) resolveRef(currentRef, ref *Ref, node, target interface{}) error {

	tgt := reflect.ValueOf(target)
	if tgt.Kind() != reflect.Ptr {
		return fmt.Errorf("resolve ref: target needs to be a pointer")
	}

	oldRef := currentRef

	if currentRef != nil {
		debugLog("resolve ref current %s new %s", currentRef.String(), ref.String())
		nextRef := nextRef(node, ref, currentRef.GetPointer())
		if nextRef == nil || nextRef.GetURL() == nil {
			return nil
		}
		var err error
		currentRef, err = currentRef.Inherits(*nextRef)
		debugLog("resolved ref current %s", currentRef.String())
		if err != nil {
			return err
		}
	}

	if currentRef == nil {
		currentRef = ref
	}

	refURL := currentRef.GetURL()
	if refURL == nil {
		return nil
	}
	if currentRef.IsRoot() {
		nv := reflect.ValueOf(node)
		reflect.Indirect(tgt).Set(reflect.Indirect(nv))
		return nil
	}

	if strings.HasPrefix(refURL.String(), "#") {
		res, _, err := ref.GetPointer().Get(node)
		if err != nil {
			res, _, err = ref.GetPointer().Get(r.root)
			if err != nil {
				return err
			}
		}
		rv := reflect.Indirect(reflect.ValueOf(res))
		tgtType := reflect.Indirect(tgt).Type()
		if rv.Type().AssignableTo(tgtType) {
			reflect.Indirect(tgt).Set(reflect.Indirect(reflect.ValueOf(res)))
		} else {
			if err := swag.DynamicJSONToStruct(rv.Interface(), target); err != nil {
				return err
			}
		}

		return nil
	}

	relativeBase := ""
	if r.options != nil && r.options.RelativeBase != "" {
		relativeBase = r.options.RelativeBase
	}
	normalizeFileRef(currentRef, relativeBase)
	normalizeFileRef(ref, relativeBase)

	data, _, _, err := r.load(currentRef.GetURL())
	if err != nil {
		return err
	}

	if ((oldRef == nil && currentRef != nil) ||
		(oldRef != nil && currentRef == nil) ||
		oldRef.String() != currentRef.String()) &&
		((oldRef == nil && ref != nil) ||
			(oldRef != nil && ref == nil) ||
			(oldRef.String() != ref.String())) {

		return r.resolveRef(currentRef, ref, data, target)
	}

	var res interface{}
	if currentRef.String() != "" {
		res, _, err = currentRef.GetPointer().Get(data)
		if err != nil {
			if strings.HasPrefix(ref.String(), "#") {
				if r.loadingRef != nil {
					rr, er := r.loadingRef.Inherits(*ref)
					if er != nil {
						return er
					}
					refURL = rr.GetURL()

					data, _, _, err = r.load(refURL)
					if err != nil {
						return err
					}
				} else {
					data = r.root
				}
			}

			res, _, err = ref.GetPointer().Get(data)
			if err != nil {
				return err
			}
		}
	} else {
		res = data
	}

	if err := swag.DynamicJSONToStruct(res, target); err != nil {
		return err
	}

	r.currentRef = currentRef

	return nil
}

func (r *schemaLoader) load(refURL *url.URL) (interface{}, url.URL, bool, error) {
	debugLog("loading schema from url: %s", refURL)
	toFetch := *refURL
	toFetch.Fragment = ""

	data, fromCache := r.cache.Get(toFetch.String())
	if !fromCache {
		b, err := r.loadDoc(toFetch.String())
		if err != nil {
			return nil, url.URL{}, false, err
		}

		if err := json.Unmarshal(b, &data); err != nil {
			return nil, url.URL{}, false, err
		}
		r.cache.Set(toFetch.String(), data)
	}

	return data, toFetch, fromCache, nil
}

func (r *schemaLoader) Resolve(ref *Ref, target interface{}) error {
	return r.resolveRef(r.currentRef, ref, r.root, target)
}

// ExpandSpec expands the references in a swagger spec
func ExpandSpec(spec *Swagger, options *ExpandOptions) error {
	resolver, err := defaultSchemaLoader(spec, nil, options, nil)
	if err != nil {
		return err
	}

	if options == nil || !options.SkipSchemas {
		for key, definition := range spec.Definitions {
			var def *Schema
			var err error
			if def, err = expandSchema(definition, []string{"#/definitions/" + key}, resolver); err != nil {
				return err
			}
			spec.Definitions[key] = *def
		}
	}

	for key, parameter := range spec.Parameters {
		if err := expandParameter(&parameter, resolver); err != nil {
			return err
		}
		spec.Parameters[key] = parameter
	}

	for key, response := range spec.Responses {
		if err := expandResponse(&response, resolver); err != nil {
			return err
		}
		spec.Responses[key] = response
	}

	if spec.Paths != nil {
		for key, path := range spec.Paths.Paths {
			if err := expandPathItem(&path, resolver); err != nil {
				return err
			}
			spec.Paths.Paths[key] = path
		}
	}

	return nil
}

// ExpandSchema expands the refs in the schema object
func ExpandSchema(schema *Schema, root interface{}, cache ResolutionCache) error {
	return ExpandSchemaWithBasePath(schema, root, cache, nil)
}

// ExpandSchemaWithBasePath expands the refs in the schema object, base path configured through expand options
func ExpandSchemaWithBasePath(schema *Schema, root interface{}, cache ResolutionCache, opts *ExpandOptions) error {
	if schema == nil {
		return nil
	}
	if root == nil {
		root = schema
	}

	nrr, _ := NewRef(schema.ID)
	var rrr *Ref
	if nrr.String() != "" {
		switch rt := root.(type) {
		case *Schema:
			rid, _ := NewRef(rt.ID)
			rrr, _ = rid.Inherits(nrr)
		case *Swagger:
			rid, _ := NewRef(rt.ID)
			rrr, _ = rid.Inherits(nrr)
		}
	}

	resolver, err := defaultSchemaLoader(root, rrr, opts, cache)
	if err != nil {
		return err
	}

	refs := []string{""}
	if rrr != nil {
		refs[0] = rrr.String()
	}
	var s *Schema
	if s, err = expandSchema(*schema, refs, resolver); err != nil {
		return err
	}
	*schema = *s
	return nil
}

func expandItems(target Schema, parentRefs []string, resolver *schemaLoader) (*Schema, error) {
	if target.Items != nil {
		if target.Items.Schema != nil {
			t, err := expandSchema(*target.Items.Schema, parentRefs, resolver)
			if err != nil {
				if target.Items.Schema.ID == "" {
					target.Items.Schema.ID = target.ID
					if err != nil {
						t, err = expandSchema(*target.Items.Schema, parentRefs, resolver)
						if err != nil {
							return nil, err
						}
					}
				}
			}
			*target.Items.Schema = *t
		}
		for i := range target.Items.Schemas {
			t, err := expandSchema(target.Items.Schemas[i], parentRefs, resolver)
			if err != nil {
				return nil, err
			}
			target.Items.Schemas[i] = *t
		}
	}
	return &target, nil
}

func expandSchema(target Schema, parentRefs []string, resolver *schemaLoader) (*Schema, error) {
	if target.Ref.String() == "" && target.Ref.IsRoot() {
		debugLog("skipping expand schema for no ref and root: %v", resolver.root)

		return resolver.root.(*Schema), nil
	}

	// t is the new expanded schema
	var t *Schema

	for target.Ref.String() != "" {
		if swag.ContainsStringsCI(parentRefs, target.Ref.String()) {
			return &target, nil
		}

		if err := resolver.Resolve(&target.Ref, &t); err != nil {
			return &target, err
		}

		parentRefs = append(parentRefs, target.Ref.String())
		target = *t
	}

	t, err := expandItems(target, parentRefs, resolver)
	if err != nil {
		return &target, err
	}
	target = *t

	for i := range target.AllOf {
		t, err := expandSchema(target.AllOf[i], parentRefs, resolver)
		if err != nil {
			return &target, err
		}
		target.AllOf[i] = *t
	}
	for i := range target.AnyOf {
		t, err := expandSchema(target.AnyOf[i], parentRefs, resolver)
		if err != nil {
			return &target, err
		}
		target.AnyOf[i] = *t
	}
	for i := range target.OneOf {
		t, err := expandSchema(target.OneOf[i], parentRefs, resolver)
		if err != nil {
			return &target, err
		}
		target.OneOf[i] = *t
	}
	if target.Not != nil {
		t, err := expandSchema(*target.Not, parentRefs, resolver)
		if err != nil {
			return &target, err
		}
		*target.Not = *t
	}
	for k := range target.Properties {
		t, err := expandSchema(target.Properties[k], parentRefs, resolver)
		if err != nil {
			return &target, err
		}
		target.Properties[k] = *t
	}
	if target.AdditionalProperties != nil && target.AdditionalProperties.Schema != nil {
		t, err := expandSchema(*target.AdditionalProperties.Schema, parentRefs, resolver)
		if err != nil {
			return &target, err
		}
		*target.AdditionalProperties.Schema = *t
	}
	for k := range target.PatternProperties {
		t, err := expandSchema(target.PatternProperties[k], parentRefs, resolver)
		if err != nil {
			return &target, err
		}
		target.PatternProperties[k] = *t
	}
	for k := range target.Dependencies {
		if target.Dependencies[k].Schema != nil {
			t, err := expandSchema(*target.Dependencies[k].Schema, parentRefs, resolver)
			if err != nil {
				return &target, err
			}
			*target.Dependencies[k].Schema = *t
		}
	}
	if target.AdditionalItems != nil && target.AdditionalItems.Schema != nil {
		t, err := expandSchema(*target.AdditionalItems.Schema, parentRefs, resolver)
		if err != nil {
			return &target, err
		}
		*target.AdditionalItems.Schema = *t
	}
	for k := range target.Definitions {
		t, err := expandSchema(target.Definitions[k], parentRefs, resolver)
		if err != nil {
			return &target, err
		}
		target.Definitions[k] = *t
	}
	return &target, nil
}

func expandPathItem(pathItem *PathItem, resolver *schemaLoader) error {
	if pathItem == nil {
		return nil
	}
	if pathItem.Ref.String() != "" {
		if err := resolver.Resolve(&pathItem.Ref, &pathItem); err != nil {
			return err
		}
	}

	for idx := range pathItem.Parameters {
		if err := expandParameter(&(pathItem.Parameters[idx]), resolver); err != nil {
			return err
		}
	}
	if err := expandOperation(pathItem.Get, resolver); err != nil {
		return err
	}
	if err := expandOperation(pathItem.Head, resolver); err != nil {
		return err
	}
	if err := expandOperation(pathItem.Options, resolver); err != nil {
		return err
	}
	if err := expandOperation(pathItem.Put, resolver); err != nil {
		return err
	}
	if err := expandOperation(pathItem.Post, resolver); err != nil {
		return err
	}
	if err := expandOperation(pathItem.Patch, resolver); err != nil {
		return err
	}
	if err := expandOperation(pathItem.Delete, resolver); err != nil {
		return err
	}
	return nil
}

func expandOperation(op *Operation, resolver *schemaLoader) error {
	if op == nil {
		return nil
	}
	for i, param := range op.Parameters {
		if err := expandParameter(&param, resolver); err != nil {
			return err
		}
		op.Parameters[i] = param
	}

	if op.Responses != nil {
		responses := op.Responses
		if err := expandResponse(responses.Default, resolver); err != nil {
			return err
		}
		for code, response := range responses.StatusCodeResponses {
			if err := expandResponse(&response, resolver); err != nil {
				return err
			}
			responses.StatusCodeResponses[code] = response
		}
	}
	return nil
}

func expandResponse(response *Response, resolver *schemaLoader) error {
	if response == nil {
		return nil
	}

	var parentRefs []string
	if response.Ref.String() != "" {
		parentRefs = append(parentRefs, response.Ref.String())
		if err := resolver.Resolve(&response.Ref, response); err != nil {
			return err
		}
	}

	if !resolver.options.SkipSchemas && response.Schema != nil {
		parentRefs = append(parentRefs, response.Schema.Ref.String())
		debugLog("response ref: %s", response.Schema.Ref)
		if err := resolver.Resolve(&response.Schema.Ref, &response.Schema); err != nil {
			return err
		}
		s, err := expandSchema(*response.Schema, parentRefs, resolver)
		if err != nil {
			return err
		}
		*response.Schema = *s
	}
	return nil
}

func expandParameter(parameter *Parameter, resolver *schemaLoader) error {
	if parameter == nil {
		return nil
	}

	var parentRefs []string
	if parameter.Ref.String() != "" {
		parentRefs = append(parentRefs, parameter.Ref.String())
		if err := resolver.Resolve(&parameter.Ref, parameter); err != nil {
			return err
		}
	}
	if !resolver.options.SkipSchemas && parameter.Schema != nil {
		parentRefs = append(parentRefs, parameter.Schema.Ref.String())
		if err := resolver.Resolve(&parameter.Schema.Ref, &parameter.Schema); err != nil {
			return err
		}
		s, err := expandSchema(*parameter.Schema, parentRefs, resolver)
		if err != nil {
			return err
		}
		*parameter.Schema = *s
	}
	return nil
}
