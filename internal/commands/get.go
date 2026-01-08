package commands

import "redis-from-scratch/internal/storage"

func Get(store *storage.Store, args []any) any {
	key := args[0].(string)

	val, ok := store.Get(key)
	if !ok {
		return nil
	}

	return val
}
