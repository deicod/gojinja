package runtime

import (
	"math"
)

type numberKind int

const (
	numberInteger numberKind = iota
	numberFloat
)

type numberValue struct {
	kind       numberKind
	intValue   int64
	floatValue float64
}

func classifyNumber(value interface{}) (numberValue, bool) {
	switch v := value.(type) {
	case int:
		return numberValue{kind: numberInteger, intValue: int64(v), floatValue: float64(v)}, true
	case int8:
		return numberValue{kind: numberInteger, intValue: int64(v), floatValue: float64(v)}, true
	case int16:
		return numberValue{kind: numberInteger, intValue: int64(v), floatValue: float64(v)}, true
	case int32:
		return numberValue{kind: numberInteger, intValue: int64(v), floatValue: float64(v)}, true
	case int64:
		return numberValue{kind: numberInteger, intValue: v, floatValue: float64(v)}, true
	case uint:
		return classifyUnsigned(uint64(v))
	case uint8:
		return classifyUnsigned(uint64(v))
	case uint16:
		return classifyUnsigned(uint64(v))
	case uint32:
		return classifyUnsigned(uint64(v))
	case uint64:
		return classifyUnsigned(v)
	case uintptr:
		return classifyUnsigned(uint64(v))
	case float32:
		return numberValue{kind: numberFloat, floatValue: float64(v)}, true
	case float64:
		return numberValue{kind: numberFloat, floatValue: v}, true
	case bool:
		if v {
			return numberValue{kind: numberInteger, intValue: 1, floatValue: 1}, true
		}
		return numberValue{kind: numberInteger, intValue: 0, floatValue: 0}, true
	default:
		return numberValue{}, false
	}
}

func classifyUnsigned(v uint64) (numberValue, bool) {
	if v <= uint64(math.MaxInt64) {
		i := int64(v)
		return numberValue{kind: numberInteger, intValue: i, floatValue: float64(i)}, true
	}
	return numberValue{kind: numberFloat, floatValue: float64(v)}, true
}

func (n numberValue) asFloat64() float64 {
	if n.kind == numberFloat {
		return n.floatValue
	}
	return n.floatValue
}

func (n numberValue) isFloat() bool {
	return n.kind == numberFloat
}

func (n numberValue) isZero() bool {
	if n.kind == numberFloat {
		return n.floatValue == 0
	}
	return n.intValue == 0
}
