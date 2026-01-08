package main

import "redis-from-scratch/internal/server"

func main() {
	server.Start(":6379")
}
