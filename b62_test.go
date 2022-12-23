package bwid

import (
	"math/big"
	"strings"
	"testing"
	"time"
)

func TestB62Len1(t *testing.T) {
	l := B62Len(0)
	if l != 1 {
		t.Fatalf("Expected 1 got %d", l)
	}
}

func TestB62Len2(t *testing.T) {
	l := B62Len(62)
	if l != 2 {
		t.Fatalf("Expected 2 got %d", l)
	}
}

func TestB62LenTodayEpoch(t *testing.T) {
	l := B62Len(time.Now().Unix())
	if l != 6 {
		t.Fatalf("Expected 6 got %d (this will overflow far into the future!!!)", l)
	}
}

func TestB62Encode0(t *testing.T) {
	o := B62EncodeFixed(0, 1)
	if o != "0" {
		t.Fatalf(`Expected "0" got "%s"`, o)
	}
}

func TestB62Encode1(t *testing.T) {
	o := B62EncodeFixed(1, 1)
	if o != "1" {
		t.Fatalf(`Expected "1" got "%s"`, o)
	}
}

func TestB62Encode10(t *testing.T) {
	o := B62EncodeFixed(10, 1)
	if o != "A" {
		t.Fatalf(`Expected "A" got "%s"`, o)
	}
}

func TestB62Encode12345(t *testing.T) {
	o := B62EncodeFixed(12345, 6)
	if o != "0003D7" {
		t.Fatalf(`Expected "0003D7" got "%s"`, o)
	}
}

func TestB62EncodeLibEqualLazy(t *testing.T) {
	o := B62EncodeFixed(12345, 3)
	o2 := big.NewInt(12345).Text(62)
	if !strings.EqualFold(o, o2) {
		t.Fatalf(`Expected "%s" got "%s"`, o2, o)
	}
}

func TestB62EncodeOverflow(t *testing.T) {
	o := B62EncodeFixed(56800235585, 6)
	if o != "000001" {
		t.Fatalf(`Expected "000001" got "%s"`, o)
	}
}

func TestB62Decode0Padded(t *testing.T) {
	n := B62Decode("000000")
	if n != 0 {
		t.Fatalf(`Expected 0 got %d`, n)
	}
}

func TestB62Decode1Padded(t *testing.T) {
	n := B62Decode("000001")
	if n != 1 {
		t.Fatalf(`Expected 1 got %d`, n)
	}
}

func TestB62Decode62Padded(t *testing.T) {
	n := B62Decode("000010")
	if n != 62 {
		t.Fatalf(`Expected 62 got %d`, n)
	}
}

func TestB62Decode63Padded(t *testing.T) {
	n := B62Decode("000011")
	if n != 63 {
		t.Fatalf(`Expected 63 got %d`, n)
	}
}

func TestB62Decodezzzzzz(t *testing.T) {
	n := B62Decode("zzzzzz")
	if n != 56800235583 {
		t.Fatalf(`Expected 56800235583 got %d`, n)
	}
}

func TestB62EncodeDecode12345(t *testing.T) {
	o := B62EncodeFixed(12345, 6)
	n := B62Decode(o)
	if n != 12345 {
		t.Fatalf(`Expected 12345 got %d`, n)
	}
}
