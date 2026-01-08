package commands

import (
	"errors"
	"time"

	"redis-from-scratch/internal/storage"
)

func Set(store *storage.Store, args []any) any {
	if len(args) < 2 {
		return errors.New("ERR wrong number of arguments for 'set' command")
	}

	key := args[0].(string)
	value := args[1].(string)

	var expiresAt int64 = 0 // 0 means no expiry

	// Handle optional EX / PX
	if len(args) > 2 {
		if len(args) != 4 {
			return errors.New("ERR syntax error")
		}

		option := args[2].(string)
		ttl := args[3].(int)

		now := time.Now().UnixMilli()

		switch option {
		case "EX":
			expiresAt = now + int64(ttl)*1000
		case "PX":
			expiresAt = now + int64(ttl)
		default:
			return errors.New("ERR syntax error")
		}
	}

	store.Set(key, value, expiresAt)
	return "OK"
}
