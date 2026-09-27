// base62 encoding helpers; see package doc in bwid.go for the alphabet

package bwid

const zeroDigit = byte(48)

func decDigitToB62(d int64) byte {
	// 0-9
	if d < 10 {
		return byte(d + 48)
	}
	// A-Z
	if d < 36 {
		return byte(d + 55)
	}
	// a-z
	return byte(d + 61)
}

func b62DigitToDec(b byte) int64 {
	// 0-9
	if b < 58 {
		return int64(b) - 48
	}
	// A-Z
	if b < 91 {
		return int64(b) - 55
	}
	// a-z
	return int64(b) - 61
}

// Increment a Base62 number by 1
func incrementB62(v string) string {
	incrementing := true
	// alloc extra byte in case new place is required
	v2 := make([]byte, len(v)+1)
	for i := len(v) - 1; i >= 0; i-- {
		i2 := i + 1
		if incrementing {
			d := b62DigitToDec(v[i])
			if d == 61 {
				// adding will overflow, next place
				v2[i2] = zeroDigit
			} else {
				// increment this place
				v2[i2] = decDigitToB62(d + 1)
				// next place remains the same
				incrementing = false
			}
		} else {
			v2[i2] = v[i]
		}
	}
	if incrementing {
		v2[0] = decDigitToB62(1)
		return string(v2)
	}
	return string(v2[1:])
}

// b62EncodeSpec returns the highest place divisor d and the
// total number of places pt for the base62 encoding of n.
func b62EncodeSpec(n int64) (d int64, pt int) {
	d = 1
	pt = 1
	// d <= n/62 avoids overflowing d*62
	for d <= n/62 {
		d *= 62
		pt++
	}
	return
}

// B62Len returns the number of base62 digits required to hold n.
// n must not be negative.
func B62Len(n int64) int {
	_, pt := b62EncodeSpec(n)
	return pt
}

// B62Encode returns n in base62 with no padding, e.g. 62 is "10".
// n must not be negative. See also the worked example in b62_test.go.
func B62Encode(n int64) string {
	d, pt := b62EncodeSpec(n)
	o := make([]byte, pt)
	pi := 0 // place idx
	for {
		decv := n / d
		o[pi] = decDigitToB62(decv)
		if d == 1 {
			break
		}
		n -= (decv * d)
		d /= 62
		pi++
	}
	return string(o)
}

// B62EncodeFixed returns n in base62, left-padded with zeros to exactly
// places digits, so values sort correctly as strings. Values too large for
// places keep only their lowest places digits, as in 1.0.x.
// n must not be negative.
func B62EncodeFixed(n int64, places int) string {
	o := B62Encode(n)
	padLen := places - len(o)
	if padLen < 0 {
		return o[-padLen:]
	}
	if padLen > 0 {
		pad := make([]byte, padLen)
		for i := 0; i < padLen; i++ {
			pad[i] = zeroDigit
		}
		o = string(pad) + o
	}
	return o
}

// B62Decode returns the value of the base62 string v. Leading zeros are
// ignored. v must contain only B62_DIGITS characters and fit in an int64;
// other input returns an unspecified value.
func B62Decode(v string) int64 {
	var o int64
	var d int64 = 1
	for pi := len(v) - 1; pi >= 0; pi-- {
		o += (b62DigitToDec(v[pi]) * d)
		d *= 62
	}
	return o
}
