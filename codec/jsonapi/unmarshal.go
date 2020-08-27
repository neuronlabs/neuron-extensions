package jsonapi

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"reflect"
	"time"

	"github.com/neuronlabs/neuron/codec"
	"github.com/neuronlabs/neuron/errors"
	"github.com/neuronlabs/neuron/log"
	"github.com/neuronlabs/neuron/mapping"
	"github.com/neuronlabs/neuron/query"
)

type storeKey struct{}

var (
	StoreKeyMarshalLinks = &storeKey{}
)

// ParamLinks defines the links parameter name.
const (
	ParamLinks      = "links"
	ParamPageSize   = "page[size]"
	ParamPageNumber = "page[number]"
)

func (c Codec) newIncludedRelation(sField *mapping.StructField, fields map[*mapping.ModelStruct]mapping.FieldSet) (includedRelation *query.IncludedRelation) {
	includedRelation = &query.IncludedRelation{StructField: sField}
	fs, ok := fields[sField.Relationship().RelatedModelStruct()]
	if ok {
		for _, field := range fs {
			switch field.Kind() {
			case mapping.KindAttribute, mapping.KindRelationshipMultiple, mapping.KindRelationshipSingle:
				includedRelation.Fieldset = append(includedRelation.Fieldset, field)
			}
		}
	} else {
		// By default set full field set with all relationships.
		includedRelation.Fieldset = append(sField.Relationship().RelatedModelStruct().Attributes(), sField.Relationship().RelatedModelStruct().RelationFields()...)
	}
	return includedRelation
}

func unmarshalPayload(in io.Reader, options *codec.UnmarshalOptions) (Payloader, error) {
	data, err := ioutil.ReadAll(in)
	if err != nil {
		return nil, err
	}
	r := bytes.NewReader(data)
	dec := json.NewDecoder(r)

	t, err := dec.Token()
	if err != nil {
		return nil, err
	}

	if t != json.Delim('{') {
		return nil, errors.Wrap(codec.ErrUnmarshalDocument, "invalid document")
	}
	if err := unmarshalPayloadFindData(dec); err != nil {
		return nil, errors.Wrap(codec.ErrUnmarshalDocument, "invalid input")
	}

	t, err = dec.Token()
	if err != nil {
		return nil, err
	}
	var payloader Payloader
	switch t {
	case json.Delim('{'):
		payloader = &SinglePayload{}
	case json.Delim('['), nil:
		payloader = &ManyPayload{}
	default:
		return nil, errors.Wrap(codec.ErrUnmarshalDocument, "invalid input")
	}
	r.Seek(0, io.SeekStart)
	dec = json.NewDecoder(r)
	if options.StrictUnmarshal {
		dec.DisallowUnknownFields()
	}
	if err = dec.Decode(payloader); err != nil {
		return nil, err
	}
	return payloader, nil
}

func unmarshalPayloadFindData(dec *json.Decoder) (err error) {
	var (
		openBracketCount int
		t                json.Token
	)
	for {
		t, err = dec.Token()
		if err != nil {
			return err
		}

		switch tk := t.(type) {
		case string:
			if tk == "data" && openBracketCount == 0 {
				return nil
			}
		case json.Delim:
			switch tk {
			case json.Delim('{'):
				openBracketCount++
			case json.Delim('}'):
				openBracketCount--
			}
		}
	}
}

var (
	singlePayloadType = reflect.TypeOf(SinglePayload{})
	manyPayloadType   = reflect.TypeOf(ManyPayload{})
)

func unmarshalHandleDecodeError(err error) error {
	// handle the incoming error
	switch e := err.(type) {
	case *json.SyntaxError:
		err := errors.WrapDet(codec.ErrUnmarshalDocument, "syntax error").
			WithDetailf("Document syntax error: '%s'. At data offset: '%d'", e.Error(), e.Offset)
		return err
	case *json.UnmarshalTypeError:
		switch e.Type {
		case singlePayloadType, manyPayloadType:
			return errors.WrapDet(codec.ErrUnmarshalDocument, "invalid jsonapi document syntax")
		}
		err := errors.WrapDet(codec.ErrUnmarshal, "invalid field type")
		var fieldType string
		switch e.Field {
		case "id", "type", "client-id":
			fieldType = e.Type.String()
		case "relationships", "attributes", "links", "meta":
			fieldType = "object"
		}
		return err.WithDetailf("Invalid type for: '%s' field. Required type '%s' but is: '%v'", e.Field, fieldType, e.Value)
	default:
		if e == io.EOF || e == io.ErrUnexpectedEOF {
			err := errors.WrapDet(codec.ErrUnmarshalDocument, "unexpected end of file occurred").
				WithDetailf("invalid document syntax")
			return err
		}
		return errors.WrapDetf(codec.ErrUnmarshal, "unknown unmarshal error: %s", e.Error())
	}
}

func unmarshalNode(mStruct *mapping.ModelStruct, data *Node, model mapping.Model, included map[string]*Node, options *codec.UnmarshalOptions) (fieldSet mapping.FieldSet, err error) {
	if data.Type != model.NeuronCollectionName() {
		err := errors.WrapDet(codec.ErrUnmarshal, "unmarshal collection name doesn't match the root struct").
			WithDetailf("unmarshal collection: '%s' doesn't match root collection:'%s'", data.Type, model.NeuronCollectionName())
		return nil, err
	}
	// Set primary key value.
	if data.ID != "" {
		fieldSet = append(fieldSet, mStruct.Primary())
		if err := model.SetPrimaryKeyStringValue(data.ID); err != nil {
			if errors.Is(err, mapping.ErrFieldValue) {
				return nil, errors.WrapDet(codec.ErrUnmarshalFieldValue, "provided invalid id value").
					WithDetail("provided invalid 'id' value")
			}
			return nil, err
		}
	}

	// Set attributes.
	if data.Attributes != nil {
		fielder, isFielder := model.(mapping.Fielder)
		if !isFielder {
			if len(mStruct.Attributes()) > 0 {
				return nil, errors.Wrap(errors.ErrInternal, "provided model is not a Fielder")
			} else if options.StrictUnmarshal {
				return nil, errors.Wrap(codec.ErrUnmarshal, "provided model doesn't have any attributes")
			}
		} else {
			// Iterate over the data attributes
			for attrName, attrValue := range data.Attributes {
				modelAttr, ok := mStruct.Attribute(attrName)
				if !ok {
					// Try to find the field by the codec struct tag.
					for _, field := range mStruct.Attributes() {
						if field.CodecName() == attrName {
							modelAttr = field
							break
						}
					}
				}

				if !ok || modelAttr.CodecSkip() {
					if options.StrictUnmarshal {
						return nil, errors.WrapDet(codec.ErrUnmarshalFieldName, "unknown field name").
							WithDetailf("provided unknown field name: '%s', for the collection: '%s'.", attrName, data.Type)
					}
					continue
				}

				fieldSet = append(fieldSet, modelAttr)
				if modelAttr.IsTime() && !(modelAttr.IsTimePointer() && attrValue == nil) {
					if modelAttr.CodecISO8601() {
						strVal, ok := attrValue.(string)
						if !ok {
							return nil, errors.WrapDet(codec.ErrUnmarshalFieldValue, "invalid ISO8601 time field").
								WithDetailf("Time field: '%s' has invalid formatting.", modelAttr.NeuronName())
						}
						t, err := time.Parse(codec.ISO8601TimeFormat, strVal)
						if err != nil {
							return nil, errors.WrapDet(codec.ErrUnmarshalFieldValue, "invalid ISO8601 time field").
								WithDetailf("Time field: '%s' has invalid formatting.", modelAttr.NeuronName())
						}
						if modelAttr.IsTimePointer() {
							attrValue = &t
						} else {
							attrValue = t
						}
					} else {
						var at int64
						switch av := attrValue.(type) {
						case float64:
							at = int64(av)
						case int64:
							at = av
						case int:
							at = int64(av)
						default:
							return nil, errors.WrapDet(codec.ErrUnmarshalFieldValue, "invalid time field value").
								WithDetailf("Time field: '%s' has invalid value.", modelAttr.NeuronName())
						}
						t := time.Unix(at, 0)
						if modelAttr.IsTimePointer() {
							attrValue = &t
						} else {
							attrValue = t
						}
					}
				}
				if err := fielder.SetFieldValue(modelAttr, attrValue); err != nil {
					if errors.Is(err, mapping.ErrFieldValue) {
						return nil, errors.WrapDet(codec.ErrUnmarshalFieldValue, "provided invalid field value").
							WithDetailf("provided invalid field: '%s' value", modelAttr.NeuronName())
					}
					return nil, err
				}
			}
		}
	}

	if data.Relationships != nil {
		for relName, relValue := range data.Relationships {
			relationshipStructField, ok := mStruct.RelationByName(relName)
			if !ok {
				for _, field := range mStruct.RelationFields() {
					if field.CodecName() == relName {
						relationshipStructField = field
						break
					}
				}
			}
			// Try to find the field by the jsonapi struct tag.
			if !ok || (ok && relationshipStructField.CodecSkip()) {
				if options.StrictUnmarshal {
					return nil, errors.WrapDet(codec.ErrUnmarshalFieldName, "unknown field name").
						WithDetailf("Provided unknown field name: '%s', for the collection: '%s'.", relName, data.Type)
				}
				continue
			}

			fieldSet = append(fieldSet, relationshipStructField)
			if relationshipStructField.Kind() == mapping.KindRelationshipMultiple {
				mr, ok := model.(mapping.MultiRelationer)
				if !ok {
					return nil, errors.Wrap(errors.ErrInternal, "model is not a multi relationer")
				}
				// to-many relationship
				relationship := new(RelationshipManyNode)

				buf := bytes.NewBuffer(nil)
				if err := json.NewEncoder(buf).Encode(data.Relationships[relName]); err != nil {
					log.Debug2f("UnmarshalNode.relationshipMultiple json.Encode failed. %v", err)
					return nil, errors.WrapDet(codec.ErrUnmarshalFieldValue, "invalid relationship format").
						WithDetailf("The value for the relationship: '%s' is of invalid form.", relName)
				}

				if err := json.NewDecoder(buf).Decode(relationship); err != nil {
					log.Debug2f("UnmarshalNode.relationshipMultiple json.Encode failed. %v", err)
					return nil, errors.WrapDet(codec.ErrUnmarshal, "invalid relationship format").
						WithDetailf("The value for the relationship: '%s' is of invalid form.", relName)
				}

				relStruct := relationshipStructField.Relationship().RelatedModelStruct()
				for _, n := range relationship.Data {
					relModel := mapping.NewModel(relStruct)
					if _, err := unmarshalNode(relStruct, fullNode(n, included), relModel, included, options); err != nil {
						log.Debug2f("unmarshalNode.RelationshipMany - unmarshalNode failed. %v", err)
						return nil, err
					}
					if err := mr.AddRelationModel(relationshipStructField, relModel); err != nil {
						if errors.Is(err, mapping.ErrInvalidRelationValue) {
							return nil, errors.WrapDet(codec.ErrUnmarshalFieldValue, "provided invalid relation value").
								WithDetailf("provided invalid relation '%s' value", relationshipStructField.NeuronName())
						}
						return nil, err
					}
				}
			} else if relationshipStructField.Kind() == mapping.KindRelationshipSingle {
				sr, ok := model.(mapping.SingleRelationer)
				if !ok {
					return nil, errors.Wrap(errors.ErrInternal, "provided model is not a single relationer")
				}
				relationship := new(RelationshipOneNode)
				buf := bytes.NewBuffer(nil)

				if err := json.NewEncoder(buf).Encode(relValue); err != nil {
					log.Debug2f("UnmarshalNode.relationshipSingle json.Encode failed. %v", err)
					return nil, errors.WrapDet(codec.ErrUnmarshalFieldValue, "invalid relationship format").
						WithDetailf("The value for the relationship: '%s' is of invalid form.", relName)
				}

				if err := json.NewDecoder(buf).Decode(relationship); err != nil {
					log.Debug2f("UnmarshalNode.RelationshipSingle json.Decode failed. %v", err)
					return nil, errors.WrapDet(codec.ErrUnmarshalFieldValue, "invalid relationship format").
						WithDetailf("The value for the relationship: '%s' is of invalid form.", relName)
				}

				if relationship.Data == nil {
					continue
				}

				relStruct := relationshipStructField.Relationship().RelatedModelStruct()
				relModel := mapping.NewModel(relStruct)

				if _, err := unmarshalNode(relStruct, fullNode(relationship.Data, included), relModel, included, options); err != nil {
					return nil, err
				}
				if err := sr.SetRelationModel(relationshipStructField, relModel); err != nil {
					if errors.Is(err, mapping.ErrInvalidRelationValue) {
						return nil, errors.WrapDet(codec.ErrUnmarshalFieldValue, "provided invalid relation value").
							WithDetailf("provided invalid relation '%s' value", relationshipStructField.NeuronName())
					}
					return nil, err
				}
			}
		}
	}
	return fieldSet, nil
}

func fullNode(n *Node, included map[string]*Node) *Node {
	if included == nil {
		return n
	}
	includedKey := includedKeyFunc(n)
	if in, ok := included[includedKey]; ok {
		return in
	}
	return n
}

func includedKeyFunc(n *Node) string {
	return fmt.Sprintf("%s,%s", n.Type, n.ID)
}
