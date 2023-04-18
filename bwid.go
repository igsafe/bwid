package bwid

import (
	"crypto/rand"
	"fmt"
	"time"
)

// match sort order of
// CHARACTER SET ascii COLLATE ascii_bin
const B62_DIGITS = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"

// number of base62 digits required to hold timestamp prefix
const TIMESTAMP_LEN = 6

func GenerateToken(length int) string {
	b := make([]byte, length)
	rand.Read(b)
	for i := 0; i < length; i++ {
		b[i] = B62_DIGITS[int(b[i])%62]
	}
	return string(b)
}

func GenerateTimestampedToken(length int) string {
	d := B62EncodeFixed(time.Now().Unix(), TIMESTAMP_LEN)
	tlen := length - TIMESTAMP_LEN
	if tlen < 1 {
		panic(fmt.Errorf("minimum timestamped token length is %d", TIMESTAMP_LEN+1))
	}
	return d + GenerateToken(tlen)
}

func GenerateObjectId() string {
	return GenerateTimestampedToken(24)
}

func GenerateBulkSeqTimestampedToken(count int64, length int) []string {
	d := B62EncodeFixed(time.Now().Unix(), TIMESTAMP_LEN)
	o := make([]string, count)
	ilen := B62Len(count)
	tlen := length - TIMESTAMP_LEN - ilen
	if tlen < 1 {
		panic(fmt.Errorf("minimum timestamped token length for %d count is %d", count, (TIMESTAMP_LEN + ilen + 1)))
	}
	for i := int64(0); i < count; i++ {
		o[i] = d + B62EncodeFixed(i, ilen) + GenerateToken(tlen)
	}
	return o
}

func GenerateBulkSeqObjectId(count int64) []string {
	return GenerateBulkSeqTimestampedToken(count, 24)
}
