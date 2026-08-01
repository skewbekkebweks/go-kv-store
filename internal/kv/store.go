package kv

import (
	"errors"
	"sort"
	"strings"
	"time"
)

var ErrInvalidTTL = errors.New("ttl < 0")

type entry struct {
	value     string
	expiresAt time.Time
}

type Store struct {
	data map[string]entry
	now  func() time.Time
}

func NewStore() *Store {
	return newStore(time.Now)
}

func newStore(now func() time.Time) *Store {
	return &Store{data: make(map[string]entry), now: now}
}

// Set returns true if the key was created and false if it was updated.
func (s *Store) Set(key, value string, ttl time.Duration) (bool, error) {
	if ttl < 0 {
		return false, ErrInvalidTTL
	}
	now := s.now()
	s.deleteIfExpired(key, now)
	_, ok := s.data[key]
	if ttl == 0 {
		s.data[key] = entry{value: value, expiresAt: time.Time{}}
	} else {
		s.data[key] = entry{value: value, expiresAt: now.Add(ttl)}
	}
	return !ok, nil
}

// Get returns true if the key exists and false otherwise.
func (s *Store) Get(key string) (string, bool) {
	s.deleteIfExpired(key, s.now())
	value, ok := s.data[key]
	return value.value, ok
}

// Delete returns true if the key existed and false otherwise.
func (s *Store) Delete(key string) bool {
	s.deleteIfExpired(key, s.now())
	_, ok := s.data[key]
	delete(s.data, key)
	return ok
}

func (s *Store) Keys(prefix string, limit int) []string {
	now := s.now()
	var res []string
	for key := range s.data {
		s.deleteIfExpired(key, now)
		if _, ok := s.data[key]; !ok {
			continue
		}
		if strings.HasPrefix(key, prefix) {
			res = append(res, key)
		}
	}

	sort.Strings(res)
	limit = min(len(res), max(0, limit))
	return res[:limit]
}

// Expire returns true if ttl updated and false otherwise.
func (s *Store) Expire(key string, ttl time.Duration) (bool, error) {
	if ttl < 0 {
		return false, ErrInvalidTTL
	}
	now := s.now()
	s.deleteIfExpired(key, now)
	item, ok := s.data[key]
	if !ok {
		return false, nil
	}
	if ttl == 0 {
		item.expiresAt = time.Time{}
	} else {
		item.expiresAt = now.Add(ttl)
	}
	s.data[key] = item
	return true, nil
}

func (s *Store) deleteIfExpired(key string, now time.Time) {
	value, ok := s.data[key]
	if !ok {
		return
	}
	if !value.expiresAt.IsZero() && !now.Before(value.expiresAt) {
		delete(s.data, key)
	}
}
