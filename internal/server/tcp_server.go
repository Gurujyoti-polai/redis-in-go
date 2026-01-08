package server

import (
	"log"
	"net"
)

func Start(addr string) {
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Redis clone listening on", addr)

	for {
		conn, err := ln.Accept()
		if err != nil {
			continue
		}
		go handleConnection(conn)
	}
}

func handleConnection(conn net.Conn) {
	defer conn.Close()

	// TEMP: fake PONG to validate TCP + redis-cli
	conn.Write([]byte("+PONG\r\n"))
}
