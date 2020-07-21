package mem

import (
	"context"
	"os"
)

type Storage struct {
	m map[string]string
}

func (s *Storage) Get(ctx context.Context, key string) (string, error) {
	if s.m == nil {
		s.m = make(map[string]string)
	}
	val, ok := s.m[key]
	if !ok {
		return "", os.ErrNotExist
	}
	return val, nil
}

func (s *Storage) Set(ctx context.Context, key, value string) error {
	if s.m == nil {
		s.m = make(map[string]string)
	}
	s.m[key] = value
	return nil
}
