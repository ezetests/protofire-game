package randomness

import (
	"math/rand"
	"time"

	"protofire-game/internal/domain"
)

type DefaultRandomGenerator struct {
	rand *rand.Rand
}

func NewDefaultRandomGenerator() *DefaultRandomGenerator {
	source := rand.NewSource(time.Now().UnixNano())
	return &DefaultRandomGenerator{
		rand: rand.New(source),
	}
}

func (g *DefaultRandomGenerator) GenerateMove() domain.Move {
	return domain.Move(g.rand.Intn(3)) // 0 = Rock, 1 = Paper, 2 = Scissors
}
