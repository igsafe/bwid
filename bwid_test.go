package bwid

import (
	"log"
	"testing"
	"time"
)

func extractBsec(token string) string {
	return token[:TIMESTAMP_LEN]
}

func extractBmic(token string) string {
	return token[TIMESTAMP_LEN : TIMESTAMP_LEN+TIMESTAMP_MICRO_LEN]
}

func extractBorder(token string, digits int) string {
	return token[TIMESTAMP_LEN+TIMESTAMP_MICRO_LEN : TIMESTAMP_LEN+TIMESTAMP_MICRO_LEN+digits]
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
		// ensure microseconds are sequential
		assertGte(t, extractBmic(nextToken), extractBmic(prevToken))
	}
}

func TestUnixTimestampLen(t *testing.T) {
	// we won't need more digits for a while.  3000 AD, still good!
	assertEqual(t, TIMESTAMP_LEN, B62Len(time.Now().Unix()))
	assertEqual(t, TIMESTAMP_LEN, B62Len(time.Date(3000, 1, 1, 0, 0, 0, 0, time.UTC).Unix()))
	// 3769 AD no bueno!
	assertEqual(t, TIMESTAMP_LEN, B62Len(time.Date(3769, 1, 1, 0, 0, 0, 0, time.UTC).Unix()))
	assertEqual(t, TIMESTAMP_LEN+1, B62Len(time.Date(3770, 1, 1, 0, 0, 0, 0, time.UTC).Unix()))
	// microseconds always fit
	assertEqual(t, TIMESTAMP_MICRO_LEN, B62Len(999999))
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
