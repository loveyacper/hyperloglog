package hyperloglog

import (
	"hash"
	"hash/fnv"
	"log"
	"math"
	"math/bits"
)

//HyperLogLog probabilistic cardinality approximation.

type HyperLoglog struct {
	hash            hash.Hash64
	reg             *registry
	log2bucketCount uint32
	alphaMM         float64
}

func NewHyperLoglog(bucketCount uint32) *HyperLoglog {
	if bucketCount == 0 || bucketCount > (1<<16) {
		panic("wrong bucketCount")
	}
	if (bucketCount & (bucketCount - 1)) != 0 {
		panic("bucketCount should be pow of 2")
	}

	reg := newRegistry(bucketCount)
	hash := fnv.New64()
	log2bucketCount := bits.TrailingZeros32(bucketCount)

	alpha := getAlphaMM(log2bucketCount, float64(bucketCount))

	return &HyperLoglog{hash: hash, reg: reg, log2bucketCount: uint32(log2bucketCount), alphaMM: alpha}
}

func (hll *HyperLoglog) Add(elem []byte) bool {
	hll.hash.Reset()
	hll.hash.Write(elem)
	hashValue := hll.hash.Sum64()
	bucketPos := hashValue >> (64 - hll.log2bucketCount)

	leadingZeros := 64 - hll.log2bucketCount
	if (hashValue << hll.log2bucketCount) != 0 {
		leadingZeros = uint32(bits.LeadingZeros64(hashValue << hll.log2bucketCount))
	}

	//log.Println("hashv bucketPos and leading : ", bucketPos, leadingZeros+1)

	return hll.reg.update(uint32(bucketPos), uint32(leadingZeros+1))
}

func (hll *HyperLoglog) Count() int {
	var sum float64 = 0
	zeros := 0 // V in the paper
	bucketCount := 1 << hll.log2bucketCount
	for i := 0; i < bucketCount; i++ {
		v := hll.reg.get(uint32(i))
		if v == 0 {
			zeros++
			// continue  if continue, big bias when small card???
		}

		v = (1 << v)
		sum += float64(1) / float64(v)
	}

	estimate := (1 / sum) * hll.alphaMM

	log.Println("estimate bucketCount alpha", estimate, bucketCount, hll.alphaMM)
	if estimate <= (5.0/2.0)*float64(bucketCount) {
		// Small Range Estimate
		// TODO why multiply by 0.7???
		return int(math.Ceil(linearCounting(bucketCount, zeros)) * 0.7)
	}

	return int(math.Ceil(estimate))
}

func getAlphaMM(p int, m float64) float64 {
	// See the paper.
	if p < 4 || p > 16 {
		panic("p should in [4,16)")
	}
	switch p {
	case 4:
		return 0.673 * m * m
	case 5:
		return 0.697 * m * m
	case 6:
		return 0.709 * m * m
	default:
		return (0.7213 / (1 + 1.079/m)) * m * m
	}
}

func linearCounting(bucketCount int, zeros int) float64 {
	count := float64(bucketCount)
	return count * math.Log2(count/float64(zeros))
}
