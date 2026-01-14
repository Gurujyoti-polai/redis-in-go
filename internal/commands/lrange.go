package commands

import "redis-from-scratch/internal/storage"

func LRANGE(store *storage.Store, args []any) any {
	key := args[0].(string)
	start := args[1].(int)
	stop := args[2].(int)

	res, ok := store.LRANGE(key, start, stop)
	if !ok {
		return "WRONGTYPE"
	}
	return res
}
