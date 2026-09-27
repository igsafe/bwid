package bwid

// Use base62 for storing numbers according to definition on Wikipedia
// 0-9 = 0-9
// 10-35 = A-Z
// 36-61 = a-z

const ZeroDigit = byte(48)

func DecDigitToB62(d uint64) byte {
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

func B62DigitToDec(b byte) uint64 {
	// 0-9
	if b < 58 {
		return uint64(b - 48)
	}
	// A-Z
	if b < 91 {
		return uint64(b - 55)
	}
	// a-z
	return uint64(b - 61)
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

// Get highest place divisor and total number of places
// for a B62 encoding
func B62EncodeSpec(n uint64) (d uint64, pt uint64) {
	d = 1  // position 1 divisor 62
	pt = 1 // total places 2
	for {
		d2 := d * 62
		if n >= d2 {
			d = d2
			pt++
			continue
		}
		return // highest position divisor, total places
	}
}

// for backward compatibility
// Calculate number of places/digits required to hold n
func B62Len(n uint64) uint64 {
	_, pt := B62EncodeSpec(n)
	return pt
}

// Encode to base62
// See also math proofs in b62_test.go
func B62Encode(n uint64) string {
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
func B62EncodeFixed(n uint64, places uint64) string {
	o := B62Encode(n)
	padLen := places - uint64(len(o))
	if padLen > 0 {
		pad := make([]byte, padLen)
		for i := uint64(0); i < padLen; i++ {
			pad[i] = ZeroDigit
		}
		o = string(pad) + o
	}
	return o
}

// Decode from base62
func B62Decode(v string) uint64 {
	var o uint64
	var d uint64 = 1
	for pi := len(v) - 1; pi >= 0; pi-- {
		o += (B62DigitToDec(v[pi]) * d)
		d *= 62
	}
	return o
}
