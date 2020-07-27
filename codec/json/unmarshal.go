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
	"github.com/neuronlabs/neuron/query"
)

type payloader interface {
	nodes() []modelNode
	pagination() *codec.PaginationLinks
}

type singlePayload struct {
	Pagination *codec.PaginationLinks
	Data       modelNode `json:"data"`
}

func (s singlePayload) nodes() []modelNode {
	return []modelNode{s.Data}
}
func (s singlePayload) pagination() *codec.PaginationLinks {
	return s.Pagination
}

type multiPayload struct {
	Pagination *codec.PaginationLinks
	Data       []modelNode `json:"data"`
}

func (m multiPayload) nodes() []modelNode {
	return m.Data
}

func (m multiPayload) pagination() *codec.PaginationLinks {
	return m.Pagination
}

type modelNode map[string]interface{}

// UnmarshalModels implements codec.Codec interface.
func (j *jsonCodec) UnmarshalModels(r io.Reader, modelStruct *mapping.ModelStruct, options *codec.UnmarshalOptions) ([]mapping.Model, error) {
	if options == nil {
		options = &codec.UnmarshalOptions{}
	}
	// Unmarshall all the data and store it in the bytes.Reader.
	data, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, err
	}

	// get the payload.
	payload, err := getPayload(data)
	if err != nil {
		return nil, err
	}
	var models []mapping.Model
	for _, n := range payload.nodes() {
		model, _, err := unmarshalNode(modelStruct, n, options)
		if err != nil {
			return nil, err
		}
		models = append(models, model)
	}
	return models, nil
}

var _ codec.QueryUnmarshaler = c

// UnmarshalQuery implements codec.QueryUnmarshaler.
func (j *jsonCodec) UnmarshalQuery(r io.Reader, modelStruct *mapping.ModelStruct, options *codec.UnmarshalOptions) (*query.Scope, error) {
	if options == nil {
		options = &codec.UnmarshalOptions{}
	}
	// Unmarshall all the data and store it in the bytes.Reader.
	data, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, err
	}

	// get the payload.
	payload, err := getPayload(data)
	if err != nil {
		return nil, err
	}
	q := query.NewScope(modelStruct)
	var models []mapping.Model
	nodes := payload.nodes()
	for i, n := range nodes {
		model, fieldSet, err := unmarshalNode(modelStruct, n, options)
		if err != nil {
			return nil, err
		}
		if len(nodes) == 1 {
			q.FieldSet = fieldSet
		} else {
			if q.BulkFieldSets == nil {
				q.BulkFieldSets = &mapping.BulkFieldSet{}
			}
			q.BulkFieldSets.Add(fieldSet, i)
		}
		models = append(models, model)
	}
	return nil, nil
}

func getPayload(data []byte) (payloader, error) {
	br := bytes.NewReader(data)
	rn, _, err := br.ReadRune()
	if err != nil {
		return nil, errors.New(codec.ClassUnmarshalDocument, "provided invalid document")
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
		return nil, errors.New(codec.ClassUnmarshalDocument, "provided invalid document")
	}
}

func unmarshalNode(mStruct *mapping.ModelStruct, node modelNode, options *codec.UnmarshalOptions) (mapping.Model, mapping.FieldSet, error) {
	if !options.StrictUnmarshal {
		return unmarshalNodeNonStrict(mStruct, node, options)
	}
	return unmarshalNodeStrict(mStruct, node, options)
}

func unmarshalNodeNonStrict(mStruct *mapping.ModelStruct, n modelNode, options *codec.UnmarshalOptions) (mapping.Model, mapping.FieldSet, error) {
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
		fieldAnnotation := getFieldAnnotation(sField)
		value, ok = n[fieldAnnotation.Name]
		if !ok {
			continue
		}
		if fieldAnnotation.IsHidden {
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
					return nil, nil, errors.NewDetf(mapping.ClassModelNotImplements, "model: '%s' doesn't implement mapping.Fielder interface", mStruct.String())
				}
			}
			if err = setAttribute(sField, value, fielder); err != nil {
				return nil, nil, err
			}
		case mapping.KindRelationshipMultiple:
			if value == nil {
				continue
			}
			if err = setRelationshipMany(mStruct, model, sField, value, options, fieldAnnotation.Name); err != nil {
				return nil, nil, err
			}
		case mapping.KindRelationshipSingle:
			if value == nil {
				continue
			}
			if err = setRelationshipSingle(mStruct, model, sField, value, options, fieldAnnotation.Name); err != nil {
				return nil, nil, err
			}
		}
	}
	return model, fieldSet, nil
}

func unmarshalNodeStrict(mStruct *mapping.ModelStruct, node modelNode, options *codec.UnmarshalOptions) (mapping.Model, mapping.FieldSet, error) {
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
			return nil, nil, errors.NewDetf(codec.ClassUnmarshal, "provided invalid field: '%s'", field)
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
					return nil, nil, errors.NewDetf(mapping.ClassModelNotImplements, "model: '%s' doesn't implement mapping.Fielder interface", mStruct.String())
				}
			}
			if err = setAttribute(sField, value, fielder); err != nil {
				return nil, nil, err
			}
		case mapping.KindRelationshipMultiple:
			if value == nil {
				continue
			}
			if err = setRelationshipMany(mStruct, model, sField, value, options, fieldAnnotation.Name); err != nil {
				return nil, nil, err
			}
		case mapping.KindRelationshipSingle:
			if value == nil {
				continue
			}
			if err = setRelationshipSingle(mStruct, model, sField, value, options, fieldAnnotation.Name); err != nil {
				return nil, nil, err
			}
		}
	}
	return model, fieldSet, nil
}

func setRelationshipSingle(mStruct *mapping.ModelStruct, model mapping.Model, sField *mapping.StructField, value interface{}, options *codec.UnmarshalOptions, name string) error {
	v, ok := value.(map[string]interface{})
	if !ok {
		return errors.NewDetf(codec.ClassUnmarshal, "provided invalid model relation: '%s' field", name)
	}
	relationNode, _, err := unmarshalNode(sField.Relationship().Struct(), v, options)
	if err != nil {
		return err
	}
	sr, ok := model.(mapping.SingleRelationer)
	if !ok {
		return errors.NewDetf(mapping.ClassModelNotImplements, "model: '%s' doesn't implement mapping.SingleRelationer interface", mStruct.String())
	}
	return sr.SetRelationModel(sField, relationNode)
}

func setRelationshipMany(mStruct *mapping.ModelStruct, model mapping.Model, sField *mapping.StructField, value interface{}, options *codec.UnmarshalOptions, name string) error {
	values, ok := value.([]interface{})
	if !ok {
		return errors.NewDetf(codec.ClassUnmarshal, "provided invalid model relation: '%s' field value", name)
	}
	var models []mapping.Model
	for _, single := range values {
		node, ok := single.(map[string]interface{})
		if !ok {
			return errors.NewDetf(codec.ClassUnmarshal, "provided invalid model relation: '%s' field value", name)
		}
		model, _, err := unmarshalNode(sField.Relationship().Struct(), node, options)
		if err != nil {
			return err
		}
		models = append(models, model)
	}
	mr, ok := model.(mapping.MultiRelationer)
	if !ok {
		return errors.NewDetf(mapping.ClassModelNotImplements, "model: '%s' doesn't implement mapping.SingleRelationer interface", mStruct.String())
	}
	return mr.SetRelationModels(sField, models...)
}

func setAttribute(sField *mapping.StructField, value interface{}, fielder mapping.Fielder) error {
	if sField.IsTime() && !(sField.IsTimePointer() && value == nil) {
		if sField.IsISO8601() {
			strVal, ok := value.(string)
			if !ok {
				return errors.NewDet(mapping.ClassFieldValue, "invalid ISO8601 time field").
					WithDetailf("Time field: '%s' has invalid formatting.", sField.NeuronName())
			}
			t, err := time.Parse(strVal, codec.ISO8601TimeFormat)
			if err != nil {
				return errors.NewDet(mapping.ClassFieldValue, "invalid ISO8601 time field").
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
				return errors.NewDet(mapping.ClassFieldValue, "invalid time field value").
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

func getFieldAnnotation(sField *mapping.StructField) codec.FieldAnnotations {
	annotations := codec.ExtractFieldAnnotations(sField, "codec")
	if annotations.Name == "" {
		annotations.Name = sField.NeuronName()
	}
	if !annotations.IsHidden {
		annotations.IsHidden = sField.IsHidden()
	}
	if !annotations.IsOmitEmpty {
		annotations.IsOmitEmpty = sField.IsOmitEmpty()
	}
	return annotations
}
