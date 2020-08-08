package migrate

import (
	"strings"

	"github.com/neuronlabs/neuron/errors"
	"github.com/neuronlabs/neuron/mapping"

	"github.com/neuronlabs/neuron-extensions/repository/postgres/log"
)

// TagSetterFunc is the function that sets the proper column info for the given tag.
// I.e. having set the 'name' - tag key to a 'NameSetter' function allows to set the Column name with tag value.
type TagSetterFunc func(*mapping.StructField, *mapping.FieldTag) error

// TagSetterFunctions is the mapping for the tags with their TagSetterFunc.
var TagSetterFunctions = map[string]TagSetterFunc{}

// RegisterTagSetter registers the TagSetter function for given tag key.
func RegisterTagSetter(key string, setter TagSetterFunc) error {
	_, ok := TagSetterFunctions[key]
	if ok {
		log.Errorf("The TagSetter function for the key: '%s' already registered.", key)
		return errors.WrapDet(errors.ErrInternal, "tag setter function is already stored")
	}
	TagSetterFunctions[key] = setter
	return nil
}

// NameSetter is the TagSetter function that sets the Column's DBName
func NameSetter(f *mapping.StructField, t *mapping.FieldTag) error {
	// get the fields column
	c, err := fieldsColumn(f)
	if err != nil {
		return err
	}

	if len(t.Values) > 0 {
		c.Name = t.Values[0]
	} else {
		log.Debugf("Provided 'name' tag with no value")
	}

	return nil
}

// IndexSetter is the TagSetter function that sets the Column's Index.
func IndexSetter(field *mapping.StructField, t *mapping.FieldTag) error {
	var (
		name string
		tp   IndexType
	)

	c, err := fieldsColumn(field)
	if err != nil {
		return err
	}

	// check if there is a name in the tag values
	for _, value := range t.Values {
		// check if the
		if i := strings.IndexRune(value, '='); i != -1 {
			k := value[:i]

			var v string
			if len(value)-1 > i {
				v = value[i+1:]

				switch k {
				case "name":
					name = v
				case "type":
					switch v {
					case BTreeTag:
						tp = BTree
					case HashTag:
						tp = Hash
					case GiSTTag:
						tp = GiST
					case GINTag:
						tp = GIN
					}
				}
			}
		} else {
			switch value {
			case BTreeTag:
				tp = BTree
			case HashTag:
				tp = Hash
			case GiSTTag:
				tp = GiST
			case GINTag:
				tp = GIN
			default:
				name = value
			}
		}
	}

	tb, err := modelsTable(field.Struct())
	if err != nil {
		return err
	}

	tb.createIndex(name, tp, c)
	return nil
}

// DataTypeSetter is the TagSetter function that sets the proper data type for given tag value or field.
func DataTypeSetter(field *mapping.StructField, t *mapping.FieldTag) error {
	if len(t.Values) > 0 {
		v, err := parseDataType(t.Values[0])
		if err != nil {
			return err
		}

		dt, ok := dataTypes[v[0]]
		if !ok {
			return err
		}

		c, err := fieldsColumn(field)
		if err != nil {
			return err
		}

		c.Type = dt

		if len(v) > 0 {
			field.StoreSet(DataTypeParametersStoreKey, v[1:])
		}
	}
	return nil
}

// ConstraintSetter is the TagSetterFunc for the constraints.
func ConstraintSetter(field *mapping.StructField, tag *mapping.FieldTag) error {
	c, err := fieldsColumn(field)
	if err != nil {
		return err
	}

	switch tag.Key {
	case cNotNull:
		log.Debugf("Adding NOT NULL constraint to column: '%s'", c.Name)
		c.Constraints = append(c.Constraints, CNotNull)
		field.StoreSet(NotNullKey, struct{}{})
	case cUnique:
		c.Constraints = append(c.Constraints, CUnique)
	}
	return nil
}

func parseDataType(v string) ([]string, error) {
	i := strings.Index(v, "(")
	if i == -1 {
		return []string{v}, nil
	} else if v[len(v)-1] != ')' {
		return nil, errors.WrapDetf(errors.ErrInternal, "invalid postgres DataType value: '%s'", v)
	}

	return append([]string{v[:i]}, strings.Split(v[i+1:len(v)-1], ",")...), nil
}
