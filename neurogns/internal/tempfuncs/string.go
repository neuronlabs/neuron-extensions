package tempfuncs

// Stringers contains stringer functions used in the templates.
var StringerFuncs = map[string]StringerFunc{
	StringerString:        DummyStringGetterFunction,
	StringerFInterface:    StringGetterInterface,
	StringerFormatFloat64: FormatFloat64,
	StringerFormatFloat32: FormatFloat32,
	StringerFormatInt:     FormatInt,
	StringerFormatInt8:    FormatInt8,
	StringerFormatInt16:   FormatInt16,
	StringerFormatInt32:   FormatInt32,
	StringerFormatInt64:   FormatInt64,
	StringerFormatUint:    FormatUint,
	StringerFormatUint8:   FormatUint8,
	StringerFormatUint16:  FormatUint16,
	StringerFormatUint32:  FormatUint32,
	StringerFormatUint64:  FormatUint64,
	StringerFormatBoolean: FormatBoolean,
}

// Stringer is the golang template string converter function.
func Stringer(name, funcName string) string {
	return StringerFuncs[funcName](name)
}

// Selector is a function that returns field with selector.
func Selector(selector, fieldName string) string {
	if selector != "" {
		fieldName = selector + "." + fieldName
	}
	return fieldName
}

// StringerFunc is a function that converts provided selector into string.
type StringerFunc func(name string) string

// StringerString is the name of the stringer string function.
const StringerString = "stringer-string"

// DummyStringGetterFunction is the stringer function that return string field value.
func DummyStringGetterFunction(name string) string {
	return name
}

// StringGetterInterface.
var _ StringerFunc = StringGetterInterface

// StringGetterInterface is the name of the stringer function that maps for StringGetterInterface.
const StringerFInterface = "stringer-interface"

func StringGetterInterface(name string) string {
	return name + ".String()"
}

//
// Floats Stringer functions
//
var (
	_ StringerFunc = FormatFloat64
	_ StringerFunc = FormatFloat32
)

const (
	StringerFormatFloat64 = "stringer-format-float64"
	StringerFormatFloat32 = "stringer-format-float32"
)

func FormatFloat64(name string) string {
	return "strconv.FormatFloat64(" + name + ", 'f', -1, 64)"
}

func FormatFloat32(name string) string {
	return "strconv.FormatFloat64(float64(" + name + "), 'f', -1, 32)"
}

// Integer stringer function names.
const (
	StringerFormatInt64 = "stringer-format-int64"
	StringerFormatInt   = "stringer-format-int"
	StringerFormatInt32 = "stringer-format-int32"
	StringerFormatInt16 = "stringer-format-int16"
	StringerFormatInt8  = "stringer-format-int8"
)

// FormatInt is the int stringer function.
func FormatInt(name string) string {
	return formatIntFunction(name, true)
}

// FormatInt64 is the int64 stringer function.
func FormatInt64(name string) string {
	return formatIntFunction(name, false)
}

// FormatInt32 is the int32 stringer function.
func FormatInt32(name string) string {
	return formatIntFunction(name, true)
}

// FormatInt16 is the int16 stringer function.
func FormatInt16(name string) string {
	return formatIntFunction(name, true)
}

// FormatInt8 is the int8 stringer function.
func FormatInt8(name string) string {
	return formatIntFunction(name, true)
}

func formatIntFunction(name string, wrap bool) string {
	fieldSelector := name
	if wrap {
		fieldSelector = "int64(" + fieldSelector + ")"
	}
	return "strconv.FormatInt(" + fieldSelector + ", 10)"
}

// Integer stringer function names.
const (
	StringerFormatUint64 = "stringer-format-uint64"
	StringerFormatUint   = "stringer-format-uint"
	StringerFormatUint32 = "stringer-format-uint32"
	StringerFormatUint16 = "stringer-format-uint16"
	StringerFormatUint8  = "stringer-format-uint8"
)

// FormatUint64 is the uint64 stringer function.
func FormatUint64(name string) string {
	return formatUintFunction(name, false)
}

// FormatUint is the uint stringer function.
func FormatUint(name string) string {
	return formatUintFunction(name, true)
}

// FormatUint32 is the uint32 stringer function.
func FormatUint32(name string) string {
	return formatUintFunction(name, true)
}

// FormatUint16 is the uint16 stringer function.
func FormatUint16(name string) string {
	return formatUintFunction(name, true)
}

// FormatUint8 is the int8 stringer function.
func FormatUint8(name string) string {
	return formatUintFunction(name, true)
}

func formatUintFunction(name string, wrap bool) string {
	fieldSelector := name
	if wrap {
		fieldSelector = "uint64(" + fieldSelector + ")"
	}
	return "strconv.FormatUint(" + fieldSelector + ", 10)"
}

const (
	StringerFormatBoolean = "stringer-format-bool"
	ParserBoolean         = "stringer-parse-bool"
)

// FormatBoolean is a stringer function for the boolean field.
func FormatBoolean(name string) string {
	return "strconv.FormatBool(" + name + ")"
}
