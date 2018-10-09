// -*- Mode: Go; indent-tabs-mode: t -*-

/*
 * Copyright (C) 2014-2016 Canonical Ltd
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

package snap

import (
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"

	"gopkg.in/yaml.v2"

	"github.com/snapcore/snapd/strutil"
	"github.com/snapcore/snapd/timeout"
)

type snapYaml struct {
	Name             string                 `yaml:"name"`
	Version          string                 `yaml:"version"`
	Type             Type                   `yaml:"type"`
	Architectures    []string               `yaml:"architectures,omitempty"`
	Assumes          []string               `yaml:"assumes"`
	Title            string                 `yaml:"title"`
	Description      string                 `yaml:"description"`
	Summary          string                 `yaml:"summary"`
	License          string                 `yaml:"license,omitempty"`
	LicenseAgreement string                 `yaml:"license-agreement,omitempty"`
	LicenseVersion   string                 `yaml:"license-version,omitempty"`
	Epoch            Epoch                  `yaml:"epoch,omitempty"`
	Base             string                 `yaml:"base,omitempty"`
	Confinement      ConfinementType        `yaml:"confinement,omitempty"`
	Environment      strutil.OrderedMap     `yaml:"environment,omitempty"`
	Plugs            map[string]interface{} `yaml:"plugs,omitempty"`
	Slots            map[string]interface{} `yaml:"slots,omitempty"`
	Apps             map[string]appYaml     `yaml:"apps,omitempty"`
	Hooks            map[string]hookYaml    `yaml:"hooks,omitempty"`
	Layout           map[string]layoutYaml  `yaml:"layout,omitempty"`

	// TypoLayouts is used to detect the use of the incorrect plural form of "layout"
	TypoLayouts typoDetector `yaml:"layouts,omitempty"`
}

type typoDetector struct {
	Hint string
}

func (td *typoDetector) UnmarshalYAML(func(interface{}) error) error {
	return fmt.Errorf("typo detected: %s", td.Hint)
}

type appYaml struct {
	Aliases []string `yaml:"aliases,omitempty"`

	Command      string   `yaml:"command"`
	CommandChain []string `yaml:"command-chain,omitempty"`

	Daemon string `yaml:"daemon"`

	StopCommand     string          `yaml:"stop-command,omitempty"`
	ReloadCommand   string          `yaml:"reload-command,omitempty"`
	PostStopCommand string          `yaml:"post-stop-command,omitempty"`
	StopTimeout     timeout.Timeout `yaml:"stop-timeout,omitempty"`
	WatchdogTimeout timeout.Timeout `yaml:"watchdog-timeout,omitempty"`
	Completer       string          `yaml:"completer,omitempty"`
	RefreshMode     string          `yaml:"refresh-mode,omitempty"`
	StopMode        StopModeType    `yaml:"stop-mode,omitempty"`

	RestartCond RestartCondition `yaml:"restart-condition,omitempty"`
	SlotNames   []string         `yaml:"slots,omitempty"`
	PlugNames   []string         `yaml:"plugs,omitempty"`

	BusName  string `yaml:"bus-name,omitempty"`
	CommonID string `yaml:"common-id,omitempty"`

	Environment strutil.OrderedMap `yaml:"environment,omitempty"`

	Sockets map[string]socketsYaml `yaml:"sockets,omitempty"`

	After  []string `yaml:"after,omitempty"`
	Before []string `yaml:"before,omitempty"`

	Timer string `yaml:"timer,omitempty"`

	Autostart string `yaml:"autostart,omitempty"`
}

type hookYaml struct {
	PlugNames    []string           `yaml:"plugs,omitempty"`
	SlotNames    []string           `yaml:"slots,omitempty"`
	Environment  strutil.OrderedMap `yaml:"environment,omitempty"`
	CommandChain []string           `yaml:"command-chain,omitempty"`
}

type layoutYaml struct {
	Bind     string `yaml:"bind,omitempty"`
	BindFile string `yaml:"bind-file,omitempty"`
	Type     string `yaml:"type,omitempty"`
	User     string `yaml:"user,omitempty"`
	Group    string `yaml:"group,omitempty"`
	Mode     string `yaml:"mode,omitempty"`
	Symlink  string `yaml:"symlink,omitempty"`
}

type socketsYaml struct {
	ListenStream string      `yaml:"listen-stream,omitempty"`
	SocketMode   os.FileMode `yaml:"socket-mode,omitempty"`
}

// InfoFromSnapYaml creates a new info based on the given snap.yaml data
func InfoFromSnapYaml(yamlData []byte) (*Info, error) {
	var y snapYaml
	// Customize hints for the typo detector.
	y.TypoLayouts.Hint = `use singular "layout" instead of plural "layouts"`
	err := yaml.Unmarshal(yamlData, &y)
	if err != nil {
		return nil, fmt.Errorf("cannot parse snap.yaml: %s", err)
	}

	snap := infoSkeletonFromSnapYaml(y)

	// Collect top-level definitions of plugs and slots
	if err := setPlugsFromSnapYaml(y, snap); err != nil {
		return nil, err
	}
	if err := setSlotsFromSnapYaml(y, snap); err != nil {
		return nil, err
	}

	// At this point snap.Plugs and snap.Slots only contain globally-declared
	// plugs and slots. We're about to change that, but we need to remember the
	// global ones for later, so save their names.
	globalPlugNames := make([]string, 0, len(snap.Plugs))
	for plugName := range snap.Plugs {
		globalPlugNames = append(globalPlugNames, plugName)
	}

	globalSlotNames := make([]string, 0, len(snap.Slots))
	for slotName := range snap.Slots {
		globalSlotNames = append(globalSlotNames, slotName)
	}

	// Collect all apps, their aliases and hooks
	if err := setAppsFromSnapYaml(y, snap); err != nil {
		return nil, err
	}
	setHooksFromSnapYaml(y, snap)

	// Bind unbound plugs to all apps and hooks
	bindUnboundPlugs(globalPlugNames, snap)

	// Bind unbound slots to all apps and hooks
	bindUnboundSlots(globalSlotNames, snap)

	// Collect layout elements.
	if y.Layout != nil {
		snap.Layout = make(map[string]*Layout, len(y.Layout))
		for path, l := range y.Layout {
			var mode os.FileMode = 0755
			if l.Mode != "" {
				m, err := strconv.ParseUint(l.Mode, 8, 32)
				if err != nil {
					return nil, err
				}
				mode = os.FileMode(m)
			}
			user := "root"
			if l.User != "" {
				user = l.User
			}
			group := "root"
			if l.Group != "" {
				group = l.Group
			}
			snap.Layout[path] = &Layout{
				Snap: snap, Path: path,
				Bind: l.Bind, Type: l.Type, Symlink: l.Symlink, BindFile: l.BindFile,
				User: user, Group: group, Mode: mode,
			}
		}
	}

	// Rename specific plugs on the core snap.
	snap.renameClashingCorePlugs()

	snap.BadInterfaces = make(map[string]string)
	SanitizePlugsSlots(snap)

	// FIXME: validation of the fields
	return snap, nil
}

// infoSkeletonFromSnapYaml initializes an Info without apps, hook, plugs, or
// slots
func infoSkeletonFromSnapYaml(y snapYaml) *Info {
	// Prepare defaults
	architectures := []string{"all"}
	if len(y.Architectures) != 0 {
		architectures = y.Architectures
	}
	typ := TypeApp
	if y.Type != "" {
		typ = y.Type
	}
	// TODO: once we have epochs transition to the snapd type for real
	if y.Name == "snapd" {
		typ = TypeSnapd
	}

	confinement := StrictConfinement
	if y.Confinement != "" {
		confinement = y.Confinement
	}

	// Construct snap skeleton without apps, hooks, plugs, or slots
	snap := &Info{
		SuggestedName:       y.Name,
		Version:             y.Version,
		Type:                typ,
		Architectures:       architectures,
		Assumes:             y.Assumes,
		OriginalTitle:       y.Title,
		OriginalDescription: y.Description,
		OriginalSummary:     y.Summary,
		License:             y.License,
		LicenseAgreement:    y.LicenseAgreement,
		LicenseVersion:      y.LicenseVersion,
		Epoch:               y.Epoch,
		Confinement:         confinement,
		Base:                y.Base,
		Apps:                make(map[string]*AppInfo),
		LegacyAliases:       make(map[string]*AppInfo),
		Hooks:               make(map[string]*HookInfo),
		Plugs:               make(map[string]*PlugInfo),
		Slots:               make(map[string]*SlotInfo),
		Environment:         y.Environment,
	}

	sort.Strings(snap.Assumes)

	return snap
}

func setPlugsFromSnapYaml(y snapYaml, snap *Info) error {
	for name, data := range y.Plugs {
		iface, label, attrs, err := convertToSlotOrPlugData("plug", name, data)
		if err != nil {
			return err
		}
		snap.Plugs[name] = &PlugInfo{
			Snap:      snap,
			Name:      name,
			Interface: iface,
			Attrs:     attrs,
			Label:     label,
		}
		if len(y.Apps) > 0 {
			snap.Plugs[name].Apps = make(map[string]*AppInfo)
		}
		if len(y.Hooks) > 0 {
			snap.Plugs[name].Hooks = make(map[string]*HookInfo)
		}
	}

	return nil
}

func setSlotsFromSnapYaml(y snapYaml, snap *Info) error {
	for name, data := range y.Slots {
		iface, label, attrs, err := convertToSlotOrPlugData("slot", name, data)
		if err != nil {
			return err
		}
		snap.Slots[name] = &SlotInfo{
			Snap:      snap,
			Name:      name,
			Interface: iface,
			Attrs:     attrs,
			Label:     label,
		}
		if len(y.Apps) > 0 {
			snap.Slots[name].Apps = make(map[string]*AppInfo)
		}
		if len(y.Hooks) > 0 {
			snap.Slots[name].Hooks = make(map[string]*HookInfo)
		}
	}

	return nil
}

func setAppsFromSnapYaml(y snapYaml, snap *Info) error {
	for appName, yApp := range y.Apps {
		// Collect all apps
		app := &AppInfo{
			Snap:            snap,
			Name:            appName,
			LegacyAliases:   yApp.Aliases,
			Command:         yApp.Command,
			CommandChain:    yApp.CommandChain,
			Daemon:          yApp.Daemon,
			StopTimeout:     yApp.StopTimeout,
			StopCommand:     yApp.StopCommand,
			ReloadCommand:   yApp.ReloadCommand,
			PostStopCommand: yApp.PostStopCommand,
			RestartCond:     yApp.RestartCond,
			BusName:         yApp.BusName,
			CommonID:        yApp.CommonID,
			Environment:     yApp.Environment,
			Completer:       yApp.Completer,
			StopMode:        yApp.StopMode,
			RefreshMode:     yApp.RefreshMode,
			Before:          yApp.Before,
			After:           yApp.After,
			Autostart:       yApp.Autostart,
			WatchdogTimeout: yApp.WatchdogTimeout,
		}
		if len(y.Plugs) > 0 || len(yApp.PlugNames) > 0 {
			app.Plugs = make(map[string]*PlugInfo)
		}
		if len(y.Slots) > 0 || len(yApp.SlotNames) > 0 {
			app.Slots = make(map[string]*SlotInfo)
		}
		if len(yApp.Sockets) > 0 {
			app.Sockets = make(map[string]*SocketInfo, len(yApp.Sockets))
		}

		snap.Apps[appName] = app
		for _, alias := range app.LegacyAliases {
			if snap.LegacyAliases[alias] != nil {
				return fmt.Errorf("cannot set %q as alias for both %q and %q", alias, snap.LegacyAliases[alias].Name, appName)
			}
			snap.LegacyAliases[alias] = app
		}
		// Bind all plugs/slots listed in this app
		for _, plugName := range yApp.PlugNames {
			plug, ok := snap.Plugs[plugName]
			if !ok {
				// Create implicit plug definitions if required
				plug = &PlugInfo{
					Snap:      snap,
					Name:      plugName,
					Interface: plugName,
					Apps:      make(map[string]*AppInfo),
				}
				snap.Plugs[plugName] = plug
			}
			app.Plugs[plugName] = plug
			plug.Apps[appName] = app
		}
		for _, slotName := range yApp.SlotNames {
			slot, ok := snap.Slots[slotName]
			if !ok {
				slot = &SlotInfo{
					Snap:      snap,
					Name:      slotName,
					Interface: slotName,
					Apps:      make(map[string]*AppInfo),
				}
				snap.Slots[slotName] = slot
			}
			app.Slots[slotName] = slot
			slot.Apps[appName] = app
		}
		for name, data := range yApp.Sockets {
			app.Sockets[name] = &SocketInfo{
				App:          app,
				Name:         name,
				ListenStream: data.ListenStream,
				SocketMode:   data.SocketMode,
			}
		}
		if yApp.Timer != "" {
			app.Timer = &TimerInfo{
				App:   app,
				Timer: yApp.Timer,
			}
		}
		// collect all common IDs
		if app.CommonID != "" {
			snap.CommonIDs = append(snap.CommonIDs, app.CommonID)
		}
	}
	return nil
}

func setHooksFromSnapYaml(y snapYaml, snap *Info) {
	for hookName, yHook := range y.Hooks {
		if !IsHookSupported(hookName) {
			continue
		}

		// Collect all hooks
		hook := &HookInfo{
			Snap:         snap,
			Name:         hookName,
			Environment:  yHook.Environment,
			CommandChain: yHook.CommandChain,
		}
		if len(y.Plugs) > 0 || len(yHook.PlugNames) > 0 {
			hook.Plugs = make(map[string]*PlugInfo)
		}
		if len(y.Slots) > 0 || len(yHook.SlotNames) > 0 {
			hook.Slots = make(map[string]*SlotInfo)
		}

		snap.Hooks[hookName] = hook
		// Bind all plugs/slots listed in this hook
		for _, plugName := range yHook.PlugNames {
			plug, ok := snap.Plugs[plugName]
			if !ok {
				// Create implicit plug definitions if required
				plug = &PlugInfo{
					Snap:      snap,
					Name:      plugName,
					Interface: plugName,
					Hooks:     make(map[string]*HookInfo),
				}
				snap.Plugs[plugName] = plug
			} else if plug.Hooks == nil {
				plug.Hooks = make(map[string]*HookInfo)
			}
			hook.Plugs[plugName] = plug
			plug.Hooks[hookName] = hook
		}
		for _, slotName := range yHook.SlotNames {
			slot, ok := snap.Slots[slotName]
			if !ok {
				// Create implicit slot definitions if required
				slot = &SlotInfo{
					Snap:      snap,
					Name:      slotName,
					Interface: slotName,
					Hooks:     make(map[string]*HookInfo),
				}
				snap.Slots[slotName] = slot
			} else if slot.Hooks == nil {
				slot.Hooks = make(map[string]*HookInfo)
			}
			hook.Slots[slotName] = slot
			slot.Hooks[hookName] = hook
		}
	}
}

func bindUnboundPlugs(plugNames []string, snap *Info) error {
	for _, plugName := range plugNames {
		plug, ok := snap.Plugs[plugName]
		if !ok {
			return fmt.Errorf("no plug named %q", plugName)
		}

		// A plug is considered unbound if it isn't being used by any apps
		// or hooks. In which case we bind them to all apps and hooks.
		if len(plug.Apps) == 0 && len(plug.Hooks) == 0 {
			for appName, app := range snap.Apps {
				app.Plugs[plugName] = plug
				plug.Apps[appName] = app
			}

			for hookName, hook := range snap.Hooks {
				hook.Plugs[plugName] = plug
				plug.Hooks[hookName] = hook
			}
		}
	}

	return nil
}

func bindUnboundSlots(slotNames []string, snap *Info) error {
	for _, slotName := range slotNames {
		slot, ok := snap.Slots[slotName]
		if !ok {
			return fmt.Errorf("no slot named %q", slotName)
		}

		// A slot is considered unbound if it isn't being used by any apps
		// or hooks. In which case we bind them to all apps and hooks.
		if len(slot.Apps) == 0 && len(slot.Hooks) == 0 {
			for appName, app := range snap.Apps {
				app.Slots[slotName] = slot
				slot.Apps[appName] = app
			}
			for hookName, hook := range snap.Hooks {
				hook.Slots[slotName] = slot
				slot.Hooks[hookName] = hook
			}
		}
	}

	return nil
}

func convertToSlotOrPlugData(plugOrSlot, name string, data interface{}) (iface, label string, attrs map[string]interface{}, err error) {
	iface = name
	switch data.(type) {
	case string:
		return data.(string), "", nil, nil
	case nil:
		return name, "", nil, nil
	case map[interface{}]interface{}:
		for keyData, valueData := range data.(map[interface{}]interface{}) {
			key, ok := keyData.(string)
			if !ok {
				err := fmt.Errorf("%s %q has attribute that is not a string (found %T)",
					plugOrSlot, name, keyData)
				return "", "", nil, err
			}
			if strings.HasPrefix(key, "$") {
				err := fmt.Errorf("%s %q uses reserved attribute %q", plugOrSlot, name, key)
				return "", "", nil, err
			}
			switch key {
			case "interface":
				value, ok := valueData.(string)
				if !ok {
					err := fmt.Errorf("interface name on %s %q is not a string (found %T)",
						plugOrSlot, name, valueData)
					return "", "", nil, err
				}
				iface = value
			case "label":
				value, ok := valueData.(string)
				if !ok {
					err := fmt.Errorf("label of %s %q is not a string (found %T)",
						plugOrSlot, name, valueData)
					return "", "", nil, err
				}
				label = value
			default:
				if attrs == nil {
					attrs = make(map[string]interface{})
				}
				value, err := normalizeYamlValue(valueData)
				if err != nil {
					return "", "", nil, fmt.Errorf("attribute %q of %s %q: %v", key, plugOrSlot, name, err)
				}
				attrs[key] = value
			}
		}
		return iface, label, attrs, nil
	default:
		err := fmt.Errorf("%s %q has malformed definition (found %T)", plugOrSlot, name, data)
		return "", "", nil, err
	}
}

// normalizeYamlValue validates values and returns a normalized version of it (map[interface{}]interface{} is turned into map[string]interface{})
func normalizeYamlValue(v interface{}) (interface{}, error) {
	switch x := v.(type) {
	case string:
		return x, nil
	case bool:
		return x, nil
	case int:
		return int64(x), nil
	case int64:
		return x, nil
	case float64:
		return x, nil
	case float32:
		return float64(x), nil
	case []interface{}:
		l := make([]interface{}, len(x))
		for i, el := range x {
			el, err := normalizeYamlValue(el)
			if err != nil {
				return nil, err
			}
			l[i] = el
		}
		return l, nil
	case map[interface{}]interface{}:
		m := make(map[string]interface{}, len(x))
		for k, item := range x {
			kStr, ok := k.(string)
			if !ok {
				return nil, fmt.Errorf("non-string key: %v", k)
			}
			item, err := normalizeYamlValue(item)
			if err != nil {
				return nil, err
			}
			m[kStr] = item
		}
		return m, nil
	case map[string]interface{}:
		m := make(map[string]interface{}, len(x))
		for k, item := range x {
			item, err := normalizeYamlValue(item)
			if err != nil {
				return nil, err
			}
			m[k] = item
		}
		return m, nil
	default:
		return nil, fmt.Errorf("invalid scalar: %v", v)
	}
}
