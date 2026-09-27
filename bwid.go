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

// base62 alphabet in ASCII order; see package doc
const B62_DIGITS = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"

// number of base62 digits required to hold timestamp prefixes
// (6 digits holds unix seconds until the year 3769)
const TIMESTAMP_LEN = 6

// number of base62 digits for the sub-second part of the timestamp
// (6 digits holds 0-999999999 nanoseconds)
//
// new in 1.1.0— resolution depends on the platform's wall clock—
//   - Linux: true nanoseconds
//   - macOS: microseconds (the last 3 decimal digits are always 0)
//   - Windows: typically 100ns steps, coarser on older systems
//
// the layout is the same on every platform, so tokens from different
// hosts still sort together; only the precision of the ordering differs.
const TIMESTAMP_NANO_LEN = 6

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

func GenerateTimestampedToken(length int) string {
	if length < TIMESTAMP_LEN+1 {
		panic(fmt.Errorf("minimum timestamped token length is %d", TIMESTAMP_LEN+1))
	}
	p := timestampPrefix(time.Now(), length)
	return p + GenerateToken(length-len(p))
}

func GenerateObjectId() string {
	return GenerateTimestampedToken(24)
}

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

func GenerateBulkSeqObjectId(count int64) []string {
	return GenerateBulkSeqTimestampedToken(count, 24)
}
