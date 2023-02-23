package ott

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"sync"
	"time"
)

type (
	Token struct {
		Data    [32]byte
		Expires time.Time
	}

	Store struct {
		sync.RWMutex
		slice         []*Token
		set           map[[32]byte]struct{}
		ttl           time.Duration
		cleanupPeriod time.Duration
	}
)

func NewToken(ttl time.Duration) *Token {
	t := new(Token)
	_, _ = rand.Read(t.Data[:])
	t.Expires = time.Now().Add(ttl)
	return t
}

func (t *Token) String() string {
	return base64.StdEncoding.EncodeToString(t.Data[:])
}

func TokenDataFromBase64(b64 string) ([32]byte, error) {
	var b [32]byte
	_, err := base64.StdEncoding.Decode(b[:], []byte(b64))
	if err != nil {
		return [32]byte{}, err
	}
	return b, nil
}

func NewStore(ttl, cleanupPeriod time.Duration) *Store {
	return &Store{
		RWMutex:       sync.RWMutex{},
		slice:         make([]*Token, 0, 1024),
		set:           make(map[[32]byte]struct{}),
		ttl:           ttl,
		cleanupPeriod: cleanupPeriod,
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

func (s *Store) RemoveExpiredLoop(ctx context.Context) {
	ticker := time.NewTicker(s.cleanupPeriod)
	for {
		select {
		case <-ctx.Done():
			ticker.Stop()
			return
		case <-ticker.C:
			s.RemoveExpired()
		}
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
			now := time.Now()
			if !now.Before(token.Expires) {
				return nil, false
			}
			return token, true
		}
	}
	return nil, false
}

func (s *Store) getExpiredIdx() int {
	now := time.Now()

	imax := -1

	if len(s.slice) > 0 && now.Before(s.slice[0].Expires) {
		return imax
	}

	for i, token := range s.slice {
		if !now.Before(token.Expires) {
			imax = i
		} else {
			break
		}
	}

	return imax
}
