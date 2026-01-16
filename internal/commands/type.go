package commands

import "redis-from-scratch/internal/storage"

func TYPE(store *storage.Store, args []any) any {
	key := args[0].(string)
	return store.TypeOf(key)
}
