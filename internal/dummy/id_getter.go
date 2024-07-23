package dummy

import (
	"math/rand/v2"
	"strconv"
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
	return strconv.FormatUint(rand.Uint64(), 10)
}
