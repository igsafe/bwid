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

func extractBorder(token string, digits uint64) string {
	return token[TIMESTAMP_LEN+TIMESTAMP_MICRO_LEN : TIMESTAMP_LEN+TIMESTAMP_MICRO_LEN+digits]
}

func assertTimestampedTokenIsNow(t *testing.T, token string) {
	t.Helper()
	now := time.Now().Format("2006-01-02T15:04")
	dsec := B62Decode(token[:TIMESTAMP_LEN])
	tokenTime := time.Unix(int64(dsec), 0)
	assertEqual(t, tokenTime.Format("2006-01-02T15:04"), now)
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
	d1 := B62Len(uint64(time.Now().Unix()))
	d2 := B62Len(uint64(time.Date(3000, 1, 1, 0, 0, 0, 0, time.UTC).Unix()))
	assertEqual(t, d1, d2)
	// 2287 AD no bueno!
	d3 := B62Len(uint64(time.Date(4000, 1, 1, 0, 0, 0, 0, time.UTC).Unix()))
	assertEqual(t, d2+1, d3)
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
		prevToken = token
		// because the string is random after the timestamp,
		// make sure the timestamp increments at least one microsecond
		// so order check does not fail on the randomness alone
		time.Sleep(time.Microsecond)
	}
}

func TestGenerateBulkSeqTimestampedToken(t *testing.T) {
	var count uint = 100000
	var tokenLen uint = 40
	tokens := GenerateBulkSeqTimestampedToken(count, tokenLen)
	log.Printf("GenerateBulkSeqTimestampedToken(n, %d)[0] %s", tokenLen, tokens[0])
	assertEqual(t, uint(len(tokens)), count)
	var prevToken string
	orderDigits := B62Len(uint64(count))
	for i, token := range tokens {
		assertEqual(t, uint(len(token)), tokenLen)
		assertTimestampedTokenIsNow(t, token)
		if prevToken != "" {
			assertTimestampedTokenOrder(t, token, prevToken)
		}
		tokenI := B62Decode(extractBorder(token, orderDigits))
		assertEqual(t, uint64(i), tokenI)
		prevToken = token
	}
}
