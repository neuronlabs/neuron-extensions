package auth

type Options struct {
	PathPrefix string
}

type Option func(o *Options)

// WithPathPrefix is an option that sets path prefix for the API.
func WithPathPrefix(pathPrefix string) Option {
	return func(o *Options) {
		o.PathPrefix = pathPrefix
	}
}
