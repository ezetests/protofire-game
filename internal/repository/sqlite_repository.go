package repository

import (
	"database/sql"
	"fmt"

	"protofire-game/internal/domain"

	_ "github.com/mattn/go-sqlite3"
)

type SQLiteRepository struct {
	db *sql.DB
}

func NewSQLiteRepository(dbPath string) (*SQLiteRepository, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("error opening database: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("error connecting to database: %w", err)
	}

	if err := createTables(db); err != nil {
		return nil, fmt.Errorf("error creating tables: %w", err)
	}

	return &SQLiteRepository{db: db}, nil
}

func createTables(db *sql.DB) error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS game_results (
			id TEXT PRIMARY KEY,
			player1 TEXT NOT NULL,
			player2 TEXT NOT NULL,
			winner TEXT NOT NULL,
			played_at DATETIME NOT NULL
		)`,
	}

	for _, query := range queries {
		if _, err := db.Exec(query); err != nil {
			return err
		}
	}

	return nil
}

func (r *SQLiteRepository) SaveGame(result *domain.Game) error {
	query := `
	INSERT INTO game_results (id, player1, player2, winner, played_at)
	VALUES (?, ?, ?, ?, ?)`

	_, err := r.db.Exec(query,
		result.ID,
		result.Player1,
		result.Player2,
		result.Winner,
		result.PlayedAt,
	)

	if err != nil {
		return fmt.Errorf("error saving game: %w", err)
	}

	return nil
}

func (r *SQLiteRepository) GetGameHistory() ([]*domain.Game, error) {
	query := `
	SELECT id, player1, player2, winner, played_at
	FROM game_results
	ORDER BY played_at DESC`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("error querying game history: %w", err)
	}
	defer rows.Close()

	var results []*domain.Game

	for rows.Next() {
		var result domain.Game
		var playedAt string

		err := rows.Scan(
			&result.ID,
			&result.Player1,
			&result.Player2,
			&result.Winner,
			&playedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("error scanning row: %w", err)
		}

		result.PlayedAt = playedAt
		results = append(results, &result)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return results, nil
}

func (r *SQLiteRepository) Close() error {
	return r.db.Close()
}
