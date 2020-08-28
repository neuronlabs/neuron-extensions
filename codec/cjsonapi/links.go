package cjsonapi

import (
	"github.com/neuronlabs/neuron/codec"
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
func (t *TopLinks) SetPaginationLinks(o *codec.PaginationLinks) {
	// if pagination links are set
	if o == nil {
		return
	}
	if o.Self != "" {
		t.Self = o.Self
	}
	if o.First != "" {
		t.First = o.First
	}
	if o.Last != "" {
		t.Last = o.Last
	}
	if o.Next != "" {
		t.Next = o.Next
	}
	if o.Prev != "" {
		t.Prev = o.Prev
	}
}

// Links is the structure used to represent a related 'links' object.
type Links map[string]interface{}

// Link is used to represent a member of the `links` object.
type Link struct {
	Href string `json:"href"`
	Meta Meta   `json:"meta,omitempty"`
}
