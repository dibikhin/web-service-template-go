package dummy

import (
	"fmt"
	"math/rand/v2"
)

type IDGetter interface {
	GetID() string
}

type idGetter struct {
}

func NewIDGetter() IDGetter {
	return &idGetter{}
}

func (g *idGetter) GetID() string {
	return fmt.Sprintf("%d", rand.Uint64())
}
