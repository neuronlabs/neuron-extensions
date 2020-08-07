package http

import (
	"crypto/tls"
)

// Options are the http server options.
type Options struct {
	VersionedAPIs map[string][]API
	APIs          []API
	Hostname      string
	Port          int
	TLSConfig     *tls.Config
}

// Option is a function that changes options in some way.
type Option func(o *Options)

// WithAPI is a server option that stores provided API in given http server.
func WithAPI(api API) Option {
	return func(o *Options) {
		o.APIs = append(o.APIs, api)
	}
}

// WithAPIVersion is a server option that stores the api for provided version.
func WithAPIVersion(version string, api API) Option {
	return func(o *Options) {
		o.VersionedAPIs[version] = append(o.VersionedAPIs[version], api)
	}
}

// WithPort sets the port option for the server.
func WithPort(port int) Option {
	return func(o *Options) {
		o.Port = port
	}
}

// WithHostname sets the hostname option for the server.
func WithHostname(hostname string) Option {
	return func(o *Options) {
		o.Hostname = hostname
	}
}

// WithTLSConfig sets the tls config option for the server.
func WithTLSConfig(tlsConfig *tls.Config) Option {
	return func(o *Options) {
		o.TLSConfig = tlsConfig
	}
}
