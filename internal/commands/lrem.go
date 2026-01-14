package commands

import "redis-from-scratch/internal/storage"

func LREM(store *storage.Store, args []any) any {
	key := args[0].(string)
	count := args[1].(int)
	value := args[2].(string)

	n := store.LREM(key, count, value)
	if n == -1 {
		return "WRONGTYPE"
	}
	return n
}
