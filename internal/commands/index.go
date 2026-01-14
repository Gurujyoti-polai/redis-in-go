package commands

import (
	"errors"
	"redis-from-scratch/internal/storage"
	"strings"
)

type CommandFunc func(*storage.Store, []any) any

type Command struct {
	handler CommandFunc
	minArgs int
	maxArgs int
}

type Router struct {
	store    *storage.Store
	commands map[string]Command
}

func NewRouter(store *storage.Store) *Router {
	return &Router{
		store: store,
		commands: map[string]Command{
			"PING":  {handler: Ping, minArgs: 0, maxArgs: 1},
			"SET":   {handler: Set, minArgs: 2, maxArgs: 2},
			"GET":   {handler: Get, minArgs: 1, maxArgs: 1},

			"LPUSH": {handler: LPUSH, minArgs: 2, maxArgs: -1},
			"RPUSH": {handler: RPUSH, minArgs: 2, maxArgs: -1},
			"LLEN":  {handler: LLEN, minArgs: 1, maxArgs: 1},
			"LRANGE":{handler: LRANGE, minArgs: 3, maxArgs: 3},
			"LREM":  {handler: LREM, minArgs: 3, maxArgs: 3},

			"BLPOP": {handler: BLPOP, minArgs: 2, maxArgs: -1},
		},

	}
}

func (r *Router) Execute(cmd []any) any {
	if len(cmd) == 0 {
		return errors.New("ERR empty command")
	}

	name := strings.ToUpper(cmd[0].(string))
	args := cmd[1:]

	entry, ok := r.commands[name]
	if !ok {
		return errors.New("ERR unknown command '" + name + "'")
	}

	if len(args) < entry.minArgs || len(args) > entry.maxArgs {
		return errors.New("ERR wrong number of arguments for '" + name + "'")
	}

	return entry.handler(r.store, args)
}
