# Redis From Scratch in Go

A small Redis-compatible, in-memory database built from scratch in Go. It accepts commands from the standard `redis-cli`, implements a subset of Redis data types, handles concurrent clients, and restores data after a restart using an append-only file (AOF).

This is an educational project. It focuses on the internals behind network servers, wire protocols, concurrent data structures, and persistence rather than full Redis compatibility.

## Features

- TCP server on Redis's default port, `6379`
- RESP request parsing and response serialization
- Concurrent client handling with goroutines
- Thread-safe string and hash storage with `sync.RWMutex`
- Append-only persistence for mutating commands
- Startup recovery by replaying persisted RESP commands
- Unit tests, malformed-input tests, and race-detector coverage

## Supported commands

| Command | Example | Response |
| --- | --- | --- |
| `PING` | `PING` | `PONG` |
| `PING message` | `PING hello` | `hello` |
| `SET` | `SET language Go` | `OK` |
| `GET` | `GET language` | `Go` or null |
| `HSET` | `HSET users u1 Ada` | `1` for new, `0` for overwrite |
| `HGET` | `HGET users u1` | `Ada` or null |
| `HGETALL` | `HGETALL users` | Alternating field/value array |

## Architecture

```text
redis-cli
    |
    | TCP / RESP
    v
connection goroutine
    |
    +--> RESP reader --> command handler --> synchronized in-memory maps
    |                         |
    |                         +--> AOF for SET and HSET
    |
    +<-- RESP writer <--------+

startup --> replay database.aof --> command handler --> restored memory
```

The server uses one goroutine per connection. Reads can proceed concurrently, while mutexes protect shared maps. Durable mutations are serialized so the order applied in memory matches the order recorded in the AOF.

## Running the project

Requirements:

- Go 1.26 or newer
- `redis-cli` for manual testing

Start the server:

```bash
go run .
```

In a second terminal:

```bash
redis-cli -p 6379 ping
redis-cli -p 6379 set language Go
redis-cli -p 6379 get language
redis-cli -p 6379 hset users u1 Ada
redis-cli -p 6379 hgetall users
```

Build a binary:

```bash
go build -o redis-from-scratch .
./redis-from-scratch
```

## Testing

Run the complete suite with Go's race detector:

```bash
go test -race ./...
```

The tests cover RESP decoding and encoding, invalid protocol input, command behavior, append-only writes, corrupt-log detection, and restart recovery.

## Persistence model

Only mutating commands (`SET` and `HSET`) are appended to `database.aof`. Each command is stored in RESP—the same format used on the network—and synchronized to disk before the server acknowledges it. On startup, the server parses and replays the file to reconstruct its in-memory state.

Syncing every mutation favors durability and simplicity over write throughput. Production Redis offers configurable AOF synchronization policies and additional recovery mechanisms.

## Project structure

```text
.
├── main.go          # TCP listener and connection lifecycle
├── resp.go          # RESP parser, serializer, reader, and writer
├── handler.go       # Command dispatch and in-memory data structures
├── server.go        # Mutation ordering and persistence coordination
├── aof.go           # Append-only writing and startup replay
└── *_test.go        # Unit and recovery tests
```

## Intentional limitations

- Implements a small subset of Redis commands and RESP2 types
- No key expiration, transactions, replication, clustering, or authentication
- No AOF compaction or repair of a partially written final command
- Fixed port and persistence filename
- Designed for learning, not production workloads

## What this project demonstrates

- Building a TCP protocol server without external libraries
- Recursive parsing and serialization of a wire protocol
- Applying Go interfaces such as `io.Reader` and `io.Writer`
- Coordinating goroutines with mutexes and the race detector
- Reasoning about durability, command ordering, and crash recovery

Built while following and extending the [Build Redis from Scratch](https://www.build-redis-from-scratch.dev/en/introduction) tutorial.
