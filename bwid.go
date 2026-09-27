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
	"sync"
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
	o := make([]byte, 0, length)
	// ~3% of bytes are rejected, so a little extra usually fills it in one read
	buf := make([]byte, length+length/16+4)
	for len(o) < length {
		// never fails on Go 1.24+, but can on older versions if the OS
		// random source is unavailable; never return non-random tokens
		if _, err := rand.Read(buf); err != nil {
			panic(fmt.Errorf("bwid: crypto/rand failed: %w", err))
		}
		o = appendB62Uniform(o, buf, length)
	}
	return string(o)
}

// appendB62Uniform appends base62 characters from random bytes in src to dst,
// up to max total. Bytes 248-255 are skipped: 248 is 4*62, so every character
// comes from exactly 4 byte values and all 62 are equally likely. (Using
// every byte with %62 would make 0-7 slightly more likely than the rest.)
func appendB62Uniform(dst, src []byte, max int) []byte {
	for _, b := range src {
		if len(dst) == max {
			break
		}
		if b < 248 {
			dst = append(dst, B62_DIGITS[b%62])
		}
	}
	return dst
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

// nextTick returns the seconds+nanoseconds prefix one nanosecond after p.
func nextTick(p string) string {
	sec := B62Decode(p[:TIMESTAMP_LEN])
	nsec := B62Decode(p[TIMESTAMP_LEN:]) + 1
	if nsec == int64(time.Second) {
		sec, nsec = sec+1, 0
	}
	return B62EncodeFixed(sec, TIMESTAMP_LEN) + B62EncodeFixed(nsec, TIMESTAMP_NANO_LEN)
}

// clock is time.Now, replaceable in tests
var clock = time.Now

// monotonicHeadLen is how many leading random digits are incremented for a
// same-tick token; see monotonic.
const monotonicHeadLen = 2

// monotonic holds the last full-layout timestamped token issued for each
// length, so the next one can be guaranteed to sort after it.
//
// What it's for: without it, two tokens made in the same clock tick come out
// in random order, and a clock step backward (e.g. an NTP correction) can make
// a new token sort before an old one. With it, every token from this process
// sorts after the previous one, so sorting by ID gives creation order and
// "everything newer than X" queries can't miss a token from this process.
//
// How: if the clock hasn't moved past the last token's timestamp, reuse that
// timestamp, add 1 to the first monotonicHeadLen random digits, and redraw
// the rest. The bumped head alone makes the token sort after the last one,
// so the tail stays unpredictable (10 fresh digits, ~59 bits, for object
// IDs). Each tick allows ~1,900 bumps on average before moving on to the
// next nanosecond. Only same-tick tokens are affected; the first token in
// each tick is fully random. Same-tick tokens are rare on Linux, where a tick
// is tens of nanoseconds, but common on macOS's microsecond clock.
//
// Limits: the guarantee is per process and is lost on restart. Across
// processes or hosts, tokens are only ordered as well as their clocks agree.
// After a clock step backward, tokens keep the last timestamp until real time
// catches up, so the time inside them is briefly stale.
var monotonic = struct {
	sync.Mutex
	last map[int]string
}{last: map[int]string{}}

// GenerateTimestampedToken returns a token of length characters that sorts
// by creation time: TIMESTAMP_LEN digits of Unix seconds, then
// TIMESTAMP_NANO_LEN digits of nanoseconds, then random characters.
//
// Within one process, each token of a given length sorts after the previous
// one, even when made in the same clock tick or after the clock steps
// backward. Across processes or hosts, tokens are ordered only as well as
// their clocks agree.
//
// Lengths shorter than TIMESTAMP_LEN+TIMESTAMP_NANO_LEN+1 (13) omit the
// nanoseconds and the ordering guarantee, and are fully random after the
// seconds, as in 1.0.x. It panics if length is less than TIMESTAMP_LEN+1 (7).
func GenerateTimestampedToken(length int) string {
	if length < TIMESTAMP_LEN+1 {
		panic(fmt.Errorf("minimum timestamped token length is %d", TIMESTAMP_LEN+1))
	}
	if length <= TIMESTAMP_LEN+TIMESTAMP_NANO_LEN {
		// seconds-only layout: a tick would be a whole second, so keep
		// these fully random rather than sequential
		p := timestampPrefix(clock(), length)
		return p + GenerateToken(length-len(p))
	}
	monotonic.Lock()
	defer monotonic.Unlock()
	p := timestampPrefix(clock(), length)
	var token string
	last, ok := monotonic.last[length]
	if ok && p <= last[:len(p)] {
		// same tick as the last token, or the clock went backward:
		// keep the last timestamp, add 1 to the first monotonicHeadLen
		// random digits, and redraw the rest
		p = last[:len(p)]
		hlen := monotonicHeadLen
		if hlen > length-len(p) {
			hlen = length - len(p)
		}
		head := incrementB62(last[len(p) : len(p)+hlen])
		if len(head) > hlen {
			// head overflowed; move to the next tick with fresh randomness
			p = nextTick(p)
			token = p + GenerateToken(length-len(p))
		} else {
			token = p + head + GenerateToken(length-len(p)-hlen)
		}
	} else {
		token = p + GenerateToken(length-len(p))
	}
	monotonic.last[length] = token
	return token
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
// Batches share ordering with GenerateTimestampedToken of the same length:
// within one process, a batch sorts after every token issued before it, and
// tokens issued after it sort after the whole batch. If the clock hasn't
// moved past the last token issued, the batch uses the nanosecond after it.
//
// The nanoseconds, and the ordering guarantee, are omitted when length leaves
// no room for them after the index digits. It panics if length is less than
// TIMESTAMP_LEN+B62Len(count)+1. count must not be negative.
func GenerateBulkSeqTimestampedToken(count int64, length int) []string {
	ilen := B62Len(count)
	if length < TIMESTAMP_LEN+ilen+1 {
		panic(fmt.Errorf("minimum timestamped token length for %d count is %d", count, (TIMESTAMP_LEN + ilen + 1)))
	}
	room := length - ilen
	var p string
	if room <= TIMESTAMP_LEN+TIMESTAMP_NANO_LEN {
		// seconds-only layout: stateless, as in 1.0.x
		p = timestampPrefix(clock(), room)
	} else {
		monotonic.Lock()
		defer monotonic.Unlock()
		p = timestampPrefix(clock(), room)
		if last, ok := monotonic.last[length]; ok && p <= last[:len(p)] {
			// not past the last token: take the next nanosecond, so every
			// token in the batch sorts after it
			p = nextTick(last[:len(p)])
		}
	}
	tlen := length - len(p) - ilen
	o := make([]string, count)
	for i := int64(0); i < count; i++ {
		o[i] = p + B62EncodeFixed(i, ilen) + GenerateToken(tlen)
	}
	if count > 0 && len(p) > TIMESTAMP_LEN {
		monotonic.last[length] = o[count-1]
	}
	return o
}

// GenerateBulkSeqObjectId returns count 24-character object IDs that share
// one timestamp and sort in slice order. See GenerateBulkSeqTimestampedToken.
func GenerateBulkSeqObjectId(count int64) []string {
	return GenerateBulkSeqTimestampedToken(count, 24)
}
