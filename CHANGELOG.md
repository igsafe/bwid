# Changelog

## 1.1.0 (unreleased)

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
- Base62 helpers: `B62Encode`, `B62EncodeSpec`, `IncrementB62`,
  `DecDigitToB62`, `B62DigitToDec`, `ZeroDigit`.
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
