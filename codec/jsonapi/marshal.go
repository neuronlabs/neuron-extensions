package jsonapi

import (
	"encoding/json"
	"io"
	"path"
	"time"

	neuronCodec "github.com/neuronlabs/neuron/codec"
	"github.com/neuronlabs/neuron/errors"

	"github.com/neuronlabs/neuron/log"
	"github.com/neuronlabs/neuron/mapping"
)

// ErrorsPayload is a serializer struct for representing a valid JSON API errors payload.
type ErrorsPayload struct {
	JSONAPI map[string]interface{} `json:"jsonapi,omitempty"`
	Errors  []*neuronCodec.Error   `json:"errors"`
}

// Marshal marshals provided value 'v' into writer 'w'
func marshalPayload(w io.Writer, payload Payloader) error {
	err := json.NewEncoder(w).Encode(payload)
	if err != nil {
		return err
	}
	return nil
}

func (j *jsonapiCodec) visitModels(models []mapping.Model, linkOptions *neuronCodec.LinkOptions) (nodes []*Node, err error) {
	var mStruct *mapping.ModelStruct
	for _, model := range models {
		if model == nil {
			continue
		}
		mStruct, err = j.c.ModelStruct(model)
		if err != nil {
			return nil, err
		}
		node, err := visitModelNode(mStruct, model, linkOptions, mStruct.StructFields())
		if err != nil {
			return nil, err
		}
		nodes = append(nodes, node)
	}
	return nodes, nil
}

func (j *jsonapiCodec) visitPayloadModels(payload *neuronCodec.Payload) (nodes []*Node, err error) {
	nodes = make([]*Node, len(payload.Data))
	var mStruct *mapping.ModelStruct

	var commonFieldset mapping.FieldSet
	switch len(payload.FieldSets) {
	case 0:
		commonFieldset = payload.ModelStruct.StructFields()
		for _, relation := range payload.IncludedRelations {
			commonFieldset = append(commonFieldset, relation.StructField)
		}
	case 1:
		commonFieldset = payload.FieldSets[0]
		// Add all included relations.
		for _, relation := range payload.IncludedRelations {
			commonFieldset = append(commonFieldset, relation.StructField)
		}
	case len(payload.Data):
	default:
		return nil, errors.NewDetf(neuronCodec.ClassMarshalPayload, "provided invalid payload fieldset number")
	}

	for i, model := range payload.Data {
		var fieldSet mapping.FieldSet
		if commonFieldset != nil {
			fieldSet = commonFieldset
		} else {
			fieldSet = payload.FieldSets[i]
			for _, relation := range payload.IncludedRelations {
				fieldSet = append(fieldSet, relation.StructField)
			}
		}
		mStruct, err = j.c.ModelStruct(model)
		if err != nil {
			return nil, err
		}

		if mStruct != payload.ModelStruct {
			return nil, errors.NewDet(neuronCodec.ClassMarshal, "expecting payload with single model type - provided multiple type models")
		}
		node, err := visitModelNode(mStruct, model, payload.MarshalLinks, fieldSet)
		if err != nil {
			return nil, err
		}
		nodes[i] = node
	}
	return nodes, nil
}

func visitModelNode(mStruct *mapping.ModelStruct, model mapping.Model, linkOptions *neuronCodec.LinkOptions, fieldSet []*mapping.StructField) (node *Node, err error) {
	node = &Node{Type: mStruct.Collection()}

	// set primary
	primStruct := mStruct.Primary()
	if !primStruct.IsHidden() && !model.IsPrimaryKeyZero() {
		node.ID, err = model.GetPrimaryKeyStringValue()
		if err != nil {
			return nil, err
		}
	}

	var (
		fielder          mapping.Fielder
		multiRelationer  mapping.MultiRelationer
		singleRelationer mapping.SingleRelationer
		ok               bool
	)

	for _, field := range fieldSet {
		isOmitEmpty, isHidden, fieldName := getFieldFlags(field, mStruct)

		// if the field is mark as hidden or '-' do not marshal that field.
		if isHidden {
			if log.CurrentLevel().IsAllowed(log.LevelDebug3) {
				log.Debug3f("jsonapi marshal: %s - field: %s not marshaled - is hidden", mStruct, field.NeuronName())
			}
			continue
		}

		// extract field's value
		// marshal differently for different field type
		switch field.Kind() {
		case mapping.KindAttribute:
			if field.IsNestedStruct() {
				// node.Attributes[fieldName] = marshalNestedStructValue(field.Nested(), fieldValue).Interface()
				continue
			}

			if node.Attributes == nil {
				node.Attributes = make(map[string]interface{})
			}
			if fielder == nil {
				fielder, ok = model.(mapping.Fielder)
				if !ok {
					return nil, errors.New(mapping.ClassModelNotImplements, "model doesn't implement fielder interface")
				}
			}

			// Check if field has zero value.
			isZero, err := fielder.IsFieldZero(field)
			if err != nil {
				return nil, err
			}
			fieldValue, err := fielder.GetFieldValue(field)
			if err != nil {
				return nil, err
			}

			if isOmitEmpty && isZero {
				if log.CurrentLevel().IsAllowed(log.LevelDebug3) {
					log.Debug3f("jsonapi marshal: %s - field: %s empty value omitted", mStruct, field.NeuronName())
				}
				continue
			}

			if field.IsPtr() && isZero {
				node.Attributes[fieldName] = nil
				continue
			}

			if !field.IsTime() {
				node.Attributes[fieldName] = fieldValue
				continue
			}

			var t time.Time
			if field.IsPtr() {
				t = *fieldValue.(*time.Time)
			} else {
				t = fieldValue.(time.Time)
			}

			if t.IsZero() && isOmitEmpty {
				if log.CurrentLevel().IsAllowed(log.LevelDebug3) {
					log.Debug3f("jsonapi marshal: %s - field: %s empty value omitted", mStruct, field.NeuronName())
				}
				continue
			}

			if field.IsISO8601() {
				if log.CurrentLevel().IsAllowed(log.LevelDebug3) {
					log.Debug3f("jsonapi marshal: %s - field: %s marshal time field using ISO8601 format", mStruct, field.NeuronName())
				}
				node.Attributes[fieldName] = t.UTC().Format(neuronCodec.ISO8601TimeFormat)
			} else {
				node.Attributes[fieldName] = t.Unix()
			}
		case mapping.KindRelationshipMultiple:
			if multiRelationer == nil {
				multiRelationer, ok = model.(mapping.MultiRelationer)
				if !ok {
					return nil, errors.New(mapping.ClassModelNotImplements, "model doesn't implement MultiRelationer")
				}
			}

			relationLen, err := multiRelationer.GetRelationLen(field)
			if err != nil {
				return nil, err
			}
			if isOmitEmpty && relationLen == 0 {
				if log.CurrentLevel().IsAllowed(log.LevelDebug3) {
					log.Debug3f("jsonapi marshal: %s - field: %s empty value omitted", mStruct, field.NeuronName())
				}
				continue
			}
			if node.Relationships == nil {
				node.Relationships = make(map[string]interface{})
			}

			relations, err := multiRelationer.GetRelationModels(field)
			if err != nil {
				return nil, err
			}

			r := &RelationshipManyNode{
				Data: make([]*Node, relationLen),
			}
			var id string
			for i, relation := range relations {
				id, err = relation.GetPrimaryKeyStringValue()
				if err != nil {
					return nil, err
				}
				r.Data[i] = &Node{Type: relation.NeuronCollectionName(), ID: id}
			}

			if linkOptions != nil {
				if linkOptions.Type == neuronCodec.ResourceLink {
					link := make(map[string]interface{})
					link["self"] = path.Join(linkOptions.BaseURL, mStruct.Collection(), node.ID, "relationships", fieldName)
					link["related"] = path.Join(linkOptions.BaseURL, mStruct.Collection(), node.ID, fieldName)
					links := Links(link)
					r.Links = &links
				}
			}
			node.Relationships[fieldName] = r
		case mapping.KindRelationshipSingle:
			if singleRelationer == nil {
				singleRelationer, ok = model.(mapping.SingleRelationer)
				if !ok {
					return nil, errors.New(mapping.ClassModelNotImplements, "model doesn't implement SingleRelationer")
				}
			}
			relation, err := singleRelationer.GetRelationModel(field)
			if err != nil {
				return nil, err
			}
			if isOmitEmpty && relation == nil {
				if log.CurrentLevel().IsAllowed(log.LevelDebug3) {
					log.Debug3f("jsonapi marshal: %s - field: %s empty value omitted", mStruct, field.NeuronName())
				}
				continue
			}

			if node.Relationships == nil {
				node.Relationships = make(map[string]interface{})
			}

			r := &RelationshipOneNode{}
			if linkOptions != nil {
				if linkOptions.Type == neuronCodec.ResourceLink {
					link := make(map[string]interface{})
					link["self"] = path.Join(linkOptions.BaseURL, mStruct.Collection(), node.ID, "relationships", fieldName)
					link["related"] = path.Join(linkOptions.BaseURL, mStruct.Collection(), node.ID, fieldName)
					links := Links(link)
					r.Links = &links
				}
			}
			if relation != nil {
				id, err := relation.GetPrimaryKeyStringValue()
				if err != nil {
					return nil, err
				}
				r.Data = &Node{Type: relation.NeuronCollectionName(), ID: id}
			}
			node.Relationships[fieldName] = r
		}
	}

	if linkOptions != nil && linkOptions.Type == neuronCodec.ResourceLink {
		links := make(map[string]interface{})
		links["self"] = path.Join(linkOptions.BaseURL, mStruct.Collection(), node.ID)
		linksObj := Links(links)
		node.Links = &(linksObj)
	}
	return node, nil
}

func getFieldFlags(field *mapping.StructField, model *mapping.ModelStruct) (bool, bool, string) {
	// define marshal flags for given field
	isOmitEmpty := field.IsOmitEmpty()
	isHidden := field.IsHidden()
	fieldName := field.NeuronName()

	// extract jsonapi field tags if exists
	tags := field.ExtractCustomFieldTags(neuronCodec.StructTag, mapping.AnnotationSeparator, " ")
	// overwrite neuron marshal flags by the 'jsonapiCodec' flags
	for _, tag := range tags {
		switch tag.Key {
		case "-":
			isHidden = true
			if log.CurrentLevel().IsAllowed(log.LevelDebug3) {
				log.Debug3f("jsonapi marshal: %s - field: %s not marshaled by field jsonapi tag \"-\"", model, field.NeuronName())
			}
		case "omitempty":
			isOmitEmpty = true
			if log.CurrentLevel().IsAllowed(log.LevelDebug3) {
				log.Debug3f("jsonapi marshal: %s - field: %s is set to omitempty with jsonapi tag", model, field.NeuronName())
			}
		default:
			// the default key value would be the field name
			fieldName = tag.Key
			if log.CurrentLevel().IsAllowed(log.LevelDebug3) {
				log.Debug3f("jsonapi marshal: %s - field: %s marshaled with key: %s by jsonapi tag", model, field.NeuronName(), fieldName)
			}
		}
	}
	return isOmitEmpty, isHidden, fieldName
}
