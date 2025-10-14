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
