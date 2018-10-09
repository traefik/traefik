// -*- Mode: Go; indent-tabs-mode: t -*-

/*
 * Copyright (C) 2015-2017 Canonical Ltd
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU General Public License version 3 as
 * published by the Free Software Foundation.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program.  If not, see <http://www.gnu.org/licenses/>.
 *
 */

package asserts

import (
	"errors"
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"
)

// AttrMatchContext has contextual helpers for evaluating attribute constraints.
type AttrMatchContext interface {
	PlugAttr(arg string) (interface{}, error)
	SlotAttr(arg string) (interface{}, error)
}

const (
	// feature label for $SLOT()/$PLUG()/$MISSING
	dollarAttrConstraintsFeature = "dollar-attr-constraints"
	// feature label for on-store/on-brand/on-model
	deviceScopeConstraintsFeature = "device-scope-constraints"
)

type attrMatcher interface {
	match(apath string, v interface{}, ctx AttrMatchContext) error

	feature(flabel string) bool
}

func chain(path, k string) string {
	if path == "" {
		return k
	}
	return fmt.Sprintf("%s.%s", path, k)
}

type compileContext struct {
	dotted string
	hadMap bool
	wasAlt bool
}

func (cc compileContext) String() string {
	return cc.dotted
}

func (cc compileContext) keyEntry(k string) compileContext {
	return compileContext{
		dotted: chain(cc.dotted, k),
		hadMap: true,
		wasAlt: false,
	}
}

func (cc compileContext) alt(alt int) compileContext {
	return compileContext{
		dotted: fmt.Sprintf("%s/alt#%d/", cc.dotted, alt+1),
		hadMap: cc.hadMap,
		wasAlt: true,
	}
}

// compileAttrMatcher compiles an attrMatcher derived from constraints,
func compileAttrMatcher(cc compileContext, constraints interface{}) (attrMatcher, error) {
	switch x := constraints.(type) {
	case map[string]interface{}:
		return compileMapAttrMatcher(cc, x)
	case []interface{}:
		if cc.wasAlt {
			return nil, fmt.Errorf("cannot nest alternative constraints directly at %q", cc)
		}
		return compileAltAttrMatcher(cc, x)
	case string:
		if !cc.hadMap {
			return nil, fmt.Errorf("first level of non alternative constraints must be a set of key-value contraints")
		}
		if strings.HasPrefix(x, "$") {
			if x == "$MISSING" {
				return missingAttrMatcher{}, nil
			}
			return compileEvalAttrMatcher(cc, x)
		}
		return compileRegexpAttrMatcher(cc, x)
	default:
		return nil, fmt.Errorf("constraint %q must be a key-value map, regexp or a list of alternative constraints: %v", cc, x)
	}
}

type mapAttrMatcher map[string]attrMatcher

func compileMapAttrMatcher(cc compileContext, m map[string]interface{}) (attrMatcher, error) {
	matcher := make(mapAttrMatcher)
	for k, constraint := range m {
		matcher1, err := compileAttrMatcher(cc.keyEntry(k), constraint)
		if err != nil {
			return nil, err
		}
		matcher[k] = matcher1
	}
	return matcher, nil
}

func matchEntry(apath, k string, matcher1 attrMatcher, v interface{}, ctx AttrMatchContext) error {
	apath = chain(apath, k)
	// every entry matcher expects the attribute to be set except for $MISSING
	if _, ok := matcher1.(missingAttrMatcher); !ok && v == nil {
		return fmt.Errorf("attribute %q has constraints but is unset", apath)
	}
	if err := matcher1.match(apath, v, ctx); err != nil {
		return err
	}
	return nil
}

func matchList(apath string, matcher attrMatcher, l []interface{}, ctx AttrMatchContext) error {
	for i, elem := range l {
		if err := matcher.match(chain(apath, strconv.Itoa(i)), elem, ctx); err != nil {
			return err
		}
	}
	return nil
}

func (matcher mapAttrMatcher) feature(flabel string) bool {
	for _, matcher1 := range matcher {
		if matcher1.feature(flabel) {
			return true
		}
	}
	return false
}

func (matcher mapAttrMatcher) match(apath string, v interface{}, ctx AttrMatchContext) error {
	switch x := v.(type) {
	case Attrer:
		// we get Atter from root-level Check (apath is "")
		for k, matcher1 := range matcher {
			v, _ := x.Lookup(k)
			if err := matchEntry("", k, matcher1, v, ctx); err != nil {
				return err
			}
		}
	case map[string]interface{}: // maps in attributes look like this
		for k, matcher1 := range matcher {
			if err := matchEntry(apath, k, matcher1, x[k], ctx); err != nil {
				return err
			}
		}
	case []interface{}:
		return matchList(apath, matcher, x, ctx)
	default:
		return fmt.Errorf("attribute %q must be a map", apath)
	}
	return nil
}

type missingAttrMatcher struct{}

func (matcher missingAttrMatcher) feature(flabel string) bool {
	return flabel == dollarAttrConstraintsFeature
}

func (matcher missingAttrMatcher) match(apath string, v interface{}, ctx AttrMatchContext) error {
	if v != nil {
		return fmt.Errorf("attribute %q is constrained to be missing but is set", apath)
	}
	return nil
}

type evalAttrMatcher struct {
	// first iteration supports just $(SLOT|PLUG)(arg)
	op  string
	arg string
}

var (
	validEvalAttrMatcher = regexp.MustCompile(`^\$(SLOT|PLUG)\((.+)\)$`)
)

func compileEvalAttrMatcher(cc compileContext, s string) (attrMatcher, error) {
	ops := validEvalAttrMatcher.FindStringSubmatch(s)
	if len(ops) == 0 {
		return nil, fmt.Errorf("cannot compile %q constraint %q: not a valid $SLOT()/$PLUG() constraint", cc, s)
	}
	return evalAttrMatcher{
		op:  ops[1],
		arg: ops[2],
	}, nil
}

func (matcher evalAttrMatcher) feature(flabel string) bool {
	return flabel == dollarAttrConstraintsFeature
}

func (matcher evalAttrMatcher) match(apath string, v interface{}, ctx AttrMatchContext) error {
	if ctx == nil {
		return fmt.Errorf("attribute %q cannot be matched without context", apath)
	}
	var comp func(string) (interface{}, error)
	switch matcher.op {
	case "SLOT":
		comp = ctx.SlotAttr
	case "PLUG":
		comp = ctx.PlugAttr
	}
	v1, err := comp(matcher.arg)
	if err != nil {
		return fmt.Errorf("attribute %q constraint $%s(%s) cannot be evaluated: %v", apath, matcher.op, matcher.arg, err)
	}
	if !reflect.DeepEqual(v, v1) {
		return fmt.Errorf("attribute %q does not match $%s(%s): %v != %v", apath, matcher.op, matcher.arg, v, v1)
	}
	return nil
}

type regexpAttrMatcher struct {
	*regexp.Regexp
}

func compileRegexpAttrMatcher(cc compileContext, s string) (attrMatcher, error) {
	rx, err := regexp.Compile("^(" + s + ")$")
	if err != nil {
		return nil, fmt.Errorf("cannot compile %q constraint %q: %v", cc, s, err)
	}
	return regexpAttrMatcher{rx}, nil
}

func (matcher regexpAttrMatcher) feature(flabel string) bool {
	return false
}

func (matcher regexpAttrMatcher) match(apath string, v interface{}, ctx AttrMatchContext) error {
	var s string
	switch x := v.(type) {
	case string:
		s = x
	case bool:
		s = strconv.FormatBool(x)
	case int64:
		s = strconv.FormatInt(x, 10)
	case []interface{}:
		return matchList(apath, matcher, x, ctx)
	default:
		return fmt.Errorf("attribute %q must be a scalar or list", apath)
	}
	if !matcher.Regexp.MatchString(s) {
		return fmt.Errorf("attribute %q value %q does not match %v", apath, s, matcher.Regexp)
	}
	return nil

}

type altAttrMatcher struct {
	alts []attrMatcher
}

func compileAltAttrMatcher(cc compileContext, l []interface{}) (attrMatcher, error) {
	alts := make([]attrMatcher, len(l))
	for i, constraint := range l {
		matcher1, err := compileAttrMatcher(cc.alt(i), constraint)
		if err != nil {
			return nil, err
		}
		alts[i] = matcher1
	}
	return altAttrMatcher{alts}, nil

}

func (matcher altAttrMatcher) feature(flabel string) bool {
	for _, alt := range matcher.alts {
		if alt.feature(flabel) {
			return true
		}
	}
	return false
}

func (matcher altAttrMatcher) match(apath string, v interface{}, ctx AttrMatchContext) error {
	var firstErr error
	for _, alt := range matcher.alts {
		err := alt.match(apath, v, ctx)
		if err == nil {
			return nil
		}
		if firstErr == nil {
			firstErr = err
		}
	}
	apathDescr := ""
	if apath != "" {
		apathDescr = fmt.Sprintf(" for attribute %q", apath)
	}
	return fmt.Errorf("no alternative%s matches: %v", apathDescr, firstErr)
}

// AttributeConstraints implements a set of constraints on the attributes of a slot or plug.
type AttributeConstraints struct {
	matcher attrMatcher
}

func (ac *AttributeConstraints) feature(flabel string) bool {
	return ac.matcher.feature(flabel)
}

// compileAttributeConstraints checks and compiles a mapping or list
// from the assertion format into AttributeConstraints.
func compileAttributeConstraints(constraints interface{}) (*AttributeConstraints, error) {
	matcher, err := compileAttrMatcher(compileContext{}, constraints)
	if err != nil {
		return nil, err
	}
	return &AttributeConstraints{matcher: matcher}, nil
}

type fixedAttrMatcher struct {
	result error
}

func (matcher fixedAttrMatcher) feature(flabel string) bool {
	return false
}

func (matcher fixedAttrMatcher) match(apath string, v interface{}, ctx AttrMatchContext) error {
	return matcher.result
}

var (
	AlwaysMatchAttributes = &AttributeConstraints{matcher: fixedAttrMatcher{nil}}
	NeverMatchAttributes  = &AttributeConstraints{matcher: fixedAttrMatcher{errors.New("not allowed")}}
)

// Attrer reflects part of the Attrer interface (see interfaces.Attrer).
type Attrer interface {
	Lookup(path string) (interface{}, bool)
}

// Check checks whether attrs don't match the constraints.
func (c *AttributeConstraints) Check(attrer Attrer, ctx AttrMatchContext) error {
	return c.matcher.match("", attrer, ctx)
}

// OnClassicConstraint specifies a constraint based whether the system is classic and optional specific distros' sets.
type OnClassicConstraint struct {
	Classic   bool
	SystemIDs []string
}

// DeviceScopeConstraint specifies a constraints based on which brand
// store, brand or model the device belongs to.
type DeviceScopeConstraint struct {
	Store []string
	Brand []string
	// Model is a list of precise "<brand>/<model>" constraints
	Model []string
}

var (
	validStoreID         = regexp.MustCompile("^[-A-Z0-9a-z_]+$")
	validBrandSlashModel = regexp.MustCompile("^(" +
		strings.Trim(validAccountID.String(), "^$") +
		")/(" +
		strings.Trim(validModel.String(), "^$") +
		")$")
	deviceScopeConstraints = map[string]*regexp.Regexp{
		"on-store": validStoreID,
		"on-brand": validAccountID,
		// on-model constraints are of the form list of
		// <brand>/<model> strings where <brand> are account
		// IDs as they appear in the respective model assertion
		"on-model": validBrandSlashModel,
	}
)

func detectDeviceScopeConstraint(cMap map[string]interface{}) bool {
	// for consistency and simplicity we support all of on-store,
	// on-brand, and on-model to appear together. The interpretation
	// layer will AND them as usual
	for field := range deviceScopeConstraints {
		if cMap[field] != nil {
			return true
		}
	}
	return false
}

func compileDeviceScopeConstraint(cMap map[string]interface{}, context string) (constr *DeviceScopeConstraint, err error) {
	// initial map size of 2: we expect usual cases to have just one of the
	// constraints or rarely 2
	deviceConstr := make(map[string][]string, 2)
	for field, validRegexp := range deviceScopeConstraints {
		vals, err := checkStringListInMap(cMap, field, fmt.Sprintf("%s in %s", field, context), validRegexp)
		if err != nil {
			return nil, err
		}
		deviceConstr[field] = vals
	}

	if len(deviceConstr) == 0 {
		return nil, fmt.Errorf("internal error: misdetected device scope constraints in %s", context)
	}
	return &DeviceScopeConstraint{
		Store: deviceConstr["on-store"],
		Brand: deviceConstr["on-brand"],
		Model: deviceConstr["on-model"],
	}, nil
}

// rules

var (
	validSnapType  = regexp.MustCompile("^(?:core|kernel|gadget|app)$")
	validDistro    = regexp.MustCompile("^[-0-9a-z._]+$")
	validSnapID    = regexp.MustCompile("^[a-z0-9A-Z]{32}$")                                        // snap-ids look like this
	validPublisher = regexp.MustCompile("^(?:[a-z0-9A-Z]{32}|[-a-z0-9]{2,28}|\\$[A-Z][A-Z0-9_]*)$") // account ids look like snap-ids or are nice identifiers, support our own special markers $MARKER

	validIDConstraints = map[string]*regexp.Regexp{
		"slot-snap-type":    validSnapType,
		"slot-snap-id":      validSnapID,
		"slot-publisher-id": validPublisher,
		"plug-snap-type":    validSnapType,
		"plug-snap-id":      validSnapID,
		"plug-publisher-id": validPublisher,
	}
)

func checkMapOrShortcut(context string, v interface{}) (m map[string]interface{}, invert bool, err error) {
	switch x := v.(type) {
	case map[string]interface{}:
		return x, false, nil
	case string:
		switch x {
		case "true":
			return nil, false, nil
		case "false":
			return nil, true, nil
		}
	}
	return nil, false, errors.New("unexpected type")
}

type constraintsHolder interface {
	setAttributeConstraints(field string, cstrs *AttributeConstraints)
	setIDConstraints(field string, cstrs []string)
	setOnClassicConstraint(onClassic *OnClassicConstraint)
	setDeviceScopeConstraint(deviceScope *DeviceScopeConstraint)
}

func baseCompileConstraints(context string, cDef constraintsDef, target constraintsHolder, attrConstraints, idConstraints []string) error {
	cMap := cDef.cMap
	if cMap == nil {
		fixed := AlwaysMatchAttributes // "true"
		if cDef.invert {               // "false"
			fixed = NeverMatchAttributes
		}
		for _, field := range attrConstraints {
			target.setAttributeConstraints(field, fixed)
		}
		return nil
	}
	defaultUsed := 0
	for _, field := range idConstraints {
		lst, err := checkStringListInMap(cMap, field, fmt.Sprintf("%s in %s", field, context), validIDConstraints[field])
		if err != nil {
			return err
		}
		if lst == nil {
			defaultUsed++
		}
		target.setIDConstraints(field, lst)
	}
	for _, field := range attrConstraints {
		cstrs := AlwaysMatchAttributes
		v := cMap[field]
		if v != nil {
			var err error
			cstrs, err = compileAttributeConstraints(cMap[field])
			if err != nil {
				return fmt.Errorf("cannot compile %s in %s: %v", field, context, err)
			}
		} else {
			defaultUsed++
		}
		target.setAttributeConstraints(field, cstrs)
	}
	onClassic := cMap["on-classic"]
	if onClassic == nil {
		defaultUsed++
	} else {
		var c *OnClassicConstraint
		switch x := onClassic.(type) {
		case string:
			switch x {
			case "true":
				c = &OnClassicConstraint{Classic: true}
			case "false":
				c = &OnClassicConstraint{Classic: false}
			}
		case []interface{}:
			lst, err := checkStringListInMap(cMap, "on-classic", fmt.Sprintf("on-classic in %s", context), validDistro)
			if err != nil {
				return err
			}
			c = &OnClassicConstraint{Classic: true, SystemIDs: lst}
		}
		if c == nil {
			return fmt.Errorf("on-classic in %s must be 'true', 'false' or a list of operating system IDs", context)
		}
		target.setOnClassicConstraint(c)
	}
	if !detectDeviceScopeConstraint(cMap) {
		defaultUsed++
	} else {
		c, err := compileDeviceScopeConstraint(cMap, context)
		if err != nil {
			return err
		}
		target.setDeviceScopeConstraint(c)
	}
	// checks whether defaults have been used for everything, which is not
	// well-formed
	// +1+1 accounts for defaults for missing on-classic plus missing
	// on-store/on-brand/on-model
	if defaultUsed == len(attributeConstraints)+len(idConstraints)+1+1 {
		return fmt.Errorf("%s must specify at least one of %s, %s, on-classic, on-store, on-brand, on-model", context, strings.Join(attrConstraints, ", "), strings.Join(idConstraints, ", "))
	}
	return nil
}

type rule interface {
	setConstraints(field string, cstrs []constraintsHolder)
}

type constraintsDef struct {
	cMap   map[string]interface{}
	invert bool
}

type subruleCompiler func(context string, def constraintsDef) (constraintsHolder, error)

func baseCompileRule(context string, rule interface{}, target rule, subrules []string, compilers map[string]subruleCompiler, defaultOutcome, invertedOutcome map[string]interface{}) error {
	rMap, invert, err := checkMapOrShortcut(context, rule)
	if err != nil {
		return fmt.Errorf("%s must be a map or one of the shortcuts 'true' or 'false'", context)
	}
	if rMap == nil {
		rMap = defaultOutcome // "true"
		if invert {
			rMap = invertedOutcome // "false"
		}
	}
	defaultUsed := 0
	// compile and set subrules
	for _, subrule := range subrules {
		v := rMap[subrule]
		var lst []interface{}
		alternatives := false
		switch x := v.(type) {
		case nil:
			v = defaultOutcome[subrule]
			defaultUsed++
		case []interface{}:
			alternatives = true
			lst = x
		}
		if lst == nil { // v is map or a string, checked below
			lst = []interface{}{v}
		}
		compiler := compilers[subrule]
		if compiler == nil {
			panic(fmt.Sprintf("no compiler for %s in %s", subrule, context))
		}
		alts := make([]constraintsHolder, len(lst))
		for i, alt := range lst {
			subctxt := fmt.Sprintf("%s in %s", subrule, context)
			if alternatives {
				subctxt = fmt.Sprintf("alternative %d of %s", i+1, subctxt)
			}
			cMap, invert, err := checkMapOrShortcut(subctxt, alt)
			if err != nil || (cMap == nil && alternatives) {
				efmt := "%s must be a map"
				if !alternatives {
					efmt = "%s must be a map or one of the shortcuts 'true' or 'false'"
				}
				return fmt.Errorf(efmt, subctxt)
			}

			cstrs, err := compiler(subctxt, constraintsDef{
				cMap:   cMap,
				invert: invert,
			})
			if err != nil {
				return err
			}
			alts[i] = cstrs
		}
		target.setConstraints(subrule, alts)
	}
	if defaultUsed == len(subrules) {
		return fmt.Errorf("%s must specify at least one of %s", context, strings.Join(subrules, ", "))
	}
	return nil
}

// PlugRule holds the rule of what is allowed, wrt installation and
// connection, for a plug of a specific interface for a snap.
type PlugRule struct {
	Interface string

	AllowInstallation []*PlugInstallationConstraints
	DenyInstallation  []*PlugInstallationConstraints

	AllowConnection []*PlugConnectionConstraints
	DenyConnection  []*PlugConnectionConstraints

	AllowAutoConnection []*PlugConnectionConstraints
	DenyAutoConnection  []*PlugConnectionConstraints
}

func (r *PlugRule) feature(flabel string) bool {
	for _, cs := range [][]*PlugInstallationConstraints{r.AllowInstallation, r.DenyInstallation} {
		for _, c := range cs {
			if c.feature(flabel) {
				return true
			}
		}
	}

	for _, cs := range [][]*PlugConnectionConstraints{r.AllowConnection, r.DenyConnection, r.AllowAutoConnection, r.DenyAutoConnection} {
		for _, c := range cs {
			if c.feature(flabel) {
				return true
			}
		}
	}

	return false
}

func castPlugInstallationConstraints(cstrs []constraintsHolder) (res []*PlugInstallationConstraints) {
	res = make([]*PlugInstallationConstraints, len(cstrs))
	for i, cstr := range cstrs {
		res[i] = cstr.(*PlugInstallationConstraints)
	}
	return res
}

func castPlugConnectionConstraints(cstrs []constraintsHolder) (res []*PlugConnectionConstraints) {
	res = make([]*PlugConnectionConstraints, len(cstrs))
	for i, cstr := range cstrs {
		res[i] = cstr.(*PlugConnectionConstraints)
	}
	return res
}

func (r *PlugRule) setConstraints(field string, cstrs []constraintsHolder) {
	if len(cstrs) == 0 {
		panic(fmt.Sprintf("cannot set PlugRule field %q to empty", field))
	}
	switch cstrs[0].(type) {
	case *PlugInstallationConstraints:
		switch field {
		case "allow-installation":
			r.AllowInstallation = castPlugInstallationConstraints(cstrs)
			return
		case "deny-installation":
			r.DenyInstallation = castPlugInstallationConstraints(cstrs)
			return
		}
	case *PlugConnectionConstraints:
		switch field {
		case "allow-connection":
			r.AllowConnection = castPlugConnectionConstraints(cstrs)
			return
		case "deny-connection":
			r.DenyConnection = castPlugConnectionConstraints(cstrs)
			return
		case "allow-auto-connection":
			r.AllowAutoConnection = castPlugConnectionConstraints(cstrs)
			return
		case "deny-auto-connection":
			r.DenyAutoConnection = castPlugConnectionConstraints(cstrs)
			return
		}
	}
	panic(fmt.Sprintf("cannot set PlugRule field %q with %T elements", field, cstrs[0]))
}

// PlugInstallationConstraints specifies a set of constraints on an interface plug relevant to the installation of snap.
type PlugInstallationConstraints struct {
	PlugSnapTypes []string

	PlugAttributes *AttributeConstraints

	OnClassic *OnClassicConstraint

	DeviceScope *DeviceScopeConstraint
}

func (c *PlugInstallationConstraints) feature(flabel string) bool {
	if flabel == deviceScopeConstraintsFeature {
		return c.DeviceScope != nil
	}
	return c.PlugAttributes.feature(flabel)
}

func (c *PlugInstallationConstraints) setAttributeConstraints(field string, cstrs *AttributeConstraints) {
	switch field {
	case "plug-attributes":
		c.PlugAttributes = cstrs
	default:
		panic("unknown PlugInstallationConstraints field " + field)
	}
}

func (c *PlugInstallationConstraints) setIDConstraints(field string, cstrs []string) {
	switch field {
	case "plug-snap-type":
		c.PlugSnapTypes = cstrs
	default:
		panic("unknown PlugInstallationConstraints field " + field)
	}
}

func (c *PlugInstallationConstraints) setOnClassicConstraint(onClassic *OnClassicConstraint) {
	c.OnClassic = onClassic
}

func (c *PlugInstallationConstraints) setDeviceScopeConstraint(deviceScope *DeviceScopeConstraint) {
	c.DeviceScope = deviceScope
}

func compilePlugInstallationConstraints(context string, cDef constraintsDef) (constraintsHolder, error) {
	plugInstCstrs := &PlugInstallationConstraints{}
	err := baseCompileConstraints(context, cDef, plugInstCstrs, []string{"plug-attributes"}, []string{"plug-snap-type"})
	if err != nil {
		return nil, err
	}
	return plugInstCstrs, nil
}

// PlugConnectionConstraints specfies a set of constraints on an
// interface plug for a snap relevant to its connection or
// auto-connection.
type PlugConnectionConstraints struct {
	SlotSnapTypes    []string
	SlotSnapIDs      []string
	SlotPublisherIDs []string

	PlugAttributes *AttributeConstraints
	SlotAttributes *AttributeConstraints

	OnClassic *OnClassicConstraint

	DeviceScope *DeviceScopeConstraint
}

func (c *PlugConnectionConstraints) feature(flabel string) bool {
	if flabel == deviceScopeConstraintsFeature {
		return c.DeviceScope != nil
	}
	return c.PlugAttributes.feature(flabel) || c.SlotAttributes.feature(flabel)
}

func (c *PlugConnectionConstraints) setAttributeConstraints(field string, cstrs *AttributeConstraints) {
	switch field {
	case "plug-attributes":
		c.PlugAttributes = cstrs
	case "slot-attributes":
		c.SlotAttributes = cstrs
	default:
		panic("unknown PlugConnectionConstraints field " + field)
	}
}

func (c *PlugConnectionConstraints) setIDConstraints(field string, cstrs []string) {
	switch field {
	case "slot-snap-type":
		c.SlotSnapTypes = cstrs
	case "slot-snap-id":
		c.SlotSnapIDs = cstrs
	case "slot-publisher-id":
		c.SlotPublisherIDs = cstrs
	default:
		panic("unknown PlugConnectionConstraints field " + field)
	}
}

func (c *PlugConnectionConstraints) setOnClassicConstraint(onClassic *OnClassicConstraint) {
	c.OnClassic = onClassic
}

func (c *PlugConnectionConstraints) setDeviceScopeConstraint(deviceScope *DeviceScopeConstraint) {
	c.DeviceScope = deviceScope
}

var (
	attributeConstraints = []string{"plug-attributes", "slot-attributes"}
	plugIDConstraints    = []string{"slot-snap-type", "slot-publisher-id", "slot-snap-id"}
)

func compilePlugConnectionConstraints(context string, cDef constraintsDef) (constraintsHolder, error) {
	plugConnCstrs := &PlugConnectionConstraints{}
	err := baseCompileConstraints(context, cDef, plugConnCstrs, attributeConstraints, plugIDConstraints)
	if err != nil {
		return nil, err
	}
	return plugConnCstrs, nil
}

var (
	defaultOutcome = map[string]interface{}{
		"allow-installation":    "true",
		"allow-connection":      "true",
		"allow-auto-connection": "true",
		"deny-installation":     "false",
		"deny-connection":       "false",
		"deny-auto-connection":  "false",
	}

	invertedOutcome = map[string]interface{}{
		"allow-installation":    "false",
		"allow-connection":      "false",
		"allow-auto-connection": "false",
		"deny-installation":     "true",
		"deny-connection":       "true",
		"deny-auto-connection":  "true",
	}

	ruleSubrules = []string{"allow-installation", "deny-installation", "allow-connection", "deny-connection", "allow-auto-connection", "deny-auto-connection"}
)

var plugRuleCompilers = map[string]subruleCompiler{
	"allow-installation":    compilePlugInstallationConstraints,
	"deny-installation":     compilePlugInstallationConstraints,
	"allow-connection":      compilePlugConnectionConstraints,
	"deny-connection":       compilePlugConnectionConstraints,
	"allow-auto-connection": compilePlugConnectionConstraints,
	"deny-auto-connection":  compilePlugConnectionConstraints,
}

func compilePlugRule(interfaceName string, rule interface{}) (*PlugRule, error) {
	context := fmt.Sprintf("plug rule for interface %q", interfaceName)
	plugRule := &PlugRule{
		Interface: interfaceName,
	}
	err := baseCompileRule(context, rule, plugRule, ruleSubrules, plugRuleCompilers, defaultOutcome, invertedOutcome)
	if err != nil {
		return nil, err
	}
	return plugRule, nil
}

// SlotRule holds the rule of what is allowed, wrt installation and
// connection, for a slot of a specific interface for a snap.
type SlotRule struct {
	Interface string

	AllowInstallation []*SlotInstallationConstraints
	DenyInstallation  []*SlotInstallationConstraints

	AllowConnection []*SlotConnectionConstraints
	DenyConnection  []*SlotConnectionConstraints

	AllowAutoConnection []*SlotConnectionConstraints
	DenyAutoConnection  []*SlotConnectionConstraints
}

func castSlotInstallationConstraints(cstrs []constraintsHolder) (res []*SlotInstallationConstraints) {
	res = make([]*SlotInstallationConstraints, len(cstrs))
	for i, cstr := range cstrs {
		res[i] = cstr.(*SlotInstallationConstraints)
	}
	return res
}

func (r *SlotRule) feature(flabel string) bool {
	for _, cs := range [][]*SlotInstallationConstraints{r.AllowInstallation, r.DenyInstallation} {
		for _, c := range cs {
			if c.feature(flabel) {
				return true
			}
		}
	}

	for _, cs := range [][]*SlotConnectionConstraints{r.AllowConnection, r.DenyConnection, r.AllowAutoConnection, r.DenyAutoConnection} {
		for _, c := range cs {
			if c.feature(flabel) {
				return true
			}
		}
	}

	return false
}

func castSlotConnectionConstraints(cstrs []constraintsHolder) (res []*SlotConnectionConstraints) {
	res = make([]*SlotConnectionConstraints, len(cstrs))
	for i, cstr := range cstrs {
		res[i] = cstr.(*SlotConnectionConstraints)
	}
	return res
}

func (r *SlotRule) setConstraints(field string, cstrs []constraintsHolder) {
	if len(cstrs) == 0 {
		panic(fmt.Sprintf("cannot set SlotRule field %q to empty", field))
	}
	switch cstrs[0].(type) {
	case *SlotInstallationConstraints:
		switch field {
		case "allow-installation":
			r.AllowInstallation = castSlotInstallationConstraints(cstrs)
			return
		case "deny-installation":
			r.DenyInstallation = castSlotInstallationConstraints(cstrs)
			return
		}
	case *SlotConnectionConstraints:
		switch field {
		case "allow-connection":
			r.AllowConnection = castSlotConnectionConstraints(cstrs)
			return
		case "deny-connection":
			r.DenyConnection = castSlotConnectionConstraints(cstrs)
			return
		case "allow-auto-connection":
			r.AllowAutoConnection = castSlotConnectionConstraints(cstrs)
			return
		case "deny-auto-connection":
			r.DenyAutoConnection = castSlotConnectionConstraints(cstrs)
			return
		}
	}
	panic(fmt.Sprintf("cannot set SlotRule field %q with %T elements", field, cstrs[0]))
}

// SlotInstallationConstraints specifies a set of constraints on an
// interface slot relevant to the installation of snap.
type SlotInstallationConstraints struct {
	SlotSnapTypes []string

	SlotAttributes *AttributeConstraints

	OnClassic *OnClassicConstraint

	DeviceScope *DeviceScopeConstraint
}

func (c *SlotInstallationConstraints) feature(flabel string) bool {
	if flabel == deviceScopeConstraintsFeature {
		return c.DeviceScope != nil
	}
	return c.SlotAttributes.feature(flabel)
}

func (c *SlotInstallationConstraints) setAttributeConstraints(field string, cstrs *AttributeConstraints) {
	switch field {
	case "slot-attributes":
		c.SlotAttributes = cstrs
	default:
		panic("unknown SlotInstallationConstraints field " + field)
	}
}

func (c *SlotInstallationConstraints) setIDConstraints(field string, cstrs []string) {
	switch field {
	case "slot-snap-type":
		c.SlotSnapTypes = cstrs
	default:
		panic("unknown SlotInstallationConstraints field " + field)
	}
}

func (c *SlotInstallationConstraints) setOnClassicConstraint(onClassic *OnClassicConstraint) {
	c.OnClassic = onClassic
}

func (c *SlotInstallationConstraints) setDeviceScopeConstraint(deviceScope *DeviceScopeConstraint) {
	c.DeviceScope = deviceScope
}

func compileSlotInstallationConstraints(context string, cDef constraintsDef) (constraintsHolder, error) {
	slotInstCstrs := &SlotInstallationConstraints{}
	err := baseCompileConstraints(context, cDef, slotInstCstrs, []string{"slot-attributes"}, []string{"slot-snap-type"})
	if err != nil {
		return nil, err
	}
	return slotInstCstrs, nil
}

// SlotConnectionConstraints specfies a set of constraints on an
// interface slot for a snap relevant to its connection or
// auto-connection.
type SlotConnectionConstraints struct {
	PlugSnapTypes    []string
	PlugSnapIDs      []string
	PlugPublisherIDs []string

	SlotAttributes *AttributeConstraints
	PlugAttributes *AttributeConstraints

	OnClassic *OnClassicConstraint

	DeviceScope *DeviceScopeConstraint
}

func (c *SlotConnectionConstraints) feature(flabel string) bool {
	if flabel == deviceScopeConstraintsFeature {
		return c.DeviceScope != nil
	}
	return c.PlugAttributes.feature(flabel) || c.SlotAttributes.feature(flabel)
}

func (c *SlotConnectionConstraints) setAttributeConstraints(field string, cstrs *AttributeConstraints) {
	switch field {
	case "plug-attributes":
		c.PlugAttributes = cstrs
	case "slot-attributes":
		c.SlotAttributes = cstrs
	default:
		panic("unknown SlotConnectionConstraints field " + field)
	}
}

func (c *SlotConnectionConstraints) setIDConstraints(field string, cstrs []string) {
	switch field {
	case "plug-snap-type":
		c.PlugSnapTypes = cstrs
	case "plug-snap-id":
		c.PlugSnapIDs = cstrs
	case "plug-publisher-id":
		c.PlugPublisherIDs = cstrs
	default:
		panic("unknown SlotConnectionConstraints field " + field)
	}
}

var (
	slotIDConstraints = []string{"plug-snap-type", "plug-publisher-id", "plug-snap-id"}
)

func (c *SlotConnectionConstraints) setOnClassicConstraint(onClassic *OnClassicConstraint) {
	c.OnClassic = onClassic
}

func (c *SlotConnectionConstraints) setDeviceScopeConstraint(deviceScope *DeviceScopeConstraint) {
	c.DeviceScope = deviceScope
}

func compileSlotConnectionConstraints(context string, cDef constraintsDef) (constraintsHolder, error) {
	slotConnCstrs := &SlotConnectionConstraints{}
	err := baseCompileConstraints(context, cDef, slotConnCstrs, attributeConstraints, slotIDConstraints)
	if err != nil {
		return nil, err
	}
	return slotConnCstrs, nil
}

var slotRuleCompilers = map[string]subruleCompiler{
	"allow-installation":    compileSlotInstallationConstraints,
	"deny-installation":     compileSlotInstallationConstraints,
	"allow-connection":      compileSlotConnectionConstraints,
	"deny-connection":       compileSlotConnectionConstraints,
	"allow-auto-connection": compileSlotConnectionConstraints,
	"deny-auto-connection":  compileSlotConnectionConstraints,
}

func compileSlotRule(interfaceName string, rule interface{}) (*SlotRule, error) {
	context := fmt.Sprintf("slot rule for interface %q", interfaceName)
	slotRule := &SlotRule{
		Interface: interfaceName,
	}
	err := baseCompileRule(context, rule, slotRule, ruleSubrules, slotRuleCompilers, defaultOutcome, invertedOutcome)
	if err != nil {
		return nil, err
	}
	return slotRule, nil
}
