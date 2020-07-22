package webdav_test

import (
	"context"
	"errors"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/wavemechanics/qdeliver/lookup"
	"github.com/wavemechanics/qdeliver/store"
	"github.com/wavemechanics/qdeliver/store/mem"
	"github.com/wavemechanics/qdeliver/store/webdav"
)

func makeHandler(db store.Storage) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {

		if len(r.URL.Query()) != 0 {
			http.Error(w, "query parameters not allowed", http.StatusBadRequest)
			return
		}

		uri := strings.TrimPrefix(r.URL.RequestURI(), "/")
		parts := strings.Split(uri, "/")
		if len(parts) != 2 {
			http.Error(w, "only dir/key allowed", http.StatusBadRequest)
			return
		}
		if parts[0] != "dg" {
			http.Error(w, "only /dg/ URI allowed", http.StatusBadRequest)
			return
		}

		ctx := context.TODO()
		if r.Method == http.MethodGet {
			val, err := db.Get(ctx, parts[1]+".txt")
			if errors.Is(err, os.ErrNotExist) {
				http.Error(w, "key not found", http.StatusNotFound)
				return
			}
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			io.WriteString(w, val)
			return
		} else if r.Method == http.MethodPut {
			buf, err := ioutil.ReadAll(r.Body)
			if err != nil {
				http.Error(w, "problem reading body", http.StatusInternalServerError)
				return
			}
			db.Set(ctx, parts[1]+".txt", string(buf))
			http.Error(w, "created", http.StatusCreated)
		} else {
			http.Error(w, "Only GET and PUT allowed", http.StatusBadRequest)
			return
		}
	}
}

func TestEmptyKey(t *testing.T) {
	var db mem.Storage

	ts := httptest.NewServer(makeHandler(&db))
	defer ts.Close()

	s, err := webdav.New(ts.URL+"/dg", "", "")
	if err != nil {
		t.Fatal(err)
	}

	ctx := context.TODO()

	if err = s.Set(ctx, "", "value"); err != store.ErrEmptyKey {
		t.Fatal("expected Set empty key to return error")
	}

	if _, err = s.Get(ctx, ""); err != store.ErrEmptyKey {
		t.Fatal("expected Gt empty key to return error")
	}
}

func TestNotFoundNoDefault(t *testing.T) {
	var db mem.Storage

	ts := httptest.NewServer(makeHandler(&db))
	defer ts.Close()

	s, err := webdav.New(ts.URL+"/dg", "", "")
	if err != nil {
		t.Fatal(err)
	}

	_, created, err := lookup.Lookup(context.TODO(), s, "missing")
	if err == nil || created {
		t.Fatal("did not expect Lookup to succeed")
	}
	if !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("expected NotFound, got %v", err)
	}
}

func TestNotFoundWithDefault(t *testing.T) {
	var db mem.Storage
	ctx := context.TODO()

	ts := httptest.NewServer(makeHandler(&db))
	defer ts.Close()

	s, err := webdav.New(ts.URL+"/dg", "", "")
	if err != nil {
		t.Fatal(err)
	}

	if err = s.Set(ctx, "default", "default value"); err != nil {
		t.Fatal("could not set up default contents")
	}

	contents, created, err := lookup.Lookup(ctx, s, "missing")
	if err != nil {
		t.Fatalf("Lookup: %v", err)
	}
	if !created {
		t.Fatalf(`Lookup: expected to create "missing"`)
	}
	if !strings.Contains(contents, "default value") {
		t.Fatalf("Lookup contents: %q, want %q", contents, "default value")
	}
}

func TestFound(t *testing.T) {
	var db mem.Storage
	ctx := context.TODO()

	ts := httptest.NewServer(makeHandler(&db))
	defer ts.Close()

	s, err := webdav.New(ts.URL+"/dg", "", "")
	if err != nil {
		t.Fatal(err)
	}

	if err = s.Set(ctx, "localpart", "some value"); err != nil {
		t.Fatal("could not set up localpart contents")
	}

	contents, created, err := lookup.Lookup(ctx, s, "localpart")
	if err != nil {
		t.Fatalf("Lookup contents: %v", err)
	}
	if created {
		t.Fatal(`Lookup: "localpart": expected created to be false`)
	}
	if contents != "some value" {
		t.Fatalf("Looking contents: %q, want %q", contents, "some value")
	}
}
