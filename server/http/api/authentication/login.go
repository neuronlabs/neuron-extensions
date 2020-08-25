package authentication

import (
	"bytes"
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/url"

	"github.com/neuronlabs/neuron/auth"
	"github.com/neuronlabs/neuron/codec"
	"github.com/neuronlabs/neuron/database"
	"github.com/neuronlabs/neuron/errors"
	"github.com/neuronlabs/neuron/mapping"
	"github.com/neuronlabs/neuron/query"

	"github.com/neuronlabs/neuron-extensions/server/http/httputil"
	"github.com/neuronlabs/neuron-extensions/server/http/log"
)

// LoginOptions are the options used while handling the login.
type LoginOptions struct {
	Header        http.Header
	Account       auth.Account
	Password      *auth.Password
	RememberToken bool
}

// LoginAccountRefresher is an interface used for refreshing - getting the account in the login process.
// The function should not check the password. It should only check if account exists and get it's value.
type LoginAccountRefresher interface {
	HandleLoginAccountRefresh(ctx context.Context, db database.DB, options *LoginOptions) error
}

// BeforeLoginer is the hook interface used before login process.
type BeforeLoginer interface {
	BeforeLogin(ctx context.Context, db database.DB, options *LoginOptions) error
}

// AfterLoginer is the hook interface used after login process.
type AfterLoginer interface {
	AfterLogin(ctx context.Context, db database.DB, options *LoginOptions) error
}

// AfterFailedLogin is an interface used to handle cases when the login fails.
type AfterFailedLoginer interface {
	AfterFailedLogin(ctx context.Context, db database.DB, options *LoginOptions) error
}

// WithTransactionLoginer is an interface that begins the transaction on the login process.
type WithTransactionLoginer interface {
	LoginWithTransaction() *query.TxOptions
}

// WithContextLoginer is an interface that provides context to the login process.
type WithContextLoginer interface {
	LoginWithContext(ctx context.Context) (context.Context, error)
}

// LoginInput is the input login structure.
type LoginInput struct {
	Username      string `json:"username"`
	Password      string `json:"password"`
	RememberToken bool   `json:"remember_token"`
}

// LoginOutput is the successful login output structure.
type LoginOutput struct {
	Meta         codec.Meta `json:"meta,omitempty"`
	AccessToken  string     `json:"access_token"`
	RefreshToken string     `json:"refresh_token,omitempty"`
	TokenType    string     `json:"token_type,omitempty"`
	ExpiresIn    int64      `json:"expires_in,omitempty"`
}

func (a *API) handleLoginEndpoint(rw http.ResponseWriter, req *http.Request) {
	input := &LoginInput{}

	username, password, hasBasicAuth := req.BasicAuth()
	switch req.Header.Get("Content-Type") {
	case "application/json":
		d := json.NewDecoder(req.Body)
		if a.Options.StrictUnmarshal {
			d.DisallowUnknownFields()
		}
		if err := d.Decode(input); err != nil {
			a.marshalErrors(rw, 400, errors.WrapDetf(codec.ErrUnmarshalDocument, "decode failed: %v", err).
				WithDetail("Provided invalid input document."))
			return
		}
	case "application/x-www-form-urlencoded":
		body, err := ioutil.ReadAll(req.Body)
		if err != nil {
			a.marshalErrors(rw, 400, errors.WrapDet(codec.ErrUnmarshalDocument, "reading failed"))
			return
		}
		q, err := url.ParseQuery(string(body))
		if err != nil {
			a.marshalErrors(rw, 400, errors.WrapDet(codec.ErrUnmarshalDocument, "parsing form failed").WithDetail("Parsing form failed."))
			return
		}

		input.Username = q.Get("username")
		input.Password = q.Get("password")
		input.RememberToken = q.Get("remember_token") == "true"
	default:
		err := req.ParseForm()
		if err != nil && !hasBasicAuth {
			a.marshalErrors(rw, 400, errors.WrapDet(codec.ErrUnmarshalDocument, "provided invalid post form").WithDetail("invalid post form"))
			return
		}
		q := req.PostForm
		input.Username = q.Get("username")
		input.Password = q.Get("password")
		input.RememberToken = q.Get("remember_token") == "true"
	}

	if hasBasicAuth {
		input.Username = username
		input.Password = password
	}

	// Validate provided username.
	if a.Options.UsernameValidator != nil {
		if err := a.Options.UsernameValidator(input.Username); err != nil {
			httpError := httputil.ErrInvalidJSONFieldValue()
			httpError.Detail = "Provided invalid username."
			if detailer, ok := err.(*errors.DetailedError); ok {
				httpError.Detail = detailer.Details
			}
			a.marshalErrors(rw, 400, httpError)
			return
		}
	}

	// Validate password.
	neuronPassword := auth.NewPassword(input.Password, a.Options.PasswordScorer)
	if a.Options.PasswordValidator != nil {
		if err := a.Options.PasswordValidator(neuronPassword); err != nil {
			httpError := httputil.ErrInvalidAuthenticationInfo()
			httpError.Detail = "username or password is not valid"
			a.marshalErrors(rw, 0, httpError)
			return
		}
	}

	account := mapping.NewModel(a.model).(auth.Account)
	account.SetUsername(input.Username)

	ctx := req.Context()
	var err error
	if contexter, ok := a.Options.AccountHandler.(WithContextLoginer); ok {
		ctx, err = contexter.LoginWithContext(ctx)
		if err != nil {
			a.marshalErrors(rw, 0, err)
			return
		}
	}

	options := &LoginOptions{
		Header:        req.Header,
		Account:       account,
		Password:      neuronPassword,
		RememberToken: input.RememberToken,
	}
	db := a.DB
	var tx *database.Tx
	transactioner, isTransactioner := a.Options.AccountHandler.(WithTransactionLoginer)
	if isTransactioner {
		tx, err = database.Begin(ctx, db, transactioner.LoginWithTransaction())
		if err != nil {
			log.Errorf("Begin transaction failed: %v", err)
			a.marshalErrors(rw, 500, httputil.ErrInternalError())
			return
		}
		db = tx
	}

	// Get the account model.
	if before, ok := a.Options.AccountHandler.(BeforeLoginer); ok {
		if err := before.BeforeLogin(ctx, db, options); err != nil {
			if isTransactioner {
				if err = tx.Rollback(); err != nil {
					log.Errorf("Rolling back transaction failed: %v", err)
				}
			}
			a.marshalErrors(rw, 0, err)
			return
		}
	}

	// Prepare account refresher.
	refresher, ok := a.Options.AccountHandler.(LoginAccountRefresher)
	if !ok {
		refresher = a.defaultHandler
	}
	if err := refresher.HandleLoginAccountRefresh(ctx, db, options); err != nil {
		if isTransactioner {
			if err = tx.Rollback(); err != nil {
				log.Errorf("Rolling back transaction failed: %v", err)
			}
		}
		if errors.Is(err, query.ErrNoResult) {
			if loginFailer, ok := a.Options.AccountHandler.(AfterLoginer); ok {
				if err = loginFailer.AfterLogin(ctx, a.DB, options); err != nil {
					a.marshalErrors(rw, 0, err)
					return
				}
			}
			httpError := httputil.ErrInvalidAuthenticationInfo()
			httpError.Detail = "username or password is not valid"
			a.marshalErrors(rw, 0, httpError)
			return
		}
		a.marshalErrors(rw, 0, err)
		return
	}

	// Check if password matches hashed password in the model.
	if err := a.Controller.Authenticator.ComparePassword(options.Account, input.Password); err != nil {
		if isTransactioner {
			if err = tx.Rollback(); err != nil {
				log.Errorf("Rolling back transaction failed: %v", err)
			}
		}
		if loginFailer, ok := a.Options.AccountHandler.(AfterLoginer); ok {
			if err = loginFailer.AfterLogin(ctx, a.DB, options); err != nil {
				a.marshalErrors(rw, 0, err)
				return
			}
		}
		httpError := httputil.ErrInvalidAuthenticationInfo()
		httpError.Detail = "username or password is not valid"
		a.marshalErrors(rw, 0, httpError)
		return
	}

	// Handle after loginer hook.
	if after, ok := a.Options.AccountHandler.(AfterLoginer); ok {
		if err := after.AfterLogin(ctx, db, options); err != nil {
			if isTransactioner {
				if err = tx.Rollback(); err != nil {
					log.Errorf("Rolling back transaction failed: %v", err)
				}
			}
			a.marshalErrors(rw, 0, err)
			return
		}
	}

	if isTransactioner {
		if err = tx.Commit(); err != nil {
			log.Errorf("Committing transaction failed: %v", err)
			a.marshalErrors(rw, 500, httputil.ErrInternalError())
			return
		}
	}

	expiration := a.Options.TokenExpiration
	if input.RememberToken {
		expiration = a.Options.RememberTokenExpiration
	}

	// Create the token for provided account.
	token, err := a.Controller.Tokener.Token(ctx, options.Account,
		auth.TokenExpirationTime(expiration),
		auth.TokenRefreshExpirationTime(a.Options.RefreshTokenExpiration),
	)
	if err != nil {
		a.marshalErrors(rw, 0, err)
		return
	}

	output := &LoginOutput{
		AccessToken:  token.AccessToken,
		RefreshToken: token.RefreshToken,
		TokenType:    token.TokenType,
		ExpiresIn:    int64(token.ExpiresIn),
	}

	a.setContentType(rw)
	buffer := &bytes.Buffer{}
	if err = json.NewEncoder(buffer).Encode(output); err != nil {
		a.marshalErrors(rw, 500, httputil.ErrInternalError())
		return
	}
	rw.WriteHeader(http.StatusCreated)
	if _, err = buffer.WriteTo(rw); err != nil {
		log.Errorf("Writing to response writer failed: %v", err)
	}
}
