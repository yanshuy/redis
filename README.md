# Build you own Redis in Go

This is a from-scratch, weekend-style Redis clone written in Go while following the Codecrafters Redis challenge. It aims to be readable and hackable more than production‑ready. I leaned into implementing core commands first, then persistence, then some extras (streams, blocking ops) without over‑engineering.

## Implemented

Data structures / commands:
- Strings: `PING`, `ECHO`, `SET`, `GET` (with `PX` / `EX` millisecond or second TTL)
- Keys/meta: `TYPE`, `KEYS`, `CONFIG GET` (for `dir`, `dbfilename`)
- Lists: `RPUSH`, `LPUSH`, `LPOP` (with count), `LLEN`, `LRANGE` (supports negative indices), `BLPOP` (basic blocking pop with timeout)
- Streams (early draft): `XADD`, `XRANGE` (simple slice-backed stream; no consumer groups / trimming yet)

Protocol / parsing:
- RESP parsing for Simple Strings, Errors, Integers, Bulk Strings, Arrays
- Command dispatcher with per‑command validation moved into `HandleXyz` helpers

Expiration:
- Per‑key TTL using millisecond deadlines
- Lazy check on access + scheduled removal via timer (simple approach)

Persistence (RDB snapshot draft):
- Minimal RDB writer for string keys (writes header, AUX fields, DB selector, hash table size, entries, EOF)
- Basic loader that reconstructs string keys (+ TTL) if the snapshot exists
- Custom length encoding (6‑bit / 14‑bit / 32‑bit) implemented

Blocking ops:
- `BLPOP` implemented with a subscribe/wait pattern (not perfectly Redis‑accurate but good enough for the challenge stage)

Tests:
- Unit-ish tests driving parsing & execution for core commands, list ops, negative LRANGE indices, blocking pop, snapshot save/restore basics

## Folder Layout
```
app/
	main.go              Entry point (flag parsing, TCP accept loop)
	request/             RESP parsing & command handlers
	store/               In-memory data store, lists, streams, snapshot (RDB)
	RESP/                RESP data type + encoder
validator/             Small experiment for slice validation
your_program.sh        Startup wrapper used by tests
```

## Running It

Quick run (defaults to dir=tmp dbfilename=rdb.snapshot on :6379):
```
go run ./app
```

Specify snapshot directory & file:
```
go run ./app --dir /tmp/redis-files --dbfilename dump.rdb
```

Using the provided script (Codecrafters harness style):
```
./your_program.sh --dir /tmp/redis-files --dbfilename dump.rdb
```

Then talk to it with redis-cli:
```
redis-cli -p 6379 PING
redis-cli SET foo bar PX 5000
redis-cli GET foo
redis-cli RPUSH mylist a b c
redis-cli LRANGE mylist 0 -1
redis-cli BLPOP mylist 3
redis-cli XADD mystream * field value
redis-cli XRANGE mystream - +
redis-cli SAVE         # (Triggers snapshot stage if wired to command, or call Go method directly)
redis-cli CONFIG GET dir
```

## Testing
```
go test ./...
```
Tests currently focus on request parsing + command semantics. Add more around snapshot round‑trip and stream edge cases if you extend the project.

