package lookup_test

import (
	"context"
	"errors"
	"os"
	"strings"
	"testing"

	"github.com/wavemechanics/deliver/lookup"
	"github.com/wavemechanics/deliver/store/mem"
)

func TestNotFoundNoDefault(t *testing.T) {
	var s mem.Storage

	_, created, err := lookup.Lookup(context.TODO(), &s, "missing")
	if err == nil || created {
		t.Fatal("did not expect Lookup to succeed")
	}
	if !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("expected Not Found, got %v", err)
	}
}

func TestNotFoundWithDefault(t *testing.T) {
	var s mem.Storage

	ctx := context.TODO()
	s.Set(ctx, "default", "default value")

	contents, created, err := lookup.Lookup(ctx, &s, "missing")
	if err != nil {
		t.Fatalf("Lookup: %v, want nil", err)
	}
	if !created {
		t.Fatal(`Lookup: "missing": should have been created`)
	}
	if !strings.Contains(contents, "default value") {
		t.Fatalf("Lookup contents: %q, want %q", contents, "default value")
	}
}

func TestFound(t *testing.T) {
	var s mem.Storage

	ctx := context.TODO()
	s.Set(ctx, "localpart", "some value")

	contents, created, err := lookup.Lookup(ctx, &s, "localpart")
	if err != nil {
		t.Fatalf("Lookup: %v, want nil", err)
	}
	if created {
		t.Fatal(`Lookup: "localpart": should not have returned created true`)
	}
	if contents != "some value" {
		t.Fatalf("Lookup contents: %q, want %q", contents, "some value")
	}
}
