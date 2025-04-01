package cli

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"protofire-game/internal/domain"
	"protofire-game/internal/usecase"
)

type GameCLI struct {
	useCase *usecase.GameUseCase
	reader  *bufio.Reader
}

func NewGameCLI(useCase *usecase.GameUseCase) *GameCLI {
	return &GameCLI{
		useCase: useCase,
		reader:  bufio.NewReader(os.Stdin),
	}
}

func (c *GameCLI) Start() {
	for {
		fmt.Println("\nRock Paper Scissors Game")
		fmt.Println("1. Player vs Player")
		fmt.Println("2. Player vs Bot")
		fmt.Println("3. View Game History")
		fmt.Println("4. Exit")
		fmt.Print("Choose an option: ")

		choice := c.readInput()

		switch choice {
		case "1":
			c.playPlayerVsPlayer()
		case "2":
			c.playPlayerVsBot()
		case "3":
			c.showHistory()
		case "4":
			fmt.Println("Thanks for playing!")
			return
		default:
			fmt.Println("Invalid option, please try again")
		}
	}
}

func (c *GameCLI) playPlayerVsPlayer() {
	fmt.Print("Enter Player 1 name: ")
	player1 := c.readPlayerName()

	fmt.Print("Enter Player 2 name: ")
	player2 := c.readPlayerName()

	c.useCase.StartNewGame(domain.PlayerVsPlayer, player1, player2)
	var currentRound int
	var game *domain.Game
	var err error

	fmt.Println("\nBest of 3 rounds! Game ends early if a player wins the first two rounds.")

	for currentRound < 3 {
		currentRound++
		fmt.Printf("\nRound %d:\n", currentRound)

		move1 := c.getPlayerMove(player1)
		move2 := c.getPlayerMove(player2)

		game, err = c.useCase.PlayRound(move1, move2)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		}

		c.displayResult(game)

		if game.Winner != "" {
			if currentRound < 3 {
				fmt.Printf("\n%s won in %d rounds!\n", game.Winner, currentRound)
			}
			return
		}
	}
}

func (c *GameCLI) playPlayerVsBot() {
	fmt.Print("Enter your name: ")
	player1 := c.readPlayerName()

	c.useCase.StartNewGame(domain.PlayerVsBot, player1, "Bot")
	var currentRound int
	var game *domain.Game
	var err error

	fmt.Println("\nBest of 3 rounds! Game ends early if a player wins the first two rounds.")

	for currentRound < 3 {
		currentRound++
		fmt.Printf("\nRound %d:\n", currentRound)

		move1 := c.getPlayerMove(player1)

		game, err = c.useCase.PlayRound(move1, 0) // 0 is a placeholder, bot move is generated in usecase
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		}

		c.displayResult(game)

		if game.Winner != "" {
			if currentRound < 3 {
				fmt.Printf("\n%s won in %d rounds!\n", game.Winner, currentRound)
			}
			return
		}
	}
}

func (c *GameCLI) getPlayerMove(player string) domain.Move {
	for {
		fmt.Printf("%s, enter your move (Rock/Paper/Scissors): ", player)
		input := strings.ToLower(c.readInput())

		switch input {
		case "rock", "r":
			return domain.Rock
		case "paper", "p":
			return domain.Paper
		case "scissors", "s":
			return domain.Scissors
		default:
			fmt.Println("Invalid move. Please enter R, P, or S (or full word)")
		}
	}
}

func (c *GameCLI) showHistory() {
	history, err := c.useCase.GetHistory()
	if err != nil {
		fmt.Printf("Error getting history: %v\n", err)
		return
	}

	if len(history) == 0 {
		fmt.Println("No games played yet!")
		return
	}

	fmt.Println("\nGame History:")
	for _, game := range history {
		fmt.Printf("\nGame ID: %s\n", game.ID)
		fmt.Printf("Players: %s vs %s\n", game.Player1, game.Player2)
		fmt.Printf("Winner: %s\n", game.Winner)
		fmt.Printf("Played at: %v\n", game.PlayedAt)
		fmt.Println("------------------------")
	}
}

func (c *GameCLI) readInput() string {
	input, _ := c.reader.ReadString('\n')
	return strings.TrimSpace(input)
}

func (c *GameCLI) readPlayerName() string {
	for {
		name := c.readInput()
		if err := c.validatePlayerName(name); err != nil {
			fmt.Printf("Invalid name: %v. Please try again: ", err)
			continue
		}
		return name
	}
}

func (c *GameCLI) validatePlayerName(name string) error {
	if len(name) == 0 {
		return fmt.Errorf("name cannot be empty")
	}
	if len(name) > 15 {
		return fmt.Errorf("name cannot be longer than 15 characters")
	}
	return nil
}

func (c *GameCLI) displayResult(result *domain.Game) {
	fmt.Printf("\nGame Result:\n")
	fmt.Printf("%s vs %s\n", result.Player1, result.Player2)

	if len(c.useCase.GetCurrentRounds()) > 0 {
		lastRound := c.useCase.GetCurrentRounds()[len(c.useCase.GetCurrentRounds())-1]
		fmt.Printf("Round moves: %s vs %s\n", lastRound.Move1, lastRound.Move2)
		if lastRound.Winner != "" {
			fmt.Printf("Round winner: %s\n", lastRound.Winner)
		}
	}

	if result.Winner != "" {
		if result.Winner == "Draw" {
			fmt.Println("The game is a draw!")
		} else {
			fmt.Printf("Game winner: %s\n", result.Winner)
		}
	}
}
