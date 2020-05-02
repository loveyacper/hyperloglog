package main

import (
    "log"
)

//HyperLogLog probabilistic cardinality approximation.


const (
    bucketCount uint32 = 1024 // the m count
    bucketBits uint32 = 10 // 2**10 = 1024
)

type registry struct {
    loglog2Capcity uint32 // if 6, then rep 64, you can count up to 2**64 objects
    m []uint32
}

func newRegistry(capacityAsLeadingZeros uint32) *registry {
    if capacityAsLeadingZeros < 4 || capacityAsLeadingZeros > 6 {
        panic("capacityAsLeadingZeros should be 4, 5, or 6, represent cardinality up to 2**16, 2*32, 2**64")
    }

    reg := &registry{}
    reg.loglog2Capcity = capacityAsLeadingZeros
    reg.m = make([]uint32, calcSizeOfUint32(bucketCount, reg.loglog2Capcity))

    return reg
}

func calcSizeOfUint32(bucketCount uint32, loglog2Capcity uint32) uint32 {
    n := bucketCount / (32 / loglog2Capcity)
    if bucketCount % (32 / loglog2Capcity) != 0 {
        return n + 1
    }

    return n
}

func (reg *registry)set(position uint32, leadingZeros uint32) {
    if position >= bucketCount {
        panic("position should be [0, 1024)")
    }

    if leadingZeros > (1 << reg.loglog2Capcity) {
        panic("wrong leadingZeros")
    }

    bucket := position / (32 / reg.loglog2Capcity)
    shift := reg.loglog2Capcity * (position - bucket * (32 / reg.loglog2Capcity))

    var leadingZerosMask uint32 = (1 << reg.loglog2Capcity) - 1
    reg.m[bucket] = (reg.m[bucket] & ^(leadingZerosMask << shift)) | (leadingZeros << shift)
}

func (reg *registry)get(position uint32) uint32 {
    if position >= bucketCount {
        panic("position should be [0, 1024)")
    }

    bucket := position / (32 / reg.loglog2Capcity)
    shift := reg.loglog2Capcity * (position - bucket * (32 / reg.loglog2Capcity))

    var leadingZerosMask uint32 = (1 << reg.loglog2Capcity) - 1
    return (reg.m[bucket] & (leadingZerosMask << shift)) >> shift
}

func (reg *registry)update(position uint32, leadingZeros uint32) {
    var curVal uint32 = reg.get(position)
    if curVal < leadingZeros {
        reg.set(position, leadingZeros)
    }
}


func (reg *registry)merge(other *registry) {
    if reg.loglog2Capcity != other.loglog2Capcity ||
       len(reg.m) != len(other.m) {
        panic("xx")
    }

    for bucket := uint32(0); bucket < bucketCount; bucket++ {
        me := reg.get(bucket)
        he := other.get(bucket)
        if he > me {
            reg.set(bucket, he)
        }
    }
}

// type hyper log

func main() {
    reg := newRegistry(6)
    reg.set(1000, 30)
    log.Println(reg.get(1000))

    reg.set(0, 20)
    log.Println(reg.get(0))

    reg.update(1000, 30)
    log.Println(reg.get(1000))

    reg.update(0, 25)
    log.Println(reg.get(0))
}

