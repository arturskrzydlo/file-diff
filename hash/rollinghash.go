package hash

import "math/rand"

const (
	WindowSize   = 63
	bhRotate     = WindowSize % 32
	bhRotateComp = 32 - bhRotate
)

type BuzHash struct {
	currentHash int
	hashes      [256]int
}

func NewBuzHash() *BuzHash {
	var hashes [256]int
	rand.Seed(42)
	for i := 0; i < 256; i++ {
		hashes[i] = rand.Int()
	}

	return &BuzHash{
		hashes: hashes,
	}
}

func (bh *BuzHash) RollingHash(oldByte byte, newByte byte) int {
	oldHash := bh.hashes[oldByte+128]
	bh.currentHash = (bh.currentHash<<1 | bh.currentHash>>31) ^
		(oldHash<<bhRotate | oldHash>>bhRotateComp) ^ bh.hashes[newByte+128]
	return bh.currentHash
}

// precalculating hashes for given window may be used to write func ?
func (bh *BuzHash) ResetHash(data []byte, pos int) {
	bh.currentHash = 0
	for i := WindowSize; i > 0; i-- {
		bh.currentHash = (bh.currentHash<<1 | bh.currentHash>>31) ^ bh.hashes[data[pos-i]+128]
	}
}
