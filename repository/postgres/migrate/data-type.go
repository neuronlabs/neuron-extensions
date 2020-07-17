package migrate

import (
	"reflect"
	"strings"
	"time"

	"github.com/neuronlabs/neuron/controller"
	"github.com/neuronlabs/neuron/errors"
	"github.com/neuronlabs/neuron/mapping"

	"github.com/neuronlabs/neuron-plugins/repository/postgres/log"
)

var (
	/** Character types */

	// FChar is the 'char' field type.
	FChar = &ParameteredDataType{SQLName: "char", DataType: DataType{Name: "char"}}
	// FVarChar is the 'varchar' field type.
	FVarChar = &ParameteredDataType{SQLName: "varchar", DataType: DataType{Name: "varchar"}}
	// FText is the 'text' field type.
	FText = &BasicDataType{SQLName: "text", DataType: DataType{Name: "text"}}

	/** Numerics */

	// FSmallInt is the 2 bytes signed 'smallint' - int16.
	FSmallInt = &BasicDataType{SQLName: "smallint", DataType: DataType{Name: "smallint"}}
	// FInteger is the 4 bytes signed 'integer' type - int32.
	FInteger = &BasicDataType{SQLName: "integer", DataType: DataType{Name: "integer"}}
	// FBigInt is the 8 bytes signed 'bigint' type - int64.
	FBigInt = &BasicDataType{SQLName: "bigint", DataType: DataType{Name: "bigint"}}
	// FDecimal is the variable 'decimal' type.
	FDecimal = &ParameteredDataType{SQLName: "decimal", DataType: DataType{Name: "decimal"}}
	// FNumeric is the ariable 'numeric' type.
	FNumeric = &ParameteredDataType{SQLName: "numeric", DataType: DataType{Name: "numeric"}}
	// FReal is the 4 bytes - 6 decimal digits precision 'real' type.
	FReal = &BasicDataType{SQLName: "real", DataType: DataType{Name: "real"}}
	// FDouble is the 8 bytes - 15 decimal digits precision 'double precision' type.
	FDouble = &BasicDataType{SQLName: "double precision", DataType: DataType{Name: "double"}}
	// FSerial is the 4 bytes - autoincrement integer 'serial' type.
	FSerial = &BasicDataType{SQLName: "serial", DataType: DataType{Name: "serial"}}
	// FBigSerial is the 8 bytes autoincrement big integer - 'bigserial' type.
	FBigSerial = &BasicDataType{SQLName: "bigserial", DataType: DataType{Name: "bigserial"}}

	/** Binary */

	// FBytea is the 1 or 4 bytes plus the actual binary string data type 'bytea'.
	FBytea = &ParameteredDataType{SQLName: "bytea", DataType: DataType{Name: "bytea"}}
	// FBoolean is the 'boolean' pq data type.
	FBoolean = &BasicDataType{SQLName: "boolean", DataType: DataType{Name: "boolean"}}

	/** Date / Time */

	// FDate is the 'date' field kind.
	FDate = &BasicDataType{SQLName: "date", DataType: DataType{Name: "date"}}
	// FTimestamp is the 'timestamp' without time zone data type.
	FTimestamp = &OptionalParameterDataType{SQLNames: []string{"timestamp"}, ParameterIndex: 1, DataType: DataType{Name: "timestamp"}}
	// FTimestampTZ is the 'timestamp with time zone' data type.
	FTimestampTZ = &OptionalParameterDataType{SQLNames: []string{"timestamp", "with time zone"}, ParameterIndex: 1, DataType: DataType{Name: "timestamptz"}}
	// FTime is the 'time' without time zone data type.
	FTime = &OptionalParameterDataType{SQLNames: []string{"time"}, ParameterIndex: 1, DataType: DataType{Name: "time"}}
	// FTimeTZ is the 'time with time zone' data type.
	FTimeTZ = &OptionalParameterDataType{SQLNames: []string{"time", "with time zone"}, ParameterIndex: 1, DataType: DataType{Name: "timetz"}}
)

// dataTypes is the array containing the data types
var (
	dataTypes     = make(map[string]DataTyper)
	defaultKindDT = map[reflect.Kind]DataTyper{
		reflect.Bool:    FBoolean,
		reflect.Int:     FInteger,
		reflect.Int8:    FSmallInt,
		reflect.Int16:   FSmallInt,
		reflect.Int32:   FInteger,
		reflect.Int64:   FBigInt,
		reflect.Uint:    FInteger,
		reflect.Uint8:   FSmallInt,
		reflect.Uint16:  FSmallInt,
		reflect.Uint32:  FInteger,
		reflect.Uint64:  FBigInt,
		reflect.String:  FText,
		reflect.Float32: FReal,
		reflect.Float64: FDouble,
	}
	defaultTypeDT = map[reflect.Type]DataTyper{
		reflect.TypeOf(time.Time{}):  FTimestamp,
		reflect.TypeOf(&time.Time{}): FTimestamp,
	}
)

// findDataType finds the data type for the provided field
func findDataType(field *mapping.StructField) (DataTyper, error) {
	c, err := fieldsColumn(field)
	if err != nil {
		return nil, err
	}

	if c.Type != nil {
		return c.Type, nil
	}
	t := field.ReflectField().Type
	if field.Kind() == mapping.KindPrimary {
		// by default for the integer primary keys set the serial or bigserial type
		switch t.Kind() {
		case reflect.Int, reflect.Int16, reflect.Int8, reflect.Int32, reflect.Uint, reflect.Uint32, reflect.Uint8, reflect.Uint16:
			return FSerial, nil
		case reflect.Int64, reflect.Uint64:
			return FBigSerial, nil
		}
	}

	if field.IsCreatedAt() || field.IsDeletedAt() || field.IsUpdatedAt() {
		return FTimestampTZ, nil
	}

	// at first check type
	dt, ok := defaultTypeDT[t]
	if !ok {
		k := t.Kind()
		if k == reflect.Ptr {
			k = t.Elem().Kind()
		}
		dt, ok = defaultKindDT[k]
		if !ok {
			err := errors.NewDet(controller.ClassInternal, "postgres field type not found.")
			err.WithDetailf("Can't find the field type. Model: '%s', Field: '%s'", field.ModelStruct().Type().Name(), field.Name())
			return nil, err
		}
	}

	c.Type = dt

	return dt, nil
}

// ExternalDataTyper is the interface that defines the columns that sets the column outside the table definition.
type ExternalDataTyper interface {
	DataTyper

	// ExternalFunction is the method used to create the column outside of the table definition.
	ExternalFunction(field *mapping.StructField) string
}

// DataTyper is the interface for basic data type methods.
type DataTyper interface {
	KeyName() string

	// GetName creates the column string used within the table definition
	GetName(field *mapping.StructField) string
}

// DataType is the pq base model defininig the data type.
type DataType struct {
	Name string
}

// KeyName gets the name of the data type.
func (d *DataType) KeyName() string {
	return d.Name
}

// BasicDataType is the InlineDataTyper that sets the basic columns on the base of it's SQLName.
type BasicDataType struct {
	DataType

	SQLName string
}

// GetName creates the inline column definition on the base of it's SQLName.
func (b *BasicDataType) GetName(*mapping.StructField) string {
	return b.SQLName
}

// compile time check of BasicDataType
var _ DataTyper = &BasicDataType{}

// ParameteredDataType is the data type that contains the variable parameters.
// i.e. varchar(2) has a single parameter '2'.
type ParameteredDataType struct {
	DataType
	SQLName  string
	Validate func(params []string) error
}

// GetName creates the inline column definition on the base of it's SQLName and Parameters.
func (p *ParameteredDataType) GetName(field *mapping.StructField) string {
	paramsValue, ok := field.StoreGet(DataTypeParametersStoreKey)
	if !ok {
		log.Warningf("The field: '%s' data type has no parameters set.", field.Name())
	}

	var params []string
	params, ok = paramsValue.([]string)
	if !ok {
		log.Errorf("The field's parameters are not an array of strings: %T", paramsValue)
	}

	if p.Validate != nil {
		err := p.Validate(params)
		if err != nil {
			log.Error(err)
		}
	}

	return p.SQLName + "(" + strings.Join(params, ",") + ")"
}

// OptionalParameterDataType is the data type that contains optional parameters.
type OptionalParameterDataType struct {
	DataType
	SQLNames       []string
	ParameterIndex int
}

// GetName creates the inline column definition on the base of it's SQLName and Parameters.
// nolint:gocritic
func (p *OptionalParameterDataType) GetName(field *mapping.StructField) string {
	var params []string

	paramsValue, ok := field.StoreGet(DataTypeParametersStoreKey)
	if !ok {
		return strings.Join(p.SQLNames, " ")
	}

	var sqlVars []string
	if params, ok = paramsValue.([]string); ok {
		param := "(" + strings.Join(params, ",") + ")"
		if p.ParameterIndex == len(p.SQLNames) {
			sqlVars = append(p.SQLNames, param)
		} else {
			sqlVars = append(p.SQLNames[:p.ParameterIndex], param)
			sqlVars = append(sqlVars, p.SQLNames[p.ParameterIndex+1:]...)
		}
	}
	return strings.Join(sqlVars, " ")
}

// RegisterDataType registers the provided datatype assigning it next id.
func RegisterDataType(dt DataTyper) error {
	return registerDataType(dt)
}

func registerDataType(dt DataTyper) error {
	// check it the data type exists
	_, ok := dataTypes[dt.KeyName()]
	if ok {
		return errors.NewDetf(controller.ClassInternal, "postgres data type: '%s' is already registered", dt.KeyName())
	}

	// set the data type at index
	dataTypes[dt.KeyName()] = dt

	return nil
}

// RegisterRefTypeDT registers default data type for provided reflect.Type.
func RegisterRefTypeDT(t reflect.Type, dt DataTyper, override ...bool) error {
	return registerRefTypeDT(t, dt, override...)
}

func registerRefTypeDT(t reflect.Type, dt DataTyper, override ...bool) error {
	var ov bool
	if len(override) > 0 {
		ov = override[0]
	}
	_, ok := defaultTypeDT[t]
	if ok && !ov {
		return errors.NewDetf(controller.ClassInternal, "default data typer is already set for given type: '%s'", t.Name())
	}

	defaultTypeDT[t] = dt
	return nil
}
