// base62 encoding helpers; see package doc in bwid.go for the alphabet

package bwid

const ZeroDigit = byte(48)

func DecDigitToB62(d int64) byte {
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

func B62DigitToDec(b byte) int64 {
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
func IncrementB62(v string) string {
	incrementing := true
	// alloc extra byte in case new place is required
	v2 := make([]byte, len(v)+1)
	for i := len(v) - 1; i >= 0; i-- {
		i2 := i + 1
		if incrementing {
			d := B62DigitToDec(v[i])
			if d == 61 {
				// adding will overflow, next place
				v2[i2] = ZeroDigit
			} else {
				// increment this place
				v2[i2] = DecDigitToB62(d + 1)
				// next place remains the same
				incrementing = false
			}
		} else {
			v2[i2] = v[i]
		}
	}
	if incrementing {
		v2[0] = DecDigitToB62(1)
		return string(v2)
	}
	return string(v2[1:])
}

// B62EncodeSpec returns the highest place divisor d and the
// total number of places pt for the base62 encoding of n.
func B62EncodeSpec(n int64) (d int64, pt int) {
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
func B62Len(n int64) int {
	_, pt := B62EncodeSpec(n)
	return pt
}

// Encode to base62
// See also math proofs in b62_test.go
func B62Encode(n int64) string {
	d, pt := B62EncodeSpec(n)
	o := make([]byte, pt)
	pi := 0 // place idx
	for {
		decv := n / d
		o[pi] = DecDigitToB62(decv)
		if d == 1 {
			break
		}
		n -= (decv * d)
		d /= 62
		pi++
	}
	return string(o)
}

// Encode to base62 with a fixed number of digits
// (useful for alpha sorts)
// values too large for places are truncated to the
// lowest places digits, as in 1.0.x
func B62EncodeFixed(n int64, places int) string {
	o := B62Encode(n)
	padLen := places - len(o)
	if padLen < 0 {
		return o[-padLen:]
	}
	if padLen > 0 {
		pad := make([]byte, padLen)
		for i := 0; i < padLen; i++ {
			pad[i] = ZeroDigit
		}
		o = string(pad) + o
	}
	return o
}

// Decode from base62
func B62Decode(v string) int64 {
	var o int64
	var d int64 = 1
	for pi := len(v) - 1; pi >= 0; pi-- {
		o += (B62DigitToDec(v[pi]) * d)
		d *= 62
	}
	return o
}
