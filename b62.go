package bwid

// Use base62 for storing numbers according to definition on Wikipedia
// 0-9 = 0-9
// 10-35 = A-Z
// 36-61 = a-z

// Calculate number of places/digits required to hold n
func B62Len(n int64) int {
	nd := n
	l := 1 // digit qty
	for nd > 61 {
		l++
		nd = nd / 62
	}
	return l
}

// Encode to base62 with a fixed number of digits
// (useful for alpha sorts)
func B62EncodeFixed(n int64, places int) string {
	o := make([]byte, places)
	for i := 1; i <= places; i++ {
		b := byte(n % 62)
		switch {
		case b < 10: // 0-9
			b += 48
		case b < 36: // A-Z
			b += 55
		default: // a-z
			b += 61
		}
		o[places-i] = b
		n /= 62
	}
	return string(o)
}

// Decode from base62
func B62Decode(o string) (n int64) {
	places := len(o)
	f := 1
	for i := 1; i <= places; i++ {
		b := int(o[places-i])
		switch {
		case b < 58: // 0-9
			b -= 48
		case b < 91: // A-Z
			b -= 55
		default: // a-z
			b -= 61
		}
		if i == 1 {
			n = int64(b)
		} else {
			n += int64(b * f)
		}
		f *= 62
	}
	return
}
