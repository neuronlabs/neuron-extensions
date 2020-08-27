package jsonapi

import (
	"encoding/json"
	"io"

	"github.com/neuronlabs/neuron/codec"
	"github.com/neuronlabs/neuron/errors"
	"github.com/neuronlabs/neuron/log"
	"github.com/neuronlabs/neuron/mapping"
	"github.com/neuronlabs/neuron/query"
)

// MarshalPayload implements codec.PayloadMarshaler interface.
func (c Codec) MarshalPayload(w io.Writer, payload *codec.Payload, options ...codec.MarshalOption) error {
	o := &codec.MarshalOptions{}
	for _, option := range options {
		option(o)
	}
	nodes, err := c.visitPayloadModels(payload, o)
	if err != nil {
		log.Debug2f("visitModels failed: %v", err)
		return err
	}

	var (
		payloader Payloader
		topLinks  *TopLinks
	)
	if payload.PaginationLinks != nil {
		topLinks = &TopLinks{}
		topLinks.SetPaginationLinks(payload.PaginationLinks)
		if payload.PaginationLinks.Total != 0 {
			if payload.Meta == nil {
				payload.Meta = codec.Meta{}
			}
			payload.Meta[KeyTotal] = payload.PaginationLinks.Total
		}
	}
	if len(nodes) == 1 && o.SingleResult {
		payloader = &SinglePayload{Data: nodes[0], Meta: payload.Meta, Links: topLinks}
	} else {
		payloader = &ManyPayload{Data: nodes, Meta: payload.Meta, Links: topLinks}
	}
	if len(payload.IncludedRelations) != 0 {
		included, err := c.extractIncludedNodes(payload, o)
		if err != nil {
			return err
		}
		payloader.SetIncluded(included)
	}
	if err = json.NewEncoder(w).Encode(payloader); err != nil {
		return err
	}
	return nil

}

func (c Codec) extractIncludedNodes(payload *codec.Payload, options *codec.MarshalOptions) ([]*Node, error) {
	collectionUniqueNodes := map[*mapping.ModelStruct]map[interface{}]*Node{}
	var nodes []*Node
	for _, included := range payload.IncludedRelations {
		for _, model := range payload.Data {
			err := c.extractIncludedModelNode(collectionUniqueNodes, included, model, options.Link)
			if err != nil {
				return nil, err
			}
		}
	}
	for _, modelNodes := range collectionUniqueNodes {
		for _, node := range modelNodes {
			nodes = append(nodes, node)
		}
	}
	return nodes, nil
}

func (c Codec) extractIncludedModelNode(collectionUniqueNodes map[*mapping.ModelStruct]map[interface{}]*Node, included *query.IncludedRelation, model mapping.Model, marshalLinks codec.LinkOptions) error {
	switch included.StructField.Kind() {
	case mapping.KindRelationshipMultiple:
		mr, ok := model.(mapping.MultiRelationer)
		if !ok {
			return errors.WrapDetf(mapping.ErrModelNotImplements, "model: %T doesn't implement MultiRelationer", model)
		}
		relations, err := mr.GetRelationModels(included.StructField)
		if err != nil {
			return err
		}
		relationStruct := included.StructField.Relationship().RelatedModelStruct()
		thisNodes := collectionUniqueNodes[relationStruct]
		if thisNodes == nil {
			thisNodes = map[interface{}]*Node{}
			collectionUniqueNodes[relationStruct] = thisNodes
		}

		for _, relation := range relations {
			if relation == nil {
				continue
			}
			for _, subIncluded := range included.IncludedRelations {
				if err = c.extractIncludedModelNode(collectionUniqueNodes, subIncluded, relation, marshalLinks); err != nil {
					return err
				}
			}
			id := relation.GetPrimaryKeyHashableValue()
			if _, ok = thisNodes[id]; ok {
				continue
			}
			node, err := visitModelNode(relationStruct, relation, marshalLinks, included.Fieldset...)
			if err != nil {
				return err
			}
			thisNodes[id] = node
		}
	case mapping.KindRelationshipSingle:
		sr, ok := model.(mapping.SingleRelationer)
		if !ok {
			return errors.WrapDetf(mapping.ErrModelNotImplements, "model: %T doesn't implement SingleRelationer", model)
		}
		relation, err := sr.GetRelationModel(included.StructField)
		if err != nil {
			return err
		}
		if relation == nil {
			return nil
		}
		relationStruct := included.StructField.Relationship().RelatedModelStruct()
		thisNodes := collectionUniqueNodes[relationStruct]
		if thisNodes == nil {
			thisNodes = map[interface{}]*Node{}
			collectionUniqueNodes[relationStruct] = thisNodes
		}
		id := relation.GetPrimaryKeyHashableValue()
		if _, ok = thisNodes[id]; ok {
			return nil
		}
		node, err := visitModelNode(relationStruct, relation, marshalLinks, included.Fieldset...)
		if err != nil {
			return err
		}
		thisNodes[id] = node

		for _, subIncluded := range included.IncludedRelations {
			if err = c.extractIncludedModelNode(collectionUniqueNodes, subIncluded, relation, marshalLinks); err != nil {
				return err
			}
		}
	}
	return nil
}
