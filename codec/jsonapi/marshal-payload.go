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
func (j *jsonapiCodec) MarshalPayload(w io.Writer, payload *codec.Payload) error {
	nodes, err := j.visitPayloadModels(payload)
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
			payload.Meta[KeyTotal] = payload.PaginationLinks.Total
		}
	}
	if len(nodes) == 1 && payload.MarshalSingularFormat {
		payloader = &SinglePayload{Data: nodes[0], Meta: payload.Meta, Links: topLinks}
	} else {
		payloader = &ManyPayload{Data: nodes, Meta: payload.Meta, Links: topLinks}
	}
	if len(payload.IncludedRelations) != 0 {
		included, err := j.extractIncludedNodes(payload)
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

func (j *jsonapiCodec) extractIncludedNodes(payload *codec.Payload) ([]*Node, error) {
	collectionUniqueNodes := map[*mapping.ModelStruct]map[interface{}]*Node{}
	var nodes []*Node
	for _, included := range payload.IncludedRelations {
		for _, model := range payload.Data {
			err := j.extractIncludedModelNode(collectionUniqueNodes, included, model, payload.MarshalLinks)
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

func (j *jsonapiCodec) extractIncludedModelNode(collectionUniqueNodes map[*mapping.ModelStruct]map[interface{}]*Node, included *query.IncludedRelation, model mapping.Model, marshalLinks *codec.LinkOptions) error {
	switch included.StructField.Kind() {
	case mapping.KindRelationshipMultiple:
		mr, ok := model.(mapping.MultiRelationer)
		if !ok {
			return errors.NewDetf(mapping.ClassModelNotImplements, "model: %T doesn't implement MultiRelationer", model)
		}
		relations, err := mr.GetRelationModels(included.StructField)
		if err != nil {
			return err
		}
		relationStruct := included.StructField.ModelStruct()
		thisNodes := collectionUniqueNodes[relationStruct]
		if thisNodes == nil {
			thisNodes = map[interface{}]*Node{}
			collectionUniqueNodes[relationStruct] = thisNodes
		}

		for _, relation := range relations {
			id := relation.GetPrimaryKeyHashableValue()
			if _, ok = thisNodes[id]; ok {
				continue
			}
			node, err := visitModelNode(relationStruct, relation, marshalLinks, included.Fieldset)
			if err != nil {
				return err
			}
			thisNodes[id] = node
		}
	case mapping.KindRelationshipSingle:
		sr, ok := model.(mapping.SingleRelationer)
		if !ok {
			return errors.NewDetf(mapping.ClassModelNotImplements, "model: %T doesn't implement SingleRelationer", model)
		}
		relation, err := sr.GetRelationModel(included.StructField)
		if err != nil {
			return err
		}
		relationStruct := included.StructField.ModelStruct()
		thisNodes := collectionUniqueNodes[relationStruct]
		if thisNodes == nil {
			thisNodes = map[interface{}]*Node{}
			collectionUniqueNodes[relationStruct] = thisNodes
		}
		id := relation.GetPrimaryKeyHashableValue()
		if _, ok = thisNodes[id]; ok {
			return nil
		}
		node, err := visitModelNode(relationStruct, relation, marshalLinks, included.Fieldset)
		if err != nil {
			return err
		}
		thisNodes[id] = node
	}
	return nil
}
