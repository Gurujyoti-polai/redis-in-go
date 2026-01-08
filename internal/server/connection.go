package server

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net"
	"strconv"

	"redis-from-scratch/internal/commands"
	"redis-from-scratch/internal/protocol"
)


func parseSimpleString(buffer []byte) (any, int, bool) {
	idx := bytes.Index(buffer, []byte("\r\n"))
	if idx == -1 {
		return nil, 0, false
	}

	value := string(buffer[1:idx])
	return value, idx + 2, true
}

func parseInteger(buffer []byte) (any, int, bool) {
	idx := bytes.Index(buffer, []byte("\r\n"))
	if idx == -1 {
		return nil, 0, false
	}

	num, _ := strconv.Atoi(string(buffer[1:idx]))
	return num, idx + 2, true
}

func parseError(buffer []byte) (any, int, bool) {
	idx := bytes.Index(buffer, []byte("\r\n"))
	if idx == -1 {
		return nil, 0, false
	}

	msg := string(buffer[1:idx])
	return errors.New(msg), idx + 2, true
}

func parseBulkString(buffer []byte) (any, int, bool) {
	headerEnd := bytes.Index(buffer, []byte("\r\n"))
	if headerEnd == -1 {
		return nil, 0, false
	}

	length, _ := strconv.Atoi(string(buffer[1:headerEnd]))

	// Null bulk string
	if length == -1 {
		return nil, headerEnd + 2, true
	}

	total := headerEnd + 2 + length + 2
	if len(buffer) < total {
		return nil, 0, false
	}

	start := headerEnd + 2
	end := start + length

	value := string(buffer[start:end])
	return value, total, true
}

func parseArray(buffer []byte) (any, int, bool) {
	headerEnd := bytes.Index(buffer, []byte("\r\n"))
	if headerEnd == -1 {
		return nil, 0, false
	}

	count, _ := strconv.Atoi(string(buffer[1:headerEnd]))
	offset := headerEnd + 2
	result := make([]any, 0, count)

	for i := 0; i < count; i++ {
		val, consumed, ok := parseRESP(buffer[offset:])
		if !ok {
			return nil, 0, false
		}
		result = append(result, val)
		offset += consumed
	}

	return result, offset, true
}


func parseRESP(buffer []byte) (any, int, bool) {
	if len(buffer) == 0 {
		return nil, 0, false
	}

	switch buffer[0] {
	case '+':
		return parseSimpleString(buffer)
	case '$':
		return parseBulkString(buffer)
	case '*':
		return parseArray(buffer)
	case ':':
		return parseInteger(buffer)
	case '-':
		return parseError(buffer)
	default:
		return nil, 0, false
	}
}




func HandleConnection(conn net.Conn, router *commands.Router) {
	// make a buffer
	tempBuffer := make([]byte,1024)
	permBuffer := []byte{}

	for {
		// append that all into the buffer
		n, err := conn.Read(tempBuffer)

		if err != nil {
			if err == io.EOF {
				fmt.Println("Client closed the connection")
			} else {
				fmt.Printf("Error reading: %s\n", err.Error())
			}
			break
		}

		// append to the permBuffer
		permBuffer = append(permBuffer, tempBuffer[:n]...)


		/// Try to parse as many commands as possible
		for {
			value, consumed, ok := parseRESP(permBuffer)
			if !ok {
				// Not enough data yet
				break
			}

			// Consume parsed bytes
			permBuffer = permBuffer[consumed:]

			// Expect RESP array for commands
			cmd, ok := value.([]any)
			if !ok {
				conn.Write([]byte("-ERR invalid command\r\n"))
				continue
			}

			// Execute command
			result := router.Execute(cmd)

			// Encode response
			resp := protocol.EncodeRESP(result)

			// Write back to client
			_, err := conn.Write(resp)
			if err != nil {
				fmt.Println("Write error:", err)
				return
			}
		}
	}
	
	
	
}