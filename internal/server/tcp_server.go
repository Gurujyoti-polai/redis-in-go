package server

import (
	"log"
	"net"

	"redis-from-scratch/internal/commands"
	"redis-from-scratch/internal/storage"
)

func Start(addr string) {
	// 1. Create shared store (single source of truth)
	store := storage.NewStore()

	// 2. Create router (command dispatcher)
	router := commands.NewRouter(store)

	// 3. Start TCP listener
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Redis clone listening on", addr)

	// 4. Accept loop (never exits)
	for {
		conn, err := ln.Accept()
		if err != nil {
			continue
		}

		// 5. One goroutine per client
		go HandleConnection(conn, router)
	}
}
