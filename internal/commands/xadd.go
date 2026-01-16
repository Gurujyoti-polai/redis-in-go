package commands

import "redis-from-scratch/internal/storage"

func XADD(store *storage.Store, args []any) any {

	if (len(args)-2)%2 != 0 {
		return "ERR wrong number of arguments for 'xadd' command"
	}

	key := args[0].(string)
	id := args[1].(string)

	fields := make(map[string]string)
	for i := 2; i < len(args); i += 2 {
		field := args[i].(string)
		value := args[i+1].(string)
		fields[field] = value
	}

	newID, ok := store.AddStreamEntry(key, id, fields)
	if !ok {
		return "ERR The ID specified in XADD is equal or smaller than the target stream top item"
	}

	return newID
}
