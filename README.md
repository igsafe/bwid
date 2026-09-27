# bwid
Go functions for generating tokens and object IDs

See [CHANGELOG.md](CHANGELOG.md) for release history.

## Why bwid?

Random IDs such as UUIDv4 are unique, but each new one lands at a random spot
in a database's B-tree index. During bulk writes this means page splits,
scattered writes, and poor cache use. bwid IDs start with a timestamp, so new
IDs sort after existing ones and inserts stay at the end of the index, while
the random part keeps them unique across hosts.

`GenerateBulkSeqObjectId(count)` goes further for batches: every ID in the
batch shares one timestamp and carries a sequence number, so the whole batch
is in order.

UUIDv7 now solves the same problem as a standard; see
[bwid or UUIDv7?](#bwid-or-uuidv7) below.

## Timestamped token layout

`GenerateObjectId()` returns a 24-character base62 token (`0-9A-Za-z`, in
ASCII order) that sorts by creation time:

| Characters | Contents |
|---|---|
| 6 | Unix seconds |
| 6 | Nanoseconds within the second |
| 12 | Random |

`GenerateTimestampedToken(length)` uses the same layout for any length of 13
or more. Lengths 7–12 fall back to the 1.0.x layout (seconds + random).

Nanosecond precision depends on the platform's wall clock: true nanoseconds on
Linux, microseconds on macOS, and typically 100ns steps on Windows. The layout
is the same everywhere, so tokens from different hosts still sort together.

## Entropy compared with UUIDs

| | Length | Random bits | Time-ordered |
|---|---|---|---|
| UUIDv4 | 36 | 122 | no |
| UUIDv7 | 36 | 74 | milliseconds |
| `GenerateObjectId()` 1.0.x | 24 | ~107 | seconds |
| `GenerateObjectId()` 1.1.x | 24 | ~71 | nanoseconds (Linux) |

**Collisions:** two object IDs can only collide if they share a timestamp, so
the random part only has to be unique within one nanosecond (one microsecond
on macOS). Generating a million IDs per second for 100 years gives at most
~5×10⁻⁷ expected collisions. In practice this could be argued as safer than
UUIDv4, where every ID competes with every other.

**Guessability:** the random part comes from Go's `crypto/rand`, and 71 bits
clears the 64-bit minimum for session secrets in both
[NIST SP 800-63B-4](https://pages.nist.gov/800-63-4/sp800-63b/session/)
(Session Bindings: "at least 64 bits" from an approved random bit generator) and the
[OWASP Session Management Cheat Sheet](https://cheatsheetseries.owasp.org/cheatsheets/Session_Management_Cheat_Sheet.html)
("at least 64 bits of entropy"). OWASP estimates 64 bits takes ~585 years to
guess at 10,000 guesses per second against 100,000 live sessions; 71 bits is
128× that. Those minimums assume short-lived session IDs, though, and are well
below the common 128-bit practice for long-lived secrets. Object IDs also
reveal their creation time. Don't rely on them as secrets (e.g. unlisted
links); use
`GenerateToken(22)` or longer (~131 bits, no timestamp) instead. If a table
needs more randomness in its IDs, use a longer `GenerateTimestampedToken`, e.g.
32 characters gives ~119 random bits.

## bwid or UUIDv7?

UUIDv7 ([RFC 9562](https://www.rfc-editor.org/rfc/rfc9562), 2024) puts a
millisecond timestamp at the front of a UUID for the same reason bwid puts one
at the front of its IDs: so inserts land at the end of the index. bwid was
written in 2022, while time-ordered UUIDs were still a draft.

**Choose UUIDv7** when you want a standard format with broad library and
database support, or the smallest index: it stores in 16 bytes, where a bwid
ID stored as a string takes 24.

**Choose bwid** when you want short, URL-safe string IDs (24 characters vs 36)
that sort correctly as plain ASCII, finer than millisecond ordering, or
sequenced bulk batches. The shorter length also carries into JSON, which has
no binary type, so UUIDs are always sent as 36-character strings. That's a
third smaller per ID uncompressed, though gzip mostly evens it out.

**Also choose bwid** if you're a 🤠 who just likes taking the scenic route.
Not every ID has to look like it came from a committee. 🦔  Have some fun.

Storing UUIDv7 in 16 bytes:

| Database | Column |
|---|---|
| PostgreSQL | `uuid` (version 18+ can generate v7 with `uuidv7()`) |
| MariaDB 10.7+ | `UUID` |
| MySQL 8.0+ | `BINARY(16)`, converting with `UUID_TO_BIN()` / `BIN_TO_UUID()` |

MySQL has no native UUID type. When storing v7 there, call `UUID_TO_BIN(id)`
without the second (swap) argument: swapping rearranges bytes to help v1
UUIDs sort, and would break v7's time ordering.

For bwid IDs, use `CHAR(24) CHARACTER SET ascii COLLATE ascii_bin` in MySQL
and MariaDB, or `text COLLATE "C"` in PostgreSQL, so they compare byte by byte.

### Performance in MySQL

- **`UUID_TO_BIN` itself is negligible.** Parsing 36 hex characters costs
  nanoseconds, and you can skip it by sending the 16 raw bytes from your app.
  Its real cost is convenience: every query needs `UUID_TO_BIN(?)` or
  `BIN_TO_UUID(id)`, and forgetting one silently matches nothing. bwid IDs are
  the same string in code, SQL, logs, and JSON.
- **Size is the real difference.** InnoDB copies the primary key into every
  secondary index entry, and foreign keys repeat it in other tables. So the
  extra 8 bytes (24 vs 16) are paid many times over. For example, an index on
  an `INT` column is roughly 30% larger with a bwid primary key. Larger indexes
  fit less of themselves in the buffer pool, which shows up under load on very
  large tables.
- **Insert locality is about the same.** Both put time first, so new rows land
  at the end of the index. UUIDv7 orders by millisecond and bwid by nanosecond
  (on Linux), and bwid's bulk generator orders a whole batch exactly.
- **Comparison speed is about the same.** `ascii_bin` compares byte by byte
  like `BINARY(16)` does; the extra length is minor next to page reads.

In short, UUIDv7 in `BINARY(16)` wins on very large tables with several
secondary indexes or many foreign keys. bwid wins on simplicity, readability,
and exact bulk ordering, which is what most apps need more.
