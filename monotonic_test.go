package bwid

import (
	"sort"
	"strings"
	"sync"
	"testing"
	"time"
)

// a fixed instant in the past, so real-clock tests that run afterward always
// see time moving forward
var frozenAt = time.Unix(1790000000, 500)

func resetMonotonic() {
	monotonic.Lock()
	monotonic.last = map[int]string{}
	monotonic.Unlock()
}

// freezeClock makes clock() return at, with fresh monotonic state, and
// restores the real clock and fresh state when the test ends
func freezeClock(t *testing.T, at time.Time) {
	t.Helper()
	resetMonotonic()
	clock = func() time.Time { return at }
	t.Cleanup(func() {
		clock = time.Now
		resetMonotonic()
	})
}

func assertB62Token(t *testing.T, token string, length int) {
	t.Helper()
	assertEqual(t, length, len(token))
	for i := 0; i < len(token); i++ {
		if strings.IndexByte(B62_DIGITS, token[i]) < 0 {
			t.Errorf("token %q has non-base62 character %q", token, token[i])
		}
	}
}

func TestNextTick(t *testing.T) {
	sec := B62EncodeFixed(1790000000, TIMESTAMP_LEN)
	next := B62EncodeFixed(1790000001, TIMESTAMP_LEN)
	assertEqual(t, sec+B62EncodeFixed(1, TIMESTAMP_NANO_LEN), nextTick(sec+B62EncodeFixed(0, TIMESTAMP_NANO_LEN)))
	// 999999999ns rolls over to the next second
	assertEqual(t, next+B62EncodeFixed(0, TIMESTAMP_NANO_LEN), nextTick(sec+B62EncodeFixed(999999999, TIMESTAMP_NANO_LEN)))
}

func TestMonotonicFrozenClock(t *testing.T) {
	freezeClock(t, frozenAt)
	prev := GenerateObjectId()
	assertB62Token(t, prev, 24)
	assertEqual(t, timestampPrefix(frozenAt, 24), prev[:12])
	bumps, tickMoves := 0, 0
	for i := 0; i < 100000; i++ {
		id := GenerateObjectId()
		assertB62Token(t, id, 24)
		if id <= prev {
			t.Fatalf("not increasing: %q after %q", id, prev)
		}
		if id[:12] == prev[:12] {
			// same tick: head is previous head + 1, tail is redrawn
			bumps++
			assertEqual(t, B62Decode(prev[12:14])+1, B62Decode(id[12:14]))
			if id[14:] == prev[14:] {
				t.Errorf("tail not redrawn: %q after %q", id, prev)
			}
		} else {
			// head overflowed: exactly one nanosecond later
			tickMoves++
			assertEqual(t, nextTick(prev[:12]), id[:12])
		}
		if t.Failed() {
			return
		}
		prev = id
	}
	// both branches must have run
	if bumps == 0 || tickMoves == 0 {
		t.Fatalf("expected head bumps and tick moves, got %d and %d", bumps, tickMoves)
	}
}

func TestMonotonicClockBackward(t *testing.T) {
	freezeClock(t, frozenAt)
	a := GenerateObjectId()
	clock = func() time.Time { return frozenAt.Add(-time.Hour) }
	b := GenerateObjectId()
	// keeps the last timestamp and still sorts after
	assertEqual(t, a[:12], b[:12])
	assertGte(t, b, a)
	if a == b {
		t.Fatal("expected a new token")
	}
	// once the clock passes the last timestamp, it's used again
	later := frozenAt.Add(time.Second)
	clock = func() time.Time { return later }
	assertEqual(t, timestampPrefix(later, 24), GenerateObjectId()[:12])
}

func TestMonotonicHeadOverflow(t *testing.T) {
	freezeClock(t, frozenAt)
	p := timestampPrefix(frozenAt, 24)

	last := p + "zz" + "AAAAAAAAAA"
	monotonic.last[24] = last
	next := GenerateObjectId()
	assertEqual(t, nextTick(p), next[:12])
	assertGte(t, next, last)

	// carry within the head
	monotonic.last[24] = p + "0z" + "AAAAAAAAAA"
	next = GenerateObjectId()
	assertEqual(t, p, next[:12])
	assertEqual(t, "10", next[12:14])
}

func TestMonotonicLength13(t *testing.T) {
	// only one random character, so the whole random part is the head
	freezeClock(t, frozenAt)
	prev := ""
	for i := 0; i < 200; i++ {
		token := GenerateTimestampedToken(13)
		assertB62Token(t, token, 13)
		if token <= prev {
			t.Fatalf("not increasing: %q after %q", token, prev)
		}
		prev = token
	}
}

func TestShortTokensStateless(t *testing.T) {
	// seconds-only lengths stay fully random, as in 1.0.x
	freezeClock(t, frozenAt)
	for length := TIMESTAMP_LEN + 1; length <= TIMESTAMP_LEN+TIMESTAMP_NANO_LEN; length++ {
		GenerateTimestampedToken(length)
		if _, ok := monotonic.last[length]; ok {
			t.Errorf("length %d should not keep monotonic state", length)
		}
	}
}

func TestMonotonicConcurrent(t *testing.T) {
	// goroutines call with no outer lock, so calls really overlap (run with
	// -race to check the locking). each goroutine's own calls happen in
	// order, so its IDs must be strictly increasing, and all IDs unique.
	const workers, each = 50, 20000
	results := make([][]string, workers)
	var wg sync.WaitGroup
	for w := 0; w < workers; w++ {
		wg.Add(1)
		go func(w int) {
			defer wg.Done()
			ids := make([]string, each)
			for i := range ids {
				ids[i] = GenerateObjectId()
			}
			results[w] = ids
		}(w)
	}
	wg.Wait()
	seen := make(map[string]bool, workers*each)
	for w, ids := range results {
		for i, id := range ids {
			if i > 0 && id <= ids[i-1] {
				t.Fatalf("goroutine %d: not increasing: %q after %q", w, id, ids[i-1])
			}
			if seen[id] {
				t.Fatalf("duplicate %q", id)
			}
			seen[id] = true
		}
	}
	assertEqual(t, workers*each, len(seen))
}

func TestMonotonicIssueOrder(t *testing.T) {
	// when calls are serialized, IDs sort in the order they were issued,
	// across goroutines
	const workers, each = 50, 2000
	var mu sync.Mutex
	var issued []string
	var wg sync.WaitGroup
	for w := 0; w < workers; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < each; i++ {
				mu.Lock()
				issued = append(issued, GenerateObjectId())
				mu.Unlock()
			}
		}()
	}
	wg.Wait()
	if !sort.StringsAreSorted(issued) {
		t.Fatal("IDs are not in the order they were issued")
	}
}

func TestMonotonicRealClock(t *testing.T) {
	prev := ""
	for i := 0; i < 1000000; i++ {
		id := GenerateObjectId()
		if id <= prev {
			t.Fatalf("not increasing at %d: %q after %q", i, id, prev)
		}
		prev = id
	}
}
