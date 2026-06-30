package waste

import (
	"crypto/rand"
	"sync"
	"time"

	"github.com/by275/noidle/internal/log"
	"golang.org/x/crypto/chacha20"
)

const cpuWorkerCount = 8

// cipherPool reduces allocation overhead by reusing ciphers
type cipherPool struct {
	mu      sync.Mutex
	ciphers []*chacha20.Cipher
}

var pool = &cipherPool{}

func (cp *cipherPool) get() *chacha20.Cipher {
	cp.mu.Lock()
	defer cp.mu.Unlock()
	if len(cp.ciphers) > 0 {
		c := cp.ciphers[len(cp.ciphers)-1]
		cp.ciphers = cp.ciphers[:len(cp.ciphers)-1]
		return c
	}
	return nil
}

func (cp *cipherPool) put(c *chacha20.Cipher) {
	cp.mu.Lock()
	defer cp.mu.Unlock()
	if len(cp.ciphers) < 16 { // limit pool size
		cp.ciphers = append(cp.ciphers, c)
	}
}

func CPU(interval time.Duration) {
	workCh := make(chan struct{})
	doneCh := make(chan struct{}, cpuWorkerCount)
	for range cpuWorkerCount {
		go runCPUWorker(workCh, doneCh)
	}

	for {
		for range cpuWorkerCount {
			workCh <- struct{}{}
		}
		for range cpuWorkerCount {
			<-doneCh
		}

		log.Logf("CPU", "Successfully wasted")
		time.Sleep(interval)
	}
}

func runCPUWorker(workCh <-chan struct{}, doneCh chan<- struct{}) {
	buffer, cipher := newCPUBufferAndCipher()

	for range workCh {
		// Perform crypto operations
		for range 64 {
			cipher.XORKeyStream(buffer, buffer)
		}

		// Occasionally rotate cipher to ensure freshness
		// but less frequently to reduce allocation overhead
		if newCipher, err := chacha20.NewUnauthenticatedCipher(buffer[:32], buffer[:24]); err == nil {
			pool.put(cipher)
			cipher = newCipher
		}

		doneCh <- struct{}{}
	}
}

func newCPUBufferAndCipher() ([]byte, *chacha20.Cipher) {
	buffer := allocatedMemory.firstBufferPrefix(4 * MiB)
	if len(buffer) == 0 {
		buffer = make([]byte, 4*MiB)
	}
	if _, err := rand.Read(buffer); err != nil {
		log.Panicf("CPU", "failed to initialize CPU buffer: %v", err)
	}

	// Try to get cipher from pool first
	if c := pool.get(); c != nil {
		return buffer, c
	}

	cipher, err := chacha20.NewUnauthenticatedCipher(buffer[:32], buffer[:24])
	if err != nil {
		log.Panicf("CPU", "failed to initialize CPU cipher: %v", err)
	}
	return buffer, cipher
}
