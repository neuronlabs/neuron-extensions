package json

import (
	"encoding/json"
	"strings"
)

type keyValue struct {
	Key   string
	Value interface{}
}

type marshaler []keyValue

// MarshalJSON implements json.Marshaler interface.
func (m marshaler) MarshalJSON() ([]byte, error) {
	if m == nil {
		return []byte("null"), nil
	}

	sb := strings.Builder{}
	sb.WriteRune('{')
	for i, kv := range m {
		sb.WriteRune('"')
		sb.WriteString(kv.Key)
		sb.WriteRune('"')
		sb.WriteRune(':')

		value, err := json.Marshal(kv.Value)
		if err != nil {
			return nil, err
		}
		sb.Write(value)
		if i != len(m)-1 {
			sb.WriteRune(',')
			sb.WriteRune(' ')
		}
	}
	return []byte(sb.String()), nil
}
