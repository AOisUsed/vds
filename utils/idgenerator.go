package utils

import (
	"fmt"
	"sync"
)

// IdGenerator 可以从0开始产生递增的id
type IdGenerator struct {
	max int
	mu  sync.Mutex
}

func NewIdGenerator() *IdGenerator {
	return &IdGenerator{max: -1}
}

func (idg *IdGenerator) Next() string {
	idg.mu.Lock()
	defer idg.mu.Unlock()
	idg.max++
	return fmt.Sprintf("%d", idg.max)
}

func (idg *IdGenerator) Max() int {
	idg.mu.Lock()
	defer idg.mu.Unlock()
	return idg.max
}

func (idg *IdGenerator) Reset() {
	idg.mu.Lock()
	defer idg.mu.Unlock()
	idg.max = -1
}
