// Code generated by neuron/generator. DO NOT EDIT.
// This file was generated at:
// Mon, 27 Jul 2020 13:30:31 +0200

package tests

import (
	"strconv"
	"time"

	"github.com/neuronlabs/neuron/errors"
	"github.com/neuronlabs/neuron/mapping"
)

// Neuron_Models stores all generated models in this package.
var Neuron_Models = []mapping.Model{
	&SimpleModel{},
	&OmitModel{},
	&Model{},
}

// Compile time check if SimpleModel implements mapping.Model interface.
var _ mapping.Model = &SimpleModel{}

// NeuronCollectionName implements mapping.Model interface method.
// Returns the name of the collection for the 'SimpleModel'.
func (s *SimpleModel) NeuronCollectionName() string {
	return "simple_models"
}

// IsPrimaryKeyZero implements mapping.Model interface method.
func (s *SimpleModel) IsPrimaryKeyZero() bool {
	return s.ID == 0
}

// GetPrimaryKeyValue implements mapping.Model interface method.
func (s *SimpleModel) GetPrimaryKeyValue() interface{} {
	return s.ID
}

// GetPrimaryKeyStringValue implements mapping.Model interface method.
func (s *SimpleModel) GetPrimaryKeyStringValue() (string, error) {
	return strconv.FormatInt(int64(s.ID), 10), nil
}

// GetPrimaryKeyAddress implements mapping.Model interface method.
func (s *SimpleModel) GetPrimaryKeyAddress() interface{} {
	return &s.ID
}

// GetPrimaryKeyHashableValue implements mapping.Model interface method.
func (s *SimpleModel) GetPrimaryKeyHashableValue() interface{} {
	return s.ID
}

// GetPrimaryKeyZeroValue implements mapping.Model interface method.
func (s *SimpleModel) GetPrimaryKeyZeroValue() interface{} {
	return 0
}

// SetPrimaryKey implements mapping.Model interface method.
func (s *SimpleModel) SetPrimaryKeyValue(value interface{}) error {
	if v, ok := value.(int); ok {
		s.ID = v
		return nil
	}
	// Check alternate types for given field.
	switch valueType := value.(type) {
	case int8:
		s.ID = int(valueType)
	case int16:
		s.ID = int(valueType)
	case int32:
		s.ID = int(valueType)
	case int64:
		s.ID = int(valueType)
	case uint:
		s.ID = int(valueType)
	case uint8:
		s.ID = int(valueType)
	case uint16:
		s.ID = int(valueType)
	case uint32:
		s.ID = int(valueType)
	case uint64:
		s.ID = int(valueType)
	case float32:
		s.ID = int(valueType)
	case float64:
		s.ID = int(valueType)
	default:
		return errors.Newf(mapping.ClassFieldValue, "provided invalid value: '%T' for the primary field for model: 'SimpleModel'", value)
	}
	return nil
}

// SetPrimaryKeyStringValue implements mapping.Model interface method.
func (s *SimpleModel) SetPrimaryKeyStringValue(value string) error {
	tmp, err := strconv.ParseInt(value, 10, mapping.IntegerBitSize)
	if err != nil {
		return err
	}
	s.ID = int(tmp)
	return nil
}

// Compile time check if SimpleModel implements mapping.Fielder interface.
var _ mapping.Fielder = &SimpleModel{}

// GetFieldsAddress gets the address of provided 'field'.
func (s *SimpleModel) GetFieldsAddress(field *mapping.StructField) (interface{}, error) {
	switch field.Index[0] {
	case 0: // ID
		return &s.ID, nil
	case 1: // Attr
		return &s.Attr, nil
	case 2: // CreatedAt
		return &s.CreatedAt, nil
	}
	return nil, errors.Newf(mapping.ClassInvalidModelField, "provided invalid field: '%s' for given model: SimpleModel'", field.Name())
}

// GetFieldZeroValue implements mapping.Fielder interface.s
func (s *SimpleModel) GetFieldZeroValue(field *mapping.StructField) (interface{}, error) {
	switch field.Index[0] {
	case 0: // ID
		return 0, nil
	case 1: // Attr
		return "", nil
	case 2: // CreatedAt
		return nil, nil
	default:
		return nil, errors.Newf(mapping.ClassInvalidModelField, "provided invalid field name: '%s'", field.Name())
	}
}

// IsFieldZero implements mapping.Fielder interface.
func (s *SimpleModel) IsFieldZero(field *mapping.StructField) (bool, error) {
	switch field.Index[0] {
	case 0: // ID
		return s.ID == 0, nil
	case 1: // Attr
		return s.Attr == "", nil
	case 2: // CreatedAt
		return s.CreatedAt == nil, nil
	}
	return false, errors.Newf(mapping.ClassInvalidModelField, "provided invalid field name: '%s'", field.Name())
}

// SetFieldZeroValue implements mapping.Fielder interface.s
func (s *SimpleModel) SetFieldZeroValue(field *mapping.StructField) error {
	switch field.Index[0] {
	case 0: // ID
		s.ID = 0
	case 1: // Attr
		s.Attr = ""
	case 2: // CreatedAt
		s.CreatedAt = nil
	default:
		return errors.Newf(mapping.ClassInvalidModelField, "provided invalid field name: '%s'", field.Name())
	}
	return nil
}

// GetHashableFieldValue implements mapping.Fielder interface.
func (s *SimpleModel) GetHashableFieldValue(field *mapping.StructField) (interface{}, error) {
	switch field.Index[0] {
	case 0: // ID
		return s.ID, nil
	case 1: // Attr
		return s.Attr, nil
	case 2: // CreatedAt
		if s.CreatedAt == nil {
			return nil, nil
		}
		return *s.CreatedAt, nil
	}
	return nil, errors.Newf(mapping.ClassInvalidModelField, "provided invalid field: '%s' for given model: 'SimpleModel'", field.Name())
}

// GetFieldValue implements mapping.Fielder interface.
func (s *SimpleModel) GetFieldValue(field *mapping.StructField) (interface{}, error) {
	switch field.Index[0] {
	case 0: // ID
		return s.ID, nil
	case 1: // Attr
		return s.Attr, nil
	case 2: // CreatedAt
		return s.CreatedAt, nil
	}
	return nil, errors.Newf(mapping.ClassInvalidModelField, "provided invalid field: '%s' for given model: SimpleModel'", field.Name())
}

// SetFieldValue implements mapping.Fielder interface.
func (s *SimpleModel) SetFieldValue(field *mapping.StructField, value interface{}) (err error) {
	switch field.Index[0] {
	case 0: // ID
		if v, ok := value.(int); ok {
			s.ID = v
			return nil
		}

		switch v := value.(type) {
		case int8:
			s.ID = int(v)
		case int16:
			s.ID = int(v)
		case int32:
			s.ID = int(v)
		case int64:
			s.ID = int(v)
		case uint:
			s.ID = int(v)
		case uint8:
			s.ID = int(v)
		case uint16:
			s.ID = int(v)
		case uint32:
			s.ID = int(v)
		case uint64:
			s.ID = int(v)
		case float32:
			s.ID = int(v)
		case float64:
			s.ID = int(v)
		default:
			return errors.Newf(mapping.ClassFieldValue, "provided invalid field type: '%T' for the field: %s", value, field.Name())
		}
		return nil
	case 1: // Attr
		if v, ok := value.(string); ok {
			s.Attr = v
			return nil
		}

		// Check alternate types for the Attr.
		if v, ok := value.([]byte); ok {
			s.Attr = string(v)
			return nil
		}
		return errors.Newf(mapping.ClassFieldValue, "provided invalid field type: '%T' for the field: %s", value, field.Name())
	case 2: // CreatedAt
		if value == nil {
			s.CreatedAt = nil
			return nil
		}
		if v, ok := value.(*time.Time); ok {
			s.CreatedAt = v
			return nil
		}
		// Check if it is non-pointer value.
		if v, ok := value.(time.Time); ok {
			s.CreatedAt = &v
			return nil
		}

		return errors.Newf(mapping.ClassFieldValue, "provided invalid field type: '%T' for the field: %s", value, field.Name())
	default:
		return errors.Newf(mapping.ClassInvalidModelField, "provided invalid field: '%s' for the model: 'SimpleModel'", field.Name())
	}
}

// SetPrimaryKeyStringValue implements mapping.Model interface method.
func (s *SimpleModel) ParseFieldsStringValue(field *mapping.StructField, value string) (interface{}, error) {
	switch field.Index[0] {
	case 0: // ID
		return strconv.ParseInt(value, 10, mapping.IntegerBitSize)
	case 1: // Attr
		return value, nil
	case 2: // CreatedAt
		var base time.Time
		temp := &base
		if err := temp.UnmarshalText([]byte(value)); err != nil {
			return "", errors.Newf(mapping.ClassFieldValue, "invalid field 'CreatedAt' value: '%v' to parse string. Err: %v", s.CreatedAt, err)
		}
		bt, err := temp.MarshalText()
		if err != nil {
			return "", errors.Newf(mapping.ClassFieldValue, "invalid field 'CreatedAt' value: '%v' to parse string. Err: %v", s.CreatedAt, err)
		}

		return string(bt), nil
	}
	return nil, errors.Newf(mapping.ClassInvalidModelField, "provided invalid field: '%s' for given model: SimpleModel'", field.Name())
}

// Compile time check if OmitModel implements mapping.Model interface.
var _ mapping.Model = &OmitModel{}

// NeuronCollectionName implements mapping.Model interface method.
// Returns the name of the collection for the 'OmitModel'.
func (o *OmitModel) NeuronCollectionName() string {
	return "omit_models"
}

// IsPrimaryKeyZero implements mapping.Model interface method.
func (o *OmitModel) IsPrimaryKeyZero() bool {
	return o.ID == 0
}

// GetPrimaryKeyValue implements mapping.Model interface method.
func (o *OmitModel) GetPrimaryKeyValue() interface{} {
	return o.ID
}

// GetPrimaryKeyStringValue implements mapping.Model interface method.
func (o *OmitModel) GetPrimaryKeyStringValue() (string, error) {
	return strconv.FormatInt(int64(o.ID), 10), nil
}

// GetPrimaryKeyAddress implements mapping.Model interface method.
func (o *OmitModel) GetPrimaryKeyAddress() interface{} {
	return &o.ID
}

// GetPrimaryKeyHashableValue implements mapping.Model interface method.
func (o *OmitModel) GetPrimaryKeyHashableValue() interface{} {
	return o.ID
}

// GetPrimaryKeyZeroValue implements mapping.Model interface method.
func (o *OmitModel) GetPrimaryKeyZeroValue() interface{} {
	return 0
}

// SetPrimaryKey implements mapping.Model interface method.
func (o *OmitModel) SetPrimaryKeyValue(value interface{}) error {
	if v, ok := value.(int); ok {
		o.ID = v
		return nil
	}
	// Check alternate types for given field.
	switch valueType := value.(type) {
	case int8:
		o.ID = int(valueType)
	case int16:
		o.ID = int(valueType)
	case int32:
		o.ID = int(valueType)
	case int64:
		o.ID = int(valueType)
	case uint:
		o.ID = int(valueType)
	case uint8:
		o.ID = int(valueType)
	case uint16:
		o.ID = int(valueType)
	case uint32:
		o.ID = int(valueType)
	case uint64:
		o.ID = int(valueType)
	case float32:
		o.ID = int(valueType)
	case float64:
		o.ID = int(valueType)
	default:
		return errors.Newf(mapping.ClassFieldValue, "provided invalid value: '%T' for the primary field for model: 'OmitModel'", value)
	}
	return nil
}

// SetPrimaryKeyStringValue implements mapping.Model interface method.
func (o *OmitModel) SetPrimaryKeyStringValue(value string) error {
	tmp, err := strconv.ParseInt(value, 10, mapping.IntegerBitSize)
	if err != nil {
		return err
	}
	o.ID = int(tmp)
	return nil
}

// Compile time check if OmitModel implements mapping.Fielder interface.
var _ mapping.Fielder = &OmitModel{}

// GetFieldsAddress gets the address of provided 'field'.
func (o *OmitModel) GetFieldsAddress(field *mapping.StructField) (interface{}, error) {
	switch field.Index[0] {
	case 0: // ID
		return &o.ID, nil
	case 1: // OmitField
		return &o.OmitField, nil
	}
	return nil, errors.Newf(mapping.ClassInvalidModelField, "provided invalid field: '%s' for given model: OmitModel'", field.Name())
}

// GetFieldZeroValue implements mapping.Fielder interface.s
func (o *OmitModel) GetFieldZeroValue(field *mapping.StructField) (interface{}, error) {
	switch field.Index[0] {
	case 0: // ID
		return 0, nil
	case 1: // OmitField
		return "", nil
	default:
		return nil, errors.Newf(mapping.ClassInvalidModelField, "provided invalid field name: '%s'", field.Name())
	}
}

// IsFieldZero implements mapping.Fielder interface.
func (o *OmitModel) IsFieldZero(field *mapping.StructField) (bool, error) {
	switch field.Index[0] {
	case 0: // ID
		return o.ID == 0, nil
	case 1: // OmitField
		return o.OmitField == "", nil
	}
	return false, errors.Newf(mapping.ClassInvalidModelField, "provided invalid field name: '%s'", field.Name())
}

// SetFieldZeroValue implements mapping.Fielder interface.s
func (o *OmitModel) SetFieldZeroValue(field *mapping.StructField) error {
	switch field.Index[0] {
	case 0: // ID
		o.ID = 0
	case 1: // OmitField
		o.OmitField = ""
	default:
		return errors.Newf(mapping.ClassInvalidModelField, "provided invalid field name: '%s'", field.Name())
	}
	return nil
}

// GetHashableFieldValue implements mapping.Fielder interface.
func (o *OmitModel) GetHashableFieldValue(field *mapping.StructField) (interface{}, error) {
	switch field.Index[0] {
	case 0: // ID
		return o.ID, nil
	case 1: // OmitField
		return o.OmitField, nil
	}
	return nil, errors.Newf(mapping.ClassInvalidModelField, "provided invalid field: '%s' for given model: 'OmitModel'", field.Name())
}

// GetFieldValue implements mapping.Fielder interface.
func (o *OmitModel) GetFieldValue(field *mapping.StructField) (interface{}, error) {
	switch field.Index[0] {
	case 0: // ID
		return o.ID, nil
	case 1: // OmitField
		return o.OmitField, nil
	}
	return nil, errors.Newf(mapping.ClassInvalidModelField, "provided invalid field: '%s' for given model: OmitModel'", field.Name())
}

// SetFieldValue implements mapping.Fielder interface.
func (o *OmitModel) SetFieldValue(field *mapping.StructField, value interface{}) (err error) {
	switch field.Index[0] {
	case 0: // ID
		if v, ok := value.(int); ok {
			o.ID = v
			return nil
		}

		switch v := value.(type) {
		case int8:
			o.ID = int(v)
		case int16:
			o.ID = int(v)
		case int32:
			o.ID = int(v)
		case int64:
			o.ID = int(v)
		case uint:
			o.ID = int(v)
		case uint8:
			o.ID = int(v)
		case uint16:
			o.ID = int(v)
		case uint32:
			o.ID = int(v)
		case uint64:
			o.ID = int(v)
		case float32:
			o.ID = int(v)
		case float64:
			o.ID = int(v)
		default:
			return errors.Newf(mapping.ClassFieldValue, "provided invalid field type: '%T' for the field: %s", value, field.Name())
		}
		return nil
	case 1: // OmitField
		if v, ok := value.(string); ok {
			o.OmitField = v
			return nil
		}

		// Check alternate types for the OmitField.
		if v, ok := value.([]byte); ok {
			o.OmitField = string(v)
			return nil
		}
		return errors.Newf(mapping.ClassFieldValue, "provided invalid field type: '%T' for the field: %s", value, field.Name())
	default:
		return errors.Newf(mapping.ClassInvalidModelField, "provided invalid field: '%s' for the model: 'OmitModel'", field.Name())
	}
}

// SetPrimaryKeyStringValue implements mapping.Model interface method.
func (o *OmitModel) ParseFieldsStringValue(field *mapping.StructField, value string) (interface{}, error) {
	switch field.Index[0] {
	case 0: // ID
		return strconv.ParseInt(value, 10, mapping.IntegerBitSize)
	case 1: // OmitField
		return value, nil
	}
	return nil, errors.Newf(mapping.ClassInvalidModelField, "provided invalid field: '%s' for given model: OmitModel'", field.Name())
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
	if v, ok := value.(int); ok {
		m.ID = v
		return nil
	}
	// Check alternate types for given field.
	switch valueType := value.(type) {
	case int8:
		m.ID = int(valueType)
	case int16:
		m.ID = int(valueType)
	case int32:
		m.ID = int(valueType)
	case int64:
		m.ID = int(valueType)
	case uint:
		m.ID = int(valueType)
	case uint8:
		m.ID = int(valueType)
	case uint16:
		m.ID = int(valueType)
	case uint32:
		m.ID = int(valueType)
	case uint64:
		m.ID = int(valueType)
	case float32:
		m.ID = int(valueType)
	case float64:
		m.ID = int(valueType)
	default:
		return errors.Newf(mapping.ClassFieldValue, "provided invalid value: '%T' for the primary field for model: 'Model'", value)
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

// Compile time check if Model implements mapping.Fielder interface.
var _ mapping.Fielder = &Model{}

// GetFieldsAddress gets the address of provided 'field'.
func (m *Model) GetFieldsAddress(field *mapping.StructField) (interface{}, error) {
	switch field.Index[0] {
	case 0: // ID
		return &m.ID, nil
	case 1: // AttrString
		return &m.AttrString, nil
	case 2: // StringPtr
		return &m.StringPtr, nil
	case 3: // Int
		return &m.Int, nil
	case 4: // CreatedAt
		return &m.CreatedAt, nil
	case 5: // UpdatedAt
		return &m.UpdatedAt, nil
	case 6: // DeletedAt
		return &m.DeletedAt, nil
	}
	return nil, errors.Newf(mapping.ClassInvalidModelField, "provided invalid field: '%s' for given model: Model'", field.Name())
}

// GetFieldZeroValue implements mapping.Fielder interface.s
func (m *Model) GetFieldZeroValue(field *mapping.StructField) (interface{}, error) {
	switch field.Index[0] {
	case 0: // ID
		return 0, nil
	case 1: // AttrString
		return "", nil
	case 2: // StringPtr
		return nil, nil
	case 3: // Int
		return 0, nil
	case 4: // CreatedAt
		return time.Time{}, nil
	case 5: // UpdatedAt
		return nil, nil
	case 6: // DeletedAt
		return nil, nil
	default:
		return nil, errors.Newf(mapping.ClassInvalidModelField, "provided invalid field name: '%s'", field.Name())
	}
}

// IsFieldZero implements mapping.Fielder interface.
func (m *Model) IsFieldZero(field *mapping.StructField) (bool, error) {
	switch field.Index[0] {
	case 0: // ID
		return m.ID == 0, nil
	case 1: // AttrString
		return m.AttrString == "", nil
	case 2: // StringPtr
		return m.StringPtr == nil, nil
	case 3: // Int
		return m.Int == 0, nil
	case 4: // CreatedAt
		return m.CreatedAt == time.Time{}, nil
	case 5: // UpdatedAt
		return m.UpdatedAt == nil, nil
	case 6: // DeletedAt
		return m.DeletedAt == nil, nil
	}
	return false, errors.Newf(mapping.ClassInvalidModelField, "provided invalid field name: '%s'", field.Name())
}

// SetFieldZeroValue implements mapping.Fielder interface.s
func (m *Model) SetFieldZeroValue(field *mapping.StructField) error {
	switch field.Index[0] {
	case 0: // ID
		m.ID = 0
	case 1: // AttrString
		m.AttrString = ""
	case 2: // StringPtr
		m.StringPtr = nil
	case 3: // Int
		m.Int = 0
	case 4: // CreatedAt
		m.CreatedAt = time.Time{}
	case 5: // UpdatedAt
		m.UpdatedAt = nil
	case 6: // DeletedAt
		m.DeletedAt = nil
	default:
		return errors.Newf(mapping.ClassInvalidModelField, "provided invalid field name: '%s'", field.Name())
	}
	return nil
}

// GetHashableFieldValue implements mapping.Fielder interface.
func (m *Model) GetHashableFieldValue(field *mapping.StructField) (interface{}, error) {
	switch field.Index[0] {
	case 0: // ID
		return m.ID, nil
	case 1: // AttrString
		return m.AttrString, nil
	case 2: // StringPtr
		if m.StringPtr == nil {
			return nil, nil
		}
		return *m.StringPtr, nil
	case 3: // Int
		return m.Int, nil
	case 4: // CreatedAt
		return m.CreatedAt, nil
	case 5: // UpdatedAt
		if m.UpdatedAt == nil {
			return nil, nil
		}
		return *m.UpdatedAt, nil
	case 6: // DeletedAt
		if m.DeletedAt == nil {
			return nil, nil
		}
		return *m.DeletedAt, nil
	}
	return nil, errors.Newf(mapping.ClassInvalidModelField, "provided invalid field: '%s' for given model: 'Model'", field.Name())
}

// GetFieldValue implements mapping.Fielder interface.
func (m *Model) GetFieldValue(field *mapping.StructField) (interface{}, error) {
	switch field.Index[0] {
	case 0: // ID
		return m.ID, nil
	case 1: // AttrString
		return m.AttrString, nil
	case 2: // StringPtr
		return m.StringPtr, nil
	case 3: // Int
		return m.Int, nil
	case 4: // CreatedAt
		return m.CreatedAt, nil
	case 5: // UpdatedAt
		return m.UpdatedAt, nil
	case 6: // DeletedAt
		return m.DeletedAt, nil
	}
	return nil, errors.Newf(mapping.ClassInvalidModelField, "provided invalid field: '%s' for given model: Model'", field.Name())
}

// SetFieldValue implements mapping.Fielder interface.
func (m *Model) SetFieldValue(field *mapping.StructField, value interface{}) (err error) {
	switch field.Index[0] {
	case 0: // ID
		if v, ok := value.(int); ok {
			m.ID = v
			return nil
		}

		switch v := value.(type) {
		case int8:
			m.ID = int(v)
		case int16:
			m.ID = int(v)
		case int32:
			m.ID = int(v)
		case int64:
			m.ID = int(v)
		case uint:
			m.ID = int(v)
		case uint8:
			m.ID = int(v)
		case uint16:
			m.ID = int(v)
		case uint32:
			m.ID = int(v)
		case uint64:
			m.ID = int(v)
		case float32:
			m.ID = int(v)
		case float64:
			m.ID = int(v)
		default:
			return errors.Newf(mapping.ClassFieldValue, "provided invalid field type: '%T' for the field: %s", value, field.Name())
		}
		return nil
	case 1: // AttrString
		if v, ok := value.(string); ok {
			m.AttrString = v
			return nil
		}

		// Check alternate types for the AttrString.
		if v, ok := value.([]byte); ok {
			m.AttrString = string(v)
			return nil
		}
		return errors.Newf(mapping.ClassFieldValue, "provided invalid field type: '%T' for the field: %s", value, field.Name())
	case 2: // StringPtr
		if value == nil {
			m.StringPtr = nil
			return nil
		}
		if v, ok := value.(*string); ok {
			m.StringPtr = v
			return nil
		}
		// Check if it is non-pointer value.
		if v, ok := value.(string); ok {
			m.StringPtr = &v
			return nil
		}

		// Check alternate types for the StringPtr.
		if v, ok := value.([]byte); ok {
			temp := string(v)
			m.StringPtr = &temp
			return nil
		}
		return errors.Newf(mapping.ClassFieldValue, "provided invalid field type: '%T' for the field: %s", value, field.Name())
	case 3: // Int
		if v, ok := value.(int); ok {
			m.Int = v
			return nil
		}

		switch v := value.(type) {
		case int8:
			m.Int = int(v)
		case int16:
			m.Int = int(v)
		case int32:
			m.Int = int(v)
		case int64:
			m.Int = int(v)
		case uint:
			m.Int = int(v)
		case uint8:
			m.Int = int(v)
		case uint16:
			m.Int = int(v)
		case uint32:
			m.Int = int(v)
		case uint64:
			m.Int = int(v)
		case float32:
			m.Int = int(v)
		case float64:
			m.Int = int(v)
		default:
			return errors.Newf(mapping.ClassFieldValue, "provided invalid field type: '%T' for the field: %s", value, field.Name())
		}
		return nil
	case 4: // CreatedAt
		if v, ok := value.(time.Time); ok {
			m.CreatedAt = v
			return nil
		}

		return errors.Newf(mapping.ClassFieldValue, "provided invalid field type: '%T' for the field: %s", value, field.Name())
	case 5: // UpdatedAt
		if value == nil {
			m.UpdatedAt = nil
			return nil
		}
		if v, ok := value.(*time.Time); ok {
			m.UpdatedAt = v
			return nil
		}
		// Check if it is non-pointer value.
		if v, ok := value.(time.Time); ok {
			m.UpdatedAt = &v
			return nil
		}

		return errors.Newf(mapping.ClassFieldValue, "provided invalid field type: '%T' for the field: %s", value, field.Name())
	case 6: // DeletedAt
		if value == nil {
			m.DeletedAt = nil
			return nil
		}
		if v, ok := value.(*time.Time); ok {
			m.DeletedAt = v
			return nil
		}
		// Check if it is non-pointer value.
		if v, ok := value.(time.Time); ok {
			m.DeletedAt = &v
			return nil
		}

		return errors.Newf(mapping.ClassFieldValue, "provided invalid field type: '%T' for the field: %s", value, field.Name())
	default:
		return errors.Newf(mapping.ClassInvalidModelField, "provided invalid field: '%s' for the model: 'Model'", field.Name())
	}
}

// SetPrimaryKeyStringValue implements mapping.Model interface method.
func (m *Model) ParseFieldsStringValue(field *mapping.StructField, value string) (interface{}, error) {
	switch field.Index[0] {
	case 0: // ID
		return strconv.ParseInt(value, 10, mapping.IntegerBitSize)
	case 1: // AttrString
		return value, nil
	case 2: // StringPtr
		return value, nil
	case 3: // Int
		return strconv.ParseInt(value, 10, mapping.IntegerBitSize)
	case 4: // CreatedAt
		temp := m.CreatedAt
		if err := m.CreatedAt.UnmarshalText([]byte(value)); err != nil {
			return "", errors.Newf(mapping.ClassFieldValue, "invalid field 'CreatedAt' value: '%v' to parse string. Err: %v", m.CreatedAt, err)
		}
		bt, err := m.CreatedAt.MarshalText()
		if err != nil {
			return "", errors.Newf(mapping.ClassFieldValue, "invalid field 'CreatedAt' value: '%v' to parse string. Err: %v", m.CreatedAt, err)
		}
		m.CreatedAt = temp
		return string(bt), nil
	case 5: // UpdatedAt
		var base time.Time
		temp := &base
		if err := temp.UnmarshalText([]byte(value)); err != nil {
			return "", errors.Newf(mapping.ClassFieldValue, "invalid field 'UpdatedAt' value: '%v' to parse string. Err: %v", m.UpdatedAt, err)
		}
		bt, err := temp.MarshalText()
		if err != nil {
			return "", errors.Newf(mapping.ClassFieldValue, "invalid field 'UpdatedAt' value: '%v' to parse string. Err: %v", m.UpdatedAt, err)
		}

		return string(bt), nil
	case 6: // DeletedAt
		var base time.Time
		temp := &base
		if err := temp.UnmarshalText([]byte(value)); err != nil {
			return "", errors.Newf(mapping.ClassFieldValue, "invalid field 'DeletedAt' value: '%v' to parse string. Err: %v", m.DeletedAt, err)
		}
		bt, err := temp.MarshalText()
		if err != nil {
			return "", errors.Newf(mapping.ClassFieldValue, "invalid field 'DeletedAt' value: '%v' to parse string. Err: %v", m.DeletedAt, err)
		}

		return string(bt), nil
	}
	return nil, errors.Newf(mapping.ClassInvalidModelField, "provided invalid field: '%s' for given model: Model'", field.Name())
}
