package redis

import (
	"context"
	"strings"
	"time"

	"github.com/go-redis/redis/v8"

	"github.com/neuronlabs/neuron/core"
	"github.com/neuronlabs/neuron/errors"
	"github.com/neuronlabs/neuron/store"
)

var (
	_ store.Store = &Redis{}
	_ core.Dialer = &Redis{}
	_ core.Closer = &Redis{}
)

// Redis is the key-value store implementation for the neuron framework.
// It implements store.Store, core.Dialer and core.Closer interfaces.
// This implementation requires store options connection url to be set.
type Redis struct {
	Options  *store.Options
	r        *redis.Client
	rOptions *redis.Options
}

// Dial implements core.Dialer interface.
func (r *Redis) Dial(ctx context.Context) error {
	// Create new client.
	r.r = redis.NewClient(r.rOptions)
	//
	_, err := r.r.Ping(ctx).Result()
	if err != nil {
		return err
	}
	return nil
}

// Close implements Closer interface.
func (r *Redis) Close(ctx context.Context) error {
	return r.r.Close()
}

// New creates new options.
func New(options ...store.Option) (*Redis, error) {
	o := &store.Options{}
	for _, option := range options {
		option(o)
	}
	if o.ConnectionURL == "" {
		return nil, errors.Wrap(store.ErrInitialization, "no connection url provided in the options")
	}
	redisOptions, err := redis.ParseURL(o.ConnectionURL)
	if err != nil {
		return nil, errors.Wrapf(store.ErrInitialization, "connection url is not valid: %v", err.Error())
	}
	r := &Redis{
		rOptions: redisOptions,
		Options:  o,
	}
	return r, nil
}

// Set implements store.Store interface.
func (r *Redis) Set(ctx context.Context, record *store.Record, options ...store.SetOption) error {
	o := &store.SetOptions{}
	for _, option := range options {
		option(o)
	}
	ttl := r.Options.DefaultExpiration
	if o.TTL != 0 {
		ttl = o.TTL
	}
	if err := r.checkInitialization(); err != nil {
		return err
	}
	if err := r.r.Set(ctx, r.getKey(record.Key), record.Value, ttl).Err(); err != nil {
		return errors.Wrap(store.ErrStore, err.Error())
	}
	return nil
}

// Get implements store.Store interface.
func (r *Redis) Get(ctx context.Context, key string) (*store.Record, error) {
	if err := r.checkInitialization(); err != nil {
		return nil, err
	}
	return r.getRecord(ctx, key)
}

// Delete implements store.Store interface.
func (r *Redis) Delete(ctx context.Context, key string) error {
	if err := r.checkInitialization(); err != nil {
		return err
	}
	if err := r.r.Del(ctx, r.getKey(key)).Err(); err != nil {
		if err == redis.Nil {
			return store.ErrRecordNotFound
		}
		return errors.Wrap(store.ErrInternal, err.Error())
	}
	return nil
}

func (r *Redis) Find(ctx context.Context, options ...store.FindOption) ([]*store.Record, error) {
	if err := r.checkInitialization(); err != nil {
		return nil, err
	}
	o := &store.FindPattern{}
	for _, option := range options {
		option(o)
	}
	prefix := r.Options.Prefix + o.Prefix
	suffix := o.Suffix + r.Options.Suffix
	var pattern string
	if prefix != "" || suffix != "" {
		pattern = prefix + "*" + suffix
	}
	keys, err := r.r.Keys(ctx, pattern).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, nil
		}
		return nil, errors.Wrap(store.ErrStore, err.Error())
	}
	var (
		records []*store.Record
		j       int
	)
	for i := range keys {
		if o.Offset > 0 {
			o.Offset--
			continue
		}
		key := keys[i]
		if r.Options.Prefix != "" {
			key = strings.TrimPrefix(key, r.Options.Prefix)
		}
		if r.Options.Suffix != "" {
			key = strings.TrimSuffix(key, r.Options.Suffix)
		}
		record, err := r.getRecord(ctx, key)
		if err != nil && err == redis.Nil {
			continue
		} else if err != nil {
			return nil, err
		}
		records = append(records, record)
		j++
		if j == o.Limit {
			break
		}
	}
	return records, nil
}

func (r *Redis) getKey(key string) string {
	return r.Options.Prefix + key + r.Options.Suffix
}

func (r *Redis) getRecord(ctx context.Context, key string) (*store.Record, error) {
	thisKey := r.getKey(key)
	value, err := r.r.Get(ctx, thisKey).Bytes()
	if err == redis.Nil {
		return nil, store.ErrRecordNotFound
	} else if err != nil {
		return nil, errors.Wrap(store.ErrInternal, err.Error())
	}
	ttl, err := r.r.TTL(ctx, thisKey).Result()
	if err != nil {
		return nil, errors.Wrap(store.ErrInternal, err.Error())
	}
	return &store.Record{
		Key:       key,
		Value:     value,
		ExpiresAt: time.Now().Add(ttl),
	}, nil
}

func (r *Redis) checkInitialization() error {
	if r.r == nil {
		return errors.Wrap(store.ErrStore, "redis store not initialized yet - it waits for the Dial to be executed")
	}
	return nil
}
