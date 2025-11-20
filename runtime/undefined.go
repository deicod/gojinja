package runtime

import "github.com/deicod/gojinja/nodes"

// undefinedType represents an internal sentinel for undefined values that can
// be safely passed through filters and tests without triggering immediate
// errors. It allows constructs like the default filter to detect when a value
// was missing.
type undefinedType interface {
	isUndefined()
	Reason() string
	ToString() (string, error)
}

type baseUndefined struct{}

func (baseUndefined) isUndefined() {}

// ChainableUndefined propagates through attribute and item lookups instead of
// raising errors. It mirrors Jinja2's ChainableUndefined, allowing expressions
// like `user.missing.attribute|default('fallback')` to resolve gracefully
// through the default filter without interrupting rendering.
type ChainableUndefined struct {
	baseUndefined
	name string
}

func (c ChainableUndefined) Reason() string {
	if c.name != "" {
		return "chainable undefined variable '" + c.name + "'"
	}
	return "chainable undefined"
}

func (c ChainableUndefined) ToString() (string, error) {
	return "", nil
}

// SilentUndefined suppresses errors when coerced to a string while still
// signalling that a name or attribute was missing. It matches Jinja2's
// SilentUndefined behaviour, returning an empty string for rendering contexts
// and allowing filters like `default` to detect the missing value.
type SilentUndefined struct {
	baseUndefined
	name string
}

func (s SilentUndefined) Reason() string {
	if s.name != "" {
		return "silent undefined variable '" + s.name + "'"
	}
	return "silent undefined"
}

func (s SilentUndefined) ToString() (string, error) {
	return "", nil
}

type StrictUndefined struct {
	baseUndefined
	name string
}

func (s StrictUndefined) Reason() string {
	if s.name != "" {
		return "strict undefined variable '" + s.name + "'"
	}
	return "strict undefined"
}

func (s StrictUndefined) ToString() (string, error) {
	return "", NewUndefinedError(s.name, nodes.Position{}, nil)
}

type DebugUndefined struct {
	baseUndefined
	name string
}

func (d DebugUndefined) Reason() string {
	if d.name != "" {
		return "undefined variable '" + d.name + "'"
	}
	return "undefined"
}

func (d DebugUndefined) ToString() (string, error) {
	return "", nil
}

var undefinedSentinel undefinedType = DebugUndefined{}

func NewUndefined(name string, strict bool) undefinedType {
	if strict {
		return StrictUndefined{name: name}
	}
	return DebugUndefined{name: name}
}

func isUndefinedValue(value interface{}) bool {
	if value == nil {
		return false
	}
	_, ok := value.(undefinedType)
	return ok
}

func isStrictUndefined(value interface{}) bool {
	_, ok := value.(StrictUndefined)
	return ok
}
