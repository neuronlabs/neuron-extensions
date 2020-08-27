package authentication

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/url"

	jsonCodec "github.com/neuronlabs/neuron-extensions/codec/json"
	"github.com/neuronlabs/neuron-extensions/server/http/httputil"
	"github.com/neuronlabs/neuron-extensions/server/http/log"
	"github.com/neuronlabs/neuron/auth"
	"github.com/neuronlabs/neuron/codec"
	"github.com/neuronlabs/neuron/database"
	"github.com/neuronlabs/neuron/errors"
	"github.com/neuronlabs/neuron/mapping"
)

// RegisterAccountOptions are the options used for registering the account.
type RegisterAccountOptions struct {
	Account    auth.Account
	Password   *auth.Password
	Meta       codec.Meta
	Attributes map[string]interface{}
}

// AccountRegisterHandler is an interface that allows to handle account insertion in a custom way.
type AccountRegisterHandler interface {
	HandleRegisterAccount(ctx context.Context, db database.DB, options *RegisterAccountOptions) error
}

// BeforeAccountRegistrar is an interface used for handling hook before insertion of account.
type BeforeAccountRegistrar interface {
	BeforeRegisterAccount(ctx context.Context, db database.DB, options *RegisterAccountOptions) error
}

// AfterAccountRegistrar is an interface used for handling hook after insertion of account.
// Within this hook, developer could set some functions that send emails, sets some additional models etc.
type AfterAccountRegistrar interface {
	AfterRegisterAccount(ctx context.Context, db database.DB, options *RegisterAccountOptions) error
}

// RegisterAccountMarshaler.
type RegisterAccountMarshaler interface {
	MarshalRegisteredAccount(ctx context.Context, options *RegisterAccountOptions) (*codec.Payload, error)
}

// AccountCreateInput is an input for the account creation.
type AccountCreateInput struct {
	Meta                 codec.Meta             `json:"meta"`
	Username             string                 `json:"username"`
	Password             string                 `json:"password"`
	PasswordConfirmation string                 `json:"password_confirmation"`
	Attributes           map[string]interface{} `json:"attributes"`
}

func (a *API) handleRegisterAccount(rw http.ResponseWriter, req *http.Request) {
	input, err := a.decodeAccountCreateInput(req)
	if err != nil {
		a.marshalErrors(rw, 0, err)
		return
	}

	// Score and analyze the password.
	password := auth.NewPassword(input.Password, a.Options.PasswordScorer)

	// Validate the password.
	if err := a.Options.PasswordValidator(password); err != nil {
		httpError := httputil.ErrInvalidJSONFieldValue()
		httpError.Detail = "Provided invalid password."
		if detailer, ok := err.(*errors.DetailedError); ok {
			httpError.Detail = detailer.Details
		}
		a.marshalErrors(rw, 400, httpError)
		return
	}

	// Validate the username.
	if err := a.Options.UsernameValidator(input.Username); err != nil {
		httpError := httputil.ErrInvalidJSONFieldValue()
		httpError.Detail = "Provided invalid username."
		if detailer, ok := err.(*errors.DetailedError); ok {
			httpError.Detail = detailer.Details
		}
		a.marshalErrors(rw, 400, httpError)
		return
	}

	ctx := req.Context()

	// Create new model and sets it's username and password.
	model := mapping.NewModel(a.model).(auth.Account)
	model.SetUsername(input.Username)

	// Hash the password (with optional salt) and set into given model.
	if err = a.Controller.Authenticator.HashAndSetPassword(model, password); err != nil {
		a.marshalErrors(rw, 0, err)
		return
	}

	registerOptions := &RegisterAccountOptions{
		Account:    model,
		Password:   password,
		Meta:       input.Meta,
		Attributes: input.Attributes,
	}
	// Execute create account
	var payload *codec.Payload
	err = database.RunInTransaction(ctx, a.DB, nil, func(db database.DB) error {
		err := a.registerAccount(ctx, db, registerOptions)
		if err != nil {
			return err
		}
		customMarshaler, ok := a.Options.AccountHandler.(RegisterAccountMarshaler)
		if !ok {
			return nil
		}
		payload, err = customMarshaler.MarshalRegisteredAccount(ctx, registerOptions)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		a.marshalErrors(rw, 0, err)
		return
	}
	if payload == nil {
		rw.WriteHeader(http.StatusNoContent)
		return
	}
	// Marshal given payload into json codec.
	payload.ModelStruct = a.model

	rw.WriteHeader(http.StatusCreated)
	cdc := jsonCodec.GetCodec(a.Controller)
	if err := cdc.MarshalPayload(rw, payload, codec.MarshalSingleModel()); err != nil {
		log.Errorf("Marshaling account payload failed: %v", err)
	}
}

func (a *API) registerAccount(ctx context.Context, db database.DB, options *RegisterAccountOptions) error {
	// Check if the username exists.
	checkHandler, ok := a.Options.AccountHandler.(CheckUsernameHandler)
	if !ok {
		checkHandler = a.defaultHandler
	}
	if err := checkHandler.HandleCheckUsername(ctx, db, options.Account); err != nil {
		return err
	}

	// Check if before insert hook exists.
	if beforeInserter, ok := a.Options.AccountHandler.(BeforeAccountRegistrar); ok {
		if err := beforeInserter.BeforeRegisterAccount(ctx, db, options); err != nil {
			return err
		}
	}

	// Execute inserter hook.
	inserter, ok := a.Options.AccountHandler.(AccountRegisterHandler)
	if !ok {
		inserter = a.defaultHandler
	}
	if err := inserter.HandleRegisterAccount(ctx, db, options); err != nil {
		return err
	}

	// Handle after insert hook.
	if afterInserter, ok := a.Options.AccountHandler.(AfterAccountRegistrar); ok {
		if err := afterInserter.AfterRegisterAccount(ctx, db, options); err != nil {
			return err
		}
	}
	return nil
}

func (a *API) decodeAccountCreateInput(req *http.Request) (*AccountCreateInput, error) {
	input := &AccountCreateInput{}
	switch req.Header.Get("Content-Type") {
	case "application/json":
		d := json.NewDecoder(req.Body)
		if a.Options.StrictUnmarshal {
			d.DisallowUnknownFields()
		}
		if err := d.Decode(input); err != nil {
			return nil, errors.WrapDetf(codec.ErrUnmarshalDocument, "decode failed: %v", err).
				WithDetail("Provided invalid input document.")
		}
	case "application/x-www-form-urlencoded":
		body, err := ioutil.ReadAll(req.Body)
		if err != nil {
			return nil, errors.WrapDet(codec.ErrUnmarshalDocument, "reading failed")
		}
		q, err := url.ParseQuery(string(body))
		if err != nil {
			return nil, errors.WrapDet(codec.ErrUnmarshalDocument, "parsing form failed").WithDetail("Parsing form failed.")
		}

		input.Username = q.Get("username")
		input.Password = q.Get("password")
		input.PasswordConfirmation = q.Get("password_confirmation")
	default:
		err := req.ParseForm()
		if err != nil {
			return nil, errors.WrapDet(codec.ErrUnmarshalDocument, "provided invalid post form").WithDetail("invalid post form")
		}
		q := req.PostForm
		input.Username = q.Get("username")
		input.Password = q.Get("password")
		input.PasswordConfirmation = q.Get("password_confirmation")
	}
	return input, nil
}

// var emailRegexp = regexp.MustCompile("^[a-zA-Z0-9.!#$%&'*+/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$")
