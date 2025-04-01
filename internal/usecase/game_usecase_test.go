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

	err := gameUseCase.StartNewGame(domain.PlayerVsPlayer, "Player1", "Player2")
	assert.NoError(t, err)

	assert.NotNil(t, gameUseCase.currentGame)
	assert.Equal(t, "Player1", gameUseCase.currentGame.Player1)
	assert.Equal(t, "Player2", gameUseCase.currentGame.Player2)
	assert.Equal(t, domain.PlayerVsPlayer, gameUseCase.currentMode)
	assert.Empty(t, gameUseCase.currentRounds)
}

func TestStartNewGameWithInvalidNames(t *testing.T) {
	repo := repository.NewMockRepository()
	randGen := randomness.NewMockRandomGenerator([]domain.Move{})
	gameUseCase := NewGameUseCase(repo, randGen)

	tests := []struct {
		name    string
		player1 string
		player2 string
		wantErr bool
		errMsg  string
	}{
		{
			name:    "empty player1 name",
			player1: "",
			player2: "Player2",
			wantErr: true,
			errMsg:  "invalid player1 name: name cannot be empty",
		},
		{
			name:    "empty player2 name",
			player1: "Player1",
			player2: "",
			wantErr: true,
			errMsg:  "invalid player2 name: name cannot be empty",
		},
		{
			name:    "player1 name too long",
			player1: "ThisNameIsTooLongForTheGame",
			player2: "Player2",
			wantErr: true,
			errMsg:  "invalid player1 name: name cannot be longer than 15 characters",
		},
		{
			name:    "player2 name too long",
			player1: "Player1",
			player2: "ThisNameIsTooLongForTheGame",
			wantErr: true,
			errMsg:  "invalid player2 name: name cannot be longer than 15 characters",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := gameUseCase.StartNewGame(domain.PlayerVsPlayer, tt.player1, tt.player2)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Equal(t, tt.errMsg, err.Error())
				assert.Nil(t, gameUseCase.currentGame)
				return
			}
			assert.NoError(t, err)
		})
	}
}

func TestPlayRoundPlayerVsPlayer(t *testing.T) {
	repo := repository.NewMockRepository()
	randGen := randomness.NewMockRandomGenerator([]domain.Move{})
	gameUseCase := NewGameUseCase(repo, randGen)

	err := gameUseCase.StartNewGame(domain.PlayerVsPlayer, "Player1", "Player2")
	assert.NoError(t, err)

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

	err := gameUseCase.StartNewGame(domain.PlayerVsBot, "Player1", "Bot")
	assert.NoError(t, err)

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
