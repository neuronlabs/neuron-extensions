package authentication

import (
	"time"

	"github.com/neuronlabs/neuron/auth"
	"github.com/neuronlabs/neuron/server"
)

// AuthenticatorOptions is the structure that contains auth API  settings.
type Options struct {
	AccountModel             auth.Account
	AccountHandler           interface{}
	PathPrefix               string
	Middlewares              []server.Middleware
	RegisterMiddlewares      []server.Middleware
	LoginMiddlewares         []server.Middleware
	LogoutMiddlewares        []server.Middleware
	RefreshTokenMiddlewares  []server.Middleware
	StrictUnmarshal          bool
	PasswordValidator        auth.PasswordValidator
	PasswordScorer           auth.PasswordScorer
	UsernameValidator        auth.UsernameValidator
	TokenExpiration          time.Duration
	RememberTokenExpiration  time.Duration
	RefreshTokenExpiration   time.Duration
	PermitRefreshTokenLogout bool
}

func defaultOptions() *Options {
	return &Options{
		PasswordScorer:          auth.DefaultPasswordScorer,
		PasswordValidator:       auth.DefaultPasswordValidator,
		UsernameValidator:       auth.DefaultUsernameValidator,
		TokenExpiration:         time.Hour * 24,
		RememberTokenExpiration: time.Hour * 24 * 7,
		RefreshTokenExpiration:  time.Hour * 24 * 30,
	}
}

type Option func(o *Options)

// WithAccountModel is an option that sets the account model within options.
func WithAccountModel(account auth.Account) Option {
	return func(o *Options) {
		o.AccountModel = account
	}
}

// WithAccountHandler is an option that sets the account handler.
func WithAccountHandler(handler interface{}) Option {
	return func(o *Options) {
		o.AccountHandler = handler
	}
}

// WithPathPrefix is an option that sets path prefix for the API.
func WithPathPrefix(pathPrefix string) Option {
	return func(o *Options) {
		o.PathPrefix = pathPrefix
	}
}

// WithStrictUnmarshal is an option that sets the 'StrictUnmarshal' setting.
func WithStrictUnmarshal(setting bool) Option {
	return func(o *Options) {
		o.StrictUnmarshal = setting
	}
}

// WithPasswordScorer sets the password scorer for the auth options.
func WithPasswordScorer(scorer auth.PasswordScorer) Option {
	return func(o *Options) {
		o.PasswordScorer = scorer
	}
}

// WithUsernameValidator sets the username validator function.
func WithUsernameValidator(validator auth.UsernameValidator) Option {
	return func(o *Options) {
		o.UsernameValidator = validator
	}
}

// WithPasswordValidator sets the password validator function.
func WithPasswordValidator(validator auth.PasswordValidator) Option {
	return func(o *Options) {
		o.PasswordValidator = validator
	}
}

// WithTokenExpiration sets the 'TokenExpiration' option for the auth service.
func WithTokenExpiration(d time.Duration) Option {
	return func(o *Options) {
		o.TokenExpiration = d
	}
}

// WithRememberMeTokenExpiration sets the 'RememberTokenExpiration' option for the auth service.
func WithRememberMeTokenExpiration(d time.Duration) Option {
	return func(o *Options) {
		o.RememberTokenExpiration = d
	}
}

// WithRefreshTokenExpiration sets the 'RefreshTokenExpiration' option for the auth service.
func WithRefreshTokenExpiration(d time.Duration) Option {
	return func(o *Options) {
		o.RefreshTokenExpiration = d
	}
}

// WithMiddlewares adds middlewares for all auth.API endpoints.
func WithMiddlewares(middlewares ...server.Middleware) Option {
	return func(o *Options) {
		o.Middlewares = append(o.Middlewares, middlewares...)
	}
}

// WithLoginMiddlewares adds middlewares for all auth.API endpoints.
func WithLoginMiddlewares(middlewares ...server.Middleware) Option {
	return func(o *Options) {
		o.LoginMiddlewares = append(o.LoginMiddlewares, middlewares...)
	}
}

// WithLogoutMiddlewares adds middlewares for all auth.API endpoints.
func WithLogoutMiddlewares(middlewares ...server.Middleware) Option {
	return func(o *Options) {
		o.LogoutMiddlewares = append(o.LogoutMiddlewares, middlewares...)
	}
}

// WithRegisterMiddlewares adds middlewares for all auth.API endpoints.
func WithRegisterMiddlewares(middlewares ...server.Middleware) Option {
	return func(o *Options) {
		o.RegisterMiddlewares = append(o.RegisterMiddlewares, middlewares...)
	}
}

// WithRefreshTokenMiddlewares adds middlewares for all auth.API endpoints.
func WithRefreshTokenMiddlewares(middlewares ...server.Middleware) Option {
	return func(o *Options) {
		o.RefreshTokenMiddlewares = append(o.RefreshTokenMiddlewares, middlewares...)
	}
}

// WithPermitRefreshTokenLogout sets the option that allows to logout using also refresh token.
func WithPermitRefreshTokenLogout(permit bool) Option {
	return func(o *Options) {
		o.PermitRefreshTokenLogout = permit
	}
}
