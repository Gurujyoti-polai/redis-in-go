package commands

import (
	"redis-from-scratch/internal/storage"
	"time"
)

func BLPOP(store *storage.Store, args []any) any {
	timeout := args[len(args)-1].(int)

	keys := []string{}
	for i := 0; i < len(args)-1; i++ {
		keys = append(keys, args[i].(string))
	}

	store.Mu.Lock()
	for _, key := range keys {
		val, ok, wrong := store.TryLPop(key)
		if wrong {
			store.Mu.Unlock()
			return "WRONGTYPE"
		}
		if ok {
			store.Mu.Unlock()
			return []any{key, val}
		}
	}

	bc := &storage.BlockedClient{
		Keys: keys,
		Ch:   make(chan []any, 1),
	}
	store.Register(keys, bc)
	store.Mu.Unlock()

	if timeout == 0 {
		return <-bc.Ch
	}

	select {
	case res := <-bc.Ch:
		return res
	case <-time.After(time.Duration(timeout) * time.Second):
		store.Mu.Lock()
		store.Unregister(bc)
		store.Mu.Unlock()
		return nil
	}
}