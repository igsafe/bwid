package bwid_test

import (
	"fmt"
	"sort"
	"time"

	"github.com/igsafe/bwid"
)

func ExampleGenerateObjectId() {
	id := bwid.GenerateObjectId()
	// e.g. "1xAv1A" + "10Hqhc" + "px67H8f38d7r": seconds, nanoseconds, random
	fmt.Println(len(id))
	// Output: 24
}

func ExampleGenerateBulkSeqObjectId() {
	ids := bwid.GenerateBulkSeqObjectId(1000)
	// one shared timestamp, then a sequence number, so the batch is in order
	fmt.Println(len(ids), sort.StringsAreSorted(ids))
	// Output: 1000 true
}

func ExampleGenerateToken() {
	// no timestamp; ~131 random bits, suitable for secrets
	secret := bwid.GenerateToken(22)
	fmt.Println(len(secret))
	// Output: 22
}

func ExampleB62Encode() {
	// digits are 0-9, then A-Z (10-35), then a-z (36-61), and each place
	// is worth 62 times the one to its right: "123" = 1*62*62 + 2*62 + 3
	fmt.Println(bwid.B62Encode(3971))
	fmt.Println(bwid.B62Encode(61))
	fmt.Println(bwid.B62Encode(62))
	// Output:
	// 123
	// z
	// 10
}

func ExampleB62EncodeFixed() {
	fmt.Println(bwid.B62EncodeFixed(62, 4))
	// too large for 2 places: keeps the lowest 2 digits of "123"
	fmt.Println(bwid.B62EncodeFixed(3971, 2))
	// Output:
	// 0010
	// 23
}

func ExampleB62Decode() {
	fmt.Println(bwid.B62Decode("123"))
	fmt.Println(bwid.B62Decode("0010"))
	// Output:
	// 3971
	// 62
}

func ExampleObjectIdTime() {
	// this ID was made on macOS, so its time stops at microseconds; IDs
	// made on Linux carry full nanoseconds
	t := bwid.ObjectIdTime("1xAv1A10Hqhcpx67H8f38d7r")
	fmt.Println(t.UTC().Format(time.RFC3339Nano))
	// Output: 2026-09-27T19:58:36.920387Z
}
