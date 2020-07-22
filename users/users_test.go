package users_test

import (
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/wavemechanics/qdeliver/users"
)

func TestLoad(t *testing.T) {
	var tests = []struct {
		path string
		ok   bool // would test for actual errors, but json errors don't like .Is
	}{
		{"noexist", false},
		{"bad-json.json", false},
		{"users.json", true},
	}

	for _, test := range tests {
		account, err := users.Load(filepath.Join("testdata", test.path))
		if test.ok {
			if err != nil {
				t.Errorf("%s: %v", test.path, err)
				continue
			}
		} else {
			if err == nil {
				t.Errorf("%s: expected error", test.path)
				continue
			}
		}
		if err != nil {
			continue
		}
		if account == nil {
			t.Errorf("%s: expected account != nil", test.path)
		}
	}
}

func TestLookup(t *testing.T) {
	var tests = []struct {
		owner  string
		domain string
		err    error
		login  string
	}{
		{"baz", "example.com", os.ErrNotExist, ""},
		{"foo", "wrong.com", os.ErrNotExist, ""},
		{"foo", "example.com", nil, "joe"},
	}

	u, err := users.Load(filepath.Join("testdata", "users.json"))
	if err != nil {
		t.Fatalf("cannot load test users file: %v", err)
	}
	for _, test := range tests {
		account, err := u.Lookup(test.owner, test.domain)
		if !errors.Is(err, test.err) {
			t.Errorf("%q, %q: %v, want %v", test.owner, test.domain, err, test.err)
			continue
		}
		if err != nil {
			continue
		}
		if account.Login != test.login {
			t.Errorf("%q, %q: %v, want %v", test.owner, test.domain, account.Login, test.login)
			continue
		}
	}
}

func TestSave(t *testing.T) {
	var err error

	udata := &users.Users{}
	if err = udata.Save("/noexist"); err == nil {
		t.Fatalf("writing to /noexist should have failed")
	}

	udata = &users.Users{
		Version: 1,
		Accounts: []users.Account{
			{
				Owner:    "owner",
				Domain:   "domain",
				URL:      "url",
				Login:    "login",
				Password: "password",
				Notify:   true,
			},
			{
				Owner:    "owner2",
				Domain:   "domain2",
				URL:      "url2",
				Login:    "login2",
				Password: "password2",
				Notify:   true,
			},
		},
	}

	dir, err := ioutil.TempDir("", "TestSave")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	path := filepath.Join(dir, "users.json")

	if err = udata.Save(path); err != nil {
		t.Fatal(err)
	}

	got, err := users.Load(path)
	if err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(udata, got) {
		t.Fatalf("%v, want %v", udata, got)
	}
}
