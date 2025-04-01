package usecase

import (
	"testing"
	"time"

	"protofire-game/internal/domain"
	"protofire-game/internal/randomness"
	"protofire-game/internal/repository"

	"github.com/stretchr/testify/assert"
)

func TestStartNewGame(t *testing.T) {
	repo := repository.NewMockRepository()
	randGen := randomness.NewMockRandomGenerator([]domain.Move{})
	gameUseCase := NewGameUseCase(repo, randGen)

	gameUseCase.StartNewGame(domain.PlayerVsPlayer, "Player1", "Player2")

	assert.NotNil(t, gameUseCase.currentGame)
	assert.Equal(t, "Player1", gameUseCase.currentGame.Player1)
	assert.Equal(t, "Player2", gameUseCase.currentGame.Player2)
	assert.Equal(t, domain.PlayerVsPlayer, gameUseCase.currentMode)
	assert.Empty(t, gameUseCase.currentRounds)
}

func TestPlayRoundPlayerVsPlayer(t *testing.T) {
	repo := repository.NewMockRepository()
	randGen := randomness.NewMockRandomGenerator([]domain.Move{})
	gameUseCase := NewGameUseCase(repo, randGen)

	gameUseCase.StartNewGame(domain.PlayerVsPlayer, "Player1", "Player2")

	// Test first round - Player1 wins
	result, err := gameUseCase.PlayRound(domain.Rock, domain.Scissors)
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Empty(t, result.Winner)
	assert.Len(t, gameUseCase.currentRounds, 1)

	// Test second round - Player1 wins again
	result, err = gameUseCase.PlayRound(domain.Paper, domain.Rock)
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "Player1", result.Winner)
	assert.Empty(t, gameUseCase.currentGame)
}

func TestPlayRoundPlayerVsBot(t *testing.T) {
	repo := repository.NewMockRepository()
	randGen := randomness.NewMockRandomGenerator([]domain.Move{domain.Rock, domain.Rock, domain.Rock})
	gameUseCase := NewGameUseCase(repo, randGen)

	gameUseCase.StartNewGame(domain.PlayerVsBot, "Player1", "Bot")

	// Test first round - Player1 wins (Paper beats Rock)
	result, err := gameUseCase.PlayRound(domain.Paper, 0)
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Empty(t, result.Winner)
	assert.Len(t, gameUseCase.currentRounds, 1)

	// Test second round - Player1 wins (Paper beats Rock)
	result, err = gameUseCase.PlayRound(domain.Paper, 0)
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "Player1", result.Winner)
	assert.Empty(t, gameUseCase.currentGame)
}

func TestPlayRoundDraw(t *testing.T) {
	repo := repository.NewMockRepository()
	randGen := randomness.NewMockRandomGenerator([]domain.Move{})
	gameUseCase := NewGameUseCase(repo, randGen)

	gameUseCase.StartNewGame(domain.PlayerVsPlayer, "Player1", "Player2")

	// Test first round - Draw
	result, err := gameUseCase.PlayRound(domain.Rock, domain.Rock)
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Empty(t, result.Winner)
	assert.Len(t, gameUseCase.currentRounds, 1)

	// Test second round - Draw
	result, err = gameUseCase.PlayRound(domain.Paper, domain.Paper)
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Empty(t, result.Winner)
	assert.Len(t, gameUseCase.currentRounds, 2)

	// Test third round - Draw
	result, err = gameUseCase.PlayRound(domain.Scissors, domain.Scissors)
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "Draw", result.Winner)
	assert.Empty(t, gameUseCase.currentGame)
}

func TestGetHistory(t *testing.T) {
	repo := repository.NewMockRepository()
	randGen := randomness.NewMockRandomGenerator([]domain.Move{})
	gameUseCase := NewGameUseCase(repo, randGen)

	testGames := []*domain.Game{
		{
			ID:       "game1",
			Player1:  "Player1",
			Player2:  "Player2",
			Winner:   "Player1",
			PlayedAt: time.Now().Format(time.RFC3339),
		},
		{
			ID:       "game2",
			Player1:  "Player3",
			Player2:  "Player4",
			Winner:   "Player4",
			PlayedAt: time.Now().Format(time.RFC3339),
		},
	}

	for _, game := range testGames {
		err := repo.SaveGame(game)
		assert.NoError(t, err)
	}

	history, err := gameUseCase.GetHistory()
	assert.NoError(t, err)
	assert.Len(t, history, 2)
	assert.Equal(t, testGames[0].ID, history[0].ID)
	assert.Equal(t, testGames[1].ID, history[1].ID)
}
