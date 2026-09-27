# Changelog

## 1.2.0 (unreleased)

Backward compatible: no signatures change, and the token layout is the same
as 1.1.0.

Developed by Claude Opus 5.5 (Anthropic) in Claude Code, including the
monotonic ordering design (with changes from human operator), its tests and
sabotage checks, and the modulo-bias fix.

### Added
- Guaranteed ordering within a process: every `GenerateObjectId` and
  `GenerateTimestampedToken` (13+ characters) sorts after the previous one of
  the same length, including across goroutines and after the clock steps
  backward. Same-tick IDs bump their first 2 random characters and redraw the
  rest, so they stay unpredictable (~59 bits).
- `ObjectIdTime(id)`: returns the time stored in an ID, with nanoseconds when
  present. Panics on invalid input.

### Changed
- Bulk batches share ordering with single IDs of the same length: a batch
  sorts after everything issued before it, and later IDs sort after the whole
  batch. A batch made in the same clock tick as the last ID takes the next
  nanosecond.
- Tokens of 7–12 characters are unchanged: fully random after the seconds, as
  in 1.0.x, with no ordering guarantee.
- CI runs the tests with the race detector.

### Fixed
- `GenerateToken` (and the random part of every token) is now uniform. It
  used `byte % 62`, which made `0`–`7` 25% more likely than other characters;
  it now rejects bytes 248–255.

## 1.1.0 (2026-09-27)

Backward compatible: every 1.0.x function keeps its signature and accepts the
same lengths.

Reviewed and developed with help from Claude Opus 5.5 (Anthropic) in Claude
Code, including the compatibility review, the base62 overflow fix, the
short-length fallback, and validation against `math/big` and fuzzing.

### Added
- Timestamped tokens now include a 6-digit nanosecond field after the seconds,
  for finer write ordering. Precision follows the platform clock: nanoseconds
  on Linux, microseconds on macOS, typically 100ns on Windows.
- `TIMESTAMP_NANO_LEN` constant.
- `B62Encode(n)`: variable-length base62 encoding (no padding).
- Package documentation, and a README covering the token layout, entropy, and
  a comparison with UUIDv7.

### Changed
- `GenerateObjectId()` layout is now 6 seconds + 6 nanoseconds + 12 random
  characters (was 6 seconds + 18 random), so random bits drop from ~107 to
  ~71. See the README for what that means for collisions and guessability.
- `GenerateTimestampedToken` and `GenerateBulkSeqTimestampedToken` lengths too
  short for the nanosecond field fall back to the 1.0.x layout (seconds +
  random), so existing lengths keep working.
- Base62 encoding rewritten, with tests checked against `math/big` and a fuzz
  test.

### Fixed
- `GenerateToken` (and everything built on it) now panics if `crypto/rand`
  fails, instead of silently returning non-random tokens. This could only
  happen on Go versions before 1.24.

## 1.0.1 (2023-04-18)

### Added
- `GenerateTimestampedToken(length)` and
  `GenerateBulkSeqTimestampedToken(count, length)` for timestamped tokens of
  any length. `GenerateObjectId` and `GenerateBulkSeqObjectId` now call them
  with 24.
- `TIMESTAMP_LEN` constant.
- Panics with a clear message when the requested length is too short.
- Tests.

## 1.0.0 (2022-12-23)

Initial release.
- `GenerateToken(length)`: random base62 token.
- `GenerateObjectId()`: 24-character ID of 6 base62 digits of Unix seconds
  followed by 18 random characters.
- `GenerateBulkSeqObjectId(count)`: a batch of IDs sharing one timestamp, with
  a sequence number so the batch sorts in order.
- Base62 helpers `B62Len`, `B62EncodeFixed`, `B62Decode`, using an ASCII-ordered
  alphabet so fixed-width IDs sort correctly byte by byte.
