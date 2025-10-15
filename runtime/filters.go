package runtime

import (
	"encoding/json"
	"fmt"
	"html"
	"math"
	"math/rand"
	"reflect"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"
)

var (
	urlizeEmailPattern       = regexp.MustCompile(`(?i)^[A-Z0-9._%+\-]+@[A-Z0-9.\-]+\.[A-Z]{2,}$`)
	urlizeURLPattern         = regexp.MustCompile(`(?i)^(?:https?://|www\.)[\w\-./~:?#[\]@!$&'()*+,;=%]+$`)
	urlizeBareDomainPattern  = regexp.MustCompile(`(?i)^(?:[a-z0-9](?:[a-z0-9-]{0,61}[a-z0-9])?\.)+[a-z]{2,63}(?:[/?#][^\s]*)?$`)
	xmlAttrInvalidKeyPattern = regexp.MustCompile(`[[:space:]/>=]`)
	uriSchemePattern         = regexp.MustCompile(`(?i)^[a-z][a-z0-9+.-]*:?$`)
)

// registerBuiltinFilters registers all built-in filters with the environment
func (env *Environment) registerBuiltinFilters() {
	// String filters
	env.AddFilter("upper", filterUpper)
	env.AddFilter("lower", filterLower)
	env.AddFilter("capitalize", filterCapitalize)
	env.AddFilter("title", filterTitle)
	env.AddFilter("trim", filterTrim)
	env.AddFilter("ltrim", filterLtrim)
	env.AddFilter("rtrim", filterRtrim)
	env.AddFilter("strip", filterTrim)
	env.AddFilter("striptags", filterStriptags)
	env.AddFilter("replace", filterReplace)
	env.AddFilter("truncate", filterTruncate)
	env.AddFilter("wordcount", filterWordcount)
	env.AddFilter("reverse", filterReverse)
	env.AddFilter("center", filterCenter)
	env.AddFilter("indent", filterIndent)
	env.AddFilter("wordwrap", filterWordwrap)

	// Number filters
	env.AddFilter("round", filterRound)
	env.AddFilter("abs", filterAbs)
	env.AddFilter("int", filterInt)
	env.AddFilter("float", filterFloat)
	env.AddFilter("default", filterDefault)

	// List filters
	env.AddFilter("length", filterLength)
	env.AddFilter("first", filterFirst)
	env.AddFilter("last", filterLast)
	env.AddFilter("join", filterJoin)
	env.AddFilter("sort", filterSort)
	env.AddFilter("unique", filterUnique)
	env.AddFilter("min", filterMin)
	env.AddFilter("max", filterMax)
	env.AddFilter("sum", filterSum)
	env.AddFilter("list", filterList)
	env.AddFilter("slice", filterSlice)
	env.AddFilter("batch", filterBatch)
	env.AddFilter("groupby", filterGroupby)
	env.AddFilter("dictsort", filterDictsort)
	env.AddFilter("dictsortcasesensitive", filterDictsortCaseSensitive)
	env.AddFilter("dictsortreversed", filterDictsortReversed)

	// Utility filters
	env.AddFilter("safe", filterSafe)
	env.AddFilter("escape", filterEscape)
	env.AddFilter("e", filterEscape)
	env.AddFilter("urlencode", filterUrlencode)
	env.AddFilter("escapejs", filterEscapeJS)
	env.AddFilter("filesizeformat", filterFilesizeformat)
	env.AddFilter("floatformat", filterFloatformat)
	env.AddFilter("pprint", filterPprint)
	env.AddFilter("format", filterFormat)
	env.AddFilter("urlize", filterUrlize)
	env.AddFilter("xmlattr", filterXMLAttr)
	env.AddFilter("forceescape", filterForceEscape)
	env.AddFilter("shuffle", filterShuffle)
	env.AddFilter("tojson", filterToJSON)
	env.AddFilter("fromjson", filterFromJSON)
	env.AddFilter("random", filterRandom)
	env.AddFilter("attr", filterAttr)
	env.AddFilter("map", filterMap)
	env.AddFilter("select", filterSelect)
	env.AddFilter("reject", filterReject)
	env.AddFilter("selectattr", filterSelectattr)
	env.AddFilter("rejectattr", filterRejectattr)
	env.AddFilter("do", filterDo)
}

// registerBuiltinTests registers all built-in tests with the environment
func (env *Environment) registerBuiltinTests() {
	env.AddTest("divisibleby", testDivisibleby)
	env.AddTest("defined", testDefined)
	env.AddTest("undefined", testUndefined)
	env.AddTest("none", testNone)
	env.AddTest("null", testNone) // Alias for none
	env.AddTest("boolean", testBoolean)
	env.AddTest("true", testTrue)
	env.AddTest("false", testFalse)
	env.AddTest("number", testNumber)
	env.AddTest("integer", testInteger)
	env.AddTest("float", testFloat)
	env.AddTest("string", testString)
	env.AddTest("sequence", testSequence)
	env.AddTest("mapping", testMapping)
	env.AddTest("iterable", testIterable)
	env.AddTest("callable", testCallable)
	env.AddTest("sameas", testSameas)
	env.AddTest("escaped", testEscaped)
	env.AddTest("module", testModule)
	env.AddTest("list", testList)
	env.AddTest("tuple", testTuple)
	env.AddTest("dict", testDict)
	env.AddTest("lower", testLowerTest)
	env.AddTest("upper", testUpperTest)
	env.AddTest("even", testEven)
	env.AddTest("odd", testOdd)
	env.AddTest("in", testInTest)
	env.AddTest("filter", testFilter)
	env.AddTest("test", testTest)
	env.AddTest("equalto", testEq)
	env.AddTest("==", testEq)
	env.AddTest("!=", testNe)
	env.AddTest("eq", testEq)
	env.AddTest("ne", testNe)
	env.AddTest("lt", testLt)
	env.AddTest("le", testLe)
	env.AddTest("gt", testGt)
	env.AddTest("ge", testGe)
	env.AddTest(">", testGt)
	env.AddTest("<", testLt)
	env.AddTest(">=", testGe)
	env.AddTest("<=", testLe)
	env.AddTest("greaterthan", testGt)
	env.AddTest("lessthan", testLt)
	env.AddTest("matching", testMatching)
	env.AddTest("search", testSearch)
	env.AddTest("startingwith", testStartingWith)
	env.AddTest("endingwith", testEndingWith)
	env.AddTest("containing", testContaining)
	env.AddTest("infinite", testInfinite)
	env.AddTest("nan", testNan)
	env.AddTest("finite", testFinite)
}

// String filters

func filterUpper(ctx *Context, value interface{}, args ...interface{}) (interface{}, error) {
	str := toString(value)
	return strings.ToUpper(str), nil
}

func filterLower(ctx *Context, value interface{}, args ...interface{}) (interface{}, error) {
	str := toString(value)
	return strings.ToLower(str), nil
}

func filterCapitalize(ctx *Context, value interface{}, args ...interface{}) (interface{}, error) {
	str := toString(value)
	if str == "" {
		return str, nil
	}
	runes := []rune(str)
	return strings.ToUpper(string(runes[0])) + strings.ToLower(string(runes[1:])), nil
}

func filterTitle(ctx *Context, value interface{}, args ...interface{}) (interface{}, error) {
	str := toString(value)
	if str == "" {
		return str, nil
	}
	return strings.Title(str), nil
}

func filterTrim(ctx *Context, value interface{}, args ...interface{}) (interface{}, error) {
	str := toString(value)
	chars := ""
	if len(args) > 0 {
		chars = toString(args[0])
	}
	if chars != "" {
		return strings.Trim(str, chars), nil
	}
	return strings.TrimSpace(str), nil
}

func filterLtrim(ctx *Context, value interface{}, args ...interface{}) (interface{}, error) {
	str := toString(value)
	chars := ""
	if len(args) > 0 {
		chars = toString(args[0])
	}
	if chars != "" {
		return strings.TrimLeft(str, chars), nil
	}
	return strings.TrimLeftFunc(str, unicode.IsSpace), nil
}

func filterRtrim(ctx *Context, value interface{}, args ...interface{}) (interface{}, error) {
	str := toString(value)
	chars := ""
	if len(args) > 0 {
		chars = toString(args[0])
	}
	if chars != "" {
		return strings.TrimRight(str, chars), nil
	}
	return strings.TrimRightFunc(str, unicode.IsSpace), nil
}

func filterStriptags(ctx *Context, value interface{}, args ...interface{}) (interface{}, error) {
	str := toString(value)
	var result strings.Builder
	inTag := false

	for _, r := range str {
		if r == '<' {
			inTag = true
		} else if r == '>' {
			inTag = false
		} else if !inTag {
			result.WriteRune(r)
		}
	}

	return result.String(), nil
}

func filterReplace(ctx *Context, value interface{}, args ...interface{}) (interface{}, error) {
	if len(args) < 2 {
		return nil, fmt.Errorf("replace filter requires at least 2 arguments")
	}

	str := toString(value)
	old := toString(args[0])
	new := toString(args[1])
	count := -1

	if len(args) > 2 {
		if c, ok := args[2].(int); ok {
			count = c
		}
	}

	return strings.Replace(str, old, new, count), nil
}

func filterTruncate(ctx *Context, value interface{}, args ...interface{}) (interface{}, error) {
	str := toString(value)
	length := 255
	killwords := false
	end := "..."

	if len(args) > 0 {
		if l, ok := args[0].(int); ok {
			length = l
		}
	}

	if len(args) > 1 {
		if kw, ok := args[1].(bool); ok {
			killwords = kw
		}
	}

	if len(args) > 2 {
		end = toString(args[2])
	}

	if len(str) <= length {
		return str, nil
	}

	if killwords {
		return str[:length-len(end)] + end, nil
	}

	// Find last space within length limit
	lastSpace := strings.LastIndex(str[:length], " ")
	if lastSpace == -1 {
		return str[:length-len(end)] + end, nil
	}

	return str[:lastSpace] + end, nil
}

func filterWordcount(ctx *Context, value interface{}, args ...interface{}) (interface{}, error) {
	str := toString(value)
	if str == "" {
		return 0, nil
	}
	words := strings.Fields(str)
	return len(words), nil
}

func filterReverse(ctx *Context, value interface{}, args ...interface{}) (interface{}, error) {
	switch v := value.(type) {
	case string:
		runes := []rune(v)
		for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
			runes[i], runes[j] = runes[j], runes[i]
		}
		return string(runes), nil
	case []interface{}:
		result := make([]interface{}, len(v))
		for i, item := range v {
			result[len(v)-1-i] = item
		}
		return result, nil
	default:
		// Try to convert to slice
		val := reflect.ValueOf(value)
		if val.Kind() == reflect.Slice || val.Kind() == reflect.Array {
			result := make([]interface{}, val.Len())
			for i := 0; i < val.Len(); i++ {
				result[val.Len()-1-i] = val.Index(i).Interface()
			}
			return result, nil
		}
		return nil, fmt.Errorf("reverse filter requires a string or sequence")
	}
}

func filterCenter(ctx *Context, value interface{}, args ...interface{}) (interface{}, error) {
	str := toString(value)
	width := 80

	if len(args) > 0 {
		if w, ok := args[0].(int); ok {
			width = w
		}
	}

	if len(str) >= width {
		return str, nil
	}

	padding := width - len(str)
	leftPadding := padding / 2
	rightPadding := padding - leftPadding

	return strings.Repeat(" ", leftPadding) + str + strings.Repeat(" ", rightPadding), nil
}

func filterIndent(ctx *Context, value interface{}, args ...interface{}) (interface{}, error) {
	str := toString(value)
	width := 4
	indentFirst := false

	if len(args) > 0 {
		if w, ok := args[0].(int); ok {
			width = w
		}
	}

	if len(args) > 1 {
		if ifirst, ok := args[1].(bool); ok {
			indentFirst = ifirst
		}
	}

	prefix := strings.Repeat(" ", width)
	lines := strings.Split(str, "\n")

	if indentFirst {
		for i := range lines {
			lines[i] = prefix + lines[i]
		}
	} else {
		for i := 1; i < len(lines); i++ {
			lines[i] = prefix + lines[i]
		}
	}

	return strings.Join(lines, "\n"), nil
}

func filterWordwrap(ctx *Context, value interface{}, args ...interface{}) (interface{}, error) {
	text := toString(value)
	kwargs, args := extractKwargs(args)

	width := 79
	breakLongWords := true
	wrapString := ""
	wrapProvided := false
	breakOnHyphens := true

	if len(args) > 0 {
		if w, ok := toInt(args[0]); ok {
			width = w
		}
	}
	if len(args) > 1 {
		breakLongWords = isTruthyValue(args[1])
	}
	if len(args) > 2 {
		wrapString = toString(args[2])
		wrapProvided = true
	}
	if len(args) > 3 {
		breakOnHyphens = isTruthyValue(args[3])
	}

	if kwargs != nil {
		if val, ok := kwargs["width"]; ok {
			if w, ok := toInt(val); ok {
				width = w
			}
		}
		if val, ok := kwargs["break_long_words"]; ok {
			breakLongWords = isTruthyValue(val)
		}
		if val, ok := kwargs["wrapstring"]; ok {
			wrapString = toString(val)
			wrapProvided = true
		}
		if val, ok := kwargs["break_on_hyphens"]; ok {
			breakOnHyphens = isTruthyValue(val)
		}
	}

	if width <= 0 {
		return nil, fmt.Errorf("wordwrap filter requires width > 0")
	}

	if !wrapProvided {
		if ctx != nil && ctx.environment != nil {
			wrapString = ctx.environment.NewlineSequence()
		}
		if wrapString == "" {
			wrapString = "\n"
		}
	}

	normalized := strings.ReplaceAll(text, "\r\n", "\n")
	normalized = strings.ReplaceAll(normalized, "\r", "\n")

	lines := strings.Split(normalized, "\n")
	if len(lines) > 0 && strings.HasSuffix(normalized, "\n") {
		lines = lines[:len(lines)-1]
	}

	if len(lines) == 0 {
		return "", nil
	}

	wrapped := make([]string, 0, len(lines))
	for _, line := range lines {
		segments := wrapParagraph(line, width, breakLongWords, breakOnHyphens)
		wrapped = append(wrapped, strings.Join(segments, wrapString))
	}

	return strings.Join(wrapped, wrapString), nil
}

// Number filters

func filterRound(ctx *Context, value interface{}, args ...interface{}) (interface{}, error) {
	var num float64
	var ok bool

	switch v := value.(type) {
	case int:
		num = float64(v)
		ok = true
	case int64:
		num = float64(v)
		ok = true
	case float64:
		num = v
		ok = true
	case float32:
		num = float64(v)
		ok = true
	default:
		if s := toString(value); s != "" {
			if parsed, err := strconv.ParseFloat(s, 64); err == nil {
				num = parsed
				ok = true
			}
		}
	}

	if !ok {
		return nil, fmt.Errorf("round filter requires a number")
	}

	precision := 0
	method := "common"

	if len(args) > 0 {
		// Try int first
		if p, ok := args[0].(int); ok {
			precision = p
		} else if p64, ok := args[0].(int64); ok {
			precision = int(p64)
		} else if pf, ok := args[0].(float64); ok {
			precision = int(pf)
		}
	}

	if len(args) > 1 {
		method = toString(args[1])
	}

	multiplier := math.Pow10(precision)
	rounded := num * multiplier

	switch method {
	case "common":
		rounded = math.Floor(rounded + 0.5)
	case "floor":
		rounded = math.Floor(rounded)
	case "ceil":
		rounded = math.Ceil(rounded)
	default:
		return nil, fmt.Errorf("unknown rounding method: %s", method)
	}

	result := rounded / multiplier

	// If precision is specified, return a formatted string to preserve decimal places
	// Otherwise return the numeric value
	if len(args) > 0 && precision >= 0 {
		return fmt.Sprintf("%.*f", precision, result), nil
	}

	return result, nil
}

func filterAbs(ctx *Context, value interface{}, args ...interface{}) (interface{}, error) {
	switch v := value.(type) {
	case int:
		if v < 0 {
			return -v, nil
		}
		return v, nil
	case int64:
		if v < 0 {
			return -v, nil
		}
		return v, nil
	case float64:
		return math.Abs(v), nil
	case float32:
		return float32(math.Abs(float64(v))), nil
	default:
		return nil, fmt.Errorf("abs filter requires a number")
	}
}

func filterInt(ctx *Context, value interface{}, args ...interface{}) (interface{}, error) {
	switch v := value.(type) {
	case int:
		return v, nil
	case int64:
		return int(v), nil
	case float64:
		return int(v), nil
	case float32:
		return int(v), nil
	case string:
		if i, err := strconv.Atoi(v); err == nil {
			return i, nil
		}
		if f, err := strconv.ParseFloat(v, 64); err == nil {
			return int(f), nil
		}
		return nil, fmt.Errorf("cannot convert '%s' to int", v)
	default:
		return nil, fmt.Errorf("int filter requires a number or string")
	}
}

func filterFloat(ctx *Context, value interface{}, args ...interface{}) (interface{}, error) {
	switch v := value.(type) {
	case int:
		return float64(v), nil
	case int64:
		return float64(v), nil
	case float64:
		return v, nil
	case float32:
		return float64(v), nil
	case string:
		if f, err := strconv.ParseFloat(v, 64); err == nil {
			return f, nil
		}
		return nil, fmt.Errorf("cannot convert '%s' to float", v)
	default:
		return nil, fmt.Errorf("float filter requires a number or string")
	}
}

func filterDefault(ctx *Context, value interface{}, args ...interface{}) (interface{}, error) {
	var defaultValue interface{} = ""
	if len(args) > 0 {
		defaultValue = args[0]
	}

	applyOnFalsy := false
	if len(args) > 1 {
		switch v := args[1].(type) {
		case bool:
			applyOnFalsy = v
		case int:
			applyOnFalsy = v != 0
		case string:
			applyOnFalsy = strings.EqualFold(v, "true")
		}
	}

	if isUndefinedValue(value) {
		return defaultValue, nil
	}

	if applyOnFalsy && !isTruthyValue(value) {
		return defaultValue, nil
	}

	return value, nil
}

// List filters

func filterLength(ctx *Context, value interface{}, args ...interface{}) (interface{}, error) {
	switch v := value.(type) {
	case string:
		return len(v), nil
	case []interface{}:
		return len(v), nil
	case map[interface{}]interface{}:
		return len(v), nil
	case map[string]interface{}:
		return len(v), nil
	default:
		// Try reflection
		val := reflect.ValueOf(value)
		switch val.Kind() {
		case reflect.Slice, reflect.Array, reflect.Map, reflect.String:
			return val.Len(), nil
		default:
			return 0, fmt.Errorf("length filter requires a sequence or mapping")
		}
	}
}

func filterFirst(ctx *Context, value interface{}, args ...interface{}) (interface{}, error) {
	switch v := value.(type) {
	case string:
		if len(v) == 0 {
			return "", nil
		}
		return string(v[0]), nil
	case []interface{}:
		if len(v) == 0 {
			return nil, nil
		}
		return v[0], nil
	default:
		// Try reflection
		val := reflect.ValueOf(value)
		if val.Kind() == reflect.Slice || val.Kind() == reflect.Array {
			if val.Len() == 0 {
				return nil, nil
			}
			return val.Index(0).Interface(), nil
		}
		return nil, fmt.Errorf("first filter requires a sequence")
	}
}

func filterLast(ctx *Context, value interface{}, args ...interface{}) (interface{}, error) {
	switch v := value.(type) {
	case string:
		if len(v) == 0 {
			return "", nil
		}
		runes := []rune(v)
		return string(runes[len(runes)-1]), nil
	case []interface{}:
		if len(v) == 0 {
			return nil, nil
		}
		return v[len(v)-1], nil
	default:
		// Try reflection
		val := reflect.ValueOf(value)
		if val.Kind() == reflect.Slice || val.Kind() == reflect.Array {
			if val.Len() == 0 {
				return nil, nil
			}
			return val.Index(val.Len() - 1).Interface(), nil
		}
		return nil, fmt.Errorf("last filter requires a sequence")
	}
}

func filterJoin(ctx *Context, value interface{}, args ...interface{}) (interface{}, error) {
	separator := ""
	if len(args) > 0 {
		separator = toString(args[0])
	}

	switch v := value.(type) {
	case []interface{}:
		strs := make([]string, len(v))
		for i, item := range v {
			strs[i] = toString(item)
		}
		return strings.Join(strs, separator), nil
	case []string:
		return strings.Join(v, separator), nil
	default:
		// Try to convert to slice of strings
		val := reflect.ValueOf(value)
		if val.Kind() == reflect.Slice || val.Kind() == reflect.Array {
			strs := make([]string, val.Len())
			for i := 0; i < val.Len(); i++ {
				strs[i] = toString(val.Index(i).Interface())
			}
			return strings.Join(strs, separator), nil
		}
		return nil, fmt.Errorf("join filter requires a sequence")
	}
}

func filterSort(ctx *Context, value interface{}, args ...interface{}) (interface{}, error) {
	reverse := false
	caseSensitive := true
	attribute := ""

	if len(args) > 0 {
		if r, ok := args[0].(bool); ok {
			reverse = r
		}
	}

	if len(args) > 1 {
		if cs, ok := args[1].(bool); ok {
			caseSensitive = cs
		}
	}

	if len(args) > 2 {
		attribute = toString(args[2])
	}

	switch v := value.(type) {
	case []interface{}:
		// Make a copy to avoid modifying original
		result := make([]interface{}, len(v))
		copy(result, v)

		if attribute != "" {
			// Sort by attribute
			sort.Slice(result, func(i, j int) bool {
				attrI, _ := getAttribute(result[i], attribute)
				attrJ, _ := getAttribute(result[j], attribute)
				cmp := compareValues(attrI, attrJ, caseSensitive)
				if reverse {
					return cmp > 0
				}
				return cmp < 0
			})
		} else {
			// Sort by value
			sort.Slice(result, func(i, j int) bool {
				cmp := compareValues(result[i], result[j], caseSensitive)
				if reverse {
					return cmp > 0
				}
				return cmp < 0
			})
		}

		return result, nil
	case []string:
		// Make a copy
		result := make([]string, len(v))
		copy(result, v)

		sort.Slice(result, func(i, j int) bool {
			cmp := compareValues(result[i], result[j], caseSensitive)
			if reverse {
				return cmp > 0
			}
			return cmp < 0
		})

		return result, nil
	default:
		return nil, fmt.Errorf("sort filter requires a sequence")
	}
}

func filterUnique(ctx *Context, value interface{}, args ...interface{}) (interface{}, error) {
	switch v := value.(type) {
	case []interface{}:
		seen := make(map[interface{}]bool)
		result := make([]interface{}, 0)
		for _, item := range v {
			if !seen[item] {
				seen[item] = true
				result = append(result, item)
			}
		}
		return result, nil
	case []string:
		seen := make(map[string]bool)
		result := make([]string, 0)
		for _, item := range v {
			if !seen[item] {
				seen[item] = true
				result = append(result, item)
			}
		}
		return result, nil
	default:
		return nil, fmt.Errorf("unique filter requires a sequence")
	}
}

func filterMin(ctx *Context, value interface{}, args ...interface{}) (interface{}, error) {
	switch v := value.(type) {
	case []interface{}:
		if len(v) == 0 {
			return nil, nil
		}
		min := v[0]
		for _, item := range v[1:] {
			if compareValues(item, min, true) < 0 {
				min = item
			}
		}
		return min, nil
	case []string:
		if len(v) == 0 {
			return "", nil
		}
		min := v[0]
		for _, item := range v[1:] {
			if compareValues(item, min, true) < 0 {
				min = item
			}
		}
		return min, nil
	default:
		return nil, fmt.Errorf("min filter requires a sequence")
	}
}

func filterMax(ctx *Context, value interface{}, args ...interface{}) (interface{}, error) {
	switch v := value.(type) {
	case []interface{}:
		if len(v) == 0 {
			return nil, nil
		}
		max := v[0]
		for _, item := range v[1:] {
			if compareValues(item, max, true) > 0 {
				max = item
			}
		}
		return max, nil
	case []string:
		if len(v) == 0 {
			return "", nil
		}
		max := v[0]
		for _, item := range v[1:] {
			if compareValues(item, max, true) > 0 {
				max = item
			}
		}
		return max, nil
	default:
		return nil, fmt.Errorf("max filter requires a sequence")
	}
}

func filterSum(ctx *Context, value interface{}, args ...interface{}) (interface{}, error) {
	start := 0.0
	if len(args) > 0 {
		if s, ok := args[0].(float64); ok {
			start = s
		}
	}

	switch v := value.(type) {
	case []interface{}:
		sum := start
		for _, item := range v {
			if num, ok := toFloat64(item); ok {
				sum += num
			}
		}
		return sum, nil
	default:
		return nil, fmt.Errorf("sum filter requires a sequence of numbers")
	}
}

func filterList(ctx *Context, value interface{}, args ...interface{}) (interface{}, error) {
	switch v := value.(type) {
	case []interface{}:
		return v, nil
	case string:
		result := make([]interface{}, len(v))
		for i, r := range v {
			result[i] = string(r)
		}
		return result, nil
	case map[interface{}]interface{}:
		result := make([]interface{}, 0, len(v))
		for _, val := range v {
			result = append(result, val)
		}
		return result, nil
	case map[string]interface{}:
		result := make([]interface{}, 0, len(v))
		for _, val := range v {
			result = append(result, val)
		}
		return result, nil
	default:
		// Try reflection
		val := reflect.ValueOf(value)
		if val.Kind() == reflect.Slice || val.Kind() == reflect.Array {
			result := make([]interface{}, val.Len())
			for i := 0; i < val.Len(); i++ {
				result[i] = val.Index(i).Interface()
			}
			return result, nil
		}
		return nil, fmt.Errorf("list filter requires a sequence or mapping")
	}
}

func filterSlice(ctx *Context, value interface{}, args ...interface{}) (interface{}, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("slice filter requires the number of slices")
	}
	slices, ok := toInt(args[0])
	if !ok || slices <= 0 {
		return nil, fmt.Errorf("slice filter requires a positive integer")
	}
	items, err := sequenceToSlice(value)
	if err != nil {
		return nil, err
	}

	fillWith := interface{}(nil)
	if len(args) > 1 {
		fillWith = args[1]
	}

	length := len(items)
	itemsPerSlice := 0
	if slices > 0 {
		itemsPerSlice = length / slices
	}
	slicesWithExtra := 0
	if slices > 0 {
		slicesWithExtra = length % slices
	}
	offset := 0
	result := make([][]interface{}, 0, slices)

	for sliceNumber := 0; sliceNumber < slices; sliceNumber++ {
		start := offset + sliceNumber*itemsPerSlice
		if start > length {
			start = length
		}
		if sliceNumber < slicesWithExtra {
			offset++
		}
		end := offset + (sliceNumber+1)*itemsPerSlice
		if end > length {
			end = length
		}
		tmp := append([]interface{}(nil), items[start:end]...)
		if fillWith != nil && sliceNumber >= slicesWithExtra {
			tmp = append(tmp, fillWith)
		}
		result = append(result, tmp)
	}
	return result, nil
}

func filterGroupby(ctx *Context, value interface{}, args ...interface{}) (interface{}, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("groupby filter requires 1 argument (attribute)")
	}

	attribute := toString(args[0])

	switch v := value.(type) {
	case []interface{}:
		groups := make(map[interface{}][]interface{})
		for _, item := range v {
			key, _ := getAttribute(item, attribute)
			groups[key] = append(groups[key], item)
		}

		result := make([]interface{}, 0, len(groups))
		for key, items := range groups {
			result = append(result, map[string]interface{}{
				"grouper": key,
				"list":    items,
			})
		}
		return result, nil
	default:
		return nil, fmt.Errorf("groupby filter requires a sequence")
	}
}

func filterDictsort(ctx *Context, value interface{}, args ...interface{}) (interface{}, error) {
	return dictsortWithDefaults(value, args, dictsortDefaults{})
}

func filterDictsortCaseSensitive(ctx *Context, value interface{}, args ...interface{}) (interface{}, error) {
	return dictsortWithDefaults(value, args, dictsortDefaults{caseSensitive: true})
}

func filterDictsortReversed(ctx *Context, value interface{}, args ...interface{}) (interface{}, error) {
	return dictsortWithDefaults(value, args, dictsortDefaults{reverse: true})
}

// Utility filters

func filterSafe(ctx *Context, value interface{}, args ...interface{}) (interface{}, error) {
	// In a full implementation, this would wrap the value in a Markup type
	// For now, just return the value
	return value, nil
}

func filterDo(ctx *Context, value interface{}, args ...interface{}) (interface{}, error) {
	if len(args) > 0 {
		return nil, fmt.Errorf("do filter does not accept arguments")
	}
	return nil, nil
}

func filterEscape(ctx *Context, value interface{}, args ...interface{}) (interface{}, error) {
	str := toString(value)
	return html.EscapeString(str), nil
}

func filterUrlencode(ctx *Context, value interface{}, args ...interface{}) (interface{}, error) {
	str := toString(value)
	// Simple URL encoding implementation
	var result strings.Builder
	for _, r := range str {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') ||
			r == '-' || r == '_' || r == '.' || r == '~' {
			result.WriteRune(r)
		} else {
			result.WriteString(fmt.Sprintf("%%%02X", r))
		}
	}
	return result.String(), nil
}

func filterEscapeJS(ctx *Context, value interface{}, args ...interface{}) (interface{}, error) {
	s := toString(value)
	if s == "" {
		return "", nil
	}

	var b strings.Builder
	b.Grow(len(s))
	for _, r := range s {
		switch r {
		case '\\':
			b.WriteString("\\\\")
		case '\n':
			b.WriteString("\\n")
		case '\r':
			b.WriteString("\\r")
		case '\t':
			b.WriteString("\\t")
		case '\b':
			b.WriteString("\\b")
		case '\f':
			b.WriteString("\\f")
		case '\v':
			b.WriteString("\\u000b")
		case '"':
			b.WriteString("\\\"")
		case '\'':
			b.WriteString("\\'")
		case '<':
			b.WriteString(fmt.Sprintf("\\u%04x", r))
		case '>':
			b.WriteString(fmt.Sprintf("\\u%04x", r))
		case '&':
			b.WriteString(fmt.Sprintf("\\u%04x", r))
		case '=':
			b.WriteString(fmt.Sprintf("\\u%04x", r))
		case '`':
			b.WriteString(fmt.Sprintf("\\u%04x", r))
		default:
			if r < 0x20 || r > 0x7e {
				if r <= 0xffff {
					b.WriteString(fmt.Sprintf("\\u%04x", r))
				} else {
					for _, rr := range string(r) {
						b.WriteString(fmt.Sprintf("\\u%04x", rr))
					}
				}
			} else {
				b.WriteRune(r)
			}
		}
	}
	return b.String(), nil
}

func filterFilesizeformat(ctx *Context, value interface{}, args ...interface{}) (interface{}, error) {
	kwargs, positional := extractKwargs(args)
	args = positional

	var (
		size float64
		ok   bool
	)

	switch v := value.(type) {
	case nil:
		return nil, fmt.Errorf("float() argument must be a string or a real number, not 'NoneType'")
	case bool:
		if v {
			size = 1
		} else {
			size = 0
		}
		ok = true
	default:
		size, ok = toFloat64(value)
	}

	if !ok {
		if str, isString := value.(string); isString {
			return nil, fmt.Errorf("could not convert string to float: '%s'", str)
		}
		return nil, fmt.Errorf("float() argument must be a string or a real number, not '%T'", value)
	}

	binary := false
	if len(args) > 0 {
		binary = isTruthyValue(args[0])
	}
	if kwargs != nil {
		if val, exists := kwargs["binary"]; exists {
			binary = isTruthyValue(val)
		}
	}

	negative := size < 0
	if negative {
		size = math.Abs(size)
	}

	if size == 0 {
		result := "0 Bytes"
		if negative {
			result = "-" + result
		}
		return result, nil
	}

	if math.Abs(size-1.0) < 1e-9 {
		result := "1 Byte"
		if negative {
			result = "-" + result
		}
		return result, nil
	}

	base := 1000.0
	units := []string{"kB", "MB", "GB", "TB", "PB", "EB", "ZB", "YB"}
	if binary {
		base = 1024.0
		units = []string{"KiB", "MiB", "GiB", "TiB", "PiB", "EiB", "ZiB", "YiB"}
	}

	if size < base {
		count := int64(math.Floor(size))
		result := fmt.Sprintf("%d Bytes", count)
		if negative {
			result = "-" + result
		}
		return result, nil
	}

	size /= base
	unitIndex := 0
	for size >= base && unitIndex < len(units)-1 {
		size /= base
		unitIndex++
	}
	if size >= base {
		for size >= base {
			size /= base
		}
		unitIndex = len(units) - 1
	}

	result := fmt.Sprintf("%.1f %s", size, units[unitIndex])
	if negative {
		result = "-" + result
	}
	return result, nil
}

func filterFloatformat(ctx *Context, value interface{}, args ...interface{}) (interface{}, error) {
	num, ok := toFloat64(value)
	if !ok {
		return value, nil
	}

	precision := -1
	trim := true
	if len(args) > 0 {
		switch v := args[0].(type) {
		case int:
			precision = v
		case int64:
			precision = int(v)
		case string:
			if strings.HasPrefix(v, "-") {
				trim = false
				v = v[1:]
			}
			if v != "" {
				if p, err := strconv.Atoi(v); err == nil {
					precision = p
				}
			}
		}
	}

	if precision < 0 {
		return strconv.FormatFloat(num, 'f', -1, 64), nil
	}

	format := fmt.Sprintf("%%.%df", precision)
	result := fmt.Sprintf(format, num)
	if trim {
		result = strings.TrimRight(result, "0")
		result = strings.TrimRight(result, ".")
	}
	return result, nil
}

func filterPprint(ctx *Context, value interface{}, args ...interface{}) (interface{}, error) {
	indent := "  "
	if len(args) > 0 {
		if s, ok := args[0].(string); ok && s != "" {
			indent = s
		}
	}

	data, err := json.MarshalIndent(value, "", indent)
	if err != nil {
		return nil, err
	}
	return string(data), nil
}

func filterFormat(ctx *Context, value interface{}, args ...interface{}) (interface{}, error) {
	format := toString(value)
	if len(args) == 1 {
		switch m := args[0].(type) {
		case map[string]interface{}:
			placeholderRE := regexp.MustCompile(`%\(([^)]+)\)[sd]`)
			ordered := make([]interface{}, 0)
			converted := placeholderRE.ReplaceAllStringFunc(format, func(match string) string {
				name := placeholderRE.FindStringSubmatch(match)[1]
				ordered = append(ordered, m[name])
				return "%v"
			})
			return fmt.Sprintf(converted, ordered...), nil
		case map[interface{}]interface{}:
			pm := make(map[string]interface{}, len(m))
			for k, v := range m {
				if key, ok := k.(string); ok {
					pm[key] = v
				}
			}
			return filterFormat(ctx, value, pm)
		}
	}
	return fmt.Sprintf(format, args...), nil
}

func filterForceEscape(ctx *Context, value interface{}, args ...interface{}) (interface{}, error) {
	return Markup(html.EscapeString(toString(value))), nil
}

func filterShuffle(ctx *Context, value interface{}, args ...interface{}) (interface{}, error) {
	items, err := sequenceToSlice(value)
	if err != nil {
		return nil, err
	}

	cpy := append([]interface{}(nil), items...)
	seed := time.Now().UnixNano()
	if len(args) > 0 {
		if s, ok := toInt(args[0]); ok {
			seed = int64(s)
		}
	}
	rnd := rand.New(rand.NewSource(seed))
	rnd.Shuffle(len(cpy), func(i, j int) {
		cpy[i], cpy[j] = cpy[j], cpy[i]
	})
	return cpy, nil
}

func filterBatch(ctx *Context, value interface{}, args ...interface{}) (interface{}, error) {
	items, err := sequenceToSlice(value)
	if err != nil {
		return nil, err
	}
	if len(args) < 1 {
		return nil, fmt.Errorf("batch filter requires a size argument")
	}
	size, ok := toInt(args[0])
	if !ok || size <= 0 {
		return nil, fmt.Errorf("batch size must be a positive integer")
	}
	filler := interface{}(nil)
	if len(args) > 1 {
		filler = args[1]
	}

	batches := make([][]interface{}, 0, (len(items)+size-1)/size)
	current := make([]interface{}, 0, size)
	for _, item := range items {
		current = append(current, item)
		if len(current) == size {
			batches = append(batches, append([]interface{}(nil), current...))
			current = current[:0]
		}
	}
	if len(current) > 0 {
		if filler != nil && len(current) < size {
			current = append(current, repeatValue(filler, size-len(current))...)
		}
		batches = append(batches, append([]interface{}(nil), current...))
	}
	return batches, nil
}

func filterToJSON(ctx *Context, value interface{}, args ...interface{}) (interface{}, error) {
	if len(args) > 0 {
		if indentStr, ok := args[0].(string); ok && indentStr != "" {
			data, err := json.MarshalIndent(value, "", indentStr)
			if err != nil {
				return nil, err
			}
			return string(data), nil
		}
	}
	data, err := json.Marshal(value)
	if err != nil {
		return nil, err
	}
	return string(data), nil
}

func filterFromJSON(ctx *Context, value interface{}, args ...interface{}) (interface{}, error) {
	str := toString(value)
	if str == "" {
		return nil, nil
	}
	var result interface{}
	if err := json.Unmarshal([]byte(str), &result); err != nil {
		return nil, err
	}
	return result, nil
}

func filterRandom(ctx *Context, value interface{}, args ...interface{}) (interface{}, error) {
	items, err := sequenceToSlice(value)
	if err != nil {
		return nil, err
	}
	if len(items) == 0 {
		return nil, fmt.Errorf("cannot choose from an empty sequence")
	}

	var rnd *rand.Rand
	if len(args) > 0 {
		if seed, ok := toInt(args[0]); ok {
			rnd = rand.New(rand.NewSource(int64(seed)))
		}
	}
	if rnd == nil {
		rnd = rand.New(rand.NewSource(time.Now().UnixNano()))
	}

	choice := items[rnd.Intn(len(items))]
	return choice, nil
}

func filterUrlize(ctx *Context, value interface{}, args ...interface{}) (interface{}, error) {
	text := toString(value)
	if text == "" {
		if ctx != nil && ctx.ShouldAutoescape() {
			return Markup(""), nil
		}
		return "", nil
	}

	trimLimit := -1
	if len(args) > 0 {
		if limit, ok := toInt(args[0]); ok {
			trimLimit = limit
		}
	}

	nofollow := false
	if len(args) > 1 {
		switch v := args[1].(type) {
		case bool:
			nofollow = v
		case string:
			nofollow = strings.Contains(strings.ToLower(v), "nofollow")
		}
	}

	var target string
	var relArg string
	if len(args) > 2 {
		target = toString(args[2])
	}
	if len(args) > 3 {
		relArg = toString(args[3])
	}

	var extraSchemes []string
	if len(args) > 4 {
		var err error
		extraSchemes, err = normalizeExtraSchemes(args[4])
		if err != nil {
			return nil, err
		}
	}

	var policyRel interface{}
	var policyTarget interface{}
	var policySchemes interface{}
	if ctx != nil && ctx.environment != nil {
		ctx.environment.mu.RLock()
		policyRel = ctx.environment.policies["urlize.rel"]
		policyTarget = ctx.environment.policies["urlize.target"]
		policySchemes = ctx.environment.policies["urlize.extra_schemes"]
		ctx.environment.mu.RUnlock()
	}

	if extraSchemes == nil {
		var err error
		extraSchemes, err = normalizeExtraSchemes(policySchemes)
		if err != nil {
			return nil, err
		}
	}

	if target == "" {
		if s, ok := policyTarget.(string); ok && s != "" {
			target = s
		}
	}

	relAttr := buildRelAttribute(nofollow, relArg, policyRel)
	targetAttr := buildAttribute("target", target)

	escaped := html.EscapeString(text)
	tokens := splitPreserveWhitespace(escaped)

	for i, token := range tokens {
		if isWhitespace(token) {
			continue
		}
		head, middle, tail := splitURLToken(token)
		if middle == "" {
			tokens[i] = token
			continue
		}

		middle = balanceTokenDelimiters(middle, &tail)

		transformed := transformURLToken(middle, trimLimit, relAttr, targetAttr, extraSchemes)
		if transformed == "" {
			tokens[i] = token
			continue
		}
		tokens[i] = head + transformed + tail
	}

	result := strings.Join(tokens, "")
	if ctx != nil && ctx.ShouldAutoescape() {
		return Markup(result), nil
	}
	return result, nil
}

func filterXMLAttr(ctx *Context, value interface{}, args ...interface{}) (interface{}, error) {
	attrs, ok := toStringInterfaceMap(value)
	if !ok {
		return nil, fmt.Errorf("xmlattr filter requires a mapping")
	}

	autospace := true
	if len(args) > 0 {
		switch v := args[0].(type) {
		case bool:
			autospace = v
		case int:
			autospace = v != 0
		case int64:
			autospace = v != 0
		case string:
			lowered := strings.ToLower(strings.TrimSpace(v))
			autospace = !(lowered == "false" || lowered == "0")
		default:
			autospace = isTruthyValue(v)
		}
	}

	keys := make([]string, 0, len(attrs))
	for k := range attrs {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	rendered := make([]string, 0, len(keys))
	for _, key := range keys {
		if xmlAttrInvalidKeyPattern.MatchString(key) {
			return nil, fmt.Errorf("invalid character in attribute name: %q", key)
		}

		val := attrs[key]
		if val == nil || isUndefinedValue(val) {
			continue
		}

		attrValue, safe := xmlAttrValue(val)
		if attrValue == "" {
			continue
		}

		name := html.EscapeString(key)
		if !safe {
			attrValue = html.EscapeString(attrValue)
		}
		rendered = append(rendered, fmt.Sprintf(`%s="%s"`, name, attrValue))
	}

	if len(rendered) == 0 {
		if ctx != nil && ctx.ShouldAutoescape() {
			return Markup(""), nil
		}
		return "", nil
	}

	result := strings.Join(rendered, " ")
	if autospace {
		result = " " + result
	}

	if ctx != nil && ctx.ShouldAutoescape() {
		return Markup(result), nil
	}
	return result, nil
}

func filterAttr(ctx *Context, value interface{}, args ...interface{}) (interface{}, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("attr filter requires 1 argument (attribute name)")
	}

	attrName := toString(args[0])
	return getAttribute(value, attrName)
}

func filterMap(ctx *Context, value interface{}, args ...interface{}) (interface{}, error) {
	kwargs, args := extractKwargs(args)
	attrName := ""
	if len(args) > 0 {
		attrName = toString(args[0])
	}
	if attrName == "" && kwargs != nil {
		if attr, ok := kwargs["attribute"]; ok {
			attrName = toString(attr)
		}
	}
	if attrName == "" {
		return nil, fmt.Errorf("map filter requires attribute name")
	}
	items, err := sequenceToSlice(value)
	if err != nil {
		return nil, fmt.Errorf("map filter requires a sequence")
	}
	result := make([]interface{}, len(items))
	for i, item := range items {
		attr, _ := getAttribute(item, attrName)
		result[i] = attr
	}
	return result, nil
}

func filterSelect(ctx *Context, value interface{}, args ...interface{}) (interface{}, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("select filter requires 1 argument (test)")
	}

	testName := toString(args[0])
	testFunc, ok := ctx.environment.GetTest(testName)
	if !ok {
		return nil, fmt.Errorf("unknown test: %s", testName)
	}

	items, err := sequenceToSlice(value)
	if err != nil {
		return nil, fmt.Errorf("select filter requires a sequence")
	}
	result := make([]interface{}, 0, len(items))
	testArgs := args[1:]
	for _, item := range items {
		if passed, err := testFunc(ctx, item, testArgs...); err == nil && passed {
			result = append(result, item)
		}
	}
	return result, nil
}

func filterReject(ctx *Context, value interface{}, args ...interface{}) (interface{}, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("reject filter requires 1 argument (test)")
	}

	testName := toString(args[0])
	testFunc, ok := ctx.environment.GetTest(testName)
	if !ok {
		return nil, fmt.Errorf("unknown test: %s", testName)
	}

	items, err := sequenceToSlice(value)
	if err != nil {
		return nil, fmt.Errorf("reject filter requires a sequence")
	}
	result := make([]interface{}, 0, len(items))
	testArgs := args[1:]
	for _, item := range items {
		if passed, err := testFunc(ctx, item, testArgs...); err != nil || !passed {
			result = append(result, item)
		}
	}
	return result, nil
}

func filterSelectattr(ctx *Context, value interface{}, args ...interface{}) (interface{}, error) {
	kwargs, args := extractKwargs(args)
	attrName := ""
	if len(args) > 0 {
		attrName = toString(args[0])
	}
	if attrName == "" && kwargs != nil {
		if attr, ok := kwargs["attribute"]; ok {
			attrName = toString(attr)
		}
	}
	if attrName == "" {
		return nil, fmt.Errorf("selectattr filter requires attribute name")
	}

	testName := ""
	testArgs := []interface{}{}
	if len(args) > 1 {
		testName = toString(args[1])
		testArgs = args[2:]
	}
	if kwargs != nil {
		if tn, ok := kwargs["test"]; ok {
			testName = toString(tn)
		}
		if val, ok := kwargs["value"]; ok {
			testArgs = append([]interface{}{val}, testArgs...)
		}
	}

	items, err := sequenceToSlice(value)
	if err != nil {
		return nil, fmt.Errorf("selectattr filter requires a sequence")
	}

	result := make([]interface{}, 0, len(items))
	for _, item := range items {
		attr, _ := getAttribute(item, attrName)

		if testName == "" {
			if attr != nil && attr != false && attr != 0 && attr != "" {
				result = append(result, item)
			}
		} else {
			testFunc, ok := ctx.environment.GetTest(testName)
			if !ok {
				continue
			}
			if passed, err := testFunc(ctx, attr, testArgs...); err == nil && passed {
				result = append(result, item)
			}
		}
	}
	return result, nil
}

func filterRejectattr(ctx *Context, value interface{}, args ...interface{}) (interface{}, error) {
	kwargs, args := extractKwargs(args)
	attrName := ""
	if len(args) > 0 {
		attrName = toString(args[0])
	}
	if attrName == "" && kwargs != nil {
		if attr, ok := kwargs["attribute"]; ok {
			attrName = toString(attr)
		}
	}
	if attrName == "" {
		return nil, fmt.Errorf("rejectattr filter requires attribute name")
	}
	testName := ""
	testArgs := []interface{}{}
	if len(args) > 1 {
		testName = toString(args[1])
		testArgs = args[2:]
	}
	if kwargs != nil {
		if tn, ok := kwargs["test"]; ok {
			testName = toString(tn)
		}
		if val, ok := kwargs["value"]; ok {
			testArgs = append([]interface{}{val}, testArgs...)
		}
	}

	items, err := sequenceToSlice(value)
	if err != nil {
		return nil, fmt.Errorf("rejectattr filter requires a sequence")
	}

	result := make([]interface{}, 0, len(items))
	for _, item := range items {
		attr, _ := getAttribute(item, attrName)

		if testName == "" {
			if attr == nil || attr == false || attr == 0 || attr == "" {
				result = append(result, item)
			}
		} else {
			testFunc, ok := ctx.environment.GetTest(testName)
			if !ok {
				result = append(result, item)
				continue
			}
			if passed, err := testFunc(ctx, attr, testArgs...); err != nil || !passed {
				result = append(result, item)
			}
		}
	}
	return result, nil
}

// Test functions

func testDivisibleby(ctx *Context, value interface{}, args ...interface{}) (bool, error) {
	if len(args) < 1 {
		return false, fmt.Errorf("divisibleby test requires 1 argument")
	}

	divisor, ok := toFloat64(args[0])
	if !ok || divisor == 0 {
		return false, nil
	}

	num, ok := toFloat64(value)
	if !ok {
		return false, nil
	}

	return math.Mod(num, divisor) == 0, nil
}

func testDefined(ctx *Context, value interface{}, args ...interface{}) (bool, error) {
	if isUndefinedValue(value) {
		return false, nil
	}
	return value != nil, nil
}

func testUndefined(ctx *Context, value interface{}, args ...interface{}) (bool, error) {
	return isUndefinedValue(value), nil
}

func testNone(ctx *Context, value interface{}, args ...interface{}) (bool, error) {
	return value == nil, nil
}

func testBoolean(ctx *Context, value interface{}, args ...interface{}) (bool, error) {
	_, ok := value.(bool)
	return ok, nil
}

func testTrue(ctx *Context, value interface{}, args ...interface{}) (bool, error) {
	if b, ok := value.(bool); ok {
		return b, nil
	}
	return false, nil
}

func testFalse(ctx *Context, value interface{}, args ...interface{}) (bool, error) {
	if b, ok := value.(bool); ok {
		return !b, nil
	}
	return false, nil
}

func testNumber(ctx *Context, value interface{}, args ...interface{}) (bool, error) {
	switch value.(type) {
	case int, int64, float64, float32:
		return true, nil
	default:
		return false, nil
	}
}

func testString(ctx *Context, value interface{}, args ...interface{}) (bool, error) {
	_, ok := value.(string)
	return ok, nil
}

func testInteger(ctx *Context, value interface{}, args ...interface{}) (bool, error) {
	if value == nil {
		return false, nil
	}
	if _, ok := value.(bool); ok {
		return false, nil
	}

	val := reflect.ValueOf(value)
	switch val.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return true, nil
	}
	return false, nil
}

func testFloat(ctx *Context, value interface{}, args ...interface{}) (bool, error) {
	switch value.(type) {
	case float32, float64:
		return true, nil
	default:
		return false, nil
	}
}

func testSequence(ctx *Context, value interface{}, args ...interface{}) (bool, error) {
	switch value.(type) {
	case []interface{}, []string, string:
		return true, nil
	default:
		val := reflect.ValueOf(value)
		return val.Kind() == reflect.Slice || val.Kind() == reflect.Array, nil
	}
}

func testMapping(ctx *Context, value interface{}, args ...interface{}) (bool, error) {
	switch value.(type) {
	case map[interface{}]interface{}, map[string]interface{}:
		return true, nil
	default:
		val := reflect.ValueOf(value)
		return val.Kind() == reflect.Map, nil
	}
}

func testIterable(ctx *Context, value interface{}, args ...interface{}) (bool, error) {
	return testSequence(ctx, value, args...)
}

func testCallable(ctx *Context, value interface{}, args ...interface{}) (bool, error) {
	kwargs, _ := extractKwargs(args)
	if kwargs != nil {
		if attr, ok := kwargs["attribute"]; ok {
			attrName := toString(attr)
			if attrName == "" || ctx == nil {
				return false, nil
			}
			resolved, err := ctx.ResolveAttribute(value, attrName)
			if err != nil {
				return false, nil
			}
			return isCallableValue(resolved), nil
		}
	}
	return isCallableValue(value), nil
}

func isCallableValue(value interface{}) bool {
	if value == nil {
		return false
	}

	switch value.(type) {
	case func(*Context, ...interface{}) (interface{}, error),
		func(...interface{}) (interface{}, error),
		func(...interface{}) interface{},
		func(*Context) interface{},
		func() interface{}:
		return true
	case *Macro:
		return true
	}

	val := reflect.ValueOf(value)
	if !val.IsValid() {
		return false
	}
	if val.Kind() != reflect.Func {
		return false
	}
	if val.IsNil() {
		return false
	}
	return true
}

func testSameas(ctx *Context, value interface{}, args ...interface{}) (bool, error) {
	if len(args) < 1 {
		return false, fmt.Errorf("sameas test requires 1 argument")
	}
	return value == args[0], nil
}

func testEscaped(ctx *Context, value interface{}, args ...interface{}) (bool, error) {
	if value == nil {
		return false, nil
	}
	if _, ok := value.(Markup); ok {
		return true, nil
	}

	type htmlRenderer interface {
		HTML() string
	}
	if _, ok := value.(htmlRenderer); ok {
		return true, nil
	}

	return false, nil
}

func testModule(ctx *Context, value interface{}, args ...interface{}) (bool, error) {
	if value == nil {
		return false, nil
	}
	_, ok := value.(*MacroNamespace)
	return ok, nil
}

func testList(ctx *Context, value interface{}, args ...interface{}) (bool, error) {
	if value == nil {
		return false, nil
	}
	val := reflect.ValueOf(value)
	switch val.Kind() {
	case reflect.Slice:
		return true, nil
	case reflect.Array:
		return true, nil
	default:
		return false, nil
	}
}

func testTuple(ctx *Context, value interface{}, args ...interface{}) (bool, error) {
	if value == nil {
		return false, nil
	}
	val := reflect.ValueOf(value)
	return val.Kind() == reflect.Array, nil
}

func testDict(ctx *Context, value interface{}, args ...interface{}) (bool, error) {
	if value == nil {
		return false, nil
	}
	val := reflect.ValueOf(value)
	return val.Kind() == reflect.Map, nil
}

func testLowerTest(ctx *Context, value interface{}, args ...interface{}) (bool, error) {
	str := toString(value)
	return strings.ToLower(str) == str, nil
}

func testUpperTest(ctx *Context, value interface{}, args ...interface{}) (bool, error) {
	str := toString(value)
	return strings.ToUpper(str) == str, nil
}

func testEven(ctx *Context, value interface{}, args ...interface{}) (bool, error) {
	if num, ok := toInt(value); ok {
		return num%2 == 0, nil
	}
	return false, nil
}

func testOdd(ctx *Context, value interface{}, args ...interface{}) (bool, error) {
	if num, ok := toInt(value); ok {
		return num%2 != 0, nil
	}
	return false, nil
}

func testInTest(ctx *Context, value interface{}, args ...interface{}) (bool, error) {
	if len(args) < 1 {
		return false, fmt.Errorf("in test requires 1 argument")
	}

	switch container := args[0].(type) {
	case []interface{}:
		for _, item := range container {
			if item == value {
				return true, nil
			}
		}
	case []string:
		if str, ok := value.(string); ok {
			for _, item := range container {
				if item == str {
					return true, nil
				}
			}
		}
	case map[interface{}]interface{}:
		_, exists := container[value]
		return exists, nil
	case map[string]interface{}:
		if str, ok := value.(string); ok {
			_, exists := container[str]
			return exists, nil
		}
	case string:
		if str, ok := value.(string); ok {
			return strings.Contains(container, str), nil
		}
	}

	return false, nil
}

func testFilter(ctx *Context, value interface{}, args ...interface{}) (bool, error) {
	if ctx == nil || ctx.environment == nil {
		return false, nil
	}
	name := strings.TrimSpace(toString(value))
	if name == "" {
		return false, nil
	}
	_, ok := ctx.environment.GetFilter(name)
	return ok, nil
}

func testTest(ctx *Context, value interface{}, args ...interface{}) (bool, error) {
	if ctx == nil || ctx.environment == nil {
		return false, nil
	}
	name := strings.TrimSpace(toString(value))
	if name == "" {
		return false, nil
	}
	_, ok := ctx.environment.GetTest(name)
	return ok, nil
}

func testEq(ctx *Context, value interface{}, args ...interface{}) (bool, error) {
	if len(args) < 1 {
		return false, fmt.Errorf("eq test requires a comparison value")
	}
	if lf, lok := toFloat64(value); lok {
		if rf, rok := toFloat64(args[0]); rok {
			return lf == rf, nil
		}
	}
	return reflect.DeepEqual(value, args[0]), nil
}

func testNe(ctx *Context, value interface{}, args ...interface{}) (bool, error) {
	res, err := testEq(ctx, value, args...)
	if err != nil {
		return false, err
	}
	return !res, nil
}

func testLt(ctx *Context, value interface{}, args ...interface{}) (bool, error) {
	if len(args) < 1 {
		return false, fmt.Errorf("lt test requires a comparison value")
	}
	return compareNumeric(value, args[0], func(a, b float64) bool { return a < b })
}

func testLe(ctx *Context, value interface{}, args ...interface{}) (bool, error) {
	if len(args) < 1 {
		return false, fmt.Errorf("le test requires a comparison value")
	}
	return compareNumeric(value, args[0], func(a, b float64) bool { return a <= b })
}

func testGt(ctx *Context, value interface{}, args ...interface{}) (bool, error) {
	if len(args) < 1 {
		return false, fmt.Errorf("gt test requires a comparison value")
	}
	return compareNumeric(value, args[0], func(a, b float64) bool { return a > b })
}

func testGe(ctx *Context, value interface{}, args ...interface{}) (bool, error) {
	if len(args) < 1 {
		return false, fmt.Errorf("ge test requires a comparison value")
	}
	return compareNumeric(value, args[0], func(a, b float64) bool { return a >= b })
}

func testMatching(ctx *Context, value interface{}, args ...interface{}) (bool, error) {
	if len(args) < 1 {
		return false, fmt.Errorf("matching test requires a pattern argument")
	}
	pattern := toString(args[0])
	re, err := regexp.Compile(pattern)
	if err != nil {
		return false, err
	}
	match := re.FindStringIndex(toString(value))
	return match != nil && match[0] == 0, nil
}

func testSearch(ctx *Context, value interface{}, args ...interface{}) (bool, error) {
	if len(args) < 1 {
		return false, fmt.Errorf("search test requires a pattern argument")
	}
	pattern := toString(args[0])
	re, err := regexp.Compile(pattern)
	if err != nil {
		return false, err
	}
	return re.FindStringIndex(toString(value)) != nil, nil
}

func testStartingWith(ctx *Context, value interface{}, args ...interface{}) (bool, error) {
	if len(args) < 1 {
		return false, fmt.Errorf("startingwith test requires at least one prefix")
	}
	val := toString(value)
	for _, arg := range args {
		if strings.HasPrefix(val, toString(arg)) {
			return true, nil
		}
	}
	return false, nil
}

func testEndingWith(ctx *Context, value interface{}, args ...interface{}) (bool, error) {
	if len(args) < 1 {
		return false, fmt.Errorf("endingwith test requires at least one suffix")
	}
	val := toString(value)
	for _, arg := range args {
		if strings.HasSuffix(val, toString(arg)) {
			return true, nil
		}
	}
	return false, nil
}

func testContaining(ctx *Context, value interface{}, args ...interface{}) (bool, error) {
	if len(args) < 1 {
		return false, fmt.Errorf("containing test requires a substring argument")
	}
	return strings.Contains(toString(value), toString(args[0])), nil
}

func testInfinite(ctx *Context, value interface{}, args ...interface{}) (bool, error) {
	if f, ok := toFloat64(value); ok {
		return math.IsInf(f, 0), nil
	}
	return false, nil
}

func testNan(ctx *Context, value interface{}, args ...interface{}) (bool, error) {
	if f, ok := toFloat64(value); ok {
		return math.IsNaN(f), nil
	}
	return false, nil
}

func testFinite(ctx *Context, value interface{}, args ...interface{}) (bool, error) {
	if f, ok := toFloat64(value); ok {
		return !math.IsNaN(f) && !math.IsInf(f, 0), nil
	}
	return false, nil
}

func compareNumeric(left, right interface{}, cmp func(a, b float64) bool) (bool, error) {
	lf, lok := toFloat64(left)
	rf, rok := toFloat64(right)
	if !lok || !rok {
		return false, fmt.Errorf("numeric comparison requires numeric values")
	}
	return cmp(lf, rf), nil
}

type dictsortDefaults struct {
	caseSensitive bool
	reverse       bool
}

type dictsortPair struct {
	key   interface{}
	value interface{}
}

func dictsortWithDefaults(value interface{}, args []interface{}, defaults dictsortDefaults) (interface{}, error) {
	kwargs, positional := extractKwargs(args)
	args = positional

	caseSensitive := defaults.caseSensitive
	by := "key"
	reverse := defaults.reverse

	if len(args) > 0 {
		caseSensitive = isTruthyValue(args[0])
	}
	if len(args) > 1 {
		by = strings.ToLower(toString(args[1]))
	}
	if len(args) > 2 {
		reverse = isTruthyValue(args[2])
	}

	if kwargs != nil {
		if val, ok := kwargs["case_sensitive"]; ok {
			caseSensitive = isTruthyValue(val)
		}
		if val, ok := kwargs["by"]; ok {
			by = strings.ToLower(toString(val))
		}
		if val, ok := kwargs["reverse"]; ok {
			reverse = isTruthyValue(val)
		}
	}

	if by == "" {
		by = "key"
	}

	switch by {
	case "key", "value", "item":
	default:
		return nil, fmt.Errorf("dictsort filter received unknown 'by' value %q", by)
	}

	pairs, err := collectDictsortPairs(value)
	if err != nil {
		return nil, err
	}

	sort.SliceStable(pairs, func(i, j int) bool {
		var cmp int
		switch by {
		case "key":
			cmp = compareValues(pairs[i].key, pairs[j].key, caseSensitive)
		case "value":
			cmp = compareValues(pairs[i].value, pairs[j].value, caseSensitive)
		case "item":
			cmp = compareValues(pairs[i].key, pairs[j].key, caseSensitive)
			if cmp == 0 {
				cmp = compareValues(pairs[i].value, pairs[j].value, caseSensitive)
			}
		}

		if reverse {
			return cmp > 0
		}
		return cmp < 0
	})

	result := make([]interface{}, len(pairs))
	for i, pair := range pairs {
		result[i] = []interface{}{pair.key, pair.value}
	}
	return result, nil
}

func collectDictsortPairs(value interface{}) ([]dictsortPair, error) {
	if value == nil || isUndefinedValue(value) {
		return []dictsortPair{}, nil
	}

	switch v := value.(type) {
	case map[string]interface{}:
		items := make([]dictsortPair, 0, len(v))
		for key, val := range v {
			items = append(items, dictsortPair{key: key, value: val})
		}
		return items, nil
	case map[interface{}]interface{}:
		items := make([]dictsortPair, 0, len(v))
		for key, val := range v {
			items = append(items, dictsortPair{key: key, value: val})
		}
		return items, nil
	case []interface{}:
		items := make([]dictsortPair, 0, len(v))
		for _, entry := range v {
			key, val, err := dictsortPairFromValue(entry)
			if err != nil {
				return nil, err
			}
			items = append(items, dictsortPair{key: key, value: val})
		}
		return items, nil
	}

	rv := reflect.ValueOf(value)
	if !rv.IsValid() {
		return nil, fmt.Errorf("dictsort filter requires a mapping or sequence of pairs")
	}

	switch rv.Kind() {
	case reflect.Map:
		items := make([]dictsortPair, 0, rv.Len())
		for _, key := range rv.MapKeys() {
			items = append(items, dictsortPair{key: key.Interface(), value: rv.MapIndex(key).Interface()})
		}
		return items, nil
	case reflect.Slice, reflect.Array:
		items := make([]dictsortPair, 0, rv.Len())
		for i := 0; i < rv.Len(); i++ {
			key, val, err := dictsortPairFromValue(rv.Index(i).Interface())
			if err != nil {
				return nil, err
			}
			items = append(items, dictsortPair{key: key, value: val})
		}
		return items, nil
	default:
		return nil, fmt.Errorf("dictsort filter requires a mapping or sequence of pairs")
	}
}

func dictsortPairFromValue(value interface{}) (interface{}, interface{}, error) {
	switch pair := value.(type) {
	case []interface{}:
		if len(pair) < 2 {
			return nil, nil, fmt.Errorf("dictsort filter requires pairs of (key, value)")
		}
		return pair[0], pair[1], nil
	case []string:
		if len(pair) < 2 {
			return nil, nil, fmt.Errorf("dictsort filter requires pairs of (key, value)")
		}
		return pair[0], pair[1], nil
	case [2]interface{}:
		return pair[0], pair[1], nil
	case map[string]interface{}:
		key, okKey := pair["key"]
		val, okVal := pair["value"]
		if okKey && okVal {
			return key, val, nil
		}
	}

	rv := reflect.ValueOf(value)
	if rv.IsValid() && (rv.Kind() == reflect.Slice || rv.Kind() == reflect.Array) {
		if rv.Len() < 2 {
			return nil, nil, fmt.Errorf("dictsort filter requires pairs of (key, value)")
		}
		return rv.Index(0).Interface(), rv.Index(1).Interface(), nil
	}

	return nil, nil, fmt.Errorf("dictsort filter requires a mapping or sequence of pairs")
}

// Utility functions shared across the runtime package

func sequenceToSlice(value interface{}) ([]interface{}, error) {
	switch v := value.(type) {
	case []interface{}:
		return append([]interface{}(nil), v...), nil
	case []string:
		result := make([]interface{}, len(v))
		for i, item := range v {
			result[i] = item
		}
		return result, nil
	case string:
		runes := []rune(v)
		result := make([]interface{}, len(runes))
		for i, r := range runes {
			result[i] = string(r)
		}
		return result, nil
	case map[interface{}]interface{}:
		result := make([]interface{}, 0, len(v))
		for key := range v {
			result = append(result, key)
		}
		return result, nil
	case map[string]interface{}:
		result := make([]interface{}, 0, len(v))
		for key := range v {
			result = append(result, key)
		}
		return result, nil
	default:
		val := reflect.ValueOf(value)
		switch val.Kind() {
		case reflect.Slice, reflect.Array:
			result := make([]interface{}, val.Len())
			for i := 0; i < val.Len(); i++ {
				result[i] = val.Index(i).Interface()
			}
			return result, nil
		}
	}
	return nil, fmt.Errorf("sequence expected")
}

func repeatValue(value interface{}, count int) []interface{} {
	result := make([]interface{}, count)
	for i := 0; i < count; i++ {
		result[i] = value
	}
	return result
}

func extractKwargs(args []interface{}) (map[string]interface{}, []interface{}) {
	if len(args) == 0 {
		return nil, args
	}
	switch kw := args[len(args)-1].(type) {
	case map[string]interface{}:
		return kw, args[:len(args)-1]
	case map[interface{}]interface{}:
		converted := make(map[string]interface{}, len(kw))
		for k, v := range kw {
			if key, ok := k.(string); ok {
				converted[key] = v
			}
		}
		return converted, args[:len(args)-1]
	default:
		return nil, args
	}
}

func toStringInterfaceMap(value interface{}) (map[string]interface{}, bool) {
	switch m := value.(type) {
	case map[string]interface{}:
		return m, true
	case map[interface{}]interface{}:
		result := make(map[string]interface{}, len(m))
		for k, v := range m {
			if key, ok := k.(string); ok {
				result[key] = v
			}
		}
		return result, true
	}

	val := reflect.ValueOf(value)
	if !val.IsValid() || val.Kind() != reflect.Map {
		return nil, false
	}
	if val.Type().Key().Kind() != reflect.String {
		return nil, false
	}

	result := make(map[string]interface{}, val.Len())
	for _, key := range val.MapKeys() {
		result[key.String()] = val.MapIndex(key).Interface()
	}
	return result, true
}

func splitPreserveWhitespace(s string) []string {
	if s == "" {
		return []string{""}
	}

	var tokens []string
	var current strings.Builder
	currentType := -1

	for _, r := range s {
		if unicode.IsSpace(r) {
			if currentType == 0 {
				tokens = append(tokens, current.String())
				current.Reset()
			}
			currentType = 1
			current.WriteRune(r)
		} else {
			if currentType == 1 {
				tokens = append(tokens, current.String())
				current.Reset()
			}
			currentType = 0
			current.WriteRune(r)
		}
	}

	if current.Len() > 0 {
		tokens = append(tokens, current.String())
	}

	return tokens
}

func wrapParagraph(line string, width int, breakLongWords, breakOnHyphens bool) []string {
	if strings.TrimSpace(line) == "" {
		return []string{""}
	}

	chunks := splitChunks(line, breakOnHyphens)
	if len(chunks) == 0 {
		return []string{""}
	}

	// normalise whitespace chunks to single spaces and remove leading/trailing spaces
	normalised := make([]string, 0, len(chunks))
	for _, chunk := range chunks {
		if chunk == "" {
			continue
		}
		if strings.TrimSpace(chunk) == "" {
			normalised = append(normalised, " ")
		} else {
			normalised = append(normalised, chunk)
		}
	}

	for len(normalised) > 0 && normalised[0] == " " {
		normalised = normalised[1:]
	}
	for len(normalised) > 0 && normalised[len(normalised)-1] == " " {
		normalised = normalised[:len(normalised)-1]
	}

	if len(normalised) == 0 {
		return []string{""}
	}

	reverseStrings(normalised)

	var lines []string

	for len(normalised) > 0 {
		// drop leading spaces when continuing paragraphs
		if len(lines) > 0 {
			for len(normalised) > 0 && normalised[len(normalised)-1] == " " {
				normalised = normalised[:len(normalised)-1]
			}
			if len(normalised) == 0 {
				break
			}
		}

		curLine := make([]string, 0)
		curLen := 0

		for len(normalised) > 0 {
			chunk := normalised[len(normalised)-1]
			chunkLen := utf8.RuneCountInString(chunk)

			if curLen+chunkLen <= width {
				curLine = append(curLine, chunk)
				curLen += chunkLen
				normalised = normalised[:len(normalised)-1]
			} else {
				break
			}
		}

		if len(normalised) > 0 {
			nextChunk := normalised[len(normalised)-1]
			if utf8.RuneCountInString(nextChunk) > width {
				handleLongWord(&normalised, &curLine, &curLen, width, breakLongWords, breakOnHyphens)
			}
		}

		for len(curLine) > 0 && curLine[len(curLine)-1] == " " {
			curLen -= 1
			curLine = curLine[:len(curLine)-1]
		}

		if len(curLine) > 0 {
			lines = append(lines, strings.Join(curLine, ""))
		} else {
			break
		}
	}

	if len(lines) == 0 {
		return []string{""}
	}

	return lines
}

func splitChunks(text string, breakOnHyphens bool) []string {
	tokens := splitPreserveWhitespace(text)
	chunks := make([]string, 0, len(tokens))

	for _, token := range tokens {
		if token == "" {
			continue
		}
		if isWhitespace(token) {
			chunks = append(chunks, token)
			continue
		}

		if !breakOnHyphens {
			chunks = append(chunks, token)
			continue
		}

		runes := []rune(token)
		start := 0
		for i, r := range runes {
			if r == '-' {
				chunks = append(chunks, string(runes[start:i+1]))
				start = i + 1
			}
		}
		if start < len(runes) {
			chunks = append(chunks, string(runes[start:]))
		}
	}

	return chunks
}

func handleLongWord(chunks *[]string, curLine *[]string, curLen *int, width int, breakLongWords, breakOnHyphens bool) {
	if width < 1 {
		width = 1
	}

	spaceLeft := width - *curLen
	if spaceLeft < 1 {
		spaceLeft = 1
	}

	stack := *chunks
	if len(stack) == 0 {
		return
	}

	chunk := stack[len(stack)-1]
	chunkRunes := []rune(chunk)

	if breakLongWords {
		end := spaceLeft
		if end > len(chunkRunes) {
			end = len(chunkRunes)
		}

		if breakOnHyphens && len(chunkRunes) > spaceLeft {
			hyphen := -1
			for i := 0; i < spaceLeft && i < len(chunkRunes); i++ {
				if chunkRunes[i] == '-' {
					hyphen = i
				}
			}
			if hyphen > 0 {
				for j := 0; j < hyphen; j++ {
					if chunkRunes[j] != '-' {
						end = hyphen + 1
						break
					}
				}
			}
		}

		splitPart := string(chunkRunes[:end])
		remainder := string(chunkRunes[end:])

		*curLine = append(*curLine, splitPart)
		*curLen += utf8.RuneCountInString(splitPart)

		if remainder == "" {
			*chunks = stack[:len(stack)-1]
		} else {
			stack[len(stack)-1] = remainder
		}
		return
	}

	if len(*curLine) == 0 {
		*curLine = append(*curLine, chunk)
		*curLen += utf8.RuneCountInString(chunk)
		*chunks = stack[:len(stack)-1]
	}
}

func reverseStrings(values []string) {
	for i, j := 0, len(values)-1; i < j; i, j = i+1, j-1 {
		values[i], values[j] = values[j], values[i]
	}
}

func isWhitespace(token string) bool {
	for _, r := range token {
		if !unicode.IsSpace(r) {
			return false
		}
	}
	return true
}

func splitURLToken(token string) (string, string, string) {
	head := ""
	middle := token
	tail := ""

	for {
		switch {
		case strings.HasPrefix(middle, "("):
			head += "("
			middle = middle[1:]
		case strings.HasPrefix(middle, "<"):
			head += "<"
			middle = middle[1:]
		case strings.HasPrefix(middle, "&lt;"):
			head += "&lt;"
			middle = middle[len("&lt;"):]
		default:
			goto trimTail
		}
		if middle == "" {
			return head, "", tail
		}
	}

trimTail:
	for {
		switch {
		case strings.HasSuffix(middle, "&gt;"):
			tail = "&gt;" + tail
			middle = middle[:len(middle)-len("&gt;")]
		case strings.HasSuffix(middle, ")"):
			tail = ")" + tail
			middle = middle[:len(middle)-1]
		case strings.HasSuffix(middle, ">"):
			tail = ">" + tail
			middle = middle[:len(middle)-1]
		case strings.HasSuffix(middle, "."):
			tail = "." + tail
			middle = middle[:len(middle)-1]
		case strings.HasSuffix(middle, ","):
			tail = "," + tail
			middle = middle[:len(middle)-1]
		case strings.HasSuffix(middle, "\n"):
			tail = "\n" + tail
			middle = middle[:len(middle)-1]
		default:
			return head, middle, tail
		}
		if middle == "" {
			return head, "", tail
		}
	}
}

func balanceTokenDelimiters(middle string, tail *string) string {
	pairs := []struct{ start, end string }{
		{"(", ")"},
		{"<", ">"},
		{"&lt;", "&gt;"},
	}

	for _, pair := range pairs {
		startCount := strings.Count(middle, pair.start)
		endCount := strings.Count(middle, pair.end)
		if startCount <= endCount {
			continue
		}
		diff := startCount - endCount
		for diff > 0 {
			idx := strings.Index(*tail, pair.end)
			if idx == -1 {
				break
			}
			idx += len(pair.end)
			middle += (*tail)[:idx]
			*tail = (*tail)[idx:]
			diff--
		}
	}

	return middle
}

func normalizeExtraSchemes(value interface{}) ([]string, error) {
	if value == nil {
		return nil, nil
	}

	switch v := value.(type) {
	case []string:
		return validateSchemes(v)
	case []interface{}:
		parts := make([]string, 0, len(v))
		for _, item := range v {
			if item == nil {
				continue
			}
			str := toString(item)
			if str != "" {
				parts = append(parts, str)
			}
		}
		return validateSchemes(parts)
	case string:
		if v == "" {
			return nil, nil
		}
		split := strings.FieldsFunc(v, func(r rune) bool {
			return r == ',' || unicode.IsSpace(r)
		})
		return validateSchemes(split)
	default:
		return nil, fmt.Errorf("extra_schemes must be a sequence of strings")
	}
}

func validateSchemes(schemes []string) ([]string, error) {
	if len(schemes) == 0 {
		return nil, nil
	}
	set := make(map[string]struct{}, len(schemes))
	result := make([]string, 0, len(schemes))
	for _, scheme := range schemes {
		scheme = strings.TrimSpace(scheme)
		if scheme == "" {
			continue
		}
		if !uriSchemePattern.MatchString(scheme) {
			return nil, fmt.Errorf("%q is not a valid URI scheme prefix", scheme)
		}
		if _, exists := set[scheme]; exists {
			continue
		}
		set[scheme] = struct{}{}
		result = append(result, scheme)
	}
	if len(result) == 0 {
		return nil, nil
	}
	sort.Strings(result)
	return result, nil
}

func buildRelAttribute(nofollow bool, relArg string, policy interface{}) string {
	relSet := make(map[string]struct{})

	if str, ok := policy.(string); ok && str != "" {
		for _, part := range strings.Fields(str) {
			if part != "" {
				relSet[part] = struct{}{}
			}
		}
	}

	if relArg != "" {
		for _, part := range strings.Fields(relArg) {
			if part != "" {
				relSet[part] = struct{}{}
			}
		}
	}

	if nofollow {
		relSet["nofollow"] = struct{}{}
	}

	if len(relSet) == 0 {
		return ""
	}

	parts := make([]string, 0, len(relSet))
	for part := range relSet {
		parts = append(parts, part)
	}
	sort.Strings(parts)

	for i, part := range parts {
		parts[i] = html.EscapeString(part)
	}

	return fmt.Sprintf(" rel=\"%s\"", strings.Join(parts, " "))
}

func buildAttribute(name, value string) string {
	if strings.TrimSpace(value) == "" {
		return ""
	}
	return fmt.Sprintf(" %s=\"%s\"", name, html.EscapeString(value))
}

func trimURLDisplay(text string, limit int) string {
	if limit < 0 {
		return text
	}
	runes := []rune(text)
	if len(runes) > limit {
		return string(runes[:limit]) + "..."
	}
	return text
}

func transformURLToken(middle string, trimLimit int, relAttr, targetAttr string, extraSchemes []string) string {
	lower := strings.ToLower(middle)

	switch {
	case urlizeURLPattern.MatchString(middle):
		href := middle
		if strings.HasPrefix(lower, "www.") {
			href = "https://" + middle
		}
		display := trimURLDisplay(middle, trimLimit)
		return fmt.Sprintf("<a href=\"%s\"%s%s>%s</a>", href, relAttr, targetAttr, display)
	case urlizeBareDomainPattern.MatchString(middle):
		display := trimURLDisplay(middle, trimLimit)
		return fmt.Sprintf("<a href=\"https://%s\"%s%s>%s</a>", middle, relAttr, targetAttr, display)
	case strings.HasPrefix(lower, "mailto:"):
		local := middle[len("mailto:"):]
		if urlizeEmailPattern.MatchString(local) {
			display := trimURLDisplay(local, trimLimit)
			return fmt.Sprintf("<a href=\"%s\">%s</a>", middle, display)
		}
	case looksLikeEmail(middle):
		return fmt.Sprintf("<a href=\"mailto:%s\">%s</a>", middle, middle)
	case matchesExtraScheme(lower, extraSchemes):
		display := trimURLDisplay(middle, trimLimit)
		return fmt.Sprintf("<a href=\"%s\"%s%s>%s</a>", middle, relAttr, targetAttr, display)
	}

	return ""
}

func looksLikeEmail(value string) bool {
	if strings.HasPrefix(value, "@") || strings.Contains(value, ":") {
		return false
	}
	if strings.HasPrefix(strings.ToLower(value), "www.") {
		return false
	}
	return urlizeEmailPattern.MatchString(value)
}

func matchesExtraScheme(value string, schemes []string) bool {
	if len(schemes) == 0 {
		return false
	}
	for _, scheme := range schemes {
		if scheme == "" {
			continue
		}
		if strings.EqualFold(value, scheme) {
			continue
		}
		if strings.HasPrefix(strings.ToLower(value), strings.ToLower(scheme)) {
			return true
		}
	}
	return false
}

func xmlAttrValue(value interface{}) (string, bool) {
	if value == nil || isUndefinedValue(value) {
		return "", false
	}

	switch v := value.(type) {
	case Markup:
		return string(v), true
	case string:
		return v, false
	case []string:
		return strings.Join(v, " "), false
	case []interface{}:
		parts := make([]string, 0, len(v))
		for _, item := range v {
			part := toString(item)
			if part != "" {
				parts = append(parts, part)
			}
		}
		return strings.Join(parts, " "), false
	default:
		return toString(value), false
	}
}

func toFloat64(val interface{}) (float64, bool) {
	switch v := val.(type) {
	case int:
		return float64(v), true
	case int64:
		return float64(v), true
	case bool:
		if v {
			return 1, true
		}
		return 0, true
	case float64:
		return v, true
	case float32:
		return float64(v), true
	case string:
		if f, err := strconv.ParseFloat(v, 64); err == nil {
			return f, true
		}
	case Markup:
		if f, err := strconv.ParseFloat(string(v), 64); err == nil {
			return f, true
		}
	}
	return 0, false
}

func toInt(val interface{}) (int, bool) {
	switch v := val.(type) {
	case int:
		return v, true
	case int64:
		return int(v), true
	case float64:
		return int(v), true
	case float32:
		return int(v), true
	case string:
		if i, err := strconv.Atoi(v); err == nil {
			return i, true
		}
	}
	return 0, false
}

func toString(value interface{}) string {
	if value == nil {
		return ""
	}

	switch v := value.(type) {
	case string:
		return v
	case undefinedType:
		str, err := v.ToString()
		if err != nil {
			return ""
		}
		return str
	case fmt.Stringer:
		return v.String()
	default:
		return fmt.Sprintf("%v", value)
	}
}

func isTruthyValue(value interface{}) bool {
	if value == nil {
		return false
	}

	if isUndefinedValue(value) {
		return false
	}

	switch v := value.(type) {
	case bool:
		return v
	case int:
		return v != 0
	case int64:
		return v != 0
	case float64:
		return v != 0
	case float32:
		return v != 0
	case string:
		return v != ""
	case []interface{}:
		return len(v) > 0
	case []string:
		return len(v) > 0
	case map[interface{}]interface{}:
		return len(v) > 0
	case map[string]interface{}:
		return len(v) > 0
	default:
		return true
	}
}

func getAttribute(obj interface{}, attr string) (interface{}, error) {
	if obj == nil {
		return nil, nil
	}

	val := reflect.ValueOf(obj)

	// Handle pointers
	if val.Kind() == reflect.Ptr {
		if val.IsNil() {
			return nil, nil
		}
		val = val.Elem()
	}

	switch val.Kind() {
	case reflect.Map:
		keyVal := reflect.ValueOf(attr)
		if !keyVal.Type().ConvertibleTo(val.Type().Key()) {
			return nil, nil
		}
		convertedKey := keyVal.Convert(val.Type().Key())
		if result := val.MapIndex(convertedKey); result.IsValid() {
			return result.Interface(), nil
		}
	case reflect.Struct:
		field := val.FieldByName(attr)
		if field.IsValid() && field.CanInterface() {
			return field.Interface(), nil
		}

		// Try methods
		method := val.MethodByName(attr)
		if method.IsValid() && method.CanInterface() {
			return method.Interface(), nil
		}
	}

	return nil, nil
}

func compareValues(a, b interface{}, caseSensitive bool) int {
	if !caseSensitive {
		if strA, ok := a.(string); ok {
			if strB, ok := b.(string); ok {
				return strings.Compare(strings.ToLower(strA), strings.ToLower(strB))
			}
		}
	}

	// Try numeric comparison
	if numA, ok := toFloat64(a); ok {
		if numB, ok := toFloat64(b); ok {
			if numA < numB {
				return -1
			} else if numA > numB {
				return 1
			}
			return 0
		}
	}

	// String comparison
	strA := toString(a)
	strB := toString(b)
	return strings.Compare(strA, strB)
}
