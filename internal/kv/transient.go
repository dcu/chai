package kv

import (
	"github.com/chaisql/chai/internal/engine"
	"github.com/cockroachdb/errors"
	"github.com/cockroachdb/pebble"
)

var _ engine.Session = (*TransientSession)(nil)

type TransientSession struct {
	db           *pebble.DB
	batch        *pebble.Batch
	store        *PebbleEngine
	maxBatchSize int
	closed       bool
}

func (s *PebbleEngine) NewTransientSession() engine.Session {
	return &TransientSession{
		db:           s.db,
		maxBatchSize: s.opts.MaxTransientBatchSize,
		store:        s,
	}
}

func (s *TransientSession) Commit() error {
	return errors.New("cannot commit in transient mode")
}

func (s *TransientSession) Close() error {
	if s.closed {
		return errors.New("already closed")
	}
	s.closed = true

	return s.batch.Close()
}

func (s *TransientSession) Insert(k, v []byte) error {
	return errors.New("cannot insert in transient mode")
}

// Put stores a key value pair. If it already exists, it overrides it.
func (s *TransientSession) Put(k, v []byte) error {
	if len(k) == 0 {
		return errors.New("cannot store empty key")
	}

	if len(v) == 0 {
		return errors.New("cannot store empty value")
	}

	if s.batch == nil {
		s.batch = s.db.NewIndexedBatch()
	}

	if s.batch.Len() > s.maxBatchSize && s.batch.Count() > 0 {
		err := s.batch.Commit(pebble.NoSync)
		if err != nil {
			return err
		}

		s.batch.Reset()
	}

	return s.batch.Set(k, v, nil)
}

// Get returns a value associated with the given key. If not found, returns ErrKeyNotFound.
func (s *TransientSession) Get(k []byte) ([]byte, error) {
	if s.batch == nil {
		return nil, errors.WithStack(engine.ErrKeyNotFound)
	}

	return get(s.batch, k)
}

// Exists returns whether a key exists and is visible by the current session.
func (s *TransientSession) Exists(k []byte) (bool, error) {
	if s.batch == nil {
		return false, nil
	}

	return exists(s.batch, k)
}

// Delete a record by key. If not found, returns ErrKeyNotFound.
func (s *TransientSession) Delete(k []byte) error {
	if s.batch == nil {
		return errors.WithStack(engine.ErrKeyNotFound)
	}

	_, closer, err := s.batch.Get(k)
	if err != nil {
		if errors.Is(err, pebble.ErrNotFound) {
			return errors.WithStack(engine.ErrKeyNotFound)
		}

		return err
	}
	err = closer.Close()
	if err != nil {
		return err
	}

	return s.batch.Delete(k, nil)
}

func (s *TransientSession) DeleteRange(start []byte, end []byte) error {
	if s.batch == nil {
		return nil
	}

	return s.batch.DeleteRange(start, end, nil)
}

func (s *TransientSession) Iterator(opts *engine.IterOptions) (engine.Iterator, error) {
	var popts *pebble.IterOptions
	if opts != nil {
		popts = &pebble.IterOptions{
			LowerBound: opts.LowerBound,
			UpperBound: opts.UpperBound,
		}
	}

	var it *pebble.Iterator
	if s.batch == nil {
		it = s.db.NewIter(popts)
	} else {
		it = s.batch.NewIter(popts)
	}

	return &iterator{
		Iterator: it,
	}, nil
}
