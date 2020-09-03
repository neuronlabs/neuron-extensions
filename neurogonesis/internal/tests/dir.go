package tests

import (
	"database/sql"
	"database/sql/driver"
	"encoding"
	"errors"
)

// SqlEnum represents sql enumerator.
type SqlEnum uint

const (
	SqlFirst SqlEnum = iota
	SqlSecond
)

// String implements fmt.Stringer interface.
func (r SqlEnum) String() string {
	switch r {
	case SqlFirst:
		return "LTR"
	case SqlSecond:
		return "RTL"
	default:
		return "UNKNOWN"
	}
}

// Valid checks the validity of the enum.
func (r SqlEnum) Valid() bool {
	switch r {
	case SqlFirst, SqlSecond:
		return true
	default:
		return false
	}
}

// compile time check for driver.Valuer interface.
var _ driver.Valuer = SqlFirst

// Value implements driver.Valuer interface.
func (r SqlEnum) Value() (driver.Value, error) {
	if r > SqlSecond {
		return nil, errors.New("unknown value")
	}
	return driver.Value(r.String()), nil
}

var _ sql.Scanner = &tmpr

// Scan implements sql.Scanner interface.
func (r *SqlEnum) Scan(value interface{}) error {
	bin, ok := value.([]byte)
	if !ok {
		return errors.New("invalid enum")
	}
	return r.UnmarshalText(bin)
}

// MarshalText implements encoding.TextMarshaler interface.
func (r SqlEnum) MarshalText() ([]byte, error) {
	return []byte(r.String()), nil
}

// compile time check for the encoding.TextUnmarshaler interface.
var (
	tmpr                          = SqlSecond
	_    encoding.TextUnmarshaler = &tmpr
)

// UnmarshalText implement encoding.TextUnmarshaler interface.
func (r *SqlEnum) UnmarshalText(text []byte) error {
	switch string(text) {
	case "First":
		*r = SqlFirst
	case "Second":
		*r = SqlSecond
	default:
		return errors.New("unknown enum to scan: " + string(text))
	}
	return nil
}
