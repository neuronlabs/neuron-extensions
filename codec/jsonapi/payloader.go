package jsonapi

import (
	neuronCodec "github.com/neuronlabs/neuron/codec"
)

// Payloader is used to encapsulate the One and Many payload types
type Payloader interface {
	GetNodes() []*Node
	GetIncluded() []*Node
	SetLinks(links *TopLinks)
	SetMeta(meta *neuronCodec.Meta)
	SetIncluded(included []*Node)
}

// compile time check for the interface implementation.
var _ Payloader = &SinglePayload{}

// SinglePayload is used to represent a generic JSON API payload where a single
// resource (Node) was included as an {} in the "data" key
type SinglePayload struct {
	Links    *TopLinks         `json:"links,omitempty"`
	Meta     *neuronCodec.Meta `json:"meta,omitempty"`
	Data     *Node             `json:"data"`
	Included []*Node           `json:"included,omitempty"`
}

func (p *SinglePayload) GetNodes() []*Node {
	return []*Node{p.Data}
}

// GetIncluded implements Payloader interface.
func (p *SinglePayload) GetIncluded() []*Node {
	return p.Included
}

// SetIncluded sets the included data for the provided payload.
// Implements Payloader interface.
func (p *SinglePayload) SetIncluded(included []*Node) {
	p.Included = included
}

// SetLinks sets the links for the single payload.
// Implements Payloader interface.
func (p *SinglePayload) SetLinks(links *TopLinks) {
	p.Links = links
}

// SetMeta sets the meta for the single payload.
// Implements Payloader interface.
func (p *SinglePayload) SetMeta(meta *neuronCodec.Meta) {
	p.Meta = meta
}

// compile time check for the interface implementation.
var _ Payloader = &ManyPayload{}

// ManyPayload is used to represent a generic JSON API payload where many
// resources (Nodes) were included in an [] in the "data" key
type ManyPayload struct {
	Links    *TopLinks         `json:"links,omitempty"`
	Meta     *neuronCodec.Meta `json:"meta,omitempty"`
	Data     []*Node           `json:"data"`
	Included []*Node           `json:"included,omitempty"`
}

func (p *ManyPayload) GetNodes() []*Node {
	return p.Data
}

// GetIncluded implements Payloader.
func (p *ManyPayload) GetIncluded() []*Node {
	return p.Included
}

// SetIncluded sets the included data for the provided payload.
// Implements Payloader interface.
func (p *ManyPayload) SetIncluded(included []*Node) {
	p.Included = included
}

// SetLinks sets the links for the ManyPaload.
// Implements Payloader interface.
func (p *ManyPayload) SetLinks(links *TopLinks) {
	p.Links = links
}

// SetMeta sets the meta for the single payload.
// Implements Payloader interface.
func (p *ManyPayload) SetMeta(meta *neuronCodec.Meta) {
	p.Meta = meta
}

// Meta is used to represent a `meta` object.
// http://jsonapi.org/format/#document-meta
type Meta map[string]interface{}

// Node is used to represent a generic JSON API Resource.
type Node struct {
	Type          string                 `json:"type"`
	ID            string                 `json:"id,omitempty"`
	Attributes    map[string]interface{} `json:"attributes,omitempty"`
	Relationships map[string]interface{} `json:"relationships,omitempty"`
	Links         *Links                 `json:"links,omitempty"`
	Meta          *neuronCodec.Meta      `json:"meta,omitempty"`
}

// RelationshipOneNode is used to represent a generic single JSON API relation.
type RelationshipOneNode struct {
	Data  *Node             `json:"data"`
	Links *Links            `json:"links,omitempty"`
	Meta  *neuronCodec.Meta `json:"meta,omitempty"`
}

// RelationshipManyNode is used to represent a generic has many JSON API relation.
type RelationshipManyNode struct {
	Data  []*Node           `json:"data"`
	Links *Links            `json:"links,omitempty"`
	Meta  *neuronCodec.Meta `json:"meta,omitempty"`
}
