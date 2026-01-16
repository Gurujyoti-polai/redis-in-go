package commands

import "redis-from-scratch/internal/storage"

func XRANGE(store *storage.Store, args []any) any {
	key := args[0].(string)
	start := args[1].(string)
	end := args[2].(string)

	entries, ok := store.RangeStream(key, start, end)
	if !ok {
		return "WRONGTYPE"
	}

	// RESP format:
	// [
	//   [ id, [ field, value, field, value ] ],
	//   ...
	// ]

	resp := []any{}
	for _, e := range entries {
		fields := []any{}
		for k, v := range e.Fields {
			fields = append(fields, k, v)
		}
		resp = append(resp, []any{e.ID, fields})
	}

	return resp
}
