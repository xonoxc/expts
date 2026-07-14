# Redis Core Speedrun (2 Hours)

> Goal: Understand the architectural core of Redis by implementing the minimum system that accepts commands over TCP and stores data in memory.

**Time Limit:** 2 Hours

**Language:** Go

---

## Success Criteria

By the end I should understand:

- How Redis accepts TCP connections.
- Why everything over the network is bytes.
- Why Redis uses RESP.
- How bytes become commands.
- How commands mutate in-memory state.
- How responses become bytes again.

---

# Architecture

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

# Scope

## Networking

Implement:

- TCP Server
- Accept Connections
- Read bytes
- Write bytes

Do NOT use HTTP.

---

## RESP

Support only Arrays of Bulk Strings.

Implement:

```
*3\r\n
$3\r\n
SET\r\n
$4\r\n
name\r\n
$4\r\n
appy\r\n
```

Need to understand:

- Array Header
- Bulk String
- CRLF

Ignore every other RESP type.

---

## Commands

Implement only:

```
PING
SET key value
GET key
DEL key
```

Nothing else.

---

## Storage

Use:

```go
map[string]string
```

No persistence.

No expiry.

No eviction.

---

## Serializer

Support:

Simple String

```
+OK\r\n
```

Bulk String

```
$4\r\n
appy\r\n
```

Null Bulk String

```
$-1\r\n
```

Integer

```
:1\r\n
```

Error

```
-ERR unknown command\r\n
```

---

# Folder Structure

```
redis/

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

# Suggested Order

1. TCP Server
2. Read bytes
3. Parse RESP
4. Command Struct
5. Dispatcher
6. HashMap
7. Serialize Response
8. Test with redis-cli

---

# Test Commands

```
redis-cli -p 6379

PING

SET name appy

GET name

DEL name
```

---

# Useful References

## RESP Specification

https://redis.io/docs/latest/develop/reference/protocol-spec/

---

## Redis CLI

https://redis.io/docs/latest/develop/tools/cli/

---

## Go net Package

https://pkg.go.dev/net

---

## bufio

https://pkg.go.dev/bufio

---

## io Package

https://pkg.go.dev/io

---

# Questions I Must Answer Afterwards

- Why does Redis use TCP?
- Why are commands bytes first?
- Why parse into Command structs?
- Why serialize responses?
- Where would persistence plug in?
- Where would replication plug in?
- Where would an event loop exist?
- Why is Redis single-threaded for command execution?
- Where is the bottleneck likely to be?

---

# Absolutely DO NOT Build

- Streams
- Transactions
- Pub/Sub
- Expiration
- Lua
- Replication
- Cluster
- RDB
- AOF
- ACL
- Modules

If I finish early:

Read how Redis actually solves one of these problems.
Do not implement it today.

---

# Mission

I am NOT building Redis.

I am building the smallest executable model that explains why Redis works.
