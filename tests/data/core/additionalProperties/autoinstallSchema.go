// Code generated by github.com/atombender/go-jsonschema, DO NOT EDIT.

package test

import "encoding/json"
import "fmt"
import "github.com/go-viper/mapstructure/v2"
import yaml "gopkg.in/yaml.v3"
import "reflect"
import "regexp"
import "strings"

type AutoinstallSchema struct {
	// ActiveDirectory corresponds to the JSON schema field "active-directory".
	ActiveDirectory *AutoinstallSchemaActiveDirectory `json:"active-directory,omitempty" yaml:"active-directory,omitempty" mapstructure:"active-directory,omitempty"`

	// Apt corresponds to the JSON schema field "apt".
	Apt *AutoinstallSchemaApt `json:"apt,omitempty" yaml:"apt,omitempty" mapstructure:"apt,omitempty"`

	// Codecs corresponds to the JSON schema field "codecs".
	Codecs *AutoinstallSchemaCodecs `json:"codecs,omitempty" yaml:"codecs,omitempty" mapstructure:"codecs,omitempty"`

	// DebconfSelections corresponds to the JSON schema field "debconf-selections".
	DebconfSelections *string `json:"debconf-selections,omitempty" yaml:"debconf-selections,omitempty" mapstructure:"debconf-selections,omitempty"`

	// Drivers corresponds to the JSON schema field "drivers".
	Drivers *AutoinstallSchemaDrivers `json:"drivers,omitempty" yaml:"drivers,omitempty" mapstructure:"drivers,omitempty"`

	// EarlyCommands corresponds to the JSON schema field "early-commands".
	EarlyCommands []interface{} `json:"early-commands,omitempty" yaml:"early-commands,omitempty" mapstructure:"early-commands,omitempty"`

	// ErrorCommands corresponds to the JSON schema field "error-commands".
	ErrorCommands []interface{} `json:"error-commands,omitempty" yaml:"error-commands,omitempty" mapstructure:"error-commands,omitempty"`

	// Identity corresponds to the JSON schema field "identity".
	Identity *AutoinstallSchemaIdentity `json:"identity,omitempty" yaml:"identity,omitempty" mapstructure:"identity,omitempty"`

	// InteractiveSections corresponds to the JSON schema field
	// "interactive-sections".
	InteractiveSections []string `json:"interactive-sections,omitempty" yaml:"interactive-sections,omitempty" mapstructure:"interactive-sections,omitempty"`

	// Kernel corresponds to the JSON schema field "kernel".
	Kernel *AutoinstallSchemaKernel `json:"kernel,omitempty" yaml:"kernel,omitempty" mapstructure:"kernel,omitempty"`

	// KernelCrashDumps corresponds to the JSON schema field "kernel-crash-dumps".
	KernelCrashDumps *AutoinstallSchemaKernelCrashDumps `json:"kernel-crash-dumps,omitempty" yaml:"kernel-crash-dumps,omitempty" mapstructure:"kernel-crash-dumps,omitempty"`

	// Keyboard corresponds to the JSON schema field "keyboard".
	Keyboard *AutoinstallSchemaKeyboard `json:"keyboard,omitempty" yaml:"keyboard,omitempty" mapstructure:"keyboard,omitempty"`

	// LateCommands corresponds to the JSON schema field "late-commands".
	LateCommands []interface{} `json:"late-commands,omitempty" yaml:"late-commands,omitempty" mapstructure:"late-commands,omitempty"`

	// Locale corresponds to the JSON schema field "locale".
	Locale *string `json:"locale,omitempty" yaml:"locale,omitempty" mapstructure:"locale,omitempty"`

	// Network corresponds to the JSON schema field "network".
	Network interface{} `json:"network,omitempty" yaml:"network,omitempty" mapstructure:"network,omitempty"`

	// Oem corresponds to the JSON schema field "oem".
	Oem *AutoinstallSchemaOem `json:"oem,omitempty" yaml:"oem,omitempty" mapstructure:"oem,omitempty"`

	// Packages corresponds to the JSON schema field "packages".
	Packages []string `json:"packages,omitempty" yaml:"packages,omitempty" mapstructure:"packages,omitempty"`

	// Proxy corresponds to the JSON schema field "proxy".
	Proxy *string `json:"proxy,omitempty" yaml:"proxy,omitempty" mapstructure:"proxy,omitempty"`

	// RefreshInstaller corresponds to the JSON schema field "refresh-installer".
	RefreshInstaller *AutoinstallSchemaRefreshInstaller `json:"refresh-installer,omitempty" yaml:"refresh-installer,omitempty" mapstructure:"refresh-installer,omitempty"`

	// Reporting corresponds to the JSON schema field "reporting".
	Reporting AutoinstallSchemaReporting `json:"reporting,omitempty" yaml:"reporting,omitempty" mapstructure:"reporting,omitempty"`

	// Shutdown corresponds to the JSON schema field "shutdown".
	Shutdown *AutoinstallSchemaShutdown `json:"shutdown,omitempty" yaml:"shutdown,omitempty" mapstructure:"shutdown,omitempty"`

	// Snaps corresponds to the JSON schema field "snaps".
	Snaps []AutoinstallSchemaSnapsElem `json:"snaps,omitempty" yaml:"snaps,omitempty" mapstructure:"snaps,omitempty"`

	// Source corresponds to the JSON schema field "source".
	Source *AutoinstallSchemaSource `json:"source,omitempty" yaml:"source,omitempty" mapstructure:"source,omitempty"`

	// Ssh corresponds to the JSON schema field "ssh".
	Ssh *AutoinstallSchemaSsh `json:"ssh,omitempty" yaml:"ssh,omitempty" mapstructure:"ssh,omitempty"`

	// Storage corresponds to the JSON schema field "storage".
	Storage AutoinstallSchemaStorage `json:"storage,omitempty" yaml:"storage,omitempty" mapstructure:"storage,omitempty"`

	// Timezone corresponds to the JSON schema field "timezone".
	Timezone *string `json:"timezone,omitempty" yaml:"timezone,omitempty" mapstructure:"timezone,omitempty"`

	// Compatibility only - use ubuntu-pro instead
	UbuntuAdvantage *AutoinstallSchemaUbuntuAdvantage `json:"ubuntu-advantage,omitempty" yaml:"ubuntu-advantage,omitempty" mapstructure:"ubuntu-advantage,omitempty"`

	// UbuntuPro corresponds to the JSON schema field "ubuntu-pro".
	UbuntuPro *AutoinstallSchemaUbuntuPro `json:"ubuntu-pro,omitempty" yaml:"ubuntu-pro,omitempty" mapstructure:"ubuntu-pro,omitempty"`

	// Updates corresponds to the JSON schema field "updates".
	Updates *AutoinstallSchemaUpdates `json:"updates,omitempty" yaml:"updates,omitempty" mapstructure:"updates,omitempty"`

	// UserData corresponds to the JSON schema field "user-data".
	UserData AutoinstallSchemaUserData `json:"user-data,omitempty" yaml:"user-data,omitempty" mapstructure:"user-data,omitempty"`

	// Version corresponds to the JSON schema field "version".
	Version int `json:"version" yaml:"version" mapstructure:"version"`

	// Zdevs corresponds to the JSON schema field "zdevs".
	Zdevs []AutoinstallSchemaZdevsElem `json:"zdevs,omitempty" yaml:"zdevs,omitempty" mapstructure:"zdevs,omitempty"`

	AdditionalProperties interface{} `mapstructure:",remain"`
}

type AutoinstallSchemaActiveDirectory struct {
	// AdminName corresponds to the JSON schema field "admin-name".
	AdminName *string `json:"admin-name,omitempty" yaml:"admin-name,omitempty" mapstructure:"admin-name,omitempty"`

	// DomainName corresponds to the JSON schema field "domain-name".
	DomainName *string `json:"domain-name,omitempty" yaml:"domain-name,omitempty" mapstructure:"domain-name,omitempty"`
}

type AutoinstallSchemaApt struct {
	// DisableComponents corresponds to the JSON schema field "disable_components".
	DisableComponents []AutoinstallSchemaAptDisableComponentsElem `json:"disable_components,omitempty" yaml:"disable_components,omitempty" mapstructure:"disable_components,omitempty"`

	// Fallback corresponds to the JSON schema field "fallback".
	Fallback *AutoinstallSchemaAptFallback `json:"fallback,omitempty" yaml:"fallback,omitempty" mapstructure:"fallback,omitempty"`

	// Geoip corresponds to the JSON schema field "geoip".
	Geoip *bool `json:"geoip,omitempty" yaml:"geoip,omitempty" mapstructure:"geoip,omitempty"`

	// MirrorSelection corresponds to the JSON schema field "mirror-selection".
	MirrorSelection *AutoinstallSchemaAptMirrorSelection `json:"mirror-selection,omitempty" yaml:"mirror-selection,omitempty" mapstructure:"mirror-selection,omitempty"`

	// Preferences corresponds to the JSON schema field "preferences".
	Preferences []AutoinstallSchemaAptPreferencesElem `json:"preferences,omitempty" yaml:"preferences,omitempty" mapstructure:"preferences,omitempty"`

	// PreserveSourcesList corresponds to the JSON schema field
	// "preserve_sources_list".
	PreserveSourcesList *bool `json:"preserve_sources_list,omitempty" yaml:"preserve_sources_list,omitempty" mapstructure:"preserve_sources_list,omitempty"`

	// Primary corresponds to the JSON schema field "primary".
	Primary []interface{} `json:"primary,omitempty" yaml:"primary,omitempty" mapstructure:"primary,omitempty"`

	// Sources corresponds to the JSON schema field "sources".
	Sources AutoinstallSchemaAptSources `json:"sources,omitempty" yaml:"sources,omitempty" mapstructure:"sources,omitempty"`
}

type AutoinstallSchemaAptDisableComponentsElem string

const AutoinstallSchemaAptDisableComponentsElemContrib AutoinstallSchemaAptDisableComponentsElem = "contrib"
const AutoinstallSchemaAptDisableComponentsElemMultiverse AutoinstallSchemaAptDisableComponentsElem = "multiverse"
const AutoinstallSchemaAptDisableComponentsElemNonFree AutoinstallSchemaAptDisableComponentsElem = "non-free"
const AutoinstallSchemaAptDisableComponentsElemRestricted AutoinstallSchemaAptDisableComponentsElem = "restricted"
const AutoinstallSchemaAptDisableComponentsElemUniverse AutoinstallSchemaAptDisableComponentsElem = "universe"

var enumValues_AutoinstallSchemaAptDisableComponentsElem = []interface{}{
	"universe",
	"multiverse",
	"restricted",
	"contrib",
	"non-free",
}

// UnmarshalJSON implements json.Unmarshaler.
func (j *AutoinstallSchemaAptDisableComponentsElem) UnmarshalJSON(value []byte) error {
	var v string
	if err := json.Unmarshal(value, &v); err != nil {
		return err
	}
	var ok bool
	for _, expected := range enumValues_AutoinstallSchemaAptDisableComponentsElem {
		if reflect.DeepEqual(v, expected) {
			ok = true
			break
		}
	}
	if !ok {
		return fmt.Errorf("invalid value (expected one of %#v): %#v", enumValues_AutoinstallSchemaAptDisableComponentsElem, v)
	}
	*j = AutoinstallSchemaAptDisableComponentsElem(v)
	return nil
}

// UnmarshalYAML implements yaml.Unmarshaler.
func (j *AutoinstallSchemaAptDisableComponentsElem) UnmarshalYAML(value *yaml.Node) error {
	var v string
	if err := value.Decode(&v); err != nil {
		return err
	}
	var ok bool
	for _, expected := range enumValues_AutoinstallSchemaAptDisableComponentsElem {
		if reflect.DeepEqual(v, expected) {
			ok = true
			break
		}
	}
	if !ok {
		return fmt.Errorf("invalid value (expected one of %#v): %#v", enumValues_AutoinstallSchemaAptDisableComponentsElem, v)
	}
	*j = AutoinstallSchemaAptDisableComponentsElem(v)
	return nil
}

type AutoinstallSchemaAptFallback string

const AutoinstallSchemaAptFallbackAbort AutoinstallSchemaAptFallback = "abort"
const AutoinstallSchemaAptFallbackContinueAnyway AutoinstallSchemaAptFallback = "continue-anyway"
const AutoinstallSchemaAptFallbackOfflineInstall AutoinstallSchemaAptFallback = "offline-install"

var enumValues_AutoinstallSchemaAptFallback = []interface{}{
	"abort",
	"continue-anyway",
	"offline-install",
}

// UnmarshalYAML implements yaml.Unmarshaler.
func (j *AutoinstallSchemaAptFallback) UnmarshalYAML(value *yaml.Node) error {
	var v string
	if err := value.Decode(&v); err != nil {
		return err
	}
	var ok bool
	for _, expected := range enumValues_AutoinstallSchemaAptFallback {
		if reflect.DeepEqual(v, expected) {
			ok = true
			break
		}
	}
	if !ok {
		return fmt.Errorf("invalid value (expected one of %#v): %#v", enumValues_AutoinstallSchemaAptFallback, v)
	}
	*j = AutoinstallSchemaAptFallback(v)
	return nil
}

// UnmarshalJSON implements json.Unmarshaler.
func (j *AutoinstallSchemaAptFallback) UnmarshalJSON(value []byte) error {
	var v string
	if err := json.Unmarshal(value, &v); err != nil {
		return err
	}
	var ok bool
	for _, expected := range enumValues_AutoinstallSchemaAptFallback {
		if reflect.DeepEqual(v, expected) {
			ok = true
			break
		}
	}
	if !ok {
		return fmt.Errorf("invalid value (expected one of %#v): %#v", enumValues_AutoinstallSchemaAptFallback, v)
	}
	*j = AutoinstallSchemaAptFallback(v)
	return nil
}

type AutoinstallSchemaAptMirrorSelection struct {
	// Primary corresponds to the JSON schema field "primary".
	Primary []string `json:"primary,omitempty" yaml:"primary,omitempty" mapstructure:"primary,omitempty"`
}

type AutoinstallSchemaAptMirrorSelectionPrimaryElem_1 struct {
	// Arches corresponds to the JSON schema field "arches".
	Arches []string `json:"arches,omitempty" yaml:"arches,omitempty" mapstructure:"arches,omitempty"`

	// Uri corresponds to the JSON schema field "uri".
	Uri string `json:"uri" yaml:"uri" mapstructure:"uri"`
}

// UnmarshalJSON implements json.Unmarshaler.
func (j *AutoinstallSchemaAptMirrorSelectionPrimaryElem_1) UnmarshalJSON(value []byte) error {
	var raw map[string]interface{}
	if err := json.Unmarshal(value, &raw); err != nil {
		return err
	}
	if _, ok := raw["uri"]; raw != nil && !ok {
		return fmt.Errorf("field uri in AutoinstallSchemaAptMirrorSelectionPrimaryElem_1: required")
	}
	type Plain AutoinstallSchemaAptMirrorSelectionPrimaryElem_1
	var plain Plain
	if err := json.Unmarshal(value, &plain); err != nil {
		return err
	}
	*j = AutoinstallSchemaAptMirrorSelectionPrimaryElem_1(plain)
	return nil
}

// UnmarshalYAML implements yaml.Unmarshaler.
func (j *AutoinstallSchemaAptMirrorSelectionPrimaryElem_1) UnmarshalYAML(value *yaml.Node) error {
	var raw map[string]interface{}
	if err := value.Decode(&raw); err != nil {
		return err
	}
	if _, ok := raw["uri"]; raw != nil && !ok {
		return fmt.Errorf("field uri in AutoinstallSchemaAptMirrorSelectionPrimaryElem_1: required")
	}
	type Plain AutoinstallSchemaAptMirrorSelectionPrimaryElem_1
	var plain Plain
	if err := value.Decode(&plain); err != nil {
		return err
	}
	*j = AutoinstallSchemaAptMirrorSelectionPrimaryElem_1(plain)
	return nil
}

type AutoinstallSchemaAptPreferencesElem struct {
	// Package corresponds to the JSON schema field "package".
	Package string `json:"package" yaml:"package" mapstructure:"package"`

	// Pin corresponds to the JSON schema field "pin".
	Pin string `json:"pin" yaml:"pin" mapstructure:"pin"`

	// PinPriority corresponds to the JSON schema field "pin-priority".
	PinPriority int `json:"pin-priority" yaml:"pin-priority" mapstructure:"pin-priority"`
}

// UnmarshalYAML implements yaml.Unmarshaler.
func (j *AutoinstallSchemaAptPreferencesElem) UnmarshalYAML(value *yaml.Node) error {
	var raw map[string]interface{}
	if err := value.Decode(&raw); err != nil {
		return err
	}
	if _, ok := raw["package"]; raw != nil && !ok {
		return fmt.Errorf("field package in AutoinstallSchemaAptPreferencesElem: required")
	}
	if _, ok := raw["pin"]; raw != nil && !ok {
		return fmt.Errorf("field pin in AutoinstallSchemaAptPreferencesElem: required")
	}
	if _, ok := raw["pin-priority"]; raw != nil && !ok {
		return fmt.Errorf("field pin-priority in AutoinstallSchemaAptPreferencesElem: required")
	}
	type Plain AutoinstallSchemaAptPreferencesElem
	var plain Plain
	if err := value.Decode(&plain); err != nil {
		return err
	}
	*j = AutoinstallSchemaAptPreferencesElem(plain)
	return nil
}

// UnmarshalJSON implements json.Unmarshaler.
func (j *AutoinstallSchemaAptPreferencesElem) UnmarshalJSON(value []byte) error {
	var raw map[string]interface{}
	if err := json.Unmarshal(value, &raw); err != nil {
		return err
	}
	if _, ok := raw["package"]; raw != nil && !ok {
		return fmt.Errorf("field package in AutoinstallSchemaAptPreferencesElem: required")
	}
	if _, ok := raw["pin"]; raw != nil && !ok {
		return fmt.Errorf("field pin in AutoinstallSchemaAptPreferencesElem: required")
	}
	if _, ok := raw["pin-priority"]; raw != nil && !ok {
		return fmt.Errorf("field pin-priority in AutoinstallSchemaAptPreferencesElem: required")
	}
	type Plain AutoinstallSchemaAptPreferencesElem
	var plain Plain
	if err := json.Unmarshal(value, &plain); err != nil {
		return err
	}
	*j = AutoinstallSchemaAptPreferencesElem(plain)
	return nil
}

type AutoinstallSchemaAptSources map[string]interface{}

type AutoinstallSchemaCodecs struct {
	// Install corresponds to the JSON schema field "install".
	Install *bool `json:"install,omitempty" yaml:"install,omitempty" mapstructure:"install,omitempty"`
}

type AutoinstallSchemaDrivers struct {
	// Install corresponds to the JSON schema field "install".
	Install *bool `json:"install,omitempty" yaml:"install,omitempty" mapstructure:"install,omitempty"`
}

type AutoinstallSchemaIdentity struct {
	// Hostname corresponds to the JSON schema field "hostname".
	Hostname string `json:"hostname" yaml:"hostname" mapstructure:"hostname"`

	// Password corresponds to the JSON schema field "password".
	Password string `json:"password" yaml:"password" mapstructure:"password"`

	// Realname corresponds to the JSON schema field "realname".
	Realname *string `json:"realname,omitempty" yaml:"realname,omitempty" mapstructure:"realname,omitempty"`

	// Username corresponds to the JSON schema field "username".
	Username string `json:"username" yaml:"username" mapstructure:"username"`
}

// UnmarshalYAML implements yaml.Unmarshaler.
func (j *AutoinstallSchemaIdentity) UnmarshalYAML(value *yaml.Node) error {
	var raw map[string]interface{}
	if err := value.Decode(&raw); err != nil {
		return err
	}
	if _, ok := raw["hostname"]; raw != nil && !ok {
		return fmt.Errorf("field hostname in AutoinstallSchemaIdentity: required")
	}
	if _, ok := raw["password"]; raw != nil && !ok {
		return fmt.Errorf("field password in AutoinstallSchemaIdentity: required")
	}
	if _, ok := raw["username"]; raw != nil && !ok {
		return fmt.Errorf("field username in AutoinstallSchemaIdentity: required")
	}
	type Plain AutoinstallSchemaIdentity
	var plain Plain
	if err := value.Decode(&plain); err != nil {
		return err
	}
	*j = AutoinstallSchemaIdentity(plain)
	return nil
}

// UnmarshalJSON implements json.Unmarshaler.
func (j *AutoinstallSchemaIdentity) UnmarshalJSON(value []byte) error {
	var raw map[string]interface{}
	if err := json.Unmarshal(value, &raw); err != nil {
		return err
	}
	if _, ok := raw["hostname"]; raw != nil && !ok {
		return fmt.Errorf("field hostname in AutoinstallSchemaIdentity: required")
	}
	if _, ok := raw["password"]; raw != nil && !ok {
		return fmt.Errorf("field password in AutoinstallSchemaIdentity: required")
	}
	if _, ok := raw["username"]; raw != nil && !ok {
		return fmt.Errorf("field username in AutoinstallSchemaIdentity: required")
	}
	type Plain AutoinstallSchemaIdentity
	var plain Plain
	if err := json.Unmarshal(value, &plain); err != nil {
		return err
	}
	*j = AutoinstallSchemaIdentity(plain)
	return nil
}

type AutoinstallSchemaKernel struct {
	// Flavor corresponds to the JSON schema field "flavor".
	Flavor *string `json:"flavor,omitempty" yaml:"flavor,omitempty" mapstructure:"flavor,omitempty"`

	// Package corresponds to the JSON schema field "package".
	Package *string `json:"package,omitempty" yaml:"package,omitempty" mapstructure:"package,omitempty"`
}

type AutoinstallSchemaKernelCrashDumps struct {
	// Enabled corresponds to the JSON schema field "enabled".
	Enabled *bool `json:"enabled" yaml:"enabled" mapstructure:"enabled"`
}

// UnmarshalJSON implements json.Unmarshaler.
func (j *AutoinstallSchemaKernelCrashDumps) UnmarshalJSON(value []byte) error {
	var raw map[string]interface{}
	if err := json.Unmarshal(value, &raw); err != nil {
		return err
	}
	if _, ok := raw["enabled"]; raw != nil && !ok {
		return fmt.Errorf("field enabled in AutoinstallSchemaKernelCrashDumps: required")
	}
	type Plain AutoinstallSchemaKernelCrashDumps
	var plain Plain
	if err := json.Unmarshal(value, &plain); err != nil {
		return err
	}
	*j = AutoinstallSchemaKernelCrashDumps(plain)
	return nil
}

// UnmarshalYAML implements yaml.Unmarshaler.
func (j *AutoinstallSchemaKernelCrashDumps) UnmarshalYAML(value *yaml.Node) error {
	var raw map[string]interface{}
	if err := value.Decode(&raw); err != nil {
		return err
	}
	if _, ok := raw["enabled"]; raw != nil && !ok {
		return fmt.Errorf("field enabled in AutoinstallSchemaKernelCrashDumps: required")
	}
	type Plain AutoinstallSchemaKernelCrashDumps
	var plain Plain
	if err := value.Decode(&plain); err != nil {
		return err
	}
	*j = AutoinstallSchemaKernelCrashDumps(plain)
	return nil
}

type AutoinstallSchemaKeyboard struct {
	// Layout corresponds to the JSON schema field "layout".
	Layout string `json:"layout" yaml:"layout" mapstructure:"layout"`

	// Toggle corresponds to the JSON schema field "toggle".
	Toggle *string `json:"toggle,omitempty" yaml:"toggle,omitempty" mapstructure:"toggle,omitempty"`

	// Variant corresponds to the JSON schema field "variant".
	Variant *string `json:"variant,omitempty" yaml:"variant,omitempty" mapstructure:"variant,omitempty"`
}

// UnmarshalJSON implements json.Unmarshaler.
func (j *AutoinstallSchemaKeyboard) UnmarshalJSON(value []byte) error {
	var raw map[string]interface{}
	if err := json.Unmarshal(value, &raw); err != nil {
		return err
	}
	if _, ok := raw["layout"]; raw != nil && !ok {
		return fmt.Errorf("field layout in AutoinstallSchemaKeyboard: required")
	}
	type Plain AutoinstallSchemaKeyboard
	var plain Plain
	if err := json.Unmarshal(value, &plain); err != nil {
		return err
	}
	*j = AutoinstallSchemaKeyboard(plain)
	return nil
}

// UnmarshalYAML implements yaml.Unmarshaler.
func (j *AutoinstallSchemaKeyboard) UnmarshalYAML(value *yaml.Node) error {
	var raw map[string]interface{}
	if err := value.Decode(&raw); err != nil {
		return err
	}
	if _, ok := raw["layout"]; raw != nil && !ok {
		return fmt.Errorf("field layout in AutoinstallSchemaKeyboard: required")
	}
	type Plain AutoinstallSchemaKeyboard
	var plain Plain
	if err := value.Decode(&plain); err != nil {
		return err
	}
	*j = AutoinstallSchemaKeyboard(plain)
	return nil
}

type AutoinstallSchemaOem struct {
	// Install corresponds to the JSON schema field "install".
	Install interface{} `json:"install" yaml:"install" mapstructure:"install"`
}

// UnmarshalJSON implements json.Unmarshaler.
func (j *AutoinstallSchemaOem) UnmarshalJSON(value []byte) error {
	var raw map[string]interface{}
	if err := json.Unmarshal(value, &raw); err != nil {
		return err
	}
	if _, ok := raw["install"]; raw != nil && !ok {
		return fmt.Errorf("field install in AutoinstallSchemaOem: required")
	}
	type Plain AutoinstallSchemaOem
	var plain Plain
	if err := json.Unmarshal(value, &plain); err != nil {
		return err
	}
	*j = AutoinstallSchemaOem(plain)
	return nil
}

// UnmarshalYAML implements yaml.Unmarshaler.
func (j *AutoinstallSchemaOem) UnmarshalYAML(value *yaml.Node) error {
	var raw map[string]interface{}
	if err := value.Decode(&raw); err != nil {
		return err
	}
	if _, ok := raw["install"]; raw != nil && !ok {
		return fmt.Errorf("field install in AutoinstallSchemaOem: required")
	}
	type Plain AutoinstallSchemaOem
	var plain Plain
	if err := value.Decode(&plain); err != nil {
		return err
	}
	*j = AutoinstallSchemaOem(plain)
	return nil
}

type AutoinstallSchemaRefreshInstaller struct {
	// Channel corresponds to the JSON schema field "channel".
	Channel *string `json:"channel,omitempty" yaml:"channel,omitempty" mapstructure:"channel,omitempty"`

	// Update corresponds to the JSON schema field "update".
	Update *bool `json:"update,omitempty" yaml:"update,omitempty" mapstructure:"update,omitempty"`
}

type AutoinstallSchemaReporting map[string]struct {
	// Type corresponds to the JSON schema field "type".
	Type string `json:"type" yaml:"type" mapstructure:"type"`

	AdditionalProperties interface{} `mapstructure:",remain"`
}

type AutoinstallSchemaShutdown string

const AutoinstallSchemaShutdownPoweroff AutoinstallSchemaShutdown = "poweroff"
const AutoinstallSchemaShutdownReboot AutoinstallSchemaShutdown = "reboot"

var enumValues_AutoinstallSchemaShutdown = []interface{}{
	"reboot",
	"poweroff",
}

// UnmarshalYAML implements yaml.Unmarshaler.
func (j *AutoinstallSchemaShutdown) UnmarshalYAML(value *yaml.Node) error {
	var v string
	if err := value.Decode(&v); err != nil {
		return err
	}
	var ok bool
	for _, expected := range enumValues_AutoinstallSchemaShutdown {
		if reflect.DeepEqual(v, expected) {
			ok = true
			break
		}
	}
	if !ok {
		return fmt.Errorf("invalid value (expected one of %#v): %#v", enumValues_AutoinstallSchemaShutdown, v)
	}
	*j = AutoinstallSchemaShutdown(v)
	return nil
}

// UnmarshalJSON implements json.Unmarshaler.
func (j *AutoinstallSchemaShutdown) UnmarshalJSON(value []byte) error {
	var v string
	if err := json.Unmarshal(value, &v); err != nil {
		return err
	}
	var ok bool
	for _, expected := range enumValues_AutoinstallSchemaShutdown {
		if reflect.DeepEqual(v, expected) {
			ok = true
			break
		}
	}
	if !ok {
		return fmt.Errorf("invalid value (expected one of %#v): %#v", enumValues_AutoinstallSchemaShutdown, v)
	}
	*j = AutoinstallSchemaShutdown(v)
	return nil
}

type AutoinstallSchemaSnapsElem struct {
	// Channel corresponds to the JSON schema field "channel".
	Channel *string `json:"channel,omitempty" yaml:"channel,omitempty" mapstructure:"channel,omitempty"`

	// Classic corresponds to the JSON schema field "classic".
	Classic *bool `json:"classic,omitempty" yaml:"classic,omitempty" mapstructure:"classic,omitempty"`

	// Name corresponds to the JSON schema field "name".
	Name string `json:"name" yaml:"name" mapstructure:"name"`
}

// UnmarshalJSON implements json.Unmarshaler.
func (j *AutoinstallSchemaSnapsElem) UnmarshalJSON(value []byte) error {
	var raw map[string]interface{}
	if err := json.Unmarshal(value, &raw); err != nil {
		return err
	}
	if _, ok := raw["name"]; raw != nil && !ok {
		return fmt.Errorf("field name in AutoinstallSchemaSnapsElem: required")
	}
	type Plain AutoinstallSchemaSnapsElem
	var plain Plain
	if err := json.Unmarshal(value, &plain); err != nil {
		return err
	}
	*j = AutoinstallSchemaSnapsElem(plain)
	return nil
}

// UnmarshalYAML implements yaml.Unmarshaler.
func (j *AutoinstallSchemaSnapsElem) UnmarshalYAML(value *yaml.Node) error {
	var raw map[string]interface{}
	if err := value.Decode(&raw); err != nil {
		return err
	}
	if _, ok := raw["name"]; raw != nil && !ok {
		return fmt.Errorf("field name in AutoinstallSchemaSnapsElem: required")
	}
	type Plain AutoinstallSchemaSnapsElem
	var plain Plain
	if err := value.Decode(&plain); err != nil {
		return err
	}
	*j = AutoinstallSchemaSnapsElem(plain)
	return nil
}

type AutoinstallSchemaSource struct {
	// Id corresponds to the JSON schema field "id".
	Id *string `json:"id,omitempty" yaml:"id,omitempty" mapstructure:"id,omitempty"`

	// SearchDrivers corresponds to the JSON schema field "search_drivers".
	SearchDrivers *bool `json:"search_drivers,omitempty" yaml:"search_drivers,omitempty" mapstructure:"search_drivers,omitempty"`
}

type AutoinstallSchemaSsh struct {
	// AllowPw corresponds to the JSON schema field "allow-pw".
	AllowPw *bool `json:"allow-pw,omitempty" yaml:"allow-pw,omitempty" mapstructure:"allow-pw,omitempty"`

	// AuthorizedKeys corresponds to the JSON schema field "authorized-keys".
	AuthorizedKeys []string `json:"authorized-keys,omitempty" yaml:"authorized-keys,omitempty" mapstructure:"authorized-keys,omitempty"`

	// InstallServer corresponds to the JSON schema field "install-server".
	InstallServer *bool `json:"install-server,omitempty" yaml:"install-server,omitempty" mapstructure:"install-server,omitempty"`
}

type AutoinstallSchemaStorage map[string]interface{}

// Compatibility only - use ubuntu-pro instead
type AutoinstallSchemaUbuntuAdvantage struct {
	// A valid token starts with a C and is followed by 23 to 29 Base58 characters.
	// See https://pkg.go.dev/github.com/btcsuite/btcutil/base58#CheckEncode
	Token *string `json:"token,omitempty" yaml:"token,omitempty" mapstructure:"token,omitempty"`
}

// UnmarshalYAML implements yaml.Unmarshaler.
func (j *AutoinstallSchemaUbuntuAdvantage) UnmarshalYAML(value *yaml.Node) error {
	type Plain AutoinstallSchemaUbuntuAdvantage
	var plain Plain
	if err := value.Decode(&plain); err != nil {
		return err
	}
	if plain.Token != nil {
		if matched, _ := regexp.MatchString(`^C[1-9A-HJ-NP-Za-km-z]+$`, string(*plain.Token)); !matched {
			return fmt.Errorf("field %s pattern match: must match %s", "Token", `^C[1-9A-HJ-NP-Za-km-z]+$`)
		}
	}
	if plain.Token != nil && len(*plain.Token) < 24 {
		return fmt.Errorf("field %s length: must be >= %d", "token", 24)
	}
	if plain.Token != nil && len(*plain.Token) > 30 {
		return fmt.Errorf("field %s length: must be <= %d", "token", 30)
	}
	*j = AutoinstallSchemaUbuntuAdvantage(plain)
	return nil
}

// UnmarshalJSON implements json.Unmarshaler.
func (j *AutoinstallSchemaUbuntuAdvantage) UnmarshalJSON(value []byte) error {
	type Plain AutoinstallSchemaUbuntuAdvantage
	var plain Plain
	if err := json.Unmarshal(value, &plain); err != nil {
		return err
	}
	if plain.Token != nil {
		if matched, _ := regexp.MatchString(`^C[1-9A-HJ-NP-Za-km-z]+$`, string(*plain.Token)); !matched {
			return fmt.Errorf("field %s pattern match: must match %s", "Token", `^C[1-9A-HJ-NP-Za-km-z]+$`)
		}
	}
	if plain.Token != nil && len(*plain.Token) < 24 {
		return fmt.Errorf("field %s length: must be >= %d", "token", 24)
	}
	if plain.Token != nil && len(*plain.Token) > 30 {
		return fmt.Errorf("field %s length: must be <= %d", "token", 30)
	}
	*j = AutoinstallSchemaUbuntuAdvantage(plain)
	return nil
}

type AutoinstallSchemaUbuntuPro struct {
	// A valid token starts with a C and is followed by 23 to 29 Base58 characters.
	// See https://pkg.go.dev/github.com/btcsuite/btcutil/base58#CheckEncode
	Token *string `json:"token,omitempty" yaml:"token,omitempty" mapstructure:"token,omitempty"`
}

// UnmarshalJSON implements json.Unmarshaler.
func (j *AutoinstallSchemaUbuntuPro) UnmarshalJSON(value []byte) error {
	type Plain AutoinstallSchemaUbuntuPro
	var plain Plain
	if err := json.Unmarshal(value, &plain); err != nil {
		return err
	}
	if plain.Token != nil {
		if matched, _ := regexp.MatchString(`^C[1-9A-HJ-NP-Za-km-z]+$`, string(*plain.Token)); !matched {
			return fmt.Errorf("field %s pattern match: must match %s", "Token", `^C[1-9A-HJ-NP-Za-km-z]+$`)
		}
	}
	if plain.Token != nil && len(*plain.Token) < 24 {
		return fmt.Errorf("field %s length: must be >= %d", "token", 24)
	}
	if plain.Token != nil && len(*plain.Token) > 30 {
		return fmt.Errorf("field %s length: must be <= %d", "token", 30)
	}
	*j = AutoinstallSchemaUbuntuPro(plain)
	return nil
}

// UnmarshalYAML implements yaml.Unmarshaler.
func (j *AutoinstallSchemaUbuntuPro) UnmarshalYAML(value *yaml.Node) error {
	type Plain AutoinstallSchemaUbuntuPro
	var plain Plain
	if err := value.Decode(&plain); err != nil {
		return err
	}
	if plain.Token != nil {
		if matched, _ := regexp.MatchString(`^C[1-9A-HJ-NP-Za-km-z]+$`, string(*plain.Token)); !matched {
			return fmt.Errorf("field %s pattern match: must match %s", "Token", `^C[1-9A-HJ-NP-Za-km-z]+$`)
		}
	}
	if plain.Token != nil && len(*plain.Token) < 24 {
		return fmt.Errorf("field %s length: must be >= %d", "token", 24)
	}
	if plain.Token != nil && len(*plain.Token) > 30 {
		return fmt.Errorf("field %s length: must be <= %d", "token", 30)
	}
	*j = AutoinstallSchemaUbuntuPro(plain)
	return nil
}

type AutoinstallSchemaUpdates string

const AutoinstallSchemaUpdatesAll AutoinstallSchemaUpdates = "all"
const AutoinstallSchemaUpdatesSecurity AutoinstallSchemaUpdates = "security"

var enumValues_AutoinstallSchemaUpdates = []interface{}{
	"security",
	"all",
}

// UnmarshalYAML implements yaml.Unmarshaler.
func (j *AutoinstallSchemaUpdates) UnmarshalYAML(value *yaml.Node) error {
	var v string
	if err := value.Decode(&v); err != nil {
		return err
	}
	var ok bool
	for _, expected := range enumValues_AutoinstallSchemaUpdates {
		if reflect.DeepEqual(v, expected) {
			ok = true
			break
		}
	}
	if !ok {
		return fmt.Errorf("invalid value (expected one of %#v): %#v", enumValues_AutoinstallSchemaUpdates, v)
	}
	*j = AutoinstallSchemaUpdates(v)
	return nil
}

// UnmarshalJSON implements json.Unmarshaler.
func (j *AutoinstallSchemaUpdates) UnmarshalJSON(value []byte) error {
	var v string
	if err := json.Unmarshal(value, &v); err != nil {
		return err
	}
	var ok bool
	for _, expected := range enumValues_AutoinstallSchemaUpdates {
		if reflect.DeepEqual(v, expected) {
			ok = true
			break
		}
	}
	if !ok {
		return fmt.Errorf("invalid value (expected one of %#v): %#v", enumValues_AutoinstallSchemaUpdates, v)
	}
	*j = AutoinstallSchemaUpdates(v)
	return nil
}

type AutoinstallSchemaUserData map[string]interface{}

type AutoinstallSchemaZdevsElem struct {
	// Enabled corresponds to the JSON schema field "enabled".
	Enabled *bool `json:"enabled,omitempty" yaml:"enabled,omitempty" mapstructure:"enabled,omitempty"`

	// Id corresponds to the JSON schema field "id".
	Id *string `json:"id,omitempty" yaml:"id,omitempty" mapstructure:"id,omitempty"`
}

// UnmarshalJSON implements json.Unmarshaler.
func (j *AutoinstallSchema) UnmarshalJSON(value []byte) error {
	var raw map[string]interface{}
	if err := json.Unmarshal(value, &raw); err != nil {
		return err
	}
	if _, ok := raw["version"]; raw != nil && !ok {
		return fmt.Errorf("field version in AutoinstallSchema: required")
	}
	type Plain AutoinstallSchema
	var plain Plain
	if err := json.Unmarshal(value, &plain); err != nil {
		return err
	}
	if 1 < plain.Version {
		return fmt.Errorf("field %s: must be <= %v", "version", 1)
	}
	if 1 > plain.Version {
		return fmt.Errorf("field %s: must be >= %v", "version", 1)
	}
	st := reflect.TypeOf(Plain{})
	for i := range st.NumField() {
		delete(raw, st.Field(i).Name)
		delete(raw, strings.Split(st.Field(i).Tag.Get("json"), ",")[0])
	}
	if err := mapstructure.Decode(raw, &plain.AdditionalProperties); err != nil {
		return err
	}
	*j = AutoinstallSchema(plain)
	return nil
}

// UnmarshalYAML implements yaml.Unmarshaler.
func (j *AutoinstallSchema) UnmarshalYAML(value *yaml.Node) error {
	var raw map[string]interface{}
	if err := value.Decode(&raw); err != nil {
		return err
	}
	if _, ok := raw["version"]; raw != nil && !ok {
		return fmt.Errorf("field version in AutoinstallSchema: required")
	}
	type Plain AutoinstallSchema
	var plain Plain
	if err := value.Decode(&plain); err != nil {
		return err
	}
	if 1 < plain.Version {
		return fmt.Errorf("field %s: must be <= %v", "version", 1)
	}
	if 1 > plain.Version {
		return fmt.Errorf("field %s: must be >= %v", "version", 1)
	}
	st := reflect.TypeOf(Plain{})
	for i := range st.NumField() {
		delete(raw, st.Field(i).Name)
		delete(raw, strings.Split(st.Field(i).Tag.Get("json"), ",")[0])
	}
	if err := mapstructure.Decode(raw, &plain.AdditionalProperties); err != nil {
		return err
	}
	*j = AutoinstallSchema(plain)
	return nil
}
