package bwid

import (
	"log"
	"testing"
	"time"
)

func extractBsec(token string) string {
	return token[:TIMESTAMP_LEN]
}

func extractBnano(token string) string {
	return token[TIMESTAMP_LEN : TIMESTAMP_LEN+TIMESTAMP_NANO_LEN]
}

func extractBorder(token string, digits int) string {
	return token[TIMESTAMP_LEN+TIMESTAMP_NANO_LEN : TIMESTAMP_LEN+TIMESTAMP_NANO_LEN+digits]
}

// assertTimestampedTokenBetween checks the token's seconds fall within
// unix times read just before and just after generating it
func assertTimestampedTokenBetween(t *testing.T, token string, before, after int64) {
	t.Helper()
	dsec := B62Decode(token[:TIMESTAMP_LEN])
	if dsec < before || dsec > after {
		t.Errorf("expected token seconds between %d and %d, got %d", before, after, dsec)
	}
}

func assertGte(t *testing.T, expectedGte, expectedLt string) {
	t.Helper()
	if expectedGte < expectedLt {
		t.Fatalf("expected '%v' >= '%v'", expectedGte, expectedLt)
	}
}

func assertTimestampedTokenOrder(t *testing.T, nextToken string, prevToken string) {
	t.Helper()
	assertGte(t, nextToken, prevToken)
	if extractBsec(nextToken) == extractBsec(prevToken) {
		// ensure nanoseconds are sequential
		assertGte(t, extractBnano(nextToken), extractBnano(prevToken))
	}
}

func TestUnixTimestampLen(t *testing.T) {
	// we won't need more digits for a while.  3000 AD, still good!
	assertEqual(t, TIMESTAMP_LEN, B62Len(time.Now().Unix()))
	assertEqual(t, TIMESTAMP_LEN, B62Len(time.Date(3000, 1, 1, 0, 0, 0, 0, time.UTC).Unix()))
	// 3769 AD no bueno!
	assertEqual(t, TIMESTAMP_LEN, B62Len(time.Date(3769, 1, 1, 0, 0, 0, 0, time.UTC).Unix()))
	assertEqual(t, TIMESTAMP_LEN+1, B62Len(time.Date(3770, 1, 1, 0, 0, 0, 0, time.UTC).Unix()))
	// nanoseconds always fit
	assertEqual(t, TIMESTAMP_NANO_LEN, B62Len(999999999))
}

func TestGenerateTokenLen(t *testing.T) {
	token := GenerateToken(24)
	log.Printf("GenerateToken(24) %s", token)
	assertEqual(t, 24, len(token))
}

func TestGenerateTokenLengths(t *testing.T) {
	for _, length := range []int{0, 1, 22, 1000} {
		assertB62Token(t, GenerateToken(length), length)
	}
}

func TestAppendB62Uniform(t *testing.T) {
	// feed every possible byte value once
	src := make([]byte, 256)
	for i := range src {
		src[i] = byte(i)
	}
	out := appendB62Uniform(nil, src, 1000)
	// 248-255 are rejected
	assertEqual(t, 248, len(out))
	// every character comes from exactly 4 byte values
	counts := map[byte]int{}
	for _, c := range out {
		counts[c]++
	}
	assertEqual(t, 62, len(counts))
	for i := 0; i < len(B62_DIGITS); i++ {
		assertEqual(t, 4, counts[B62_DIGITS[i]])
	}
	// stops at max, and appends to what's already there
	assertEqual(t, "ab0123", string(appendB62Uniform([]byte("ab"), src, 6)))
	// all-rejected input adds nothing
	assertEqual(t, "", string(appendB62Uniform(nil, []byte{248, 255}, 10)))
}

func TestGenerateTimestampedTokenLen(t *testing.T) {
	token := GenerateTimestampedToken(24)
	log.Printf("GenerateTimestampedToken(24) %s", token)
	assertEqual(t, 24, len(token))
}

func TestGenerateTimestampedTokenNanoRange(t *testing.T) {
	for i := 0; i < 1000; i++ {
		n := B62Decode(extractBnano(GenerateTimestampedToken(24)))
		if n < 0 || n > 999999999 {
			t.Fatalf("nanoseconds out of range: %d", n)
		}
	}
}

func assertPanics(t *testing.T, f func()) {
	t.Helper()
	defer func() {
		if recover() == nil {
			t.Errorf("expected panic")
		}
	}()
	f()
}

func TestGenerateTimestampedTokenShortFallback(t *testing.T) {
	// shorter than seconds + nanoseconds + 1 random digit falls back
	// to seconds only, as in 1.0.x
	for length := TIMESTAMP_LEN + 1; length <= TIMESTAMP_LEN+TIMESTAMP_NANO_LEN+1; length++ {
		before := time.Now().Unix()
		token := GenerateTimestampedToken(length)
		after := time.Now().Unix()
		assertEqual(t, length, len(token))
		assertTimestampedTokenBetween(t, token, before, after)
	}
	assertPanics(t, func() { GenerateTimestampedToken(TIMESTAMP_LEN) })
}

func TestTimestampPrefix(t *testing.T) {
	// 1234567890 = "1LY7VK", 999999999 = "15ftgF"
	fixed := time.Unix(1234567890, 999999999)
	assertEqual(t, "1LY7VK", timestampPrefix(fixed, TIMESTAMP_LEN+1))
	assertEqual(t, "1LY7VK", timestampPrefix(fixed, TIMESTAMP_LEN+TIMESTAMP_NANO_LEN))
	assertEqual(t, "1LY7VK15ftgF", timestampPrefix(fixed, TIMESTAMP_LEN+TIMESTAMP_NANO_LEN+1))
	assertEqual(t, "1LY7VK15ftgF", timestampPrefix(fixed, 24))
	// zero nanoseconds are still padded to full width
	assertEqual(t, "1LY7VK000000", timestampPrefix(time.Unix(1234567890, 0), 24))
}

func TestGenerateBulkSeqTimestampedTokenShortFallback(t *testing.T) {
	// 100 needs 2 order digits
	before := time.Now().Unix()
	tokens := GenerateBulkSeqTimestampedToken(100, TIMESTAMP_LEN+2+1)
	after := time.Now().Unix()
	for i, token := range tokens {
		assertEqual(t, TIMESTAMP_LEN+2+1, len(token))
		assertTimestampedTokenBetween(t, token, before, after)
		assertEqual(t, int64(i), B62Decode(token[TIMESTAMP_LEN:TIMESTAMP_LEN+2]))
	}
	assertPanics(t, func() { GenerateBulkSeqTimestampedToken(100, TIMESTAMP_LEN+2) })
}

func TestGenerateTimestampedToken(t *testing.T) {
	var prevToken string
	for i := 0; i < 100000; i++ {
		before := time.Now().Unix()
		token := GenerateTimestampedToken(24)
		after := time.Now().Unix()
		assertTimestampedTokenBetween(t, token, before, after)
		if prevToken != "" {
			assertTimestampedTokenOrder(t, token, prevToken)
		}
		if t.Failed() {
			return
		}
		// no sleep needed: same-tick tokens are ordered by the monotonic
		// head, even on macOS's microsecond clock
		prevToken = token
	}
}

func TestGenerateBulkSeqTimestampedToken(t *testing.T) {
	var count int64 = 100000
	var tokenLen = 40
	before := time.Now().Unix()
	tokens := GenerateBulkSeqTimestampedToken(count, tokenLen)
	after := time.Now().Unix()
	log.Printf("GenerateBulkSeqTimestampedToken(n, %d)[0] %s", tokenLen, tokens[0])
	assertEqual(t, count, int64(len(tokens)))
	var prevToken string
	orderDigits := B62Len(count)
	for i, token := range tokens {
		assertEqual(t, tokenLen, len(token))
		assertTimestampedTokenBetween(t, token, before, after)
		if prevToken != "" {
			assertTimestampedTokenOrder(t, token, prevToken)
		}
		tokenI := B62Decode(extractBorder(token, orderDigits))
		assertEqual(t, int64(i), tokenI)
		if t.Failed() {
			return
		}
		prevToken = token
	}
}
