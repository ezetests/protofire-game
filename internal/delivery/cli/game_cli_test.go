package cli

import (
	"bufio"
	"bytes"
	"errors"
	"io"
	"os"
	"strings"
	"testing"

	"protofire-game/internal/domain"
	"protofire-game/internal/usecase"
)

// MockGameRepository implements domain.GameRepository for testing
type MockGameRepository struct {
	history []*domain.Game
	err     error
}

func (m *MockGameRepository) SaveGame(result *domain.Game) error {
	return m.err
}

func (m *MockGameRepository) GetGameHistory() ([]*domain.Game, error) {
	return m.history, m.err
}

// MockRandomGenerator implements domain.RandomGenerator for testing
type MockRandomGenerator struct {
	move domain.Move
}

func (m *MockRandomGenerator) GenerateMove() domain.Move {
	return m.move
}

func TestNewGameCLI(t *testing.T) {
	repo := &MockGameRepository{}
	randGen := &MockRandomGenerator{}
	useCase := usecase.NewGameUseCase(repo, randGen)
	cli := NewGameCLI(useCase)
	if cli == nil {
		t.Error("NewGameCLI returned nil")
	}
}

func TestValidatePlayerName(t *testing.T) {
	repo := &MockGameRepository{}
	randGen := &MockRandomGenerator{}
	useCase := usecase.NewGameUseCase(repo, randGen)
	cli := NewGameCLI(useCase)

	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{"valid name", "John", false},
		{"empty name", "", true},
		{"too long name", "ThisNameIsTooLongForTheGame", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := cli.validatePlayerName(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("validatePlayerName() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestGetPlayerMove(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected domain.Move
	}{
		{"rock", "rock", domain.Rock},
		{"paper", "paper", domain.Paper},
		{"scissors", "scissors", domain.Scissors},
		{"r", "r", domain.Rock},
		{"p", "p", domain.Paper},
		{"s", "s", domain.Scissors},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &MockGameRepository{}
			randGen := &MockRandomGenerator{}
			useCase := usecase.NewGameUseCase(repo, randGen)
			cli := NewGameCLI(useCase)
			cli.reader = bufio.NewReader(strings.NewReader(tt.input + "\n"))

			got := cli.getPlayerMove("TestPlayer")
			if got != tt.expected {
				t.Errorf("getPlayerMove() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestDisplayResult(t *testing.T) {
	repo := &MockGameRepository{}
	randGen := &MockRandomGenerator{}
	useCase := usecase.NewGameUseCase(repo, randGen)
	cli := NewGameCLI(useCase)

	game := &domain.Game{
		Player1: "Player1",
		Player2: "Player2",
		Winner:  "Player1",
	}

	// Capture stdout
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	cli.displayResult(game)

	// Restore stdout
	w.Close()

	// Read the output
	var buf bytes.Buffer
	io.Copy(&buf, r)

	output := buf.String()
	expectedStrings := []string{
		"Game Result:",
		"Player1 vs Player2",
		"Game winner: Player1",
	}

	for _, expected := range expectedStrings {
		if !strings.Contains(output, expected) {
			t.Errorf("displayResult() output missing expected string: %s", expected)
		}
	}

	os.Stdout = old
}

func TestShowHistory(t *testing.T) {
	repo := &MockGameRepository{
		history: []*domain.Game{
			{
				Player1: "Player1",
				Player2: "Player2",
				Winner:  "Player1",
			},
		},
	}
	randGen := &MockRandomGenerator{}
	useCase := usecase.NewGameUseCase(repo, randGen)
	cli := NewGameCLI(useCase)

	// Capture stdout
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	cli.showHistory()

	// Restore stdout
	w.Close()

	// Read the output
	var buf bytes.Buffer
	io.Copy(&buf, r)

	output := buf.String()
	expectedStrings := []string{
		"Game History:",
		"Players: Player1 vs Player2",
		"Winner: Player1",
	}

	for _, expected := range expectedStrings {
		if !strings.Contains(output, expected) {
			t.Errorf("showHistory() output missing expected string: %s", expected)
		}
	}

	os.Stdout = old
}

func TestShowHistoryError(t *testing.T) {
	repo := &MockGameRepository{
		err: errors.New("database error"),
	}
	randGen := &MockRandomGenerator{}
	useCase := usecase.NewGameUseCase(repo, randGen)
	cli := NewGameCLI(useCase)

	// Capture stdout
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	cli.showHistory()

	// Restore stdout
	w.Close()

	// Read the output
	var buf bytes.Buffer
	io.Copy(&buf, r)

	output := buf.String()
	if !strings.Contains(output, "Error getting history: database error") {
		t.Error("showHistory() did not display error message correctly")
	}

	os.Stdout = old
}
