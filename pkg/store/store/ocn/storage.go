// SPDX-FileCopyrightText: 2022-present Intel Corporation
// SPDX-FileCopyrightText: 2020-present Open Networking Foundation <info@opennetworking.org>
//
// SPDX-License-Identifier: Apache-2.0

package ocnstorage

import (
	"context"
	"sync"

	"github.com/google/uuid"
	"github.com/onosproject/onos-lib-go/pkg/errors"
	"github.com/onosproject/onos-lib-go/pkg/logging"
	"github.com/onosproject/cco-mon/pkg/store/event"
	"github.com/onosproject/cco-mon/pkg/store/store/storage"
	"github.com/onosproject/cco-mon/pkg/store/store/watcher"
	meastype "github.com/onosproject/rrm-son-lib/pkg/model/measurement/type"
)

var log = logging.GetLogger()

// NewStore generates a store object to save Ocn into a map
func NewStore() Store {
	watchers := watcher.NewWatchers()
	return &store{
		storage:  make(map[storage.IDs]*OcnMap),
		watchers: watchers,
	}
}

// Store has all functions in this store
type Store interface {
	// Put puts key and its value
	Put(ctx context.Context, key storage.IDs, value *OcnMap) (*OcnMap, error)

	// Get gets the element with key
	Get(ctx context.Context, key storage.IDs) (*OcnMap, error)

	// ListElements gets all elements in this store
	ListElements(ctx context.Context, ch chan<- *OcnMap) error

	// ListKeys gets all keys in this store
	ListKeys(ctx context.Context, ch chan<- storage.IDs) error

	// Update updates an element
	Update(ctx context.Context, entry *OcnMap, key storage.IDs) error

	// Delete deletes an element
	Delete(ctx context.Context, key storage.IDs) error

	// Watch watches the event of this store
	Watch(ctx context.Context, ch chan<- event.Event) error

	// Print prints the map in this store for debugging
	Print()

	// PutInnerMapElem puts inner key and its value into the inner map
	PutInnerMapElem(ctx context.Context, key storage.IDs, innerKey storage.IDs, value meastype.QOffsetRange) error

	// PutInnerMapElems puts multiple OCNs into the inner map
	PutInnerMapElems(ctx context.Context, key storage.IDs, ocns map[storage.IDs]meastype.QOffsetRange) error

	// GetInnerMapElem gets inner element with inner key
	GetInnerMapElem(ctx context.Context, key storage.IDs, innerKey storage.IDs) (meastype.QOffsetRange, error)

	// UpdateInnerMapElem gets inner element with inner key
	UpdateInnerMapElem(ctx context.Context, key storage.IDs, innerKey storage.IDs, value meastype.QOffsetRange) error

	// ListAllInnerElement gets all inner element in this store
	ListAllInnerElement(ctx context.Context, ch chan<- Entry) error

	// ListInnerElement gets all inner element in this store
	ListInnerElement(ctx context.Context, key storage.IDs, ch chan<- InnerEntry) error

	// DeleteInnerElement deletes an inner element
	DeleteInnerElement(ctx context.Context, key storage.IDs, innerKey storage.IDs) error
}

type store struct {
	storage  map[storage.IDs]*OcnMap
	mu       sync.RWMutex
	watchers *watcher.Watchers
}

func (s *store) Put(_ context.Context, key storage.IDs, value *OcnMap) (*OcnMap, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.storage[key] = value
	s.watchers.Send(event.Event{
		Key:   key,
		Value: value,
		Type:  storage.Created,
	})
	return value, nil
}

func (s *store) Get(_ context.Context, key storage.IDs) (*OcnMap, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if v, ok := s.storage[key]; ok {
		return v, nil
	}
	return nil, errors.New(errors.NotFound, "the storage entry does not exist")
}

func (s *store) ListElements(_ context.Context, ch chan<- *OcnMap) error {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if len(s.storage) == 0 {
		return errors.New(errors.NotFound, "no storage entries stored")
	}

	for _, entry := range s.storage {
		ch <- entry
	}
	close(ch)
	return nil
}

func (s *store) ListKeys(_ context.Context, ch chan<- storage.IDs) error {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if len(s.storage) == 0 {
		return errors.New(errors.NotFound, "no storage entries stored")
	}

	for key := range s.storage {
		ch <- key
	}
	close(ch)
	return nil
}

func (s *store) Update(_ context.Context, entry *OcnMap, key storage.IDs) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.storage[key]; ok {
		s.storage[key] = entry
		s.watchers.Send(event.Event{
			Key:   key,
			Value: entry,
			Type:  storage.Updated,
		})
	}

	return errors.New(errors.NotFound, "no storage entry does not exist; put the entry first")
}

func (s *store) Delete(_ context.Context, key storage.IDs) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.storage, key)
	return nil
}

func (s *store) Watch(ctx context.Context, ch chan<- event.Event) error {
	id := uuid.New()
	err := s.watchers.AddWatcher(id, ch)
	if err != nil {
		log.Error(err)
		close(ch)
		return err
	}
	go func() {
		<-ctx.Done()
		err = s.watchers.RemoveWatcher(id)
		if err != nil {
			log.Error(err)
		}
		close(ch)
	}()
	return nil
}

func (s *store) Print() {
	s.mu.Lock()
	defer s.mu.Unlock()
	for k, v := range s.storage {
		log.Infof("key - %v / value - %v", k, v)
	}
}

func (s *store) PutInnerMapElem(_ context.Context, key storage.IDs, innerKey storage.IDs, value meastype.QOffsetRange) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.storage[key]; !ok {
		return errors.NewNotFound("inner map does not exist")
	}
	s.storage[key].Value[innerKey] = value
	return nil
}

func (s *store) PutInnerMapElems(_ context.Context, key storage.IDs, ocns map[storage.IDs]meastype.QOffsetRange) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	// verify if there is a key
	if _, ok := s.storage[key]; !ok {
		return errors.NewNotFound("inner map does not exist")
	}
	for k, v := range ocns {
		s.storage[key].Value[k] = v
	}
	return nil
}

func (s *store) GetInnerMapElem(_ context.Context, key storage.IDs, innerKey storage.IDs) (meastype.QOffsetRange, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if _, ok := s.storage[key].Value[innerKey]; !ok {
		return 0, errors.NewNotFound("element does not exist")
	}
	return s.storage[key].Value[innerKey], nil
}

func (s *store) UpdateInnerMapElem(_ context.Context, key storage.IDs, innerKey storage.IDs, value meastype.QOffsetRange) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.storage[key]; !ok {
		return errors.NewNotFound("inner map does not exist")
	}
	s.storage[key].Value[innerKey] = value
	return nil
}

func (s *store) ListAllInnerElement(_ context.Context, ch chan<- Entry) error {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if len(s.storage) == 0 {
		return errors.New(errors.NotFound, "no storage entries stored")
	}

	for k, v := range s.storage {
		if len(v.Value) == 0 {
			return errors.New(errors.NotFound, "no inner map in storage stored")
		}
		for ik, iv := range v.Value {
			ch <- Entry{
				Key: k,
				Value: InnerEntry{
					Key:   ik,
					Value: iv,
				},
			}
		}
	}
	close(ch)
	return nil
}

func (s *store) ListInnerElement(_ context.Context, key storage.IDs, ch chan<- InnerEntry) error {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if _, ok := s.storage[key]; !ok {
		return errors.NewNotFound("no inner map in storage stored")
	}
	for k, v := range s.storage[key].Value {
		ch <- InnerEntry{
			Key:   k,
			Value: v,
		}
	}
	close(ch)
	return nil
}

func (s *store) DeleteInnerElement(_ context.Context, key storage.IDs, innerKey storage.IDs) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.storage[key].Value, innerKey)
	return nil
}
