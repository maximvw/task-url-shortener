package memory

import (
	"fmt"
	"golang_project/internal/storage"
)

type Storage struct {
	memory map[string]string
}

func New() (*Storage, error) {
	// const op = "storage.memory.New"

	memory := make(map[string]string)

	return &Storage{memory: memory}, nil
}

func (s *Storage) SaveURL(urlToSave string, alias string) error {
	const op = "storage.memory.SaveURL"

	for _, alias_value := range s.memory {
		if alias_value == alias {
			return fmt.Errorf("%s: %w", op, storage.ErrAliasExists)
		}
	}

	if _, ok := s.memory[urlToSave]; ok {
		return fmt.Errorf("%s: %w", op, storage.ErrURLExists)
	}

	s.memory[urlToSave] = alias
	return nil
}

func (s *Storage) GetURL(alias string) (string, error) {
	const op = "storage.memory.GetURL"

	var resURL string

	for url_key, alias_value := range s.memory {
		if alias_value == alias {
			resURL = url_key
			break
		}
	}
	if resURL == "" {
		return "", fmt.Errorf("%s: %w", op, storage.ErrURLNotFound)
	}

	return resURL, nil
}
