package randomness

import (
	"protofire-game/internal/domain"
)

type MockRandomGenerator struct {
	Moves []domain.Move
	Index int
}

func NewMockRandomGenerator(moves []domain.Move) *MockRandomGenerator {
	return &MockRandomGenerator{
		Moves: moves,
		Index: 0,
	}
}

func (m *MockRandomGenerator) GenerateMove() domain.Move {
	if m.Index >= len(m.Moves) {
		m.Index = 0
	}
	move := m.Moves[m.Index]
	m.Index++
	return move
}
