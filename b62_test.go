package bwid

import (
	"testing"
)

func assertEqual[V comparable](t *testing.T, expected, got V) {
	t.Helper()
	if expected != got {
		t.Fatalf("expected '%v' got '%v'", expected, got)
	}
}

// the math:
// values 0-61 are one digit straight from the alphabet
// 0 = "0"
// (62-1) = 61 = "z"
// values 62-123 are two digits starting with "1"
// (62*1) = 62 = "10"
// 63 = "11"
// ((62*2)-1) = 123 = "1z"
// (62*2) = 124 = "20"
// moving toward three digits...
// ((62*62)-1) = 3843 = "zz"
// (62*62) = 3844 = "100"
// ...

func TestB62EncProof(t *testing.T) {
	// ex: how to encode "123"?
	// from the math above we know that the decimal value is
	// 3 + (2 * 62) + (1 * (62*62)) = 3971
	assertEqual(t, 3+(2*62)+(1*62*62), B62Decode("123"))
	assertEqual(t, 3971, B62Decode("123"))
	// so how to encode 3971?
	// find out highest place divisor
	// is 3971 > 62?  yes, so it has > 1 position
	// is 3971 > (62*62)?  yes, so it has > 2 positions
	// is 3971 > (62*62*62)?  NO, so it has 3 positions
	assertEqual(t, 3, B62Len(3971))
	// for 3rd position (leftmost)
	// thanks to int division in go
	// which rounds quotients down to nearest integer
	// simply divide 3971 by divisor of position 3 (62*62)
	// 3971/(62*62) = 1 (rounded down by int division in go)
	assertEqual(t, 1, 3971/(62*62))
	assertEqual(t, '1', DecDigitToB62(1))
	// for 2nd position
	// subtract decimal value of position 3 (1) times divisor (62*62)
	// 3971 - (1*62*62) = 127
	// then divide that remainder by divisor or position 2 (62)
	// 127/(62) = 2 (rounded down by int division in go)
	assertEqual(t, 2, (3971-(1*62*62))/(62))
	assertEqual(t, '2', DecDigitToB62(2))
	// for 1st position
	// subtract decimal value of position 2 (2) times divisor (62)
	// 127 - (2*62) = 3
	assertEqual(t, 3, 127-(2*62))
	assertEqual(t, '3', DecDigitToB62(3))
}

func TestDecDigitToB62(t *testing.T) {
	assertEqual(t, '0', DecDigitToB62(0))
	assertEqual(t, '1', DecDigitToB62(1))
	assertEqual(t, 'A', DecDigitToB62(10))
	assertEqual(t, 'Z', DecDigitToB62(35))
	assertEqual(t, 'a', DecDigitToB62(36))
	assertEqual(t, 'z', DecDigitToB62(61))
}

func TestB62DigitToDec(t *testing.T) {
	assertEqual(t, 0, B62DigitToDec('0'))
	assertEqual(t, 1, B62DigitToDec('1'))
	assertEqual(t, 10, B62DigitToDec('A'))
	assertEqual(t, 35, B62DigitToDec('Z'))
	assertEqual(t, 36, B62DigitToDec('a'))
	assertEqual(t, 61, B62DigitToDec('z'))
}

func TestIncrementB62(t *testing.T) {
	assertEqual(t, "1", IncrementB62("0"))
	assertEqual(t, "2", IncrementB62("1"))
	assertEqual(t, "10", IncrementB62("z"))
	assertEqual(t, "11", IncrementB62("10"))
	assertEqual(t, "B0", IncrementB62("Az"))
	assertEqual(t, "a0", IncrementB62("Zz"))
	assertEqual(t, "x9", IncrementB62("x8"))
	assertEqual(t, "100", IncrementB62("zz"))
	assertEqual(t, "1000", IncrementB62("zzz"))
}

func TestB62EncodeSpec(t *testing.T) {
	_, t0 := B62EncodeSpec(0) // "0"
	assertEqual(t, 1, t0)
	_, t1 := B62EncodeSpec(1) // "1"
	assertEqual(t, 1, t1)
	_, t61 := B62EncodeSpec(61) // "z"
	assertEqual(t, 1, t61)
	_, t62 := B62EncodeSpec(62) // "10"
	assertEqual(t, 2, t62)
	_, t63 := B62EncodeSpec(63) // "11"
	assertEqual(t, 2, t63)
	_, t124 := B62EncodeSpec(124) // "20"
	assertEqual(t, 2, t124)
	_, t3843 := B62EncodeSpec(3843) // "zz"
	assertEqual(t, 2, t3843)
	_, t3844 := B62EncodeSpec(3844) // "100"
	assertEqual(t, 3, t3844)
	_, t238327 := B62EncodeSpec(238327) // "zzz"
	assertEqual(t, 3, t238327)
	_, t238328 := B62EncodeSpec(238328) // "1000"
	assertEqual(t, 4, t238328)
}

func TestB62Encode(t *testing.T) {
	assertEqual(t, "0", B62Encode(0))
	assertEqual(t, "1", B62Encode(1))
	assertEqual(t, "z", B62Encode(61))
	assertEqual(t, "10", B62Encode(62))
	assertEqual(t, "11", B62Encode(63))
	assertEqual(t, "20", B62Encode(124))
	assertEqual(t, "zz", B62Encode(3843))
	assertEqual(t, "100", B62Encode(3844))
	assertEqual(t, "zzz", B62Encode(238327))
	assertEqual(t, "1000", B62Encode(238328))
}

func TestB62EncodeFixed(t *testing.T) {
	assertEqual(t, "0", B62EncodeFixed(0, 1))
	assertEqual(t, "00", B62EncodeFixed(0, 2))
	assertEqual(t, "0000000000", B62EncodeFixed(0, 10))
	assertEqual(t, "0000000001", B62EncodeFixed(1, 10))
	assertEqual(t, "000000000z", B62EncodeFixed(61, 10))
	assertEqual(t, "0000000010", B62EncodeFixed(62, 10))
	assertEqual(t, "0000000011", B62EncodeFixed(63, 10))
	assertEqual(t, "0000000020", B62EncodeFixed(124, 10))
	assertEqual(t, "00000000zz", B62EncodeFixed(3843, 10))
	assertEqual(t, "0000000100", B62EncodeFixed(3844, 10))
	assertEqual(t, "0000000zzz", B62EncodeFixed(238327, 10))
	assertEqual(t, "0000001000", B62EncodeFixed(238328, 10))
}

func TestB62Decode(t *testing.T) {
	assertEqual(t, 0, B62Decode("0"))
	assertEqual(t, 0, B62Decode("0000000000"))
	assertEqual(t, 1, B62Decode("01"))
	assertEqual(t, 61, B62Decode("z"))
	assertEqual(t, 62, B62Decode("010"))
	assertEqual(t, 63, B62Decode("11"))
	assertEqual(t, 124, B62Decode("0020"))
	assertEqual(t, 3843, B62Decode("0zz"))
	assertEqual(t, 3844, B62Decode("100"))
	assertEqual(t, 238327, B62Decode("zzz"))
	assertEqual(t, 238328, B62Decode("01000"))
}
