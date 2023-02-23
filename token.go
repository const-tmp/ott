package ott

import (
	"crypto/rand"
	"encoding/base64"
	"sync"
	"time"
)

type (
	Token struct {
		Data    [32]byte
		Expires int64
	}

	Store struct {
		sync.RWMutex
		slice []*Token
		set   map[[32]byte]struct{}
		ttl   int64
	}
)

func NewToken(ttl int64) *Token {
	t := new(Token)
	_, _ = rand.Read(t.Data[:])
	t.Expires = time.Now().Add(time.Duration(ttl)).Unix()
	return t
}

func (t *Token) String() string {
	return base64.StdEncoding.EncodeToString(t.Data[:])
}

func NewStore(ttlSec int64) *Store {
	return &Store{
		RWMutex: sync.RWMutex{},
		slice:   make([]*Token, 0, 1024),
		set:     make(map[[32]byte]struct{}),
		ttl:     ttlSec,
	}
}

func (s *Store) NewToken() *Token {
	t := NewToken(s.ttl)
	s.add(t)
	return t
}

func (s *Store) Exists(data [32]byte) bool {
	s.RLock()
	_, ok := s.set[data]
	s.RUnlock()
	return ok
}

func (s *Store) Pop(data [32]byte) (*Token, bool) {
	if !s.Exists(data) {
		return nil, false
	}
	return s.pop(data)
}

func (s *Store) RemoveExpired() {
	s.Lock()
	defer s.Unlock()
	if idx := s.getExpiredIdx(); idx != -1 {
		s.removeLeft(idx + 1)
	}
}

func (s *Store) add(token *Token) {
	s.Lock()
	s.slice = append(s.slice, token)
	s.set[token.Data] = struct{}{}
	s.Unlock()
}

func (s *Store) removeLeft(idx int) {
	for _, token := range s.slice[:idx] {
		delete(s.set, token.Data)
	}
	s.slice = s.slice[idx:]
}

func (s *Store) pop(data [32]byte) (*Token, bool) {
	s.Lock()
	defer s.Unlock()
	delete(s.set, data)
	for i, token := range s.slice {
		if token.Data == data {
			s.slice = append(s.slice[:i], s.slice[i:]...)
			now := time.Now().Unix()
			if token.Expires <= now {
				return nil, false
			}
			return token, true
		}
	}
	return nil, false
}

func (s *Store) getExpiredIdx() int {
	now := time.Now().Unix()

	imax := -1

	if len(s.slice) > 0 && s.slice[0].Expires > now {
		return imax
	}

	for i, token := range s.slice {
		if token.Expires <= now {
			imax = i
		} else {
			break
		}
	}

	return imax
}
