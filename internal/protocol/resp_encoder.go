package protocol

import (
	"bytes"
	"strconv"
)

var crlf = []byte("\r\n")

func EncodeRESP(value any) []byte {
	switch v := value.(type) {

	case nil:
		return []byte("$-1\r\n")

	case error:
		return append([]byte("-"+v.Error()), crlf...)

	case string:
		return encodeBulkString(v)

	case int:
		return []byte(":" + strconv.Itoa(v) + "\r\n")

	case []any:
		return encodeArray(v)

	default:
		return []byte("-ERR unknown response type\r\n")
	}
}

func encodeBulkString(s string) []byte {
	var buf bytes.Buffer
	buf.WriteByte('$')
	buf.WriteString(strconv.Itoa(len(s)))
	buf.Write(crlf)
	buf.WriteString(s)
	buf.Write(crlf)
	return buf.Bytes()
}

func encodeArray(arr []any) []byte {
	var buf bytes.Buffer
	buf.WriteByte('*')
	buf.WriteString(strconv.Itoa(len(arr)))
	buf.Write(crlf)

	for _, item := range arr {
		buf.Write(EncodeRESP(item))
	}

	return buf.Bytes()
}
