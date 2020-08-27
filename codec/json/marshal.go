package json

import (
	"encoding/json"
	"io"
	"time"

	"github.com/neuronlabs/neuron/codec"
	"github.com/neuronlabs/neuron/errors"
	"github.com/neuronlabs/neuron/mapping"
)

// MarshalPayload implements codec.PayloadMarshaler interface.
func (c Codec) MarshalPayload(w io.Writer, payload *codec.Payload, options ...codec.MarshalOption) error {
	o := &codec.MarshalOptions{}
	for _, option := range options {
		option(o)
	}
	if o.SingleResult {
		if len(payload.Data) > 1 {
			return errors.WrapDetf(errors.ErrInternal, "marshaling singular format with multiple data models")
		}
		pl := marshalSinglePayload{
			Meta:  payload.Meta,
			Links: payload.PaginationLinks,
		}
		if len(payload.Data) == 1 {
			node, err := c.marshalModel(payload.Data[0])
			if err != nil {
				return err
			}
			pl.Data = node
		}
		if err := json.NewEncoder(w).Encode(pl); err != nil {
			return errors.WrapDetf(codec.ErrMarshal, "marshaling payload failed: %v", err)
		}
	} else {
		mr := marshalMultiPayload{
			Links: payload.PaginationLinks,
			Meta:  payload.Meta,
		}
		for _, model := range payload.Data {
			node, err := c.marshalModel(model)
			if err != nil {
				return err
			}
			mr.Data = append(mr.Data, node)
		}
		if err := json.NewEncoder(w).Encode(mr); err != nil {
			return errors.WrapDetf(codec.ErrMarshal, "marshaling payload failed: %v", err)
		}
	}
	return nil
}

// MarshalModel implements codec.ModelMarshaler interface.
func (c Codec) MarshalModel(model mapping.Model, _ ...codec.MarshalOption) ([]byte, error) {
	marshaler, err := c.marshalModel(model)
	if err != nil {
		return nil, err
	}
	return marshaler.MarshalJSON()
}

// MarshalModels implements codec.ModelMarshaler interface.
func (c Codec) MarshalModels(models []mapping.Model, _ ...codec.MarshalOption) ([]byte, error) {
	var err error
	nodes := make([]marshaler, len(models))
	for i, model := range models {
		if model == nil {
			nodes[i] = nil
			continue
		}
		nodes[i], err = c.marshalModel(model)
		if err != nil {
			return nil, err
		}
	}
	return json.Marshal(nodes)
}

func (c Codec) marshalModel(model mapping.Model) (marshaler, error) {
	mStruct, err := c.c.ModelStruct(model)
	if err != nil {
		return nil, err
	}

	fielder, ok := model.(mapping.Fielder)
	if !ok {
		return nil, errors.WrapDetf(mapping.ErrModelNotImplements, "provided model: '%s' doesn't implement Fielder interface", mStruct)
	}

	var result marshaler
	// Iterate over primary key, attributes and foreign keys.
	for _, sField := range mStruct.StructFields() {
		if sField.CodecSkip() {
			continue
		}
		switch sField.Kind() {
		case mapping.KindPrimary, mapping.KindAttribute, mapping.KindForeignKey:
			isZero, err := fielder.IsFieldZero(sField)
			if err != nil {
				return nil, err
			}
			if isZero && sField.CodecOmitEmpty() {
				continue
			}
			fieldValue, err := fielder.GetFieldValue(sField)
			if err != nil {
				return nil, err
			}

			if sField.IsTimePointer() && fieldValue != nil {
				tv := fieldValue.(*time.Time)
				if sField.CodecISO8601() {
					fieldValue = tv.Format(codec.ISO8601TimeFormat)
				} else {
					fieldValue = tv.Unix()
				}
			} else if sField.IsTime() {
				tv := fieldValue.(time.Time)
				if sField.CodecISO8601() {
					fieldValue = tv.Format(codec.ISO8601TimeFormat)
				} else {
					fieldValue = tv.Unix()
				}
			}
			result = append(result, keyValue{Key: sField.CodecName(), Value: fieldValue})
		case mapping.KindRelationshipSingle:
			sr, ok := model.(mapping.SingleRelationer)
			if !ok {
				return nil, errors.WrapDetf(mapping.ErrModelNotImplements, "model: '%s' doesn't implement SingleRelationer", mStruct)
			}
			relation, err := sr.GetRelationModel(sField)
			if err != nil {
				return nil, err
			}
			if relation == nil && sField.CodecOmitEmpty() {
				continue
			}
			if relation == nil {
				result = append(result, keyValue{Key: sField.CodecName(), Value: nil})
				continue
			}
			relationNode, err := c.marshalModel(relation)
			if err != nil {
				return nil, err
			}
			result = append(result, keyValue{Key: sField.CodecName(), Value: relationNode})
		case mapping.KindRelationshipMultiple:
			mr, ok := model.(mapping.MultiRelationer)
			if !ok {
				return nil, errors.WrapDetf(mapping.ErrModelNotImplements, "model: '%s' doesn't implement MultiRelationer", mStruct)
			}
			models, err := mr.GetRelationModels(sField)
			if err != nil {
				return nil, err
			}
			if len(models) == 0 && sField.CodecOmitEmpty() {
				continue
			}
			relationNodes := make([]marshaler, len(models))
			for i, relation := range models {
				if relation == nil {
					relationNodes[i] = nil
					continue
				}
				relationNode, err := c.marshalModel(relation)
				if err != nil {
					return nil, err
				}
				relationNodes[i] = relationNode
			}
			result = append(result, keyValue{Key: sField.CodecName(), Value: relationNodes})
		}
	}
	return result, nil
}

type marshalSinglePayload struct {
	Data  marshaler              `json:"data"`
	Links *codec.PaginationLinks `json:"links,omitempty"`
	Meta  codec.Meta             `json:"meta,omitempty"`
}

type marshalMultiPayload struct {
	Data  []marshaler            `json:"data"`
	Links *codec.PaginationLinks `json:"links,omitempty"`
	Meta  codec.Meta             `json:"meta,omitempty"`
}
