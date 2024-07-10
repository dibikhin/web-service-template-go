package dummy

import (
	"fmt"
	"math/rand/v2"
)

type IDGenerator interface {
	NewID() string
}

type randIDGenerator struct {
}

func NewRandIDGenerator() IDGenerator {
	return &randIDGenerator{}
}

func (g *randIDGenerator) NewID() string {
	return fmt.Sprintf("%d", rand.Uint64())
}
