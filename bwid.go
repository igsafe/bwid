// Package bwid generates random and time-sortable tokens.
//
// Numbers are encoded in base62 using ASCII order—
//
//	0-9   = 0-9
//	10-35 = A-Z
//	36-61 = a-z
//
// This differs from math/big's Text(62), which puts lowercase letters before
// uppercase. Using ASCII order means fixed-width encodings sort the same as
// their numeric values under byte-wise comparison, e.g. Go string comparison
// or MySQL's ascii_bin collation.
package bwid

import (
	"crypto/rand"
	"fmt"
	"time"
)

// B62_DIGITS is the base62 alphabet, in ASCII order; see the package doc.
const B62_DIGITS = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"

// TIMESTAMP_LEN is the number of base62 digits of Unix seconds at the start
// of every timestamped token (6 digits holds seconds until the year 3769).
const TIMESTAMP_LEN = 6

// TIMESTAMP_NANO_LEN is the number of base62 digits of nanoseconds that
// follow the seconds, when the token is long enough (6 digits holds
// 0-999999999).
//
// new in 1.1.0— resolution depends on the platform's wall clock—
//   - Linux: true nanoseconds
//   - macOS: microseconds (the last 3 decimal digits are always 0)
//   - Windows: typically 100ns steps, coarser on older systems
//
// the layout is the same on every platform, so tokens from different
// hosts still sort together; only the precision of the ordering differs.
const TIMESTAMP_NANO_LEN = 6

// GenerateToken returns length random base62 characters from crypto/rand,
// with no timestamp. Use it for secrets such as API keys or unlisted links;
// 22 characters gives about 131 bits of randomness. It panics if crypto/rand
// fails.
func GenerateToken(length int) string {
	b := make([]byte, length)
	// never fails on Go 1.24+, but can on older versions if the OS
	// random source is unavailable; never return non-random tokens
	if _, err := rand.Read(b); err != nil {
		panic(fmt.Errorf("bwid: crypto/rand failed: %w", err))
	}
	for i := 0; i < length; i++ {
		b[i] = B62_DIGITS[int(b[i])%62]
	}
	return string(b)
}

// timestampPrefix encodes now as seconds, followed by nanoseconds when
// room leaves space for at least one more digit.  shorter tokens fall
// back to seconds only, as in 1.0.x, so existing lengths keep working.
// a given room always produces the same layout.
func timestampPrefix(now time.Time, room int) string {
	p := B62EncodeFixed(now.Unix(), TIMESTAMP_LEN)
	if room > TIMESTAMP_LEN+TIMESTAMP_NANO_LEN {
		p += B62EncodeFixed(int64(now.Nanosecond()), TIMESTAMP_NANO_LEN)
	}
	return p
}

// GenerateTimestampedToken returns a token of length characters that sorts
// by creation time: TIMESTAMP_LEN digits of Unix seconds, then
// TIMESTAMP_NANO_LEN digits of nanoseconds, then random characters.
//
// Lengths shorter than TIMESTAMP_LEN+TIMESTAMP_NANO_LEN+1 (13) omit the
// nanoseconds, as in 1.0.x. It panics if length is less than
// TIMESTAMP_LEN+1 (7).
func GenerateTimestampedToken(length int) string {
	if length < TIMESTAMP_LEN+1 {
		panic(fmt.Errorf("minimum timestamped token length is %d", TIMESTAMP_LEN+1))
	}
	p := timestampPrefix(time.Now(), length)
	return p + GenerateToken(length-len(p))
}

// GenerateObjectId returns a 24-character timestamped token: 6 characters of
// seconds, 6 of nanoseconds, and 12 random (about 71 bits). See
// GenerateTimestampedToken.
func GenerateObjectId() string {
	return GenerateTimestampedToken(24)
}

// GenerateBulkSeqTimestampedToken returns count tokens of length characters
// that share one timestamp and sort in slice order. Each token is the
// timestamp (as in GenerateTimestampedToken), then its index in
// B62Len(count) base62 digits, then random characters.
//
// The nanoseconds are omitted when length leaves no room for them after the
// index digits. It panics if length is less than
// TIMESTAMP_LEN+B62Len(count)+1. count must not be negative.
func GenerateBulkSeqTimestampedToken(count int64, length int) []string {
	ilen := B62Len(count)
	if length < TIMESTAMP_LEN+ilen+1 {
		panic(fmt.Errorf("minimum timestamped token length for %d count is %d", count, (TIMESTAMP_LEN + ilen + 1)))
	}
	p := timestampPrefix(time.Now(), length-ilen)
	tlen := length - len(p) - ilen
	o := make([]string, count)
	for i := int64(0); i < count; i++ {
		o[i] = p + B62EncodeFixed(i, ilen) + GenerateToken(tlen)
	}
	return o
}

// GenerateBulkSeqObjectId returns count 24-character object IDs that share
// one timestamp and sort in slice order. See GenerateBulkSeqTimestampedToken.
func GenerateBulkSeqObjectId(count int64) []string {
	return GenerateBulkSeqTimestampedToken(count, 24)
}
