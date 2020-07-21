package store

import (
	"context"

	"github.com/wavemechanics/etype"
)

const ErrEmptyKey = etype.Sentinel("empty key not allowed")

type Storage interface {
	Get(ctx context.Context, key string) (string, error)
	Set(ctx context.Context, key, value string) error
}
