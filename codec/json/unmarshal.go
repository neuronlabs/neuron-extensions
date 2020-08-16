package json

import (
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
	"time"

	"github.com/neuronlabs/neuron/codec"
	"github.com/neuronlabs/neuron/errors"
	"github.com/neuronlabs/neuron/mapping"
)

// UnmarshalModels implements codec.Codec interface.
func (c Codec) UnmarshalModels(data []byte, options codec.UnmarshalOptions) ([]mapping.Model, error) {
	// get the payload.
	payload, err := getPayload(data)
	if err != nil {
		return nil, err
	}
	var models []mapping.Model
	for _, n := range payload.nodes() {
		model, _, err := c.unmarshalNode(options.ModelStruct, n, options)
		if err != nil {
			return nil, err
		}
		models = append(models, model)
	}
	return models, nil
}

// UnmarshalPayload implements codec.PayloadMarshaler interface.
func (c Codec) UnmarshalPayload(r io.Reader, options codec.UnmarshalOptions) (*codec.Payload, error) {
	data, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, errors.WrapDetf(codec.ErrUnmarshal, "reading input failed: %v", err)
	}

	inputPayload, err := getPayload(data)
	if err != nil {
		return nil, err
	}
	payload := codec.Payload{
		Meta:            inputPayload.meta(),
		PaginationLinks: inputPayload.pagination(),
	}
	for _, n := range inputPayload.nodes() {
		model, fieldSet, err := c.unmarshalNode(options.ModelStruct, n, options)
		if err != nil {
			return nil, err
		}
		payload.Data = append(payload.Data, model)
		payload.FieldSets = append(payload.FieldSets, fieldSet)
	}
	return &payload, nil
}

func getPayload(data []byte) (payloader, error) {
	br := bytes.NewReader(data)
	rn, _, err := br.ReadRune()
	if err != nil {
		return nil, errors.Wrap(codec.ErrUnmarshalDocument, "provided invalid document")
	}
	br.Seek(0, io.SeekStart)
	switch rn {
	case '{':
		payload := singlePayload{}
		if err = json.NewDecoder(br).Decode(&payload); err != nil {
			return nil, err
		}
		return payload, nil
	case '[':
		payload := multiPayload{}
		if err := json.NewDecoder(br).Decode(&payload); err != nil {
			return nil, err
		}
		return payload, nil
	default:
		return nil, errors.Wrap(codec.ErrUnmarshalDocument, "provided invalid document")
	}
}

func (c Codec) unmarshalNode(mStruct *mapping.ModelStruct, node modelNode, options codec.UnmarshalOptions) (mapping.Model, mapping.FieldSet, error) {
	if !options.StrictUnmarshal {
		return c.unmarshalNodeNonStrict(mStruct, node, options)
	}
	return c.unmarshalNodeStrict(mStruct, node, options)
}

func (c Codec) unmarshalNodeNonStrict(mStruct *mapping.ModelStruct, n modelNode, options codec.UnmarshalOptions) (mapping.Model, mapping.FieldSet, error) {
	model := mapping.NewModel(mStruct)
	var (
		sField   *mapping.StructField
		ok       bool
		fieldSet mapping.FieldSet
		fielder  mapping.Fielder
		err      error
		value    interface{}
	)
	for _, sField = range mStruct.StructFields() {
		value, ok = n[sField.CodecName()]
		if !ok {
			continue
		}
		if sField.CodecSkip() {
			continue
		}
		fieldSet = append(fieldSet, sField)
		switch sField.Kind() {
		case mapping.KindPrimary:
			if err = model.SetPrimaryKeyValue(value); err != nil {
				return nil, nil, err
			}
		case mapping.KindAttribute, mapping.KindForeignKey:
			if fielder == nil {
				fielder, ok = model.(mapping.Fielder)
				if !ok {
					return nil, nil, errors.WrapDetf(mapping.ErrModelNotImplements, "model: '%s' doesn't implement mapping.Fielder interface", mStruct.String())
				}
			}
			if err = setAttribute(sField, value, fielder); err != nil {
				return nil, nil, err
			}
		case mapping.KindRelationshipMultiple:
			if value == nil {
				continue
			}
			if err = c.setRelationshipMany(mStruct, model, sField, value, options, sField.CodecName()); err != nil {
				return nil, nil, err
			}
		case mapping.KindRelationshipSingle:
			if value == nil {
				continue
			}
			if err = c.setRelationshipSingle(mStruct, model, sField, value, options, sField.CodecName()); err != nil {
				return nil, nil, err
			}
		}
	}
	return model, fieldSet, nil
}

func (c Codec) unmarshalNodeStrict(mStruct *mapping.ModelStruct, node modelNode, options codec.UnmarshalOptions) (mapping.Model, mapping.FieldSet, error) {
	model := mapping.NewModel(mStruct)
	var (
		sField   *mapping.StructField
		ok       bool
		fieldSet mapping.FieldSet
		fielder  mapping.Fielder
		err      error
	)
	for field, value := range node {
		var fieldAnnotation codec.FieldAnnotations
		sField, ok = mStruct.FieldByName(field)
		if !ok {
			for _, structField := range mStruct.StructFields() {
				fieldAnnotation = codec.ExtractFieldAnnotations(structField, "codec")
				if fieldAnnotation.Name == field {
					sField = structField
					break
				}
			}
		} else {
			fieldAnnotation = codec.ExtractFieldAnnotations(sField, "codec")
		}
		if sField == nil || fieldAnnotation.IsHidden {
			return nil, nil, errors.WrapDetf(codec.ErrUnmarshal, "provided invalid field: '%s'", field)
		}

		fieldSet = append(fieldSet, sField)
		switch sField.Kind() {
		case mapping.KindPrimary:
			if err = model.SetPrimaryKeyValue(value); err != nil {
				return nil, nil, err
			}
		case mapping.KindAttribute, mapping.KindForeignKey:
			if fielder == nil {
				fielder, ok = model.(mapping.Fielder)
				if !ok {
					return nil, nil, errors.WrapDetf(mapping.ErrModelNotImplements, "model: '%s' doesn't implement mapping.Fielder interface", mStruct.String())
				}
			}
			if err = setAttribute(sField, value, fielder); err != nil {
				return nil, nil, err
			}
		case mapping.KindRelationshipMultiple:
			if value == nil {
				continue
			}
			if err = c.setRelationshipMany(mStruct, model, sField, value, options, fieldAnnotation.Name); err != nil {
				return nil, nil, err
			}
		case mapping.KindRelationshipSingle:
			if value == nil {
				continue
			}
			if err = c.setRelationshipSingle(mStruct, model, sField, value, options, fieldAnnotation.Name); err != nil {
				return nil, nil, err
			}
		}
	}
	return model, fieldSet, nil
}

func (c Codec) setRelationshipSingle(mStruct *mapping.ModelStruct, model mapping.Model, sField *mapping.StructField, value interface{}, options codec.UnmarshalOptions, name string) error {
	v, ok := value.(map[string]interface{})
	if !ok {
		return errors.WrapDetf(codec.ErrUnmarshal, "provided invalid model relation: '%s' field", name)
	}
	relationNode, _, err := c.unmarshalNode(sField.Relationship().RelatedModelStruct(), v, options)
	if err != nil {
		return err
	}
	sr, ok := model.(mapping.SingleRelationer)
	if !ok {
		return errors.WrapDetf(mapping.ErrModelNotImplements, "model: '%s' doesn't implement mapping.SingleRelationer interface", mStruct.String())
	}
	return sr.SetRelationModel(sField, relationNode)
}

func (c Codec) setRelationshipMany(mStruct *mapping.ModelStruct, model mapping.Model, sField *mapping.StructField, value interface{}, options codec.UnmarshalOptions, name string) error {
	values, ok := value.([]interface{})
	if !ok {
		return errors.WrapDetf(codec.ErrUnmarshal, "provided invalid model relation: '%s' field value", name)
	}
	var models []mapping.Model
	for _, single := range values {
		node, ok := single.(map[string]interface{})
		if !ok {
			return errors.WrapDetf(codec.ErrUnmarshal, "provided invalid model relation: '%s' field value", name)
		}
		model, _, err := c.unmarshalNode(sField.Relationship().RelatedModelStruct(), node, options)
		if err != nil {
			return err
		}
		models = append(models, model)
	}
	mr, ok := model.(mapping.MultiRelationer)
	if !ok {
		return errors.WrapDetf(mapping.ErrModelNotImplements, "model: '%s' doesn't implement mapping.SingleRelationer interface", mStruct.String())
	}
	return mr.SetRelationModels(sField, models...)
}

func setAttribute(sField *mapping.StructField, value interface{}, fielder mapping.Fielder) error {
	if sField.IsTime() && !(sField.IsTimePointer() && value == nil) {
		if sField.IsISO8601() {
			strVal, ok := value.(string)
			if !ok {
				return errors.WrapDet(mapping.ErrFieldValue, "invalid ISO8601 time field").
					WithDetailf("Time field: '%s' has invalid formatting.", sField.NeuronName())
			}
			t, err := time.Parse(strVal, codec.ISO8601TimeFormat)
			if err != nil {
				return errors.WrapDet(mapping.ErrFieldValue, "invalid ISO8601 time field").
					WithDetailf("Time field: '%s' has invalid formatting.", sField.NeuronName())
			}
			if sField.IsTimePointer() {
				value = &t
			} else {
				value = t
			}
		} else {
			var at int64
			switch av := value.(type) {
			case float64:
				at = int64(av)
			case int64:
				at = av
			case int:
				at = int64(av)
			default:
				return errors.WrapDet(mapping.ErrFieldValue, "invalid time field value").
					WithDetailf("Time field: '%s' has invalid value.", sField.NeuronName())
			}
			t := time.Unix(at, 0)
			if sField.IsTimePointer() {
				value = &t
			} else {
				value = t
			}
		}
	}
	if err := fielder.SetFieldValue(sField, value); err != nil {
		return err
	}
	return nil
}
