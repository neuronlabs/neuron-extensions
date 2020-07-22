package jsonapi

import (
	neuronCodec "github.com/neuronlabs/neuron/codec"
)

// TopLinks is used to represent a `links` object.
// http://jsonapi.org/format/#document-links
type TopLinks struct {
	Self    string `json:"self,omitempty"`
	Related string `json:"related,omitempty"`

	First string `json:"first,omitempty"`
	Prev  string `json:"prev,omitempty"`
	Next  string `json:"next,omitempty"`
	Last  string `json:"last,omitempty"`
}

// SetPaginationLinks sets the pagination links from the marshal options
func (t *TopLinks) SetPaginationLinks(o *neuronCodec.MarshalOptions) {
	// if pagination links are set
	pLinks := o.Link.PaginationLinks
	if pLinks == nil {
		return
	}
	base := t.Self
	if pLinks.Self != "" {
		t.Self += "?" + pLinks.Self
	}
	if pLinks.First != "" {
		t.First = base + "?" + pLinks.First
	}
	if pLinks.Last != "" {
		t.Last = base + "?" + pLinks.Last
	}
	if pLinks.Next != "" {
		t.Next = base + "?" + pLinks.Next
	}
	if pLinks.Prev != "" {
		t.Prev = base + "?" + pLinks.Prev
	}
}

// Links is the structure used to represent a related 'links' object.
type Links map[string]interface{}

// Link is used to represent a member of the `links` object.
type Link struct {
	Href string `json:"href"`
	Meta Meta   `json:"meta,omitempty"`
}
