package domain

type Move int

const (
	Rock Move = iota
	Paper
	Scissors
)

type GameType int

const (
	PlayerVsPlayer GameType = iota
	PlayerVsBot
)

type Game struct {
	ID       string
	Player1  string
	Player2  string
	Winner   string
	PlayedAt string
}

type RoundResult struct {
	Move1  Move
	Move2  Move
	Winner string
}

type GameRepository interface {
	SaveGame(result *Game) error
	GetGameHistory() ([]*Game, error)
}

type RandomGenerator interface {
	GenerateMove() Move
}

func (m Move) String() string {
	switch m {
	case Rock:
		return "Rock"
	case Paper:
		return "Paper"
	case Scissors:
		return "Scissors"
	default:
		return "Unknown"
	}
}

func (m GameType) String() string {
	switch m {
	case PlayerVsPlayer:
		return "Player vs Player"
	case PlayerVsBot:
		return "Player vs Bot"
	default:
		return "Unknown"
	}
}

func DetermineWinner(move1, move2 Move) int {
	if move1 == move2 {
		return 0 // Draw
	}

	if (move1 == Rock && move2 == Scissors) ||
		(move1 == Paper && move2 == Rock) ||
		(move1 == Scissors && move2 == Paper) {
		return 1 // Player 1 wins
	}

	return 2 // Player 2 wins
}
