# bwid

Small Go library for time-sortable IDs and random tokens. Public API is in
`bwid.go` (generators) and `b62.go` (base62 encoding); design notes and
trade-offs are in `README.md`.

## Commands

```sh
gofmt -l .      # must print nothing
go vet ./...
go test -race ./...
# Linux (production target; macOS clocks are microsecond-only):
docker run --rm -v "$PWD":/src -w /src golang:1.18 go test ./...
# longer fuzz run:
go test -run '^$' -fuzz FuzzB62RoundTrip -fuzztime 1m .
```

## Constraints

- **v1 API is frozen.** Never change or remove an exported name or signature
  (that needs a `/v2` module). Keep new helpers unexported unless there's a
  real caller.
- **Go 1.18 minimum** (`go.mod`, and CI tests it). Don't use newer standard
  library features or builtins (e.g. `min`/`max`, `slices`, `maps`,
  `math/rand/v2`).
- **Sort order is the point of the library.** The base62 alphabet must stay in
  ASCII order (`0-9A-Za-z`), and timestamp fields must stay fixed-width, so
  tokens sort correctly as bytes.
- **Token layout is compatible across versions:** 6 digits of seconds, then 6
  of nanoseconds when the length allows (13+), then random. Shorter lengths
  (7–12) fall back to the 1.0.x layout. Don't change this without a version
  plan.
- **In-process ordering is guaranteed** for 13+ character tokens: shared
  state (`monotonic`, per length, behind one mutex) makes each token sort after
  the last, with same-tick tokens bumping a 2-digit head and redrawing the
  rest. Bulk batches share that state. Any change here must keep the lock and
  the strict-ordering tests passing under `-race`. 7–12 character tokens stay
  stateless and fully random.
- Randomness must come from `crypto/rand`, uniformly (reject bytes >= 248,
  never `% 62` alone).

## Tests

- `assertEqual(t, expected, got)`, in that order; it uses `t.Errorf`, so long
  loops stop early with `if t.Failed() { return }`.
- Keep tests strict: prefer exact checks (e.g. bracketing `time.Now()` before
  and after) over tolerances. Don't delete or loosen a test without asking the
  maintainer; `TestB62EncProof` is intentionally kept as a worked explanation.
- Base62 encoding is checked against `math/big` (with case swapped) and by
  `FuzzB62RoundTrip`.
- Anything involving clock precision or ordering must be verified on Linux.
- Time-dependent tests use `freezeClock(t, at)` (in `monotonic_test.go`),
  which swaps the unexported `clock` and resets the monotonic state; frozen
  times must be in the past so later real-clock tests aren't affected.
- Concurrency tests must call the generators with no outer lock, or they
  can't catch a missing lock. CI runs `go test -race`.

## Changes and releases

- Add user-facing changes to the top section of `CHANGELOG.md`.
- Follow `RELEASING.md` for versioning and releases. Never move or delete a
  pushed tag.
