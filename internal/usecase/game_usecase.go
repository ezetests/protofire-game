package usecase

import (
	"fmt"
	"time"

	"protofire-game/internal/domain"

	"github.com/google/uuid"
)

type GameUseCase struct {
	repository      domain.GameRepository
	randomGenerator domain.RandomGenerator
	currentGame     *domain.Game
	currentMode     domain.GameType
	currentRounds   []domain.RoundResult
}

func NewGameUseCase(repo domain.GameRepository, randGen domain.RandomGenerator) *GameUseCase {
	return &GameUseCase{
		repository:      repo,
		randomGenerator: randGen,
	}
}

func (g *GameUseCase) StartNewGame(mode domain.GameType, player1, player2 string) error {
	if err := domain.ValidatePlayerName(player1); err != nil {
		return fmt.Errorf("invalid player1 name: %w", err)
	}
	if err := domain.ValidatePlayerName(player2); err != nil {
		return fmt.Errorf("invalid player2 name: %w", err)
	}

	g.currentGame = &domain.Game{
		ID:       uuid.New().String(),
		Player1:  player1,
		Player2:  player2,
		PlayedAt: time.Now().Format(time.RFC3339),
	}
	g.currentMode = mode
	g.currentRounds = make([]domain.RoundResult, 0)
	return nil
}

func (g *GameUseCase) PlayRound(move1, move2 domain.Move) (*domain.Game, error) {
	if g.currentGame == nil {
		return nil, fmt.Errorf("no game in progress")
	}

	if g.currentMode == domain.PlayerVsBot {
		move2 = g.randomGenerator.GenerateMove()
	}

	winner := domain.DetermineWinner(move1, move2)
	var winnerName string
	switch winner {
	case 0:
		winnerName = "Draw"
	case 1:
		winnerName = g.currentGame.Player1
	case 2:
		winnerName = g.currentGame.Player2
	}

	round := domain.RoundResult{
		Move1:  move1,
		Move2:  move2,
		Winner: winnerName,
	}

	g.currentRounds = append(g.currentRounds, round)

	if len(g.currentRounds) == 2 {
		if g.currentRounds[0].Winner == g.currentRounds[1].Winner &&
			g.currentRounds[0].Winner != "Draw" {
			g.currentGame.Winner = g.currentRounds[0].Winner
			if err := g.repository.SaveGame(g.currentGame); err != nil {
				return nil, fmt.Errorf("failed to save game: %w", err)
			}
			result := g.currentGame
			g.currentGame = nil
			g.currentRounds = nil
			return result, nil
		}
	}

	if len(g.currentRounds) == 3 {
		p1Wins := 0
		p2Wins := 0
		draws := 0

		for _, r := range g.currentRounds {
			if r.Winner == g.currentGame.Player1 {
				p1Wins++
			} else if r.Winner == g.currentGame.Player2 {
				p2Wins++
			} else {
				draws++
			}
		}

		if p1Wins > p2Wins {
			g.currentGame.Winner = g.currentGame.Player1
		} else if p2Wins > p1Wins {
			g.currentGame.Winner = g.currentGame.Player2
		} else {
			g.currentGame.Winner = "Draw"
		}

		if err := g.repository.SaveGame(g.currentGame); err != nil {
			return nil, fmt.Errorf("failed to save game: %w", err)
		}
		result := g.currentGame
		g.currentGame = nil
		g.currentRounds = nil
		return result, nil
	}

	return g.currentGame, nil
}

func (g *GameUseCase) GetHistory() ([]*domain.Game, error) {
	return g.repository.GetGameHistory()
}

func (g *GameUseCase) GetCurrentRounds() []domain.RoundResult {
	return g.currentRounds
}
