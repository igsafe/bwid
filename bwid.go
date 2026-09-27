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
const TIMESTAMP_MICRO_LEN = 4

func GenerateToken(length int) string {
	b := make([]byte, length)
	rand.Read(b)
	for i := 0; i < length; i++ {
		b[i] = B62_DIGITS[int(b[i])%62]
	}
	return string(b)
}

func GenerateTimestampedToken(length int) string {
	now := time.Now()
	dsec := B62EncodeFixed(now.Unix(), TIMESTAMP_LEN)
	// new in 1.1.0—  encode the microseconds as next 4 digits for further
	// index write order accuracy.
	// the whole timestamp could be done in 9 digits if combined
	// but i don't want to change the appearance of these tokens
	// right now.
	dmic := B62EncodeFixed(now.UnixMicro()%1000000, TIMESTAMP_MICRO_LEN)
	tlen := length - TIMESTAMP_LEN - TIMESTAMP_MICRO_LEN
	if tlen < 1 {
		panic(fmt.Errorf("minimum timestamped token length is %d", TIMESTAMP_LEN+TIMESTAMP_MICRO_LEN+1))
	}
	return dsec + dmic + GenerateToken(tlen)
}

func GenerateObjectId() string {
	return GenerateTimestampedToken(24)
}

func GenerateBulkSeqTimestampedToken(count int64, length int) []string {
	now := time.Now()
	dsec := B62EncodeFixed(now.Unix(), TIMESTAMP_LEN)
	dmic := B62EncodeFixed(now.UnixMicro()%1000000, TIMESTAMP_MICRO_LEN)
	o := make([]string, count)
	ilen := B62Len(count)
	tlen := length - TIMESTAMP_LEN - TIMESTAMP_MICRO_LEN - ilen
	if tlen < 1 {
		panic(fmt.Errorf("minimum timestamped token length for %d count is %d", count, (TIMESTAMP_LEN + TIMESTAMP_MICRO_LEN + ilen + 1)))
	}
	for i := int64(0); i < count; i++ {
		o[i] = dsec + dmic + B62EncodeFixed(i, ilen) + GenerateToken(tlen)
	}
	return o
}

func GenerateBulkSeqObjectId(count int64) []string {
	return GenerateBulkSeqTimestampedToken(count, 24)
}
