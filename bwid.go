package bwid

import (
	"crypto/rand"
	"time"
)

// match sort order of
// CHARACTER SET ascii COLLATE ascii_bin
const B62_DIGITS = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"

func GenerateToken(length int) string {
	b := make([]byte, length)
	rand.Read(b)
	for i := 0; i < length; i++ {
		b[i] = B62_DIGITS[int(b[i])%62]
	}
	return string(b)
}

func GenerateObjectId() string {
	d := B62EncodeFixed(time.Now().Unix(), 6)
	return d + GenerateToken(18)
}

func GenerateBulkSeqObjectId(count int64) []string {
	d := B62EncodeFixed(time.Now().Unix(), 6)
	o := make([]string, count)
	ilen := B62Len(count)
	tlen := 18 - ilen
	for i := int64(0); i < count; i++ {
		o[i] = d + B62EncodeFixed(i, ilen) + GenerateToken(tlen)
	}
	return o
}
