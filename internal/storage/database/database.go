package database

import (
	"database/sql"
	"errors"
	"fmt"
	"url-shortener/internal/storage"
	"github.com/jmoiron/sqlx"
)

type Storage struct {
	db *sqlx.DB
}

func NewStorage(db *sqlx.DB) *Storage {
	return &Storage{db: db}
}

type Results struct {
	URL   string `db:"url" json:"url"`
	Alias string `db:"alias" json:"alias"`
}

func (s *Storage) SaveURL(urlToSave, alias string) (int64, error) {
	const op = "storage.database.SaveURL"
	query := `INSERT INTO url(url, alias) VALUES($1, $2) RETURNING id`

	var id int64
	err := s.db.QueryRow(query, urlToSave, alias).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	return id, nil
}


func (s *Storage) GetURL(alias string) (string, error) {
	const op = "storage.database.database"
	query := `SELECT url FROM url WHERE alias = $1`
	var result string

	err := s.db.QueryRow(query, alias).Scan(&result)

	if errors.Is(err, sql.ErrNoRows) {
		return "", storage.ErrURLNotFound
	}

	if err != nil {
		return "", fmt.Errorf("%s: %w", op, err)
	}

	return result, nil
}

func (s *Storage) GetAll() ([]Results, error) {
	const op = "storage.database.database"
	res := make([]Results, 0)
	query := `SELECT url, alias FROM url`

	err := s.db.Select(&res, query)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	return res, nil
}

func (s *Storage) Delete(alias string) error {
	const op = "storage.database.database"
	query := `DELETE FROM url WHERE alias = $1`

	result, err := s.db.Exec(query, alias)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	row, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	if row == 0 {
		return fmt.Errorf("%s: %w", op, storage.ErrURLNotFound)
	}
	return nil
}