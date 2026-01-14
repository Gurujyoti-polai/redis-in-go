package commands

import "redis-from-scratch/internal/storage"

func RPUSH(store *storage.Store, args []any) any {
	key := args[0].(string)

	values := []string{}
	for _, v := range args[1:] {
		values = append(values, v.(string))
	}

	n := store.RPUSH(key, values)
	if n == -1 {
		return "WRONGTYPE"
	}
	return n
}
