package lookup

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/wavemechanics/deliver/store"
)

// Lookup returns the delivery instructions for localpart in storage s.
// If localpart doesn't exist, it will be created if default instructions exist.
// created will be true if a new key for localpart was created.
//
func Lookup(ctx context.Context, s store.Storage, localpart string) (instruction string, created bool, err error) {
	contents, err := s.Get(ctx, localpart)
	if err == nil {
		return contents, false, nil
	}

	if !errors.Is(err, os.ErrNotExist) {
		return "", false, err
	}

	contents, err = s.Get(ctx, "default")
	if err != nil {
		return "", false, err
	}

	sender := os.Getenv("SENDER")
	timestamp := time.Now().UTC().Format(time.RFC3339Nano)
	contents += fmt.Sprintf("\n# Sender: %s\n# Timestamp: %s\n", sender, timestamp)

	err = s.Set(ctx, localpart, contents)
	if err != nil {
		return "", false, err
	}

	return contents, true, nil
}
