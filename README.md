# fizzbuzz-server

A REST API implementing a generalised fizz-buzz, built with Go's standard
library only. Returns numbers 1–N with configurable divisor/string pairs.

## Requirements

Go 1.25.5 (declared in `go.mod`)

## Run it

```bash
go run ./cmd/server
# or
make run
```

Listens on `:8080` by default. Override with the `PORT` env var:

```bash
PORT=9000 go run ./cmd/server
```

`PORT` must be a valid integer in the range 1–65535; an invalid value exits
immediately with a logged error.

## Test it

```bash
make test
# or
go test ./... -race -cover
```

Coverage (from actual run, `go test ./... -race -cover`):

| Package | Coverage |
| --- | --- |
| `internal/fizzbuzz` | 100.0% |
| `internal/stats` | 100.0% |
| `internal/server` | 94.7% |
| `cmd/server` | 0.0% (main, not unit-tested by convention) |

The remaining 5.3% in `internal/server` is the `writeJSON` encode-error
branch, which requires a broken `io.Writer` to reach and is not worth the
added test complexity.

## API

All responses have `Content-Type: application/json; charset=utf-8`. All
error bodies have the shape `{"error": "..."}`.

---

### `GET /fizzbuzz`

Produces the fizz-buzz sequence for the range `[1, limit]`.

**Query parameters** (all required):

| Parameter | Type | Constraint |
| --- | --- | --- |
| `int1` | integer | > 0 |
| `int2` | integer | > 0 |
| `limit` | integer | 1 – 1,000,000 (inclusive) |
| `str1` | string | non-empty, ≤ 100 characters |
| `str2` | string | non-empty, ≤ 100 characters |

`str1` and `str2` are present in the query string but not validated as
integers, empty string is rejected by `Validate`, absent key returns an
empty string (treated as empty by the validator).

**Success, `200 OK`**

```json
{"result": ["1","2","fizz","4","buzz","fizz","7","8","fizz","buzz","11","fizz","13","14","fizzbuzz"]}
```

```bash
curl "localhost:8080/fizzbuzz?int1=3&int2=5&limit=15&str1=fizz&str2=buzz"
```

**Errors, `400 Bad Request`**

Validation is checked in the order below; only the first failing condition
is reported per request.

| Condition | `error` field |
| --- | --- |
| `int1` absent or empty | `int1 is required` |
| `int1` not a valid integer | `int1 must be a valid integer` |
| `int2` absent or empty | `int2 is required` |
| `int2` not a valid integer | `int2 must be a valid integer` |
| `limit` absent or empty | `limit is required` |
| `limit` not a valid integer | `limit must be a valid integer` |
| `int1` ≤ 0 | `int1 must be a positive integer` |
| `int2` ≤ 0 | `int2 must be a positive integer` |
| `limit` < 1 or > 1,000,000 | `limit must be between 1 and 1000000` |
| `str1` empty | `str1 must not be empty` |
| `str2` empty | `str2 must not be empty` |
| `str1` > 100 characters | `str1 must be no longer than 100 characters` |
| `str2` > 100 characters | `str2 must be no longer than 100 characters` |

Integer parameters (`int1`, `int2`, `limit`) are parsed before domain
validation runs, so a missing or non-numeric value returns the parse error
before the range check is reached.

**Other status codes for this route**

| Status | Condition |
| --- | --- |
| `405 Method Not Allowed` | Any method other than GET, body: `{"error":"method not allowed, use GET"}` |
| `500 Internal Server Error` | Unexpected panic (caught by recovery middleware) |

---

### `GET /statistics`

Returns the `/fizzbuzz` parameter set that has been requested most often,
and its hit count. Only successfully validated requests are counted.

**Success, `200 OK`**

```json
{"int1":3,"int2":5,"limit":15,"str1":"fizz","str2":"buzz","hits":42}
```

```bash
curl "localhost:8080/statistics"
```

**Errors**

| Status | Condition | `error` field |
| --- | --- | --- |
| `404 Not Found` | No validated requests recorded yet | `no requests have been recorded yet` |
| `405 Method Not Allowed` | Non-GET method | `method not allowed, use GET` |

In case of a tie between two parameter sets with equal hit counts, the
winner is arbitrary (map-iteration order).

---

### `GET /healthz`

Liveness probe for load balancers and orchestrators.

**Success, `200 OK`** (always, as long as the process is up)

```json
{"status":"ok"}
```

```bash
curl "localhost:8080/healthz"
```

| Status | Condition |
|---|---|
| `405 Method Not Allowed` | Non-GET method, body: `{"error":"method not allowed, use GET"}` |

---

### Unknown paths

Any path not listed above returns:

```
404  {"error":"no such endpoint"}
```

## Docker

```bash
# Build
make docker
# or
docker build -t fizzbuzz-server .

# Run
docker run -p 8080:8080 fizzbuzz-server
```

Two-stage build: `golang:1.25-alpine` compiles a fully static binary
(`CGO_ENABLED=0`); the runtime image is
`gcr.io/distroless/static-debian12:nonroot`, no shell, no package manager,
runs as a non-root user. This minimises the production attack surface.

## Design decisions

- **Standard library only.** `net/http`'s `ServeMux` (Go 1.22+) supports
  method-aware routing (`"GET /fizzbuzz"`) natively. There is no problem
  here that a third-party router solves.

- **`MaxLimit = 1,000,000` and `MaxStrLen = 100`** (constants in
  `internal/fizzbuzz/fizzbuzz.go`). An unbounded `limit` is a trivial DoS
  vector, a single request could force the server to allocate gigabytes.
  An unbounded `str1`/`str2` causes proportional memory growth per unique
  key stored in the stats map. The caps are generous enough for any
  realistic use.

- **`RWMutex` for the stats store** (`internal/stats/stats.go`). Concurrent
  `GET /fizzbuzz` requests write to the store; concurrent `GET /statistics`
  requests only read. A `sync.RWMutex` lets multiple readers proceed in
  parallel without blocking each other, only writes (`Increment`) acquire
  the exclusive lock.

- **`Most()` is O(n) in distinct request types.** For a fizz-buzz service
  the number of distinct parameter combinations is small in practice; a
  full map scan is negligible. The doc comment in `stats.go` documents this
  and the straightforward upgrade path (a max-heap) if it ever mattered.

- **Tie-breaking in `Most()` is arbitrary.** Go's map iteration order is
  deliberately randomised; among tied entries, whichever one the runtime
  visits first wins. This is documented in the `Most()` doc comment and is
  the correct behaviour for "show one most-popular request" when there is no
  meaningful tiebreaker.

- **Graceful shutdown on `SIGINT`/`SIGTERM`** with a 10-second drain window
  (`context.WithTimeout(…, 10*time.Second)` in `main.go`). Ensures
  in-flight requests complete before the process exits, relevant for
  `docker stop` and Kubernetes pod eviction.

- **Panic recovery middleware** (`internal/server/middleware.go`). Converts
  any unexpected panic in a handler into a `500` response instead of
  crashing the entire server process, preserving availability for other
  concurrent requests.

## Project structure

```
fizzbuzz-server/
├── cmd/server/main.go              # entrypoint: port config, HTTP server lifecycle, graceful shutdown
├── internal/
│   ├── fizzbuzz/
│   │   ├── fizzbuzz.go             # core algorithm (Generate), validation (Validate), constants
│   │   └── fizzbuzz_test.go        # unit tests for Generate and Validate
│   ├── server/
│   │   ├── router.go               # route registration and middleware chain wiring
│   │   ├── fizzbuzz_handler.go     # GET /fizzbuzz: parse, validate, generate, record stats
│   │   ├── stats_handler.go        # GET /statistics: read and decode the most-used request
│   │   ├── health_handler.go       # GET /healthz: liveness probe
│   │   ├── middleware.go           # request ID, structured logging, panic recovery
│   │   ├── respond.go              # writeJSON / writeError helpers
│   │   └── server_test.go          # HTTP-layer tests using httptest (no live server)
│   └── stats/
│       ├── stats.go                # thread-safe in-memory hit counter (RWMutex-guarded map)
│       └── stats_test.go           # unit + concurrent stress tests for Store
├── Dockerfile                      # two-stage build: golang:1.25-alpine -> distroless
├── Makefile                        # run, build, test, lint, docker targets
├── go.mod                          # module declaration, no external dependencies
├── .gitignore
└── .dockerignore
```
