package goid

import (
	"crypto/rand"
	"math/big"
	"sync/atomic"
	"time"
)

type ID3 struct {
	id          int64
	delta       uint16
	randomDelta uint16
	node        uint16
	nodeBits    uint8
	bits        uint8
}

func (i *ID3) Generate() int64 {
	for {
		old := atomic.LoadInt64(&i.id)
		nt := time.Now().UnixMilli()
		ncbits := 53 - i.bits
		lt := (old >> ncbits) & ((1 << i.bits) - 1)
		cBits := ncbits - i.nodeBits
		mask := uint16((1 << cBits) - 1)
		ct := uint16(old) & mask
		if nt < lt {
			time.Sleep(time.Microsecond)
			continue
		}
		if nt == lt {
			ct += i.getDelta()
			if ct > mask {
				continue
			}
		} else {
			ct = i.getDelta()
		}

		now := (nt << ncbits) | int64(ct)
		if i.nodeBits > 0 {
			now |= int64(i.node) << cBits
		}
		if atomic.CompareAndSwapInt64(&i.id, old, now) {
			return now
		}
	}
}

func (i *ID3) getDelta() uint16 {
	if i.randomDelta > 0 {
		if de, err := rand.Int(rand.Reader, big.NewInt(int64(i.randomDelta))); err == nil {
			return uint16(de.Int64())
		}
	}
	return i.delta
}

func (i *ID3) SetDelta(d uint16) {
	if d == 0 || d >= (1<<(53-i.bits-i.nodeBits)-1) {
		panic("delta too large or invalid")
	}
	i.delta = d
}

func (i *ID3) GetDelta() uint16 {
	return i.delta
}

func (i *ID3) SetRandomDelta(r uint16) {
	if r == 0 || r >= (1<<(53-i.bits-i.nodeBits)-1) {
		panic("random delta too large or invalid")
	}
	i.randomDelta = r
}

func (i *ID3) GetRandomDelta() uint16 {
	return i.randomDelta
}

func (i *ID3) SetNode(node uint16, nodeBits uint8) {
	if nodeBits < 2 || nodeBits > (53-i.bits-2) ||
		node == 0 || node > (1<<nodeBits-1) ||
		i.delta >= (1<<(53-i.bits-nodeBits)-1) ||
		i.randomDelta >= (1<<(53-i.bits-nodeBits)-1) {
		panic("node or nodeBits is invalid")
	}
	i.node, i.nodeBits = node, nodeBits
}

func (i *ID3) GetNode() (node uint16, nodeBits uint8) {
	return i.node, i.nodeBits
}

func (i *ID3) SetBits(bits uint8) {
	if bits < 42 || bits > 43 ||
		i.delta >= (1<<(53-bits-i.nodeBits)-1) ||
		i.randomDelta >= (1<<(53-bits-i.nodeBits)-1) {
		panic("bits is invalid")
	}
	i.bits = bits
}

func NewID3() *ID3 {
	return &ID3{
		delta: 1,
		bits:  42,
	}
}

var _id3 = NewID3()

func GetID3() *ID3 {
	return _id3
}

func GenID3() int64 {
	return _id3.Generate()
}

func ResolveID3(id int64, oid *ID3) (timestamp int64, counter uint16) {
	return id >> (53 - oid.bits), uint16(id) & uint16((1<<(53-oid.bits-oid.nodeBits))-1)
}