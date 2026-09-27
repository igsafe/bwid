# TODO

Small ideas deferred from the 1.1.0 and 1.2.0 work. None are urgent.

- **Friendlier panic for a negative bulk count.**
  `GenerateBulkSeqTimestampedToken` and `GenerateBulkSeqObjectId` document
  that `count` must not be negative, but a negative value currently panics
  inside Go's `make` with a `makeslice` error. Check it up front and panic
  with a clear `bwid:` message instead, like the other input checks.

- **Shared test value table for base62.**
  `TestB62Encode`, `TestB62EncodeFixed`, `TestB62Decode` and
  `TestB62EncodeSpec` each repeat the same values. One table of
  `{n, "b62"}` pairs could drive all of them, plus `incrementB62`
  (`incrementB62(enc(n)) == enc(n+1)`). Keep `TestB62EncProof` as it is.

- **CI runs twice for pull requests from branches in this repo.**
  `.github/workflows/test.yml` triggers on every `push` and every
  `pull_request`, so a PR from a same-repo branch runs both. Changing
  `push:` to `push: { branches: [main] }` would run branches through their
  PR, and `main` after merging.

- **`B62Len(count)` uses one extra index digit for exact powers of 62.**
  Bulk batches size their index digits with `B62Len(count)`, but indexes run
  from 0 to `count-1`. So a batch of exactly 62 gets 2 digits where 1 would
  do (likewise 3844, and so on), costing one random character. Harmless;
  using `B62Len(count-1)` would fix it, but it changes the layout of those
  batch sizes, so it needs a version note.
