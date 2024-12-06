package multilimiter

import (
	"context"
	"sync"
	"time"

	"github.com/pkg/errors"
	"github.com/sethvargo/go-limiter"
	"github.com/sethvargo/go-limiter/memorystore"
)

// MultiStore is a type for multiple limiter.Store
type MultiStore struct {
	multi []*struct {
		store            limiter.Store
		previouslyFailed bool
	}
	mutex sync.RWMutex
}

// New creates a new *MultiStore
func New(configs []*memorystore.Config) (*MultiStore, error) {
	m := &MultiStore{}
	for _, c := range configs {
		ms, err := memorystore.New(c)
		if err != nil {
			return nil, err
		}
		m.multi = append(
			m.multi, &struct {
				store            limiter.Store
				previouslyFailed bool
			}{
				store: ms,
			},
		)
	}
	return m, nil
}

// Take implements the limiter.Store interface
func (m *MultiStore) Take(ctx context.Context, key string) (ok bool, reset time.Time, firstFail bool, err error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	for _, s := range m.multi {
		_, _, resetT, okk, errr := s.store.Take(ctx, key)
		if errr != nil {
			err = errors.WithStack(errr)
			return
		}
		if okk {
			s.previouslyFailed = false
		} else {
			ok = false
			firstFail = !s.previouslyFailed
			s.previouslyFailed = true
			reset = time.Unix(int64(resetT/uint64(time.Second)), int64(resetT%uint64(time.Second)))
			return
		}
	}
	ok = true
	return
}

// NewDefaultMultiStore creates a *MultiStore with the default internals
func NewDefaultMultiStore() (*MultiStore, error) {
	return New(
		[]*memorystore.Config{
			{
				Tokens:      10,
				Interval:    time.Second,
				SweepMinTTL: time.Hour,
			},
			{
				Tokens:      20,
				Interval:    5 * time.Minute,
				SweepMinTTL: time.Hour,
			},
			{
				Tokens:      50,
				Interval:    time.Hour,
				SweepMinTTL: 4 * time.Hour,
			},
			{
				Tokens:      150,
				Interval:    24 * time.Hour,
				SweepMinTTL: 48 * time.Hour,
			},
		},
	)
}
