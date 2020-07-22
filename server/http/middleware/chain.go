package middleware

import (
	"net/http"
)

type Middleware func(http.Handler) http.Handler

type Chain []Middleware

// Handle handles the chain of middlewares for given 'handle' http.Handle.
func (c Chain) Handle(handle http.Handler) http.Handler {
	for i := range c {
		handle = c[len(c)-1-i](handle)
	}
	return handle
}
