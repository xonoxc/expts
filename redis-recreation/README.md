# Redis Recreation

A minimal Redis clone written in Go. Implements the core subset of Redis — enough to understand how Redis works under the hood without the complexity of the full system.

**Language:** Go

---

## What It Is

A TCP server that speaks the RESP protocol, accepts commands, and stores data in memory. A stripped-down Redis that does the essential things and nothing more.

---

## Scope

### Features

- TCP server accepting multiple connections
- RESP protocol parsing and serialization
- In-memory key-value storage

### Commands

```
PING                  — connectivity check
SET key value         — store a value
GET key               — retrieve a value
DEL key               — delete a key
```

### What It Does NOT Have

- Persistence (no RDB, no AOF)
- Expiration / TTL
- Eviction policies
- Pub/Sub
- Transactions
- Streams
- Replication
- Clustering
- Lua scripting
- ACL / authentication
- Modules

---

## Architecture

```
         TCP Socket
              │
              ▼
     Read []byte from Conn
              │
              ▼
        RESP Parser
              │
              ▼
    Command { Name, Args }
              │
              ▼
      Command Dispatcher
              │
              ▼
       In-Memory HashMap
              │
              ▼
      RESP Serializer
              │
              ▼
       conn.Write(bytes)
```

---

## Folder Structure

```
cmd/
    main.go

internal/
    server/
        server.go

    resp/
        parser.go
        serializer.go

    command/
        command.go
        dispatcher.go

    store/
        store.go
```

---

## Running

```
go run ./cmd
```

The server listens on port `6379` by default.

---

## Testing

Connect with `redis-cli`:

```
redis-cli -p 6379

PING
SET name appy
GET name
DEL name
```

---

## References

- [RESP Specification](https://redis.io/docs/latest/develop/reference/protocol-spec/)
- [Redis CLI](https://redis.io/docs/latest/develop/tools/cli/)
- [Go net Package](https://pkg.go.dev/net)
- [Go bufio Package](https://pkg.go.dev/bufio)
- [Go io Package](https://pkg.go.dev/io)
