package commands

import "redis-from-scratch/internal/storage"

func Ping(store *storage.Store, args []any) any {
	if len(args) == 1 {
		return args[0]
	}
	return "PONG"
}