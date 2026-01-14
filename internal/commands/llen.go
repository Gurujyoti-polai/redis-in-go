package commands

import "redis-from-scratch/internal/storage"

func LLEN(store *storage.Store, args []any) any {
	n := store.LLEN(args[0].(string))
	if n == -1 {
		return "WRONGTYPE"
	}
	return n
}
