package dejavu

import (
	"crypto/sha256"
	"sync"
)

// DejaVu witnesses data and recalls if seen before.
type DejaVu interface {

	// Witness data and add to memory. Returns true if previously seen.
	Witness(data []byte) bool

	// WitnessDigest is equivalent to the Winness method but bypasses hashing
	// the data. Use this to improve performance if you already happen
	// to have the sha256 digest.
	WitnessDigest(dataDigest [sha256.Size]byte) bool
}

//////////////////////////////////
// Deterministic implementation //
//////////////////////////////////

type deterministic struct {
	buffer [][sha256.Size]byte       // ring buffer
	size   int                       // ring buffer size
	index  int                       // current ring buffer index
	lookup map[[sha256.Size]byte]int // digest -> newest index (optimization)
	mutex  *sync.Mutex
}

// NewDejaVuDeterministic creates a deterministic DejaVu memory. Will remember
// most recent entries within given entrie limit and forget older entries.
func NewDejaVuDeterministic(entrieLimit uint) DejaVu {
	return &deterministic{
		buffer: make([][sha256.Size]byte, entrieLimit),
		size:   int(entrieLimit),
		index:  0,
		lookup: make(map[[sha256.Size]byte]int),
		mutex:  new(sync.Mutex),
	}
}

func (d *deterministic) WitnessDigest(dataDigest [sha256.Size]byte) bool {
	d.mutex.Lock()
	_, familiar := d.lookup[dataDigest] // check if previously seen

	// rm oldest lookup key if no newer entry
	maxed := len(d.buffer) == d.size // overwriting oldest entry
	if maxed && (d.lookup[d.buffer[d.index]] == d.index) {
		delete(d.lookup, d.buffer[d.index]) // no newer entries
	}

	// add entry and update index/lookup
	d.buffer[d.index] = dataDigest
	d.lookup[dataDigest] = d.index
	d.index = (d.index + 1) % d.size

	d.mutex.Unlock()
	return familiar
}

func (d *deterministic) Witness(data []byte) bool {
	return d.WitnessDigest(sha256.Sum256(data))
}
