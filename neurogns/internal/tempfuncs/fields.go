package tempfuncs

import (
	"fmt"
	"strings"

	"github.com/neuronlabs/neuron-extensions/neurogns/input"
)

func IsFieldWrappedSlice(field *input.Field) bool {
	if strings.ContainsRune(field.Type, '[') || !field.IsSlice {
		return false
	}
	var isAnySlice bool
	for _, wrapped := range field.WrappedTypes {
		if IsWrappedTypeSlice(wrapped) {
			isAnySlice = true
			break
		}
	}
	return isAnySlice
}

func IsWrappedTypeSlice(wrapped string) bool {
	return strings.ContainsRune(wrapped, '[')
}

func FieldsWrappedTypeElem(field *input.Field) string {
	for _, wrapped := range field.WrappedTypes {
		if IsWrappedTypeSlice(wrapped) {
			i := strings.IndexRune(wrapped, ']')
			if i == -1 {
				panic("Provided wrapped type: '%s' is not an array")
			}
			return wrapped[i+1:]
		}
	}
	panic(fmt.Sprintf("no wrapped type slice field found for the field: '%s'", field.Name))
}
