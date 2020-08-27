// Code generated by neuron/generator. DO NOT EDIT.
// This file was generated at:
// Fri, 28 Aug 2020 01:12:27 +0200

package tests

import (
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/neuronlabs/neuron-extensions/neurogonesis/internal/tests/external"
	"github.com/neuronlabs/neuron/errors"
	"github.com/neuronlabs/neuron/mapping"
	"github.com/neuronlabs/neuron/query"
)

// Neuron_Models stores all generated models in this package.
var Neuron_Models = []mapping.Model{
	&Car{},
	&User{},
}

// Compile time check if Car implements mapping.Model interface.
var _ mapping.Model = &Car{}

// NeuronCollectionName implements mapping.Model interface method.
// Returns the name of the collection for the 'Car'.
func (c *Car) NeuronCollectionName() string {
	return "cars"
}

// IsPrimaryKeyZero implements mapping.Model interface method.
func (c *Car) IsPrimaryKeyZero() bool {
	return c.ID == ""
}

// GetPrimaryKeyValue implements mapping.Model interface method.
func (c *Car) GetPrimaryKeyValue() interface{} {
	return c.ID
}

// GetPrimaryKeyStringValue implements mapping.Model interface method.
func (c *Car) GetPrimaryKeyStringValue() (string, error) {
	return c.ID, nil
}

// GetPrimaryKeyAddress implements mapping.Model interface method.
func (c *Car) GetPrimaryKeyAddress() interface{} {
	return &c.ID
}

// GetPrimaryKeyHashableValue implements mapping.Model interface method.
func (c *Car) GetPrimaryKeyHashableValue() interface{} {
	return c.ID
}

// GetPrimaryKeyZeroValue implements mapping.Model interface method.
func (c *Car) GetPrimaryKeyZeroValue() interface{} {
	return ""
}

// SetPrimaryKey implements mapping.Model interface method.
func (c *Car) SetPrimaryKeyValue(value interface{}) error {
	if v, ok := value.(string); ok {
		c.ID = v
		return nil
	}
	// Check alternate types for given field.
	if v, ok := value.([]byte); ok {
		c.ID = string(v)
		return nil
	}
	return errors.Wrapf(mapping.ErrFieldValue, "provided invalid value: '%T' for the primary field for model: '%T'",
		value, c)
}

// SetPrimaryKeyStringValue implements mapping.Model interface method.
func (c *Car) SetPrimaryKeyStringValue(value string) error {
	c.ID = value
	return nil
}

// SetFrom implements FromSetter interface.
func (c *Car) SetFrom(model mapping.Model) error {
	if model == nil {
		return errors.Wrap(query.ErrInvalidInput, "provided nil model to set from")
	}
	from, ok := model.(*Car)
	if !ok {
		return errors.WrapDetf(mapping.ErrModelNotMatch, "provided model doesn't match the input: %T", model)
	}
	*c = *from
	return nil
}

// Compile time check if Car implements mapping.Fielder interface.
var _ mapping.Fielder = &Car{}

// GetFieldsAddress gets the address of provided 'field'.
func (c *Car) GetFieldsAddress(field *mapping.StructField) (interface{}, error) {
	switch field.Index[0] {
	case 0: // ID
		return &c.ID, nil
	case 1: // Plates
		return &c.Plates, nil
	}
	return nil, errors.Wrapf(mapping.ErrInvalidModelField, "provided invalid field: '%s' for given model: Car'", field.Name())
}

// GetFieldZeroValue implements mapping.Fielder interface.s
func (c *Car) GetFieldZeroValue(field *mapping.StructField) (interface{}, error) {
	switch field.Index[0] {
	case 0: // ID
		return "", nil
	case 1: // Plates
		return "", nil
	default:
		return nil, errors.Wrapf(mapping.ErrInvalidModelField, "provided invalid field name: '%s'", field.Name())
	}
}

// IsFieldZero implements mapping.Fielder interface.
func (c *Car) IsFieldZero(field *mapping.StructField) (bool, error) {
	switch field.Index[0] {
	case 0: // ID
		return c.ID == "", nil
	case 1: // Plates
		return c.Plates == "", nil
	}
	return false, errors.Wrapf(mapping.ErrInvalidModelField, "provided invalid field name: '%s'", field.Name())
}

// SetFieldZeroValue implements mapping.Fielder interface.s
func (c *Car) SetFieldZeroValue(field *mapping.StructField) error {
	switch field.Index[0] {
	case 0: // ID
		c.ID = ""
	case 1: // Plates
		c.Plates = ""
	default:
		return errors.Wrapf(mapping.ErrInvalidModelField, "provided invalid field name: '%s'", field.Name())
	}
	return nil
}

// GetHashableFieldValue implements mapping.Fielder interface.
func (c *Car) GetHashableFieldValue(field *mapping.StructField) (interface{}, error) {
	switch field.Index[0] {
	case 0: // ID
		return c.ID, nil
	case 1: // Plates
		return c.Plates, nil
	}
	return nil, errors.Wrapf(mapping.ErrInvalidModelField, "provided invalid field: '%s' for given model: 'Car'", field.Name())
}

// GetFieldValue implements mapping.Fielder interface.
func (c *Car) GetFieldValue(field *mapping.StructField) (interface{}, error) {
	switch field.Index[0] {
	case 0: // ID
		return c.ID, nil
	case 1: // Plates
		return c.Plates, nil
	}
	return nil, errors.Wrapf(mapping.ErrInvalidModelField, "provided invalid field: '%s' for given model: Car'", field.Name())
}

// SetFieldValue implements mapping.Fielder interface.
func (c *Car) SetFieldValue(field *mapping.StructField, value interface{}) (err error) {
	switch field.Index[0] {
	case 0: // ID
		if v, ok := value.(string); ok {
			c.ID = v
			return nil
		}
		if field.DatabaseNotNull() {
			isZero := c.ID == ""
			if isZero {
				c.ID = ""
				return nil
			}
		}

		// Check alternate types for the ID.
		if v, ok := value.([]byte); ok {
			c.ID = string(v)
			return nil
		}
		return errors.Wrapf(mapping.ErrFieldValue, "provided invalid field type: '%T' for the field: %s", value, field.Name())
	case 1: // Plates
		if v, ok := value.(string); ok {
			c.Plates = v
			return nil
		}
		if field.DatabaseNotNull() {
			isZero := c.Plates == ""
			if isZero {
				c.Plates = ""
				return nil
			}
		}

		// Check alternate types for the Plates.
		if v, ok := value.([]byte); ok {
			c.Plates = string(v)
			return nil
		}
		return errors.Wrapf(mapping.ErrFieldValue, "provided invalid field type: '%T' for the field: %s", value, field.Name())
	default:
		return errors.Wrapf(mapping.ErrInvalidModelField, "provided invalid field: '%s' for the model: 'Car'", field.Name())
	}
}

// SetPrimaryKeyStringValue implements mapping.Model interface method.
func (c *Car) ParseFieldsStringValue(field *mapping.StructField, value string) (interface{}, error) {
	switch field.Index[0] {
	case 0: // ID
		return value, nil
	case 1: // Plates
		return value, nil
	}
	return nil, errors.Wrapf(mapping.ErrInvalidModelField, "provided invalid field: '%s' for given model: Car'", field.Name())
}

// Compile time check if User implements mapping.Model interface.
var _ mapping.Model = &User{}

// NeuronCollectionName implements mapping.Model interface method.
// Returns the name of the collection for the 'User'.
func (u *User) NeuronCollectionName() string {
	return "users"
}

// IsPrimaryKeyZero implements mapping.Model interface method.
func (u *User) IsPrimaryKeyZero() bool {
	return u.ID == uuid.UUID([16]byte{})
}

// GetPrimaryKeyValue implements mapping.Model interface method.
func (u *User) GetPrimaryKeyValue() interface{} {
	return u.ID
}

// GetPrimaryKeyStringValue implements mapping.Model interface method.
func (u *User) GetPrimaryKeyStringValue() (string, error) {
	id, err := u.ID.MarshalText()
	if err != nil {
		return "", errors.Wrapf(mapping.ErrFieldValue, "invalid primary field value: %v to parse string. Err: %v", u.ID, err)
	}
	return string(id), nil
}

// GetPrimaryKeyAddress implements mapping.Model interface method.
func (u *User) GetPrimaryKeyAddress() interface{} {
	return &u.ID
}

// GetPrimaryKeyHashableValue implements mapping.Model interface method.
func (u *User) GetPrimaryKeyHashableValue() interface{} {
	return u.ID
}

// GetPrimaryKeyZeroValue implements mapping.Model interface method.
func (u *User) GetPrimaryKeyZeroValue() interface{} {
	return uuid.UUID([16]byte{})
}

// SetPrimaryKey implements mapping.Model interface method.
func (u *User) SetPrimaryKeyValue(value interface{}) error {
	if v, ok := value.(uuid.UUID); ok {
		u.ID = v
		return nil
	} else if v, ok := value.([16]byte); ok {
		u.ID = uuid.UUID(v)
	}
	return errors.Wrapf(mapping.ErrFieldValue, "provided invalid value: '%T' for the primary field for model: '%T'",
		value, u)
}

// SetPrimaryKeyStringValue implements mapping.Model interface method.
func (u *User) SetPrimaryKeyStringValue(value string) error {
	if err := u.ID.UnmarshalText([]byte(value)); err != nil {
		return errors.Wrapf(mapping.ErrFieldValue, "invalid primary field value: %v to parse string. Err: %v", u.ID, err)
	}
	return nil
}

// SetFrom implements FromSetter interface.
func (u *User) SetFrom(model mapping.Model) error {
	if model == nil {
		return errors.Wrap(query.ErrInvalidInput, "provided nil model to set from")
	}
	from, ok := model.(*User)
	if !ok {
		return errors.WrapDetf(mapping.ErrModelNotMatch, "provided model doesn't match the input: %T", model)
	}
	*u = *from
	return nil
}

// Compile time check if User implements mapping.Fielder interface.
var _ mapping.Fielder = &User{}

// GetFieldsAddress gets the address of provided 'field'.
func (u *User) GetFieldsAddress(field *mapping.StructField) (interface{}, error) {
	switch field.Index[0] {
	case 0: // ID
		return &u.ID, nil
	case 1: // CreatedAt
		return &u.CreatedAt, nil
	case 2: // DeletedAt
		return &u.DeletedAt, nil
	case 3: // Name
		return &u.Name, nil
	case 4: // Age
		return &u.Age, nil
	case 5: // IntArray
		return &u.IntArray, nil
	case 6: // Bytes
		return &u.Bytes, nil
	case 7: // PtrBytes
		return &u.PtrBytes, nil
	case 8: // Wrapped
		return &u.Wrapped, nil
	case 9: // PtrWrapped
		return &u.PtrWrapped, nil
	case 12: // FavoriteCarID
		return &u.FavoriteCarID, nil
	}
	return nil, errors.Wrapf(mapping.ErrInvalidModelField, "provided invalid field: '%s' for given model: User'", field.Name())
}

// GetFieldZeroValue implements mapping.Fielder interface.s
func (u *User) GetFieldZeroValue(field *mapping.StructField) (interface{}, error) {
	switch field.Index[0] {
	case 0: // ID
		return uuid.UUID([16]byte{}), nil
	case 1: // CreatedAt
		return time.Time{}, nil
	case 2: // DeletedAt
		return nil, nil
	case 3: // Name
		return nil, nil
	case 4: // Age
		return 0, nil
	case 5: // IntArray
		return nil, nil
	case 6: // Bytes
		return nil, nil
	case 7: // PtrBytes
		return nil, nil
	case 8: // Wrapped
		return external.Int(0), nil
	case 9: // PtrWrapped
		return nil, nil
	case 12: // FavoriteCarID
		return "", nil
	default:
		return nil, errors.Wrapf(mapping.ErrInvalidModelField, "provided invalid field name: '%s'", field.Name())
	}
}

// IsFieldZero implements mapping.Fielder interface.
func (u *User) IsFieldZero(field *mapping.StructField) (bool, error) {
	switch field.Index[0] {
	case 0: // ID
		return u.ID == uuid.UUID([16]byte{}), nil
	case 1: // CreatedAt
		return u.CreatedAt == time.Time{}, nil
	case 2: // DeletedAt
		return u.DeletedAt == nil, nil
	case 3: // Name
		return u.Name == nil, nil
	case 4: // Age
		return u.Age == 0, nil
	case 5: // IntArray
		return len(u.IntArray) == 0, nil
	case 6: // Bytes
		return len(u.Bytes) == 0, nil
	case 7: // PtrBytes
		return u.PtrBytes == nil, nil
	case 8: // Wrapped
		return u.Wrapped == external.Int(0), nil
	case 9: // PtrWrapped
		return u.PtrWrapped == nil, nil
	case 12: // FavoriteCarID
		return u.FavoriteCarID == "", nil
	}
	return false, errors.Wrapf(mapping.ErrInvalidModelField, "provided invalid field name: '%s'", field.Name())
}

// SetFieldZeroValue implements mapping.Fielder interface.s
func (u *User) SetFieldZeroValue(field *mapping.StructField) error {
	switch field.Index[0] {
	case 0: // ID
		u.ID = uuid.UUID([16]byte{})
	case 1: // CreatedAt
		u.CreatedAt = time.Time{}
	case 2: // DeletedAt
		u.DeletedAt = nil
	case 3: // Name
		u.Name = nil
	case 4: // Age
		u.Age = 0
	case 5: // IntArray
		u.IntArray = nil
	case 6: // Bytes
		u.Bytes = nil
	case 7: // PtrBytes
		u.PtrBytes = nil
	case 8: // Wrapped
		u.Wrapped = external.Int(0)
	case 9: // PtrWrapped
		u.PtrWrapped = nil
	case 12: // FavoriteCarID
		u.FavoriteCarID = ""
	default:
		return errors.Wrapf(mapping.ErrInvalidModelField, "provided invalid field name: '%s'", field.Name())
	}
	return nil
}

// GetHashableFieldValue implements mapping.Fielder interface.
func (u *User) GetHashableFieldValue(field *mapping.StructField) (interface{}, error) {
	switch field.Index[0] {
	case 0: // ID
		return u.ID, nil
	case 1: // CreatedAt
		return u.CreatedAt, nil
	case 2: // DeletedAt
		if u.DeletedAt == nil {
			return nil, nil
		}
		return *u.DeletedAt, nil
	case 3: // Name
		if u.Name == nil {
			return nil, nil
		}
		return *u.Name, nil
	case 4: // Age
		return u.Age, nil
	case 5: // IntArray
		return u.IntArray, nil
	case 6: // Bytes
		return string(u.Bytes), nil
	case 7: // PtrBytes
		if u.PtrBytes == nil {
			return nil, nil
		}
		return *u.PtrBytes, nil
	case 8: // Wrapped
		return u.Wrapped, nil
	case 9: // PtrWrapped
		if u.PtrWrapped == nil {
			return nil, nil
		}
		return *u.PtrWrapped, nil
	case 12: // FavoriteCarID
		return u.FavoriteCarID, nil
	}
	return nil, errors.Wrapf(mapping.ErrInvalidModelField, "provided invalid field: '%s' for given model: 'User'", field.Name())
}

// GetFieldValue implements mapping.Fielder interface.
func (u *User) GetFieldValue(field *mapping.StructField) (interface{}, error) {
	switch field.Index[0] {
	case 0: // ID
		return u.ID, nil
	case 1: // CreatedAt
		return u.CreatedAt, nil
	case 2: // DeletedAt
		return u.DeletedAt, nil
	case 3: // Name
		return u.Name, nil
	case 4: // Age
		return u.Age, nil
	case 5: // IntArray
		return u.IntArray, nil
	case 6: // Bytes
		return u.Bytes, nil
	case 7: // PtrBytes
		return u.PtrBytes, nil
	case 8: // Wrapped
		return u.Wrapped, nil
	case 9: // PtrWrapped
		return u.PtrWrapped, nil
	case 12: // FavoriteCarID
		return u.FavoriteCarID, nil
	}
	return nil, errors.Wrapf(mapping.ErrInvalidModelField, "provided invalid field: '%s' for given model: User'", field.Name())
}

// SetFieldValue implements mapping.Fielder interface.
func (u *User) SetFieldValue(field *mapping.StructField, value interface{}) (err error) {
	switch field.Index[0] {
	case 0: // ID
		if v, ok := value.(uuid.UUID); ok {
			u.ID = v
			return nil
		}
		if generic, ok := value.([]interface{}); ok {
			if len(generic) > 16 {
				return errors.Wrapf(mapping.ErrFieldValue, "provided too many values for the field: 'ID")
			}
			for i, item := range generic {
				if v, ok := item.(byte); ok {
					u.ID[i] = v
					continue
				}

			}
			return nil
		}
		// Checked wrapped types.
		if v, ok := value.([16]byte); ok {
			u.ID = uuid.UUID(v)
			return nil
		}
		return errors.Wrapf(mapping.ErrFieldValue, "provided invalid field type: '%T' for the field: %s", value, field.Name())
	case 1: // CreatedAt
		if v, ok := value.(time.Time); ok {
			u.CreatedAt = v
			return nil
		}
		if field.DatabaseNotNull() {
			isZero := u.CreatedAt == time.Time{}
			if isZero {
				u.CreatedAt = time.Time{}
				return nil
			}
		}

		return errors.Wrapf(mapping.ErrFieldValue, "provided invalid field type: '%T' for the field: %s", value, field.Name())
	case 2: // DeletedAt
		if value == nil {
			u.DeletedAt = nil
			return nil
		}
		if v, ok := value.(*time.Time); ok {
			u.DeletedAt = v
			return nil
		}
		// Check if it is non-pointer value.
		if v, ok := value.(time.Time); ok {
			u.DeletedAt = &v
			return nil
		}

		return errors.Wrapf(mapping.ErrFieldValue, "provided invalid field type: '%T' for the field: %s", value, field.Name())
	case 3: // Name
		if value == nil {
			u.Name = nil
			return nil
		}
		if v, ok := value.(*string); ok {
			u.Name = v
			return nil
		}
		// Check if it is non-pointer value.
		if v, ok := value.(string); ok {
			u.Name = &v
			return nil
		}

		// Check alternate types for the Name.
		if v, ok := value.([]byte); ok {
			temp := string(v)
			u.Name = &temp
			return nil
		}
		return errors.Wrapf(mapping.ErrFieldValue, "provided invalid field type: '%T' for the field: %s", value, field.Name())
	case 4: // Age
		if v, ok := value.(int); ok {
			u.Age = v
			return nil
		}
		if field.DatabaseNotNull() {
			isZero := u.Age == 0
			if isZero {
				u.Age = 0
				return nil
			}
		}

		switch v := value.(type) {
		case int8:
			u.Age = int(v)
		case int16:
			u.Age = int(v)
		case int32:
			u.Age = int(v)
		case int64:
			u.Age = int(v)
		case uint:
			u.Age = int(v)
		case uint8:
			u.Age = int(v)
		case uint16:
			u.Age = int(v)
		case uint32:
			u.Age = int(v)
		case uint64:
			u.Age = int(v)
		case float32:
			u.Age = int(v)
		case float64:
			u.Age = int(v)
		default:
			return errors.Wrapf(mapping.ErrFieldValue, "provided invalid field type: '%T' for the field: %s", value, field.Name())
		}
		return nil
	case 5: // IntArray
		if value == nil {
			u.IntArray = nil
			return nil
		}
		if v, ok := value.([]int); ok {
			u.IntArray = v
			return nil
		}
		if generic, ok := value.([]interface{}); ok {
			for _, item := range generic {
				if v, ok := item.(int); ok {
					u.IntArray = append(u.IntArray, v)
					continue
				}
				switch v := item.(type) {
				case int64:
					u.IntArray = append(u.IntArray, int(v))
				case uint:
					u.IntArray = append(u.IntArray, int(v))
				case int32:
					u.IntArray = append(u.IntArray, int(v))
				case int16:
					u.IntArray = append(u.IntArray, int(v))
				case int8:
					u.IntArray = append(u.IntArray, int(v))
				case float64:
					u.IntArray = append(u.IntArray, int(v))
				default:
					return errors.Wrapf(mapping.ErrFieldValue, "provided invalid field type: '%T' for the field: %s", value, field.Name())
				}
			}
			return nil
		}
		return errors.Wrapf(mapping.ErrFieldValue, "provided invalid field type: '%T' for the field: %s", value, field.Name())
	case 6: // Bytes
		if value == nil {
			u.Bytes = nil
			return nil
		}
		if v, ok := value.([]byte); ok {
			u.Bytes = v
			return nil
		}
		if generic, ok := value.([]interface{}); ok {
			for _, item := range generic {
				if v, ok := item.(byte); ok {
					u.Bytes = append(u.Bytes, v)
					continue
				}

			}
			return nil
		}
		// Check alternate types for the Bytes.
		if v, ok := value.(string); ok {
			u.Bytes = []byte(v)
			return nil
		}
		return errors.Wrapf(mapping.ErrFieldValue, "provided invalid field type: '%T' for the field: %s", value, field.Name())
	case 7: // PtrBytes
		if value == nil {
			u.PtrBytes = nil
			return nil
		}
		if v, ok := value.(*[]byte); ok {
			u.PtrBytes = v
			return nil
		}
		if generic, ok := value.([]interface{}); ok {
			if u.PtrBytes == nil {
				temp := []byte{}
				u.PtrBytes = &temp
			}
			for _, item := range generic {
				if v, ok := item.(byte); ok {
					*u.PtrBytes = append(*u.PtrBytes, v)
					continue
				}

			}
			return nil
		}
		// Check alternate types for the PtrBytes.
		if v, ok := value.(string); ok {
			temp := []byte(v)
			u.PtrBytes = &temp
			return nil
		}
		return errors.Wrapf(mapping.ErrFieldValue, "provided invalid field type: '%T' for the field: %s", value, field.Name())
	case 8: // Wrapped
		if v, ok := value.(external.Int); ok {
			u.Wrapped = v
			return nil
		}
		if field.DatabaseNotNull() {
			isZero := u.Wrapped == external.Int(0)
			if isZero {
				u.Wrapped = external.Int(0)
				return nil
			}
		}

		// Checked wrapped types.
		if v, ok := value.(int); ok {
			u.Wrapped = external.Int(v)
			return nil
		}
		switch v := value.(type) {
		case int8:
			u.Wrapped = external.Int(v)
		case int16:
			u.Wrapped = external.Int(v)
		case int32:
			u.Wrapped = external.Int(v)
		case int64:
			u.Wrapped = external.Int(v)
		case uint:
			u.Wrapped = external.Int(v)
		case uint8:
			u.Wrapped = external.Int(v)
		case uint16:
			u.Wrapped = external.Int(v)
		case uint32:
			u.Wrapped = external.Int(v)
		case uint64:
			u.Wrapped = external.Int(v)
		case float32:
			u.Wrapped = external.Int(v)
		case float64:
			u.Wrapped = external.Int(v)
		default:
			return errors.Wrapf(mapping.ErrFieldValue, "provided invalid field type: '%T' for the field: %s", value, field.Name())
		}
		return nil
	case 9: // PtrWrapped
		if value == nil {
			u.PtrWrapped = nil
			return nil
		}
		if v, ok := value.(*external.Int); ok {
			u.PtrWrapped = v
			return nil
		}
		// Check if it is non-pointer value.
		if v, ok := value.(external.Int); ok {
			u.PtrWrapped = &v
			return nil
		}

		// Checked wrapped types.
		if v, ok := value.(int); ok {
			temp := external.Int(v)
			u.PtrWrapped = &temp
			return nil
		}
		switch v := value.(type) {
		case int8:
			temp := external.Int(v)
			u.PtrWrapped = &temp
		case *int8:
			temp := external.Int(*v)
			u.PtrWrapped = &temp
		case int16:
			temp := external.Int(v)
			u.PtrWrapped = &temp
		case *int16:
			temp := external.Int(*v)
			u.PtrWrapped = &temp
		case int32:
			temp := external.Int(v)
			u.PtrWrapped = &temp
		case *int32:
			temp := external.Int(*v)
			u.PtrWrapped = &temp
		case int64:
			temp := external.Int(v)
			u.PtrWrapped = &temp
		case *int64:
			temp := external.Int(*v)
			u.PtrWrapped = &temp
		case uint:
			temp := external.Int(v)
			u.PtrWrapped = &temp
		case *uint:
			temp := external.Int(*v)
			u.PtrWrapped = &temp
		case uint8:
			temp := external.Int(v)
			u.PtrWrapped = &temp
		case *uint8:
			temp := external.Int(*v)
			u.PtrWrapped = &temp
		case uint16:
			temp := external.Int(v)
			u.PtrWrapped = &temp
		case *uint16:
			temp := external.Int(*v)
			u.PtrWrapped = &temp
		case uint32:
			temp := external.Int(v)
			u.PtrWrapped = &temp
		case *uint32:
			temp := external.Int(*v)
			u.PtrWrapped = &temp
		case uint64:
			temp := external.Int(v)
			u.PtrWrapped = &temp
		case *uint64:
			temp := external.Int(*v)
			u.PtrWrapped = &temp
		case float32:
			temp := external.Int(v)
			u.PtrWrapped = &temp
		case *float32:
			temp := external.Int(*v)
			u.PtrWrapped = &temp
		case float64:
			temp := external.Int(v)
			u.PtrWrapped = &temp
		case *float64:
			temp := external.Int(*v)
			u.PtrWrapped = &temp
		default:
			return errors.Wrapf(mapping.ErrFieldValue, "provided invalid field type: '%T' for the field: %s", value, field.Name())
		}
		return nil
	case 12: // FavoriteCarID
		if v, ok := value.(string); ok {
			u.FavoriteCarID = v
			return nil
		}
		if field.DatabaseNotNull() {
			isZero := u.FavoriteCarID == ""
			if isZero {
				u.FavoriteCarID = ""
				return nil
			}
		}

		// Check alternate types for the FavoriteCarID.
		if v, ok := value.([]byte); ok {
			u.FavoriteCarID = string(v)
			return nil
		}
		return errors.Wrapf(mapping.ErrFieldValue, "provided invalid field type: '%T' for the field: %s", value, field.Name())
	default:
		return errors.Wrapf(mapping.ErrInvalidModelField, "provided invalid field: '%s' for the model: 'User'", field.Name())
	}
}

// SetPrimaryKeyStringValue implements mapping.Model interface method.
func (u *User) ParseFieldsStringValue(field *mapping.StructField, value string) (interface{}, error) {
	switch field.Index[0] {
	case 0: // ID
		temp := u.ID
		if err := u.ID.UnmarshalText([]byte(value)); err != nil {
			return "", errors.Wrapf(mapping.ErrFieldValue, "invalid field 'ID' value: '%v' to parse string. Err: %v", u.ID, err)
		}
		bt, err := u.ID.MarshalText()
		if err != nil {
			return "", errors.Wrapf(mapping.ErrFieldValue, "invalid field 'ID' value: '%v' to parse string. Err: %v", u.ID, err)
		}
		u.ID = temp
		return string(bt), nil
	case 1: // CreatedAt
		temp := u.CreatedAt
		if err := u.CreatedAt.UnmarshalText([]byte(value)); err != nil {
			return "", errors.Wrapf(mapping.ErrFieldValue, "invalid field 'CreatedAt' value: '%v' to parse string. Err: %v", u.CreatedAt, err)
		}
		bt, err := u.CreatedAt.MarshalText()
		if err != nil {
			return "", errors.Wrapf(mapping.ErrFieldValue, "invalid field 'CreatedAt' value: '%v' to parse string. Err: %v", u.CreatedAt, err)
		}
		u.CreatedAt = temp
		return string(bt), nil
	case 2: // DeletedAt
		var base time.Time
		temp := &base
		if err := temp.UnmarshalText([]byte(value)); err != nil {
			return "", errors.Wrapf(mapping.ErrFieldValue, "invalid field 'DeletedAt' value: '%v' to parse string. Err: %v", u.DeletedAt, err)
		}
		bt, err := temp.MarshalText()
		if err != nil {
			return "", errors.Wrapf(mapping.ErrFieldValue, "invalid field 'DeletedAt' value: '%v' to parse string. Err: %v", u.DeletedAt, err)
		}

		return string(bt), nil
	case 3: // Name
		return value, nil
	case 4: // Age
		return strconv.ParseInt(value, 10, mapping.IntegerBitSize)
	case 5: // IntArray
		return strconv.ParseInt(value, 10, mapping.IntegerBitSize)
	case 6: // Bytes
		return "", errors.Wrap(mapping.ErrFieldNotParser, "field 'Bytes' doesn't have string setter.")
	case 7: // PtrBytes
		return "", errors.Wrap(mapping.ErrFieldNotParser, "field 'PtrBytes' doesn't have string setter.")
	case 8: // Wrapped
		return strconv.ParseInt(value, 10, mapping.IntegerBitSize)
	case 9: // PtrWrapped
		return strconv.ParseInt(value, 10, mapping.IntegerBitSize)
	case 12: // FavoriteCarID
		return value, nil
	}
	return nil, errors.Wrapf(mapping.ErrInvalidModelField, "provided invalid field: '%s' for given model: User'", field.Name())
}

// Compile time check if User implements mapping.SingleRelationer interface.
var _ mapping.SingleRelationer = &User{}

// GetRelationModel implements mapping.SingleRelationer interface.
func (u *User) GetRelationModel(relation *mapping.StructField) (mapping.Model, error) {
	switch relation.Index[0] {
	case 11: // FavoriteCar

		return &u.FavoriteCar, nil
	case 15: // Sister
		if u.Sister == nil {
			return nil, nil
		}
		return u.Sister, nil
	case 10: // External
		if u.External == nil {
			return nil, nil
		}
		return u.External, nil
	default:
		return nil, errors.Wrapf(mapping.ErrInvalidRelationField, "provided invalid relation: '%s' for model: '%T'", relation, u)
	}
}

// SetRelationModel implements mapping.SingleRelationer interface.
func (u *User) SetRelationModel(relation *mapping.StructField, model mapping.Model) error {
	switch relation.Index[0] {
	case 11: // FavoriteCar
		if model == nil {
			u.FavoriteCar = Car{}
			return nil
		} else if favoriteCar, ok := model.(*Car); ok {
			u.FavoriteCar = *favoriteCar
			return nil
		}
		return errors.Wrapf(mapping.ErrInvalidRelationValue, "provided invalid model value: '%T' for relation FavoriteCar", model)
	case 15: // Sister
		if model == nil {
			u.Sister = nil
			return nil
		} else if sister, ok := model.(*User); ok {
			u.Sister = sister
			return nil
		}
		return errors.Wrapf(mapping.ErrInvalidRelationValue, "provided invalid model value: '%T' for relation Sister", model)
	case 10: // External
		if model == nil {
			u.External = nil
			return nil
		} else if external, ok := model.(*external.Model); ok {
			u.External = external
			return nil
		}
		return errors.Wrapf(mapping.ErrInvalidRelationValue, "provided invalid model value: '%T' for relation External", model)
	default:
		return errors.Wrapf(mapping.ErrInvalidRelationField, "provided invalid relation: '%s' for model: '%T'", relation, u)
	}
}

// Compile time check for the mapping.MultiRelationer interface implementation.
var _ mapping.MultiRelationer = &User{}

// AddRelationModel implements mapping.MultiRelationer interface.
func (u *User) AddRelationModel(relation *mapping.StructField, model mapping.Model) error {
	switch relation.Index[0] {
	case 13: // Cars
		car, ok := model.(*Car)
		if !ok {
			return errors.Wrapf(mapping.ErrInvalidRelationValue, "provided invalid value type: '%T'  for the field: 'Cars'", model)
		}
		u.Cars = append(u.Cars, car)
	case 14: // Sons
		user, ok := model.(*User)
		if !ok {
			return errors.Wrapf(mapping.ErrInvalidRelationValue, "provided invalid value type: '%T'  for the field: 'Sons'", model)
		}
		u.Sons = append(u.Sons, user)
	default:
		return errors.Wrapf(mapping.ErrInvalidRelationField, "provided invalid relation: '%T' for the model 'User'", model)
	}
	return nil
}

// GetRelationModels implements mapping.MultiRelationer interface.
func (u *User) GetRelationModels(relation *mapping.StructField) (models []mapping.Model, err error) {
	switch relation.Index[0] {
	case 13: // Cars
		for _, model := range u.Cars {
			models = append(models, model)
		}
	case 14: // Sons
		for _, model := range u.Sons {
			models = append(models, model)
		}
	default:
		return nil, errors.Wrapf(mapping.ErrInvalidRelationField, "provided invalid relation: '%s' for model: '%T'", relation, u)
	}
	return models, nil
}

// GetRelationModelAt implements mapping.MultiRelationer interface.
func (u *User) GetRelationModelAt(relation *mapping.StructField, index int) (models mapping.Model, err error) {
	switch relation.Index[0] {
	case 13: // Cars
		if index > len(u.Cars)-1 {
			return nil, errors.Wrapf(mapping.ErrInvalidRelationIndex, "index out of possible range. Model: 'User', Field Cars")
		}
		return u.Cars[index], nil
	case 14: // Sons
		if index > len(u.Sons)-1 {
			return nil, errors.Wrapf(mapping.ErrInvalidRelationIndex, "index out of possible range. Model: 'User', Field Sons")
		}
		return u.Sons[index], nil
	default:
		return nil, errors.Wrapf(mapping.ErrInvalidRelationField, "provided invalid relation: '%s' for model: '%T'", relation, u)
	}
	return models, nil
}

// GetRelationLen implements mapping.MultiRelationer interface.
func (u *User) GetRelationLen(relation *mapping.StructField) (int, error) {
	switch relation.Index[0] {
	case 13: // Cars
		return len(u.Cars), nil
	case 14: // Sons
		return len(u.Sons), nil
	default:
		return 0, errors.Wrapf(mapping.ErrInvalidRelationField, "provided invalid relation: '%s' for model: '%T'", relation, u)
	}
}

// SetRelationModels implements mapping.MultiRelationer interface.
func (u *User) SetRelationModels(relation *mapping.StructField, models ...mapping.Model) error {
	switch relation.Index[0] {
	case 13: // Cars
		temp := make([]*Car, len(models))
		for i, model := range models {
			car, ok := model.(*Car)
			if !ok {
				return errors.Wrapf(mapping.ErrInvalidRelationValue, "provided invalid value type: '%T'  for the field: 'Cars'", model)
			}
			temp[i] = car
		}
		u.Cars = temp
	case 14: // Sons
		temp := make([]*User, len(models))
		for i, model := range models {
			user, ok := model.(*User)
			if !ok {
				return errors.Wrapf(mapping.ErrInvalidRelationValue, "provided invalid value type: '%T'  for the field: 'Sons'", model)
			}
			temp[i] = user
		}
		u.Sons = temp
	default:
		return errors.Wrapf(mapping.ErrInvalidRelationField, "provided invalid relation: '%s' for the model 'User'", relation.String())
	}
	return nil
}
