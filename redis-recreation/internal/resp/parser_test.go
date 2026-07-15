package resp

import (
	"reflect"
	"testing"
)

func assertEqual(t *testing.T, got, want []string) {
	t.Helper()

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v want %v", got, want)
	}
}

func TestParseSET(t *testing.T) {
	parser := NewParser()

	buf := []byte("*3\r\n$3\r\nSET\r\n$3\r\nfoo\r\n$3\r\nbar\r\n")

	got, err := parser.Parse(buf)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	assertEqual(t, got, []string{
		"SET",
		"foo",
		"bar",
	})
}

func TestParseGET(t *testing.T) {
	parser := NewParser()

	buf := []byte("*2\r\n$3\r\nGET\r\n$3\r\nfoo\r\n")

	got, err := parser.Parse(buf)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	assertEqual(t, got, []string{
		"GET",
		"foo",
	})
}

func TestParseDEL(t *testing.T) {
	parser := NewParser()

	buf := []byte("*2\r\n$3\r\nDEL\r\n$3\r\nfoo\r\n")

	got, err := parser.Parse(buf)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	assertEqual(t, got, []string{
		"DEL",
		"foo",
	})
}

func TestEmptyBulkString(t *testing.T) {
	parser := NewParser()

	buf := []byte("*1\r\n$0\r\n\r\n")

	got, err := parser.Parse(buf)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	assertEqual(t, got, []string{""})
}

func TestParserReuse(t *testing.T) {
	parser := NewParser()

	buf := []byte("*2\r\n$3\r\nGET\r\n$3\r\nfoo\r\n")

	for range 5 {
		got, err := parser.Parse(buf)
		if err != nil {
			t.Fatal(err)
		}

		assertEqual(t, got, []string{
			"GET",
			"foo",
		})
	}
}

func TestEmptyBuffer(t *testing.T) {
	parser := NewParser()

	_, err := parser.Parse(nil)

	if err != ErrEmptyBuffer {
		t.Fatalf("expected %v got %v", ErrEmptyBuffer, err)
	}
}

func TestInvalidRootType(t *testing.T) {
	parser := NewParser()

	buf := []byte("$3\r\nSET\r\n")

	_, err := parser.Parse(buf)

	if err != ErrMalFormedBytes {
		t.Fatalf("expected %v got %v", ErrMalFormedBytes, err)
	}
}

func TestInvalidBulkType(t *testing.T) {
	parser := NewParser()

	buf := []byte("*1\r\n*1\r\n")

	_, err := parser.Parse(buf)

	if err != ErrMalFormedBytes {
		t.Fatalf("expected %v got %v", ErrMalFormedBytes, err)
	}
}

func TestIncompleteArrayHeader(t *testing.T) {
	parser := NewParser()

	buf := []byte("*3")

	_, err := parser.Parse(buf)

	if err != ErrIncomplete {
		t.Fatalf("expected %v got %v", ErrIncomplete, err)
	}
}

func TestIncompleteBulkHeader(t *testing.T) {
	parser := NewParser()

	buf := []byte("*1\r\n$3")

	_, err := parser.Parse(buf)

	if err != ErrIncomplete {
		t.Fatalf("expected %v got %v", ErrIncomplete, err)
	}
}

func TestBulkLengthMismatch(t *testing.T) {
	parser := NewParser()

	buf := []byte("*1\r\n$5\r\nabc\r\n")

	_, err := parser.Parse(buf)

	if err != ErrMalFormedBytes {
		t.Fatalf("expected %v got %v", ErrMalFormedBytes, err)
	}
}

func TestInvalidArrayLength(t *testing.T) {
	parser := NewParser()

	buf := []byte("*x\r\n")

	_, err := parser.Parse(buf)

	if err != ErrMalFormedBytes {
		t.Fatalf("expected %v got %v", ErrMalFormedBytes, err)
	}
}

func TestInvalidBulkLength(t *testing.T) {
	parser := NewParser()

	buf := []byte("*1\r\n$x\r\nabc\r\n")

	_, err := parser.Parse(buf)

	if err != ErrMalFormedBytes {
		t.Fatalf("expected %v got %v", ErrMalFormedBytes, err)
	}
}

// ---------- Known limitation ----------
//
// RESP bulk strings are binary-safe.
// Your parser currently uses readLine() instead of readN(),
// so payloads containing CRLF are not supported yet.
//
// Uncomment after implementing readN().
//
// func TestBulkStringContainingCRLF(t *testing.T) {
// 	parser := NewParser()
//
// 	buf := []byte("*1\r\n$11\r\nhello\r\nyou\r\n")
//
// 	got, err := parser.Parse(buf)
// 	if err != nil {
// 		t.Fatal(err)
// 	}
//
// 	assertEqual(t, got, []string{
// 		"hello\r\nyou",
// 	})
// }
