// Code generated by neurogonesis. DO NOT EDIT.
// This file was generated at:
// Wed, 02 Sep 2020 15:43:38 +0200

package migrate

import (
	"strconv"
	"time"

	"github.com/neuronlabs/neuron/errors"
	"github.com/neuronlabs/neuron/mapping"
)

// Neuron_Models stores all generated models in this package.
var Neuron_Models = []mapping.Model{
	&BasicModel{},
	&Model{},
}

// Compile time check if BasicModel implements mapping.Model interface.
var _ mapping.Model = &BasicModel{}

// NeuronCollectionName implements mapping.Model interface method.
// Returns the name of the collection for the 'BasicModel'.
func (b *BasicModel) NeuronCollectionName() string {
	return "basic_models"
}

// IsPrimaryKeyZero implements mapping.Model interface method.
func (b *BasicModel) IsPrimaryKeyZero() bool {
	return b.ID == 0
}

// GetPrimaryKeyValue implements mapping.Model interface method.
func (b *BasicModel) GetPrimaryKeyValue() interface{} {
	return b.ID
}

// GetPrimaryKeyStringValue implements mapping.Model interface method.
func (b *BasicModel) GetPrimaryKeyStringValue() (string, error) {
	return strconv.FormatInt(int64(b.ID), 10), nil
}

// GetPrimaryKeyAddress implements mapping.Model interface method.
func (b *BasicModel) GetPrimaryKeyAddress() interface{} {
	return &b.ID
}

// GetPrimaryKeyHashableValue implements mapping.Model interface method.
func (b *BasicModel) GetPrimaryKeyHashableValue() interface{} {
	return b.ID
}

// GetPrimaryKeyZeroValue implements mapping.Model interface method.
func (b *BasicModel) GetPrimaryKeyZeroValue() interface{} {
	return 0
}

// SetPrimaryKey implements mapping.Model interface method.
func (b *BasicModel) SetPrimaryKeyValue(value interface{}) error {
	if _v, ok := value.(int); ok {
		b.ID = _v
		return nil
	}
	// Check alternate types for given field.
	switch _valueType := value.(type) {
	case int8:
		b.ID = int(_valueType)
	case int16:
		b.ID = int(_valueType)
	case int32:
		b.ID = int(_valueType)
	case int64:
		b.ID = int(_valueType)
	case uint:
		b.ID = int(_valueType)
	case uint8:
		b.ID = int(_valueType)
	case uint16:
		b.ID = int(_valueType)
	case uint32:
		b.ID = int(_valueType)
	case uint64:
		b.ID = int(_valueType)
	case float32:
		b.ID = int(_valueType)
	case float64:
		b.ID = int(_valueType)
	default:
		return errors.Wrapf(mapping.ErrFieldValue, "provided invalid value: '%T' for the primary field for model: 'BasicModel'", value)
	}
	return nil
}

// SetPrimaryKeyStringValue implements mapping.Model interface method.
func (b *BasicModel) SetPrimaryKeyStringValue(value string) error {
	tmp, err := strconv.ParseInt(value, 10, mapping.IntegerBitSize)
	if err != nil {
		return err
	}
	b.ID = int(tmp)
	return nil
}

// SetFrom implements FromSetter interface.
func (b *BasicModel) SetFrom(model mapping.Model) error {
	if model == nil {
		return errors.Wrap(mapping.ErrNilModel, "provided nil model to set from")
	}
	from, ok := model.(*BasicModel)
	if !ok {
		return errors.WrapDetf(mapping.ErrModelNotMatch, "provided model doesn't match the input: %T", model)
	}
	*b = *from
	return nil
}

// StructFieldValues gets the value for specified 'field'.
func (b *BasicModel) StructFieldValue(field *mapping.StructField) (interface{}, error) {
	switch field.Index[0] {
	case 0: // ID
		return b.ID, nil
	case 1: // String
		return b.String, nil
	case 2: // Timed
		return b.Timed, nil
	case 3: // PtrTime
		return b.PtrTime, nil
	case 4: // Int
		return b.Int, nil
	case 5: // Int16
		return b.Int16, nil
	case 6: // Varchar20
		return b.Varchar20, nil
	case 7: // Float32
		return b.Float32, nil
	case 8: // IntArray
		return b.IntArray, nil
	case 9: // IntSlice
		return b.IntSlice, nil
	default:
		return nil, errors.Wrapf(mapping.ErrInvalidModelField, "provided invalid field: '%s' for given model: BasicModel'", field.Name())
	}
}

// Compile time check if BasicModel implements mapping.Fielder interface.
var _ mapping.Fielder = &BasicModel{}

// GetFieldsAddress gets the address of provided 'field'.
func (b *BasicModel) GetFieldsAddress(field *mapping.StructField) (interface{}, error) {
	switch field.Index[0] {
	case 0: // ID
		return &b.ID, nil
	case 1: // String
		return &b.String, nil
	case 2: // Timed
		return &b.Timed, nil
	case 3: // PtrTime
		return &b.PtrTime, nil
	case 4: // Int
		return &b.Int, nil
	case 5: // Int16
		return &b.Int16, nil
	case 6: // Varchar20
		return &b.Varchar20, nil
	case 7: // Float32
		return &b.Float32, nil
	case 8: // IntArray
		return &b.IntArray, nil
	case 9: // IntSlice
		return &b.IntSlice, nil
	}
	return nil, errors.Wrapf(mapping.ErrInvalidModelField, "provided invalid field: '%s' for given model: BasicModel'", field.Name())
}

// GetFieldZeroValue implements mapping.Fielder interface.s
func (b *BasicModel) GetFieldZeroValue(field *mapping.StructField) (interface{}, error) {
	switch field.Index[0] {
	case 0: // ID
		return 0, nil
	case 1: // String
		return "", nil
	case 2: // Timed
		return time.Time{}, nil
	case 3: // PtrTime
		return nil, nil
	case 4: // Int
		return 0, nil
	case 5: // Int16
		return 0, nil
	case 6: // Varchar20
		return "", nil
	case 7: // Float32
		return 0, nil
	case 8: // IntArray
		return [3]int{}, nil
	case 9: // IntSlice
		return nil, nil
	default:
		return nil, errors.Wrapf(mapping.ErrInvalidModelField, "provided invalid field name: '%s'", field.Name())
	}
}

// IsFieldZero implements mapping.Fielder interface.
func (b *BasicModel) IsFieldZero(field *mapping.StructField) (bool, error) {
	switch field.Index[0] {
	case 0: // ID
		return b.ID == 0, nil
	case 1: // String
		return b.String == "", nil
	case 2: // Timed
		return b.Timed == time.Time{}, nil
	case 3: // PtrTime
		return b.PtrTime == nil, nil
	case 4: // Int
		return b.Int == 0, nil
	case 5: // Int16
		return b.Int16 == 0, nil
	case 6: // Varchar20
		return b.Varchar20 == "", nil
	case 7: // Float32
		return b.Float32 == 0, nil
	case 8: // IntArray
		return b.IntArray == [3]int{}, nil
	case 9: // IntSlice
		return len(b.IntSlice) == 0, nil
	}
	return false, errors.Wrapf(mapping.ErrInvalidModelField, "provided invalid field name: '%s'", field.Name())
}

// SetFieldZeroValue implements mapping.Fielder interface.s
func (b *BasicModel) SetFieldZeroValue(field *mapping.StructField) error {
	switch field.Index[0] {
	case 0: // ID
		b.ID = 0
	case 1: // String
		b.String = ""
	case 2: // Timed
		b.Timed = time.Time{}
	case 3: // PtrTime
		b.PtrTime = nil
	case 4: // Int
		b.Int = 0
	case 5: // Int16
		b.Int16 = 0
	case 6: // Varchar20
		b.Varchar20 = ""
	case 7: // Float32
		b.Float32 = 0
	case 8: // IntArray
		b.IntArray = [3]int{}
	case 9: // IntSlice
		b.IntSlice = nil
	default:
		return errors.Wrapf(mapping.ErrInvalidModelField, "provided invalid field name: '%s'", field.Name())
	}
	return nil
}

// GetHashableFieldValue implements mapping.Fielder interface.
func (b *BasicModel) GetHashableFieldValue(field *mapping.StructField) (interface{}, error) {
	switch field.Index[0] {
	case 0: // ID
		return b.ID, nil
	case 1: // String
		return b.String, nil
	case 2: // Timed
		return b.Timed, nil
	case 3: // PtrTime
		if b.PtrTime == nil {
			return nil, nil
		}
		return *b.PtrTime, nil
	case 4: // Int
		return b.Int, nil
	case 5: // Int16
		return b.Int16, nil
	case 6: // Varchar20
		return b.Varchar20, nil
	case 7: // Float32
		return b.Float32, nil
	case 8: // IntArray
		return b.IntArray, nil
	case 9: // IntSlice
		return b.IntSlice, nil
	}
	return nil, errors.Wrapf(mapping.ErrInvalidModelField, "provided invalid field: '%s' for given model: 'BasicModel'", field.Name())
}

// GetFieldValue implements mapping.Fielder interface.
func (b *BasicModel) GetFieldValue(field *mapping.StructField) (interface{}, error) {
	switch field.Index[0] {
	case 0: // ID
		return b.ID, nil
	case 1: // String
		return b.String, nil
	case 2: // Timed
		return b.Timed, nil
	case 3: // PtrTime
		return b.PtrTime, nil
	case 4: // Int
		return b.Int, nil
	case 5: // Int16
		return b.Int16, nil
	case 6: // Varchar20
		return b.Varchar20, nil
	case 7: // Float32
		return b.Float32, nil
	case 8: // IntArray
		return b.IntArray, nil
	case 9: // IntSlice
		return b.IntSlice, nil
	}
	return nil, errors.Wrapf(mapping.ErrInvalidModelField, "provided invalid field: '%s' for given model: BasicModel'", field.Name())
}

// SetFieldValue implements mapping.Fielder interface.
func (b *BasicModel) SetFieldValue(field *mapping.StructField, value interface{}) (err error) {
	switch field.Index[0] {
	case 0: // ID
		if _v, ok := value.(int); ok {
			b.ID = _v
			return nil
		}
		if field.DatabaseNotNull() && value == nil {
			b.ID = 0
			return nil
		}

		switch _v := value.(type) {
		case int8:
			b.ID = int(_v)
		case int16:
			b.ID = int(_v)
		case int32:
			b.ID = int(_v)
		case int64:
			b.ID = int(_v)
		case uint:
			b.ID = int(_v)
		case uint8:
			b.ID = int(_v)
		case uint16:
			b.ID = int(_v)
		case uint32:
			b.ID = int(_v)
		case uint64:
			b.ID = int(_v)
		case float32:
			b.ID = int(_v)
		case float64:
			b.ID = int(_v)
		default:
			return errors.Wrapf(mapping.ErrFieldValue, "provided invalid field type: '%T' for the field: %s", value, field.Name())
		}
		return nil
	case 1: // String
		if _v, ok := value.(string); ok {
			b.String = _v
			return nil
		}
		if field.DatabaseNotNull() && value == nil {
			b.String = ""
			return nil
		}

		// Check alternate types for the String.
		if _v, ok := value.([]byte); ok {
			b.String = string(_v)
			return nil
		}
		return errors.Wrapf(mapping.ErrFieldValue, "provided invalid field type: '%T' for the field: %s", value, field.Name())
	case 2: // Timed
		if _v, ok := value.(time.Time); ok {
			b.Timed = _v
			return nil
		}
		if field.DatabaseNotNull() && value == nil {
			b.Timed = time.Time{}
			return nil
		}

		return errors.Wrapf(mapping.ErrFieldValue, "provided invalid field type: '%T' for the field: %s", value, field.Name())
	case 3: // PtrTime
		if value == nil {
			b.PtrTime = nil
			return nil
		}
		if _v, ok := value.(*time.Time); ok {
			b.PtrTime = _v
			return nil
		}
		// Check if it is non-pointer value.
		if _v, ok := value.(time.Time); ok {
			b.PtrTime = &_v
			return nil
		}

		return errors.Wrapf(mapping.ErrFieldValue, "provided invalid field type: '%T' for the field: %s", value, field.Name())
	case 4: // Int
		if _v, ok := value.(int); ok {
			b.Int = _v
			return nil
		}
		if field.DatabaseNotNull() && value == nil {
			b.Int = 0
			return nil
		}

		switch _v := value.(type) {
		case int8:
			b.Int = int(_v)
		case int16:
			b.Int = int(_v)
		case int32:
			b.Int = int(_v)
		case int64:
			b.Int = int(_v)
		case uint:
			b.Int = int(_v)
		case uint8:
			b.Int = int(_v)
		case uint16:
			b.Int = int(_v)
		case uint32:
			b.Int = int(_v)
		case uint64:
			b.Int = int(_v)
		case float32:
			b.Int = int(_v)
		case float64:
			b.Int = int(_v)
		default:
			return errors.Wrapf(mapping.ErrFieldValue, "provided invalid field type: '%T' for the field: %s", value, field.Name())
		}
		return nil
	case 5: // Int16
		if _v, ok := value.(int16); ok {
			b.Int16 = _v
			return nil
		}
		if field.DatabaseNotNull() && value == nil {
			b.Int16 = 0
			return nil
		}

		switch _v := value.(type) {
		case int:
			b.Int16 = int16(_v)
		case int8:
			b.Int16 = int16(_v)
		case int32:
			b.Int16 = int16(_v)
		case int64:
			b.Int16 = int16(_v)
		case uint:
			b.Int16 = int16(_v)
		case uint8:
			b.Int16 = int16(_v)
		case uint16:
			b.Int16 = int16(_v)
		case uint32:
			b.Int16 = int16(_v)
		case uint64:
			b.Int16 = int16(_v)
		case float32:
			b.Int16 = int16(_v)
		case float64:
			b.Int16 = int16(_v)
		default:
			return errors.Wrapf(mapping.ErrFieldValue, "provided invalid field type: '%T' for the field: %s", value, field.Name())
		}
		return nil
	case 6: // Varchar20
		if _v, ok := value.(string); ok {
			b.Varchar20 = _v
			return nil
		}
		if field.DatabaseNotNull() && value == nil {
			b.Varchar20 = ""
			return nil
		}

		// Check alternate types for the Varchar20.
		if _v, ok := value.([]byte); ok {
			b.Varchar20 = string(_v)
			return nil
		}
		return errors.Wrapf(mapping.ErrFieldValue, "provided invalid field type: '%T' for the field: %s", value, field.Name())
	case 7: // Float32
		if _v, ok := value.(float32); ok {
			b.Float32 = _v
			return nil
		}
		if field.DatabaseNotNull() && value == nil {
			b.Float32 = 0
			return nil
		}

		switch _v := value.(type) {
		case int:
			b.Float32 = float32(_v)
		case int8:
			b.Float32 = float32(_v)
		case int16:
			b.Float32 = float32(_v)
		case int32:
			b.Float32 = float32(_v)
		case int64:
			b.Float32 = float32(_v)
		case uint:
			b.Float32 = float32(_v)
		case uint8:
			b.Float32 = float32(_v)
		case uint16:
			b.Float32 = float32(_v)
		case uint32:
			b.Float32 = float32(_v)
		case uint64:
			b.Float32 = float32(_v)
		case float64:
			b.Float32 = float32(_v)
		default:
			return errors.Wrapf(mapping.ErrFieldValue, "provided invalid field type: '%T' for the field: %s", value, field.Name())
		}
		return nil
	case 8: // IntArray
		if _v, ok := value.([3]int); ok {
			b.IntArray = _v
			return nil
		}
		if generic, ok := value.([]interface{}); ok {
			if len(generic) > 3 {
				return errors.Wrapf(mapping.ErrFieldValue, "provided too many values for the field: 'IntArray")
			}
			for i, item := range generic {
				if _v, ok := item.(int); ok {
					b.IntArray[i] = _v
					continue
				}
				switch _v := item.(type) {
				case int64:
					b.IntArray[i] = int(_v)
				case uint:
					b.IntArray[i] = int(_v)
				case int32:
					b.IntArray[i] = int(_v)
				case int16:
					b.IntArray[i] = int(_v)
				case int8:
					b.IntArray[i] = int(_v)
				case float64:
					b.IntArray[i] = int(_v)
				default:
					return errors.Wrapf(mapping.ErrFieldValue, "provided invalid field type: '%T' for the field: %s", value, field.Name())
				}
			}
			return nil
		}
		return errors.Wrapf(mapping.ErrFieldValue, "provided invalid field type: '%T' for the field: %s", value, field.Name())
	case 9: // IntSlice
		if value == nil {
			b.IntSlice = nil
			return nil
		}
		if _v, ok := value.([]int); ok {
			b.IntSlice = _v
			return nil
		}
		if generic, ok := value.([]interface{}); ok {
			for _, item := range generic {
				if _v, ok := item.(int); ok {
					b.IntSlice = append(b.IntSlice, _v)
					continue
				}
				switch _v := item.(type) {
				case int64:
					b.IntSlice = append(b.IntSlice, int(_v))
				case uint:
					b.IntSlice = append(b.IntSlice, int(_v))
				case int32:
					b.IntSlice = append(b.IntSlice, int(_v))
				case int16:
					b.IntSlice = append(b.IntSlice, int(_v))
				case int8:
					b.IntSlice = append(b.IntSlice, int(_v))
				case float64:
					b.IntSlice = append(b.IntSlice, int(_v))
				default:
					return errors.Wrapf(mapping.ErrFieldValue, "provided invalid field type: '%T' for the field: %s", value, field.Name())
				}
			}
			return nil
		}
		return errors.Wrapf(mapping.ErrFieldValue, "provided invalid field type: '%T' for the field: %s", value, field.Name())
	default:
		return errors.Wrapf(mapping.ErrInvalidModelField, "provided invalid field: '%s' for the model: 'BasicModel'", field.Name())
	}
}

// SetPrimaryKeyStringValue implements mapping.Model interface method.
func (b *BasicModel) ParseFieldsStringValue(field *mapping.StructField, value string) (interface{}, error) {
	switch field.Index[0] {
	case 0: // ID
		return strconv.ParseInt(value, 10, mapping.IntegerBitSize)
	case 1: // String
		return value, nil
	case 2: // Timed
		temp := b.Timed
		if err := b.Timed.UnmarshalText([]byte(value)); err != nil {
			return "", errors.Wrapf(mapping.ErrFieldValue, "invalid field 'Timed' value: '%v' to parse string. Err: %v", b.Timed, err)
		}
		bt, err := b.Timed.MarshalText()
		if err != nil {
			return "", errors.Wrapf(mapping.ErrFieldValue, "invalid field 'Timed' value: '%v' to parse string. Err: %v", b.Timed, err)
		}
		b.Timed = temp
		return string(bt), nil
	case 3: // PtrTime
		var base time.Time
		temp := &base
		if err := temp.UnmarshalText([]byte(value)); err != nil {
			return "", errors.Wrapf(mapping.ErrFieldValue, "invalid field 'PtrTime' value: '%v' to parse string. Err: %v", b.PtrTime, err)
		}
		bt, err := temp.MarshalText()
		if err != nil {
			return "", errors.Wrapf(mapping.ErrFieldValue, "invalid field 'PtrTime' value: '%v' to parse string. Err: %v", b.PtrTime, err)
		}

		return string(bt), nil
	case 4: // Int
		return strconv.ParseInt(value, 10, mapping.IntegerBitSize)
	case 5: // Int16
		return strconv.ParseInt(value, 10, 16)
	case 6: // Varchar20
		return value, nil
	case 7: // Float32
		return strconv.ParseFloat(value, 64)
	case 8: // IntArray
		return strconv.ParseInt(value, 10, mapping.IntegerBitSize)
	case 9: // IntSlice
		return strconv.ParseInt(value, 10, mapping.IntegerBitSize)
	}
	return nil, errors.Wrapf(mapping.ErrInvalidModelField, "provided invalid field: '%s' for given model: BasicModel'", field.Name())
}

// Compile time check if Model implements mapping.Model interface.
var _ mapping.Model = &Model{}

// NeuronCollectionName implements mapping.Model interface method.
// Returns the name of the collection for the 'Model'.
func (m *Model) NeuronCollectionName() string {
	return "models"
}

// IsPrimaryKeyZero implements mapping.Model interface method.
func (m *Model) IsPrimaryKeyZero() bool {
	return m.ID == 0
}

// GetPrimaryKeyValue implements mapping.Model interface method.
func (m *Model) GetPrimaryKeyValue() interface{} {
	return m.ID
}

// GetPrimaryKeyStringValue implements mapping.Model interface method.
func (m *Model) GetPrimaryKeyStringValue() (string, error) {
	return strconv.FormatInt(int64(m.ID), 10), nil
}

// GetPrimaryKeyAddress implements mapping.Model interface method.
func (m *Model) GetPrimaryKeyAddress() interface{} {
	return &m.ID
}

// GetPrimaryKeyHashableValue implements mapping.Model interface method.
func (m *Model) GetPrimaryKeyHashableValue() interface{} {
	return m.ID
}

// GetPrimaryKeyZeroValue implements mapping.Model interface method.
func (m *Model) GetPrimaryKeyZeroValue() interface{} {
	return 0
}

// SetPrimaryKey implements mapping.Model interface method.
func (m *Model) SetPrimaryKeyValue(value interface{}) error {
	if _v, ok := value.(int); ok {
		m.ID = _v
		return nil
	}
	// Check alternate types for given field.
	switch _valueType := value.(type) {
	case int8:
		m.ID = int(_valueType)
	case int16:
		m.ID = int(_valueType)
	case int32:
		m.ID = int(_valueType)
	case int64:
		m.ID = int(_valueType)
	case uint:
		m.ID = int(_valueType)
	case uint8:
		m.ID = int(_valueType)
	case uint16:
		m.ID = int(_valueType)
	case uint32:
		m.ID = int(_valueType)
	case uint64:
		m.ID = int(_valueType)
	case float32:
		m.ID = int(_valueType)
	case float64:
		m.ID = int(_valueType)
	default:
		return errors.Wrapf(mapping.ErrFieldValue, "provided invalid value: '%T' for the primary field for model: 'Model'", value)
	}
	return nil
}

// SetPrimaryKeyStringValue implements mapping.Model interface method.
func (m *Model) SetPrimaryKeyStringValue(value string) error {
	tmp, err := strconv.ParseInt(value, 10, mapping.IntegerBitSize)
	if err != nil {
		return err
	}
	m.ID = int(tmp)
	return nil
}

// SetFrom implements FromSetter interface.
func (m *Model) SetFrom(model mapping.Model) error {
	if model == nil {
		return errors.Wrap(mapping.ErrNilModel, "provided nil model to set from")
	}
	from, ok := model.(*Model)
	if !ok {
		return errors.WrapDetf(mapping.ErrModelNotMatch, "provided model doesn't match the input: %T", model)
	}
	*m = *from
	return nil
}

// StructFieldValues gets the value for specified 'field'.
func (m *Model) StructFieldValue(field *mapping.StructField) (interface{}, error) {
	switch field.Index[0] {
	case 0: // ID
		return m.ID, nil
	case 1: // Attr
		return m.Attr, nil
	case 2: // SnakeCased
		return m.SnakeCased, nil
	case 3: // CreatedAt
		return m.CreatedAt, nil
	case 4: // UpdatedAt
		return m.UpdatedAt, nil
	case 5: // DeletedAt
		return m.DeletedAt, nil
	default:
		return nil, errors.Wrapf(mapping.ErrInvalidModelField, "provided invalid field: '%s' for given model: Model'", field.Name())
	}
}

// Compile time check if Model implements mapping.Fielder interface.
var _ mapping.Fielder = &Model{}

// GetFieldsAddress gets the address of provided 'field'.
func (m *Model) GetFieldsAddress(field *mapping.StructField) (interface{}, error) {
	switch field.Index[0] {
	case 0: // ID
		return &m.ID, nil
	case 1: // Attr
		return &m.Attr, nil
	case 2: // SnakeCased
		return &m.SnakeCased, nil
	case 3: // CreatedAt
		return &m.CreatedAt, nil
	case 4: // UpdatedAt
		return &m.UpdatedAt, nil
	case 5: // DeletedAt
		return &m.DeletedAt, nil
	}
	return nil, errors.Wrapf(mapping.ErrInvalidModelField, "provided invalid field: '%s' for given model: Model'", field.Name())
}

// GetFieldZeroValue implements mapping.Fielder interface.s
func (m *Model) GetFieldZeroValue(field *mapping.StructField) (interface{}, error) {
	switch field.Index[0] {
	case 0: // ID
		return 0, nil
	case 1: // Attr
		return "", nil
	case 2: // SnakeCased
		return "", nil
	case 3: // CreatedAt
		return time.Time{}, nil
	case 4: // UpdatedAt
		return time.Time{}, nil
	case 5: // DeletedAt
		return nil, nil
	default:
		return nil, errors.Wrapf(mapping.ErrInvalidModelField, "provided invalid field name: '%s'", field.Name())
	}
}

// IsFieldZero implements mapping.Fielder interface.
func (m *Model) IsFieldZero(field *mapping.StructField) (bool, error) {
	switch field.Index[0] {
	case 0: // ID
		return m.ID == 0, nil
	case 1: // Attr
		return m.Attr == "", nil
	case 2: // SnakeCased
		return m.SnakeCased == "", nil
	case 3: // CreatedAt
		return m.CreatedAt == time.Time{}, nil
	case 4: // UpdatedAt
		return m.UpdatedAt == time.Time{}, nil
	case 5: // DeletedAt
		return m.DeletedAt == nil, nil
	}
	return false, errors.Wrapf(mapping.ErrInvalidModelField, "provided invalid field name: '%s'", field.Name())
}

// SetFieldZeroValue implements mapping.Fielder interface.s
func (m *Model) SetFieldZeroValue(field *mapping.StructField) error {
	switch field.Index[0] {
	case 0: // ID
		m.ID = 0
	case 1: // Attr
		m.Attr = ""
	case 2: // SnakeCased
		m.SnakeCased = ""
	case 3: // CreatedAt
		m.CreatedAt = time.Time{}
	case 4: // UpdatedAt
		m.UpdatedAt = time.Time{}
	case 5: // DeletedAt
		m.DeletedAt = nil
	default:
		return errors.Wrapf(mapping.ErrInvalidModelField, "provided invalid field name: '%s'", field.Name())
	}
	return nil
}

// GetHashableFieldValue implements mapping.Fielder interface.
func (m *Model) GetHashableFieldValue(field *mapping.StructField) (interface{}, error) {
	switch field.Index[0] {
	case 0: // ID
		return m.ID, nil
	case 1: // Attr
		return m.Attr, nil
	case 2: // SnakeCased
		return m.SnakeCased, nil
	case 3: // CreatedAt
		return m.CreatedAt, nil
	case 4: // UpdatedAt
		return m.UpdatedAt, nil
	case 5: // DeletedAt
		if m.DeletedAt == nil {
			return nil, nil
		}
		return *m.DeletedAt, nil
	}
	return nil, errors.Wrapf(mapping.ErrInvalidModelField, "provided invalid field: '%s' for given model: 'Model'", field.Name())
}

// GetFieldValue implements mapping.Fielder interface.
func (m *Model) GetFieldValue(field *mapping.StructField) (interface{}, error) {
	switch field.Index[0] {
	case 0: // ID
		return m.ID, nil
	case 1: // Attr
		return m.Attr, nil
	case 2: // SnakeCased
		return m.SnakeCased, nil
	case 3: // CreatedAt
		return m.CreatedAt, nil
	case 4: // UpdatedAt
		return m.UpdatedAt, nil
	case 5: // DeletedAt
		return m.DeletedAt, nil
	}
	return nil, errors.Wrapf(mapping.ErrInvalidModelField, "provided invalid field: '%s' for given model: Model'", field.Name())
}

// SetFieldValue implements mapping.Fielder interface.
func (m *Model) SetFieldValue(field *mapping.StructField, value interface{}) (err error) {
	switch field.Index[0] {
	case 0: // ID
		if _v, ok := value.(int); ok {
			m.ID = _v
			return nil
		}
		if field.DatabaseNotNull() && value == nil {
			m.ID = 0
			return nil
		}

		switch _v := value.(type) {
		case int8:
			m.ID = int(_v)
		case int16:
			m.ID = int(_v)
		case int32:
			m.ID = int(_v)
		case int64:
			m.ID = int(_v)
		case uint:
			m.ID = int(_v)
		case uint8:
			m.ID = int(_v)
		case uint16:
			m.ID = int(_v)
		case uint32:
			m.ID = int(_v)
		case uint64:
			m.ID = int(_v)
		case float32:
			m.ID = int(_v)
		case float64:
			m.ID = int(_v)
		default:
			return errors.Wrapf(mapping.ErrFieldValue, "provided invalid field type: '%T' for the field: %s", value, field.Name())
		}
		return nil
	case 1: // Attr
		if _v, ok := value.(string); ok {
			m.Attr = _v
			return nil
		}
		if field.DatabaseNotNull() && value == nil {
			m.Attr = ""
			return nil
		}

		// Check alternate types for the Attr.
		if _v, ok := value.([]byte); ok {
			m.Attr = string(_v)
			return nil
		}
		return errors.Wrapf(mapping.ErrFieldValue, "provided invalid field type: '%T' for the field: %s", value, field.Name())
	case 2: // SnakeCased
		if _v, ok := value.(string); ok {
			m.SnakeCased = _v
			return nil
		}
		if field.DatabaseNotNull() && value == nil {
			m.SnakeCased = ""
			return nil
		}

		// Check alternate types for the SnakeCased.
		if _v, ok := value.([]byte); ok {
			m.SnakeCased = string(_v)
			return nil
		}
		return errors.Wrapf(mapping.ErrFieldValue, "provided invalid field type: '%T' for the field: %s", value, field.Name())
	case 3: // CreatedAt
		if _v, ok := value.(time.Time); ok {
			m.CreatedAt = _v
			return nil
		}
		if field.DatabaseNotNull() && value == nil {
			m.CreatedAt = time.Time{}
			return nil
		}

		return errors.Wrapf(mapping.ErrFieldValue, "provided invalid field type: '%T' for the field: %s", value, field.Name())
	case 4: // UpdatedAt
		if _v, ok := value.(time.Time); ok {
			m.UpdatedAt = _v
			return nil
		}
		if field.DatabaseNotNull() && value == nil {
			m.UpdatedAt = time.Time{}
			return nil
		}

		return errors.Wrapf(mapping.ErrFieldValue, "provided invalid field type: '%T' for the field: %s", value, field.Name())
	case 5: // DeletedAt
		if value == nil {
			m.DeletedAt = nil
			return nil
		}
		if _v, ok := value.(*time.Time); ok {
			m.DeletedAt = _v
			return nil
		}
		// Check if it is non-pointer value.
		if _v, ok := value.(time.Time); ok {
			m.DeletedAt = &_v
			return nil
		}

		return errors.Wrapf(mapping.ErrFieldValue, "provided invalid field type: '%T' for the field: %s", value, field.Name())
	default:
		return errors.Wrapf(mapping.ErrInvalidModelField, "provided invalid field: '%s' for the model: 'Model'", field.Name())
	}
}

// SetPrimaryKeyStringValue implements mapping.Model interface method.
func (m *Model) ParseFieldsStringValue(field *mapping.StructField, value string) (interface{}, error) {
	switch field.Index[0] {
	case 0: // ID
		return strconv.ParseInt(value, 10, mapping.IntegerBitSize)
	case 1: // Attr
		return value, nil
	case 2: // SnakeCased
		return value, nil
	case 3: // CreatedAt
		temp := m.CreatedAt
		if err := m.CreatedAt.UnmarshalText([]byte(value)); err != nil {
			return "", errors.Wrapf(mapping.ErrFieldValue, "invalid field 'CreatedAt' value: '%v' to parse string. Err: %v", m.CreatedAt, err)
		}
		bt, err := m.CreatedAt.MarshalText()
		if err != nil {
			return "", errors.Wrapf(mapping.ErrFieldValue, "invalid field 'CreatedAt' value: '%v' to parse string. Err: %v", m.CreatedAt, err)
		}
		m.CreatedAt = temp
		return string(bt), nil
	case 4: // UpdatedAt
		temp := m.UpdatedAt
		if err := m.UpdatedAt.UnmarshalText([]byte(value)); err != nil {
			return "", errors.Wrapf(mapping.ErrFieldValue, "invalid field 'UpdatedAt' value: '%v' to parse string. Err: %v", m.UpdatedAt, err)
		}
		bt, err := m.UpdatedAt.MarshalText()
		if err != nil {
			return "", errors.Wrapf(mapping.ErrFieldValue, "invalid field 'UpdatedAt' value: '%v' to parse string. Err: %v", m.UpdatedAt, err)
		}
		m.UpdatedAt = temp
		return string(bt), nil
	case 5: // DeletedAt
		var base time.Time
		temp := &base
		if err := temp.UnmarshalText([]byte(value)); err != nil {
			return "", errors.Wrapf(mapping.ErrFieldValue, "invalid field 'DeletedAt' value: '%v' to parse string. Err: %v", m.DeletedAt, err)
		}
		bt, err := temp.MarshalText()
		if err != nil {
			return "", errors.Wrapf(mapping.ErrFieldValue, "invalid field 'DeletedAt' value: '%v' to parse string. Err: %v", m.DeletedAt, err)
		}

		return string(bt), nil
	}
	return nil, errors.Wrapf(mapping.ErrInvalidModelField, "provided invalid field: '%s' for given model: Model'", field.Name())
}
