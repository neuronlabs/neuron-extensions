package tempfuncs

// AlternateTypes is the mapping between basic types and it's Alternate values.
var AlternateTypes = map[string][]string{
	"int":     {"int64", "uint", "int32", "int16", "int8", "float64"},
	"int8":    {"int64", "int", "uint", "int32", "int16", "uint8"},
	"int16":   {"int64", "int", "int32", "uint16", "int8", "uint"},
	"int32":   {"int64", "uint", "uint32", "int16", "int8", "float64"},
	"int64":   {"uint64", "uint", "int", "int32", "int16", "int8", "float64"},
	"uint":    {"uint64", "int", "uint32", "uint16", "uint8", "float64"},
	"uint8":   {"uint64", "int", "uint", "uint32", "uint16", "int8"},
	"uint16":  {"uint64", "int", "uint", "uint32", "int16", "uint8"},
	"uint32":  {"uint64", "int", "uint", "int32", "uint16", "uint8", "float64"},
	"uint64":  {"int64", "int", "uint", "uint32", "uint16", "uint8", "float64"},
	"float64": {"float32", "int", "uint", "int64", "uint64"},
	"float32": {"float64", "int", "uint", "int64", "uint64"},
}

// GetAlternateTypes is a template function that gets
func GetAlternateTypes(tp string) []string {
	return AlternateTypes[tp]
}
