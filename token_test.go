package ott

import (
	"context"
	"encoding/base64"
	"fmt"
	"math/rand"
	"sync"
	"testing"
	"time"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

func TestToken(t *testing.T) {
	t.Log(NewToken(1))
}

func TestSlice(t *testing.T) {
	slice := []int{1, 2, 3, 4, 5}
	for i := 0; i < 5; i++ {
		t.Log(slice[:i], slice[i+1:])
	}
}

func TestStore(t *testing.T) {
	store := NewStore(5)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	wg := new(sync.WaitGroup)
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go g(ctx, store, wg, i)
	}

	wg.Wait()

	for i := 0; i < 6; i++ {
		time.Sleep(time.Duration(i) * time.Second)
		fmt.Printf("len = %d\n", len(store.slice))
		store.RemoveExpired()
	}
}

func g(ctx context.Context, store *Store, wg *sync.WaitGroup, n int) {
Loop:
	for {
		select {
		case <-ctx.Done():
			wg.Done()
			break Loop
		default:
			action(store, n)
			time.Sleep(time.Duration(rand.Int63n(1000)) * time.Millisecond)
		}
	}
}

func action(store *Store, i int) {
	n := rand.Int31n(10)
	switch {
	case n < 4:
		token := store.NewToken()
		fmt.Printf("created token: %s\n", base64.StdEncoding.EncodeToString(token.Data[:]))
	case n > 6:
		fmt.Println("removing expired")
		store.RemoveExpired()
	default:
		fmt.Printf("goroutine %d len = %d\n", i, len(store.slice))
	}
}

func BenchmarkStore(b *testing.B) {
	store := NewStore(5)

	for i := 0; i < b.N; i++ {
		token := store.NewToken()
		t, ok := store.Pop(token.Data)
		if !ok {
			b.Error("not ok", t.Data)
		}
		if token.Data != t.Data {
			b.Error("tokens are not equal")
		}
	}
}
