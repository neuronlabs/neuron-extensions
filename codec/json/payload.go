package json

import (
	"github.com/neuronlabs/neuron/codec"
)

type payloader interface {
	nodes() []modelNode
	pagination() *codec.PaginationLinks
	meta() codec.Meta
}

type singlePayload struct {
	Pagination *codec.PaginationLinks `json:"links"`
	Data       modelNode              `json:"data"`
	Meta       codec.Meta             `json:"meta"`
}

func (s singlePayload) nodes() []modelNode {
	return []modelNode{s.Data}
}
func (s singlePayload) pagination() *codec.PaginationLinks {
	return s.Pagination
}

func (s singlePayload) meta() codec.Meta {
	return s.Meta
}

type multiPayload struct {
	Pagination *codec.PaginationLinks `json:"links"`
	Data       []modelNode            `json:"data"`
	Meta       codec.Meta             `json:"meta"`
}

func (m multiPayload) nodes() []modelNode {
	return m.Data
}

func (m multiPayload) pagination() *codec.PaginationLinks {
	return m.Pagination
}

func (m multiPayload) meta() codec.Meta {
	return m.Meta
}

type modelNode map[string]interface{}
