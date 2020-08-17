package tempfuncs

// Parsers is the mapping between parsing name and its functions.
var Parsers = map[string]StringParserFunc{
	ParserFloat64:     ParseFloat64,
	ParserFloat32:     ParseFloat32,
	ParserInt64:       ParseInt64,
	ParserInt:         ParseInt,
	ParserInt32:       ParseInt32,
	ParserInt16:       ParseInt16,
	ParserInt8:        ParseInt8,
	ParserUint64:      ParseUint64,
	ParserUint:        ParseUint,
	ParserUint32:      ParseUint32,
	ParserUint16:      ParseUint16,
	ParserUint8:       ParseUint8,
	ParserBoolean:     ParseBoolean,
	DummyStringParser: ParseDummyString,
}

// Parser gets the parser function and executes with given 'name'.
func Parser(name, funcName string) string {
	return Parsers[funcName](name)
}

// ParserWrapper is a template function that checks if given fieldType should be wrapped when used for given parser 'funcName'.
func ParserWrapper(funcName, fieldType string) bool {
	switch funcName {
	case ParserFloat64:
		if fieldType != "float64" {
			return true
		}
	case ParserFloat32:
		return true
	case ParserInt64, ParserInt, ParserInt32, ParserInt16, ParserInt8:
		if fieldType != "int64" {
			return true
		}
	case ParserUint64, ParserUint, ParserUint32, ParserUint16, ParserUint8:
		if fieldType != "uint64" {
			return true
		}
	}
	return false
}

// StringParserFunc is a function that converts provided selector and parses with a possible error into given value.
type StringParserFunc func(name string) string

const (
	ParserFloat64 = "float64"
	ParserFloat32 = "float32"
)

var (
	_ StringerFunc = ParseFloat32
	_ StringerFunc = ParseFloat64
)

// ParseFloat64 gets the float64 parser function.
func ParseFloat64(name string) string {
	return "strconv.ParseFloat(" + name + ", 64)"
}

// ParseFloat32 gets the float32 parser function.
func ParseFloat32(name string) string {
	return "strconv.ParseFloat(" + name + ", 64)"
}

const (
	ParserInt64 = "int64"
	ParserInt   = "int"
	ParserInt32 = "int32"
	ParserInt16 = "int16"
	ParserInt8  = "int8"
)

// ParseInt gets a string parses for int.
func ParseInt(name string) string {
	return parseIntFunction(name, "mapping.IntegerBitSize", false)
}

// ParseInt64 gets a string parses for int64.
func ParseInt64(name string) string {
	return parseIntFunction(name, "64", false)
}

// ParseInt32 gets a string parses for int32.
func ParseInt32(name string) string {
	return parseIntFunction(name, "32", false)
}

// ParseInt16 gets a string parses for int16.
func ParseInt16(name string) string {
	return parseIntFunction(name, "16", false)
}

// ParseInt8 gets a string parses for int8.
func ParseInt8(name string) string {
	return parseIntFunction(name, "8", false)
}

func parseIntFunction(name, bitSize string, wrap bool) string {
	if wrap {
		name = "int64(" + name + ")"
	}
	return "strconv.ParseInt(" + name + ", 10," + bitSize + ")"
}

const (
	ParserUint64 = "uint64"
	ParserUint   = "uint"
	ParserUint32 = "uint32"
	ParserUint16 = "uint16"
	ParserUint8  = "uint8"
)

// ParseUint is uint stringer function.
func ParseUint(name string) string {
	return parseUintFunction(name, "mapping.IntegerBitSize", false)
}

// ParseUint64 is uint64 stringer function.
func ParseUint64(name string) string {
	return parseUintFunction(name, "64", false)
}

// ParseUint32 is uint32 stringer function.
func ParseUint32(name string) string {
	return parseUintFunction(name, "32", false)
}

// ParseUint16 is uint64 stringer function.
func ParseUint16(name string) string {
	return parseUintFunction(name, "16", false)
}

// ParseUint8 is uint8 stringer function.
func ParseUint8(name string) string {
	return parseUintFunction(name, "8", false)
}

func parseUintFunction(name, bitSize string, wrap bool) string {
	if wrap {
		name = "uint64(" + name + ")"
	}
	return "strconv.ParseUint(" + name + ", 10," + bitSize + ")"
}

// ParseBoolean is a stringer function that parses boolean.
func ParseBoolean(name string) string {
	return "strconv.ParseBool(" + name + ")"
}

// DummyStringParser is dummy field string parser
const DummyStringParser = "string"

func ParseDummyString(name string) string {
	return name + ", nil"
}
