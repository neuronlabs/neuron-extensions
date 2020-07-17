package jsonapi

import (
	"encoding/json"
	"io"
	"path"
	"time"

	"github.com/neuronlabs/neuron/codec"
	"github.com/neuronlabs/neuron/errors"

	"github.com/neuronlabs/neuron/log"
	"github.com/neuronlabs/neuron/mapping"
	"github.com/neuronlabs/neuron/query"
)

// ErrorsPayload is a serializer struct for representing a valid JSON API errors payload.
type ErrorsPayload struct {
	JSONAPI map[string]interface{} `json:"jsonapi,omitempty"`
	Errors  []*codec.Error         `json:"errors"`
}

// Marshal marshals provided value 'v' into writer 'w'
func marshalPayload(w io.Writer, payload Payloader) error {
	err := json.NewEncoder(w).Encode(payload)
	if err != nil {
		return err
	}
	return nil
}

func queryPayload(s *query.Scope, o *codec.MarshalOptions) (Payloader, error) {
	var (
		payload Payloader
		err     error
	)

	if len(s.Models) == 0 {
		if o != nil && o.SingleResult {
			payload = &ManyPayload{Data: []*Node{}}
		} else {
			payload = &SinglePayload{Data: nil}
		}
		return payload, nil
	}

	if len(s.Models) > 0 {
		payload, err = marshalQueryManyModels(s, o)
	} else {
		payload, err = marshalQuerySingleModel(s, o)
	}
	if err != nil {
		return nil, err
	}

	if o != nil && len(o.RelationQueries) > 0 {
		var includedNodes []*Node
		for _, q := range o.RelationQueries {
			nodes, err := includedRelationsNodes(q, o)
			if err != nil {
				return nil, err
			}
			includedNodes = append(includedNodes, nodes...)
		}
		payload.SetIncluded(includedNodes)
	}
	return payload, nil
}

func includedRelationsNodes(s *query.Scope, o *codec.MarshalOptions) ([]*Node, error) {
	fieldSet := s.FieldSet
	for _, included := range s.IncludedRelations {
		if len(included.Fieldset) == 1 && included.Fieldset[0].IsPrimary() {
			fieldSet = append(fieldSet, included.StructField)
		}
	}

	var nodes []*Node
	for _, model := range s.Models {
		if model == nil {
			continue
		}
		node, err := visitModelNode(s.ModelStruct, model, o, fieldSet)
		if err != nil {
			return nil, err
		}
		nodes = append(nodes, node)
	}
	return nodes, nil
}

func marshalQuerySingleModel(s *query.Scope, o *codec.MarshalOptions) (*SinglePayload, error) {
	var model mapping.Model
	if len(s.Models) > 0 {
		model = s.Models[0]
	}
	// Get the node for given model.
	fieldSet := s.FieldSet
	for _, include := range s.IncludedRelations {
		fieldSet = append(fieldSet, include.StructField)
	}

	n, err := visitModelNode(s.ModelStruct, model, o, fieldSet)
	if err != nil {
		return nil, err
	}

	var links *TopLinks
	if o != nil {
		switch o.Link.Type {
		case codec.ResourceLink:
			// By default the top level should contain 'self' value.
			links = &TopLinks{Self: path.Join(o.Link.BaseURL, o.Link.Collection, o.Link.RootID)}
		case codec.RelatedLink:
			links = &TopLinks{Self: path.Join(o.Link.BaseURL, o.Link.Collection, o.Link.RootID, o.Link.RelatedField)}
		case codec.RelationshipLink:
			links = &TopLinks{
				Self:    path.Join(o.Link.BaseURL, o.Link.Collection, o.Link.RootID, "relationships", o.Link.RelatedField),
				Related: path.Join(o.Link.BaseURL, o.Link.Collection, o.Link.RootID, o.Link.RelatedField),
			}
		}
	}
	return &SinglePayload{Data: n, Links: links}, nil
}

func marshalQueryManyModels(s *query.Scope, o *codec.MarshalOptions) (*ManyPayload, error) {
	n, err := visitQueryManyNodes(s, o)
	if err != nil {
		return nil, err
	}

	var (
		links *TopLinks
		meta  *codec.Meta
	)
	if o != nil {
		switch o.Link.Type {
		case codec.ResourceLink:
			links = &TopLinks{Self: path.Join(o.Link.BaseURL, o.Link.Collection)}
			links.SetPaginationLinks(o)
			if pLinks := o.Link.PaginationLinks; pLinks != nil {
				if pLinks.Total != 0 {
					meta = &codec.Meta{KeyTotal: pLinks.Total}
				}
			}
		case codec.RelatedLink:
			links = &TopLinks{Self: path.Join(o.Link.BaseURL, o.Link.Collection, o.Link.RootID, o.Link.RelatedField)}
			links.SetPaginationLinks(o)
			if pLinks := o.Link.PaginationLinks; pLinks != nil {
				if pLinks.Total != 0 {
					meta = &codec.Meta{KeyTotal: pLinks.Total}
				}
			}
		case codec.RelationshipLink:
			links = &TopLinks{
				Related: path.Join(o.Link.BaseURL, o.Link.Collection, o.Link.RootID, o.Link.RelatedField),
				Self:    path.Join(o.Link.BaseURL, o.Link.Collection, o.Link.RootID, "relationships", o.Link.RelatedField),
			}
			if pLinks := o.Link.PaginationLinks; pLinks != nil {
				if pLinks.Total != 0 {
					meta = &codec.Meta{KeyTotal: pLinks.Total}
				}
			}
		}
	}
	return &ManyPayload{Data: n, Links: links, Meta: meta}, nil
}

func visitQueryManyNodes(s *query.Scope, o *codec.MarshalOptions) ([]*Node, error) {
	nodes := make([]*Node, len(s.Models))
	fieldSet := s.FieldSet
	for _, included := range s.IncludedRelations {
		fieldSet = append(fieldSet, included.StructField)
	}
	for i, model := range s.Models {
		node, err := visitModelNode(s.ModelStruct, model, o, fieldSet)
		if err != nil {
			return nil, err
		}
		nodes[i] = node
	}
	return nodes, nil
}

func visitModels(mStruct *mapping.ModelStruct, models []mapping.Model, o *codec.MarshalOptions) ([]*Node, error) {
	var nodes []*Node
	for _, model := range models {
		if model == nil {
			continue
		}
		node, err := visitModelNode(mStruct, model, o, mStruct.Fields())
		if err != nil {
			return nil, err
		}
		nodes = append(nodes, node)
	}
	return nodes, nil
}

func visitModelNode(mStruct *mapping.ModelStruct, model mapping.Model, o *codec.MarshalOptions, fieldSet []*mapping.StructField) (node *Node, err error) {
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
				node.Attributes[fieldName] = t.UTC().Format(ISO8601TimeFormat)
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

			if o != nil {
				if o.Link.Type == codec.ResourceLink {
					link := make(map[string]interface{})
					link["self"] = path.Join(o.Link.BaseURL, mStruct.Collection(), node.ID, "relationships", fieldName)
					link["related"] = path.Join(o.Link.BaseURL, mStruct.Collection(), node.ID, fieldName)
					links := Links(link)
					r.Links = &links
				}
				if optionMeta, ok := o.RelationshipMeta[fieldName]; ok {
					r.Meta = &optionMeta
				}
			}
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
			if o != nil {
				if o.Link.Type == codec.ResourceLink {
					link := make(map[string]interface{})
					link["self"] = path.Join(o.Link.BaseURL, mStruct.Collection(), node.ID, "relationships", fieldName)
					link["related"] = path.Join(o.Link.BaseURL, mStruct.Collection(), node.ID, fieldName)
					links := Links(link)
					r.Links = &links
				}
				if optionMeta, ok := o.RelationshipMeta[fieldName]; ok {
					r.Meta = &optionMeta
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

	if o != nil && o.Link.Type == codec.ResourceLink {
		links := make(map[string]interface{})
		links["self"] = path.Join(o.Link.BaseURL, mStruct.Collection(), node.ID)
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
	tags := field.ExtractCustomFieldTags(codec.StructTag, mapping.AnnotationSeparator, " ")
	// overwrite neuron marshal flags by the 'codec' flags
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
