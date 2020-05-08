package hyperloglog

import (
	"math"
	"math/rand"
	"testing"
	"time"
)

// for test
const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
const (
	letterIdxBits = 6                    // 6 bits to represent a letter index
	letterIdxMask = 1<<letterIdxBits - 1 // All 1-bits, as many as letterIdxBits
	letterIdxMax  = 63 / letterIdxBits   // # of letter indices fitting in 63 bits
)

var src = rand.NewSource(time.Now().UnixNano())

func randBytes(n int) []byte {
	b := make([]byte, n)
	// A src.Int63() generates 63 random bits, enough for letterIdxMax characters!
	for i, cache, remain := n-1, src.Int63(), letterIdxMax; i >= 0; {
		if remain == 0 {
			cache, remain = src.Int63(), letterIdxMax
		}
		if idx := int(cache & letterIdxMask); idx < len(letterBytes) {
			b[i] = letterBytes[idx]
			i--
		}
		cache >>= letterIdxBits
		remain--
	}
	return b
}

func TestRegistry(t *testing.T) {
	reg := newRegistry(1024)
	reg.set(1000, 30)
	if v := reg.get(1000); v != 30 {
		t.Errorf("wrong %v expect 30", v)
	}

	reg.set(0, 20)
	if v := reg.get(0); v != 20 {
		t.Errorf("wrong %v expect 20", v)
	}

	reg.update(1000, 40)
	if v := reg.get(1000); v != 40 {
		t.Errorf("wrong %v expect 40", v)
	}

	reg.update(0, 25)
	if v := reg.get(0); v != 25 {
		t.Errorf("wrong %v expect 25", v)
	}
}

func TestMerge(t *testing.T) {
	reg := newRegistry(1024)
	reg.set(100, 30)

	reg2 := newRegistry(1024)
	reg2.set(100, 20)

	reg.merge(reg2)
	if v := reg.get(100); v != 30 {
		t.Errorf("wrong %v expect 30", v)
	}
}

func TestLargeEstimate(t *testing.T) {
	for n := 0; n <= 10*10000; n += 2000 {
		hll := NewHyperLoglog(16384) // bias: 1.04/128 = 0.81%
		for i := 0; i < n; i++ {
			b := randBytes(128)
			hll.Add(b)
		}

		esti := hll.Count()
		bias := float64(esti-n) / float64(n)
		bias = math.Abs(bias * 100)
		t.Logf("bias %.2f%%: real %d, estimate %d", bias, n, esti)

		if bias > 5.0 {
			t.Errorf("bias %.6f%%, should not exceed 5%%", bias)
		}
	}
}
