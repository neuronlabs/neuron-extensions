package tokener

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"time"

	"github.com/neuronlabs/neuron/errors"
	"github.com/neuronlabs/neuron/log"
	"github.com/neuronlabs/neuron/store"
)

// StoreToken is the token's store value.
type StoreToken struct {
	MappedTokens []string   `json:"mapped_tokens"`
	RevokedAt    *time.Time `json:"is_revoked"`
	ExpiresAt    time.Time  `json:"expires_at"`
}

func (t *Tokener) getStoreToken(ctx context.Context, token string) (*StoreToken, error) {
	key := t.tokenStoreKey(token)
	record, err := t.Store.Get(ctx, key)
	if err != nil {
		return nil, err
	}
	st := &StoreToken{}
	if err := json.Unmarshal(record.Value, st); err != nil {
		log.Errorf("[jwt-tokener] unmarshal store token failed: %v", err)
		return nil, errors.Wrap(store.ErrInternal, "store token malformed")
	}
	return st, nil
}

func (t *Tokener) setStoreToken(ctx context.Context, token string, sToken *StoreToken) error {
	ttl := t.c.Now().Sub(sToken.ExpiresAt)
	value, err := json.Marshal(sToken)
	if err != nil {
		log.Errorf("[jwt-tokener] marshal store token failed: %v", err)
		return errors.Wrap(store.ErrInternal, "tokener marshal store token failed")
	}

	key := t.tokenStoreKey(token)
	record := &store.Record{
		Key:       key,
		Value:     value,
		ExpiresAt: sToken.ExpiresAt,
	}
	return t.Store.Set(ctx, record, store.SetWithTTL(ttl))

}

func (t *Tokener) tokenStoreKey(token string) string {
	h := sha256.New()
	h.Write([]byte(token))
	return hex.EncodeToString(h.Sum(nil))
}
