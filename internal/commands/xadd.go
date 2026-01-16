package commands

import "redis-from-scratch/internal/storage"

func XADD(store *storage.Store, args []any) any {
	key := args[0].(string)
	id := args[1].(string)

	fields := make(map[string]string)
	for i := 2; i < len(args); i += 2 {
		fields[args[i].(string)] = args[i+1].(string)
	}

	newID, ok := store.AddStreamEntry(key, id, fields)
	if !ok {
		return "WRONGTYPE"
	}

	return newID
}
