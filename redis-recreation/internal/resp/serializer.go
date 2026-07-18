package resp

import (
	"fmt"
)

func SerializeBulkString(s string) []byte {
	return fmt.Appendf([]byte{}, "$%d\r\n%s\r\n", len(s), s)
}

func SerializeNull() []byte {
	return []byte("$-1\r\n")
}

func SerializeSimpleString(s string) []byte {
	return fmt.Appendf([]byte{}, "+%s\r\n", s)
}

func SerializeError(msg string) []byte {
	return fmt.Appendf([]byte{}, "-Error %s\r\n", msg)
}

func SerializeInteger(n int) []byte {
	return fmt.Appendf([]byte{}, ":%d\r\n", n)
}
