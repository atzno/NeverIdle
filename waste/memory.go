package waste

import (
	"crypto/rand"
	"sync"

	"github.com/by275/noidle/internal/log"
)

const (
	KiB = 1024
	MiB = 1024 * KiB
	GiB = 1024 * MiB
)

type GiBObject struct {
	B [GiB]byte
}

type memoryStore struct {
	mu      sync.RWMutex
	buffers []*GiBObject
}

var allocatedMemory memoryStore

func Memory(gib int) {
	if gib <= 0 {
		return
	}

	buffers := make([]*GiBObject, 0, gib)
	for i := 0; i < gib; i++ {
		o := new(GiBObject)
		if _, err := rand.Read(o.B[:]); err != nil {
			log.Panicf("MEM", "failed to initialize reserved memory: %v", err)
		}
		buffers = append(buffers, o)
	}
	allocatedMemory.set(buffers)
}

func (m *memoryStore) set(buffers []*GiBObject) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.buffers = buffers
}

// firstBufferPrefix returns a slice of the first buffer's beginning without copying
// to reduce memory overhead. Caller must not modify the returned slice.
func (m *memoryStore) firstBufferPrefix(size int) []byte {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if len(m.buffers) == 0 {
		return nil
	}

	if size > len(m.buffers[0].B) {
		size = len(m.buffers[0].B)
	}
	// Return direct slice without copying for efficiency
	return m.buffers[0].B[:size]
}
