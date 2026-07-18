package command

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/xonoxc/expts/redis-recreation/internal/resp"
	"github.com/xonoxc/expts/redis-recreation/internal/store"
)

func newTestStore() *store.Store {
	return store.New()
}

func TestDispatchSETAndGet(t *testing.T) {
	st := newTestStore()
	d := NewDispatcher(st)

	got, err := d.Dispatch(Command{Name: "SET", Args: []string{"foo", "bar"}})
	if err != nil {
		t.Fatalf("SET returned error: %v", err)
	}
	if string(got) != string(resp.SerializeSimpleString("OK")) {
		t.Fatalf("expected OK got %s", got)
	}

	got, err = d.Dispatch(Command{Name: "GET", Args: []string{"foo"}})
	if err != nil {
		t.Fatalf("GET returned error: %v", err)
	}
	if string(got) != string(resp.SerializeBulkString("bar")) {
		t.Fatalf("expected $3\\r\\nbar\\r\\n got %s", got)
	}
}

func TestDispatchGETNonExistent(t *testing.T) {
	st := newTestStore()
	d := NewDispatcher(st)

	got, err := d.Dispatch(Command{Name: "GET", Args: []string{"nope"}})
	if err != nil {
		t.Fatalf("GET returned error: %v", err)
	}
	if string(got) != string(resp.SerializeNull()) {
		t.Fatalf("expected null got %s", got)
	}
}

func TestDispatchDELETE(t *testing.T) {
	st := newTestStore()
	d := NewDispatcher(st)

	d.Dispatch(Command{Name: "SET", Args: []string{"k", "v"}})
	got, err := d.Dispatch(Command{Name: "DELETE", Args: []string{"k"}})
	if err != nil {
		t.Fatalf("DELETE returned error: %v", err)
	}
	if string(got) != string(resp.SerializeInteger(1)) {
		t.Fatalf("expected :1\\r\\n got %s", got)
	}

	got, err = d.Dispatch(Command{Name: "GET", Args: []string{"k"}})
	if err != nil {
		t.Fatalf("GET returned error: %v", err)
	}
	if string(got) != string(resp.SerializeNull()) {
		t.Fatalf("expected null after delete got %s", got)
	}
}

func TestDispatchSETWithEX(t *testing.T) {
	st := newTestStore()
	d := NewDispatcher(st)

	_, err := d.Dispatch(Command{Name: "SET", Args: []string{"k", "v", "EX", "2"}})
	if err != nil {
		t.Fatalf("SET EX returned error: %v", err)
	}

	got, err := d.Dispatch(Command{Name: "GET", Args: []string{"k"}})
	if err != nil {
		t.Fatalf("GET returned error: %v", err)
	}
	if string(got) != string(resp.SerializeBulkString("v")) {
		t.Fatalf("expected $1\\r\\nv\\r\\n got %s", got)
	}
}

func TestDispatchSETWithInvalidEXKeyword(t *testing.T) {
	st := newTestStore()
	d := NewDispatcher(st)

	_, err := d.Dispatch(Command{Name: "SET", Args: []string{"k", "v", "PX", "100"}})
	if err != ErrInvalidSyntax {
		t.Fatalf("expected ErrInvalidSyntax got %v", err)
	}
}

func TestDispatchSETWithNonNumericExpiration(t *testing.T) {
	st := newTestStore()
	d := NewDispatcher(st)

	_, err := d.Dispatch(Command{Name: "SET", Args: []string{"k", "v", "EX", "abc"}})
	if err != ErrInvalidExpirationTime {
		t.Fatalf("expected ErrInvalidExpirationTime got %v", err)
	}
}

func TestDispatchGETWithTooManyArgs(t *testing.T) {
	st := newTestStore()
	d := NewDispatcher(st)

	_, err := d.Dispatch(Command{Name: "GET", Args: []string{"k", "extra"}})
	if err != ErrNotEnoughArguements {
		t.Fatalf("expected ErrNotEnoughArguements got %v", err)
	}
}

func TestDispatchGETWithNoArgs(t *testing.T) {
	st := newTestStore()
	d := NewDispatcher(st)

	_, err := d.Dispatch(Command{Name: "GET", Args: []string{}})
	if err != ErrNotEnoughArguements {
		t.Fatalf("expected ErrNotEnoughArguements got %v", err)
	}
}

func TestDispatchUnknownCommand(t *testing.T) {
	st := newTestStore()
	d := NewDispatcher(st)

	_, err := d.Dispatch(Command{Name: "PING", Args: []string{}})
	if err != ErrrUnkownCommand {
		t.Fatalf("expected ErrrUnkownCommand got %v", err)
	}
}

func TestDispatchSETWrongArgCount(t *testing.T) {
	st := newTestStore()
	d := NewDispatcher(st)

	_, err := d.Dispatch(Command{Name: "SET", Args: []string{"k"}})
	if err == nil {
		t.Fatal("expected error for SET with 1 arg")
	}
}

func TestDispatchSETWithOddArgCount(t *testing.T) {
	st := newTestStore()
	d := NewDispatcher(st)

	_, err := d.Dispatch(Command{Name: "SET", Args: []string{"k", "v", "extra"}})
	if err == nil {
		t.Fatal("expected error for SET with 3 args")
	}
}

func TestDispatchSETOverwritesPreviousValue(t *testing.T) {
	st := newTestStore()
	d := NewDispatcher(st)

	d.Dispatch(Command{Name: "SET", Args: []string{"k", "first"}})
	d.Dispatch(Command{Name: "SET", Args: []string{"k", "second"}})

	got, _ := d.Dispatch(Command{Name: "GET", Args: []string{"k"}})
	if string(got) != string(resp.SerializeBulkString("second")) {
		t.Fatalf("expected second got %s", got)
	}
}

func TestDispatchSETExValueExpires(t *testing.T) {
	st := newTestStore()
	d := NewDispatcher(st)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	var wg sync.WaitGroup
	st.StartStoreGC(ctx, &wg)

	_, err := d.Dispatch(Command{Name: "SET", Args: []string{"k", "v", "EX", "1"}})
	if err != nil {
		t.Fatalf("SET EX returned error: %v", err)
	}

	time.Sleep(2 * time.Second)

	got, err := d.Dispatch(Command{Name: "GET", Args: []string{"k"}})
	if err != nil {
		t.Fatalf("GET returned error: %v", err)
	}
	if string(got) != string(resp.SerializeNull()) {
		t.Fatalf("expected null after expiration got %s", got)
	}
}

func TestCommandNewCommand(t *testing.T) {
	cmd := NewCommand("SET", []string{"k", "v"})
	if cmd.Name != "SET" {
		t.Fatalf("expected SET got %s", cmd.Name)
	}
	if len(cmd.Args) != 2 || cmd.Args[0] != "k" || cmd.Args[1] != "v" {
		t.Fatalf("expected [k v] got %v", cmd.Args)
	}
}
