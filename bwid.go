package bwid

import (
	"crypto/rand"
	"fmt"
	"time"
)

// match sort order of
// CHARACTER SET ascii COLLATE ascii_bin
const B62_DIGITS = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"

// number of base62 digits required to hold timestamp prefixes
var TIMESTAMP_LEN uint64 = B62Len(uint64(time.Now().Unix()))
var TIMESTAMP_MICRO_LEN uint64 = B62Len(999999)

func GenerateToken(length uint) string {
	b := make([]byte, length)
	rand.Read(b)
	for i := uint(0); i < length; i++ {
		b[i] = B62_DIGITS[int(b[i])%62]
	}
	return string(b)
}

func GenerateTimestampedToken(length uint) string {
	now := time.Now()
	dsec := B62EncodeFixed(uint64(now.Unix()), TIMESTAMP_LEN)
	// new in 1.1.0—  encode the microseconds as next 4 digits for further
	// index write order accuracy.
	// the whole timestamp could be done in 9 digits if combined
	// but i don't want to change the appearance of these tokens
	// right now.
	dmic := B62EncodeFixed(uint64(now.UnixMicro()%1000000), TIMESTAMP_MICRO_LEN)
	tlen := length - uint(TIMESTAMP_LEN) - uint(TIMESTAMP_MICRO_LEN)
	if tlen < 1 {
		panic(fmt.Errorf("minimum timestamped token length is %d", TIMESTAMP_LEN+1))
	}
	return dsec + dmic + GenerateToken(tlen)
}

func GenerateObjectId() string {
	return GenerateTimestampedToken(24)
}

func GenerateBulkSeqTimestampedToken(count uint, length uint) []string {
	now := time.Now()
	dsec := B62EncodeFixed(uint64(now.Unix()), TIMESTAMP_LEN)
	dmic := B62EncodeFixed(uint64(now.UnixMicro()%1000000), TIMESTAMP_MICRO_LEN)
	o := make([]string, count)
	ilen := B62Len(uint64(count))
	tlen := length - uint(TIMESTAMP_LEN) - uint(TIMESTAMP_MICRO_LEN) - uint(ilen)
	if tlen < 1 {
		panic(fmt.Errorf("minimum timestamped token length for %d count is %d", count, (TIMESTAMP_LEN + ilen + 1)))
	}
	for i := uint64(0); i < uint64(count); i++ {
		o[i] = dsec + dmic + B62EncodeFixed(i, ilen) + GenerateToken(tlen)
	}
	return o
}

func GenerateBulkSeqObjectId(count uint) []string {
	return GenerateBulkSeqTimestampedToken(count, 24)
}
