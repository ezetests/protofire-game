package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/joho/godotenv"

	"protofire-game/internal/delivery/cli"
	"protofire-game/internal/domain"
	service "protofire-game/internal/randomness"
	"protofire-game/internal/repository"
	"protofire-game/internal/usecase"
)

var dataDir string

func initSQLiteRepository() (domain.GameRepository, error) {
	if dataDir == "" {
		return nil, fmt.Errorf("DATA_DIR is not set")
	}

	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create data directory: %v", err)
	}

	dbPath := filepath.Join(dataDir, "protofire-game.db")
	return repository.NewSQLiteRepository(dbPath)
}

func initOnChainRepository() (domain.GameRepository, error) {
	return repository.NewOnChainRepository()
}

func main() {

	if err := godotenv.Load(); err != nil {
		log.Printf("Warning: Error loading .env file: %v", err)
	}

	fmt.Println("\nSelect Storage Type:")
	fmt.Println("1. SQLite")
	fmt.Println("2. On-Chain")
	fmt.Print("Choose storage type: ")

	var choice string
	fmt.Scanln(&choice)

	var repo domain.GameRepository
	var err error

	switch choice {
	case "1":
		fmt.Println("Using SQLite storage")
		repo, err = initSQLiteRepository()
	case "2":
		fmt.Println("Using On-Chain storage")
		repo, err = initOnChainRepository()
	default:
		fmt.Println("Invalid choice. Using default SQLite storage")
		repo, err = initSQLiteRepository()
	}

	if err != nil {
		log.Fatalf("Failed to initialize repository: %v", err)
	}

	defer func() {
		if closer, ok := repo.(interface{ Close() error }); ok {
			if err := closer.Close(); err != nil {
				fmt.Printf("Error closing repository: %v\n", err)
			}
		}
	}()

	randGen := service.NewDefaultRandomGenerator()
	gameUseCase := usecase.NewGameUseCase(repo, randGen)

	gameCLI := cli.NewGameCLI(gameUseCase)
	gameCLI.Start()
}
