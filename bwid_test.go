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

func assertTimestampedTokenIsNow(t *testing.T, token string) {
	t.Helper()
	now := time.Now().Format("2006-01-02T15:04")
	dsec := B62Decode(token[:TIMESTAMP_LEN])
	tokenTime := time.Unix(dsec, 0)
	assertEqual(t, now, tokenTime.Format("2006-01-02T15:04"))
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
		token := GenerateTimestampedToken(length)
		assertEqual(t, length, len(token))
		assertTimestampedTokenIsNow(t, token)
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
	tokens := GenerateBulkSeqTimestampedToken(100, TIMESTAMP_LEN+2+1)
	for i, token := range tokens {
		assertEqual(t, TIMESTAMP_LEN+2+1, len(token))
		assertTimestampedTokenIsNow(t, token)
		assertEqual(t, int64(i), B62Decode(token[TIMESTAMP_LEN:TIMESTAMP_LEN+2]))
	}
	assertPanics(t, func() { GenerateBulkSeqTimestampedToken(100, TIMESTAMP_LEN+2) })
}

func TestGenerateTimestampedToken(t *testing.T) {
	var prevToken string
	for i := 0; i < 100000; i++ {
		token := GenerateTimestampedToken(24)
		assertTimestampedTokenIsNow(t, token)
		if prevToken != "" {
			assertTimestampedTokenOrder(t, token, prevToken)
		}
		if t.Failed() {
			return
		}
		prevToken = token
		// because the string is random after the timestamp,
		// make sure the timestamp increments at least one microsecond
		// so order check does not fail on the randomness alone
		// (microseconds, not nanoseconds, because that's the macOS
		// wall clock resolution)
		time.Sleep(time.Microsecond)
	}
}

func TestGenerateBulkSeqTimestampedToken(t *testing.T) {
	var count int64 = 100000
	var tokenLen = 40
	tokens := GenerateBulkSeqTimestampedToken(count, tokenLen)
	log.Printf("GenerateBulkSeqTimestampedToken(n, %d)[0] %s", tokenLen, tokens[0])
	assertEqual(t, count, int64(len(tokens)))
	var prevToken string
	orderDigits := B62Len(count)
	for i, token := range tokens {
		assertEqual(t, tokenLen, len(token))
		assertTimestampedTokenIsNow(t, token)
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
