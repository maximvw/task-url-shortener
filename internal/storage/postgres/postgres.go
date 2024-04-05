package postgres

import (
	"database/sql"
	"errors"
	"fmt"
	"golang_project/internal/config"
	"golang_project/internal/storage"

	"github.com/lib/pq"
	// "golang_project/internal/storage/postgres"
	// _ "golang_project/internal/storage"
)

type Storage struct {
	db *sql.DB
}

func New(pg_cfg *config.PostgresCfg) (*Storage, error) {
	const op = "storage.postgres.New"

	psqlconn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		pg_cfg.Host, pg_cfg.Port, pg_cfg.User, pg_cfg.Password, pg_cfg.Database)

	db, err := sql.Open("postgres", psqlconn)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	createTable := `
	CREATE TABLE IF NOT EXISTS url(
		alias TEXT NOT NULL PRIMARY KEY,
		url TEXT NOT NULL UNIQUE);
	`
	err = applyToDataBase(db, createTable, op+".Create")
	if err != nil {
		return nil, err
	}

	createIdx := `
	CREATE INDEX IF NOT EXISTS idx_alias ON url(alias);
	`

	err = applyToDataBase(db, createIdx, op+".Index")
	if err != nil {
		return nil, err
	}

	return &Storage{db: db}, nil
}

func applyToDataBase(db *sql.DB, strQuery string, op string) error {
	stmt, err := db.Prepare(strQuery)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	_, err = stmt.Exec()
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}

func (s *Storage) SaveURL(urlToSave string, alias string) error {
	const op = "storage.postgres.SaveURL"

	stmt, err := s.db.Prepare("SELECT url FROM url WHERE alias = $1")
	if err != nil {
		return fmt.Errorf("%s: prepare statement: %w", op, err)
	}
	var resURL string

	err = stmt.QueryRow(alias).Scan(&resURL)
	if !errors.Is(err, sql.ErrNoRows) {
		return fmt.Errorf("%s: %w", op, storage.ErrAliasExists)
	}

	stmt, err = s.db.Prepare(`insert into "url"("url", "alias") values($1, $2)`)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	_, err = stmt.Exec(urlToSave, alias)
	if err != nil {
		if pgerr, ok := err.(*pq.Error); ok && pgerr.Code == "23505" {
			return fmt.Errorf("%s: %w", op, storage.ErrURLExists)
		}

		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (s *Storage) GetURL(alias string) (string, error) {
	const op = "storage.postgres.GetURL"

	stmt, err := s.db.Prepare("SELECT url FROM url WHERE alias = $1")
	if err != nil {
		return "", fmt.Errorf("%s: prepare statement: %w", op, err)
	}

	var resURL string

	err = stmt.QueryRow(alias).Scan(&resURL)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", fmt.Errorf("%s: %w", op, storage.ErrURLNotFound)
		}
		return "", fmt.Errorf("%s: execute statement: %w", op, err)
	}

	return resURL, nil
}
