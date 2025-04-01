package repository

import (
	"protofire-game/internal/domain"
)

type MockRepository struct {
	Games []*domain.Game
}

func NewMockRepository() *MockRepository {
	return &MockRepository{
		Games: make([]*domain.Game, 0),
	}
}

func (m *MockRepository) SaveGame(result *domain.Game) error {
	m.Games = append(m.Games, result)
	return nil
}

func (m *MockRepository) GetGameHistory() ([]*domain.Game, error) {
	return m.Games, nil
}

func (m *MockRepository) Close() error {
	return nil
}
