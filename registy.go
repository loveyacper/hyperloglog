package hyperloglog

import (
    "fmt"
    "log"
)

const (
    loglog2Capcity uint32 = 6 // rep 64, you can count up to 2**64 objects
)

type registry struct {
    m []uint32
}

func newRegistry(bucketCount uint32) *registry {
    if bucketCount == 0 || bucketCount > 16384 {
        panic("bucketCount should be [1, 16384]")
    }

    if bucketCount & (bucketCount -1) != 0 {
        panic("bucketCount should be power of 2")
    }

    reg := &registry{}
    reg.m = make([]uint32, calcSizeOfUint32(bucketCount, loglog2Capcity))

    return reg
}

func (reg *registry)bucketCount() uint32 {
    l := uint32(len(reg.m))
    return l * (32 / loglog2Capcity)
}

func calcSizeOfUint32(bucketCount uint32, loglog2Capcity uint32) uint32 {
    n := bucketCount / (32 / loglog2Capcity)
    if bucketCount % (32 / loglog2Capcity) != 0 {
        return n + 1
    }

    return n
}

func (reg *registry)set(position uint32, leadingZeros uint32) {
    if position >= reg.bucketCount() {
        panic(fmt.Sprintf("wrong position %v, expect < %v", position, reg.bucketCount()))
    }

    if leadingZeros > 64 {
        panic("wrong leadingZeros")
    }

    bucket := position / (32 / loglog2Capcity)
    shift := loglog2Capcity * (position - bucket * (32 / loglog2Capcity))

    var leadingZerosMask uint32 = (1 << loglog2Capcity) - 1
    reg.m[bucket] = (reg.m[bucket] & ^(leadingZerosMask << shift)) | (leadingZeros << shift)
}

func (reg *registry)get(position uint32) uint32 {
    if position >= reg.bucketCount() {
        panic(fmt.Sprintf("wrong position %v, expect < %v", position, reg.bucketCount()))
    }

    bucket := position / (32 / loglog2Capcity)
    shift := loglog2Capcity * (position - bucket * (32 / loglog2Capcity))

    var leadingZerosMask uint32 = (1 << loglog2Capcity) - 1
    return (reg.m[bucket] & (leadingZerosMask << shift)) >> shift
}

func (reg *registry)update(position uint32, leadingZeros uint32) bool {
    var curVal uint32 = reg.get(position)
    if curVal < leadingZeros {
        reg.set(position, leadingZeros)
        return true
    }

    return false
}

func (reg *registry)merge(other *registry) {
    if len(reg.m) != len(other.m) {
        panic("m should be same for merge")
    }

    for bucket := uint32(0); bucket < reg.bucketCount(); bucket++ {
        me := reg.get(bucket)
        he := other.get(bucket)
        if me != 0 || he != 0 {
            log.Printf("%v %v %v\n", bucket, me, he)
        }
        if he > me {
            reg.set(bucket, he)
        }
    }
}

