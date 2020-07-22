package app_test

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/wavemechanics/qdeliver/app"
	"github.com/wavemechanics/qdeliver/internal/webdavd"
	"github.com/wavemechanics/qdeliver/users"
)

// TestFlags tests command line parsing problems
func TestFlags(t *testing.T) {
	var tests = []struct {
		args []string
		exit int
	}{
		{[]string{}, 2},
		{[]string{"local"}, 2},
		{[]string{"-z"}, 2},
		{[]string{"", ""}, 2},
	}

	for _, test := range tests {
		exit := app.Run(test.args)
		fmt.Fprintf(os.Stderr, "exit=%d\n", exit)
		if exit != test.exit {
			t.Errorf("%v: exit %d, want %d", test.args, exit, test.exit)
		}
	}
}

// TestUsers tests user lookup errors
func TestUsers(t *testing.T) {
	var tests = []struct {
		db        string
		localpart string
		domain    string
		exit      int
	}{
		{"testdata/missing.json", "owner", "domain", 1},
		{"testdata/bad-json.json", "owner", "domain", 1},
		{"testdata/users.json", "wrongowner", "domain", 100},
	}

	for _, test := range tests {
		args := []string{
			"--db", test.db,
			"--handler", "testdata/dummy-handler.sh",
			test.localpart, test.domain,
		}
		exit := app.Run(args)
		if exit != test.exit {
			t.Errorf("%v: exit %d, want %d", test, exit, test.exit)
		}
	}
}

// TestWebdav tests webdav errors
func TestWebdav(t *testing.T) {
	var tests = []struct {
		owner string
	}{
		{"userwithbadurl"},
		{"userwithdownserver"},
		{"userwithwrongpass"},
	}

	domain := "example.com"

	dir, err := ioutil.TempDir("", "TestWebdav")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	// create instruction that will pass so we know if these tests succeeded unexpectedly
	err = ioutil.WriteFile(filepath.Join(dir, "userwithwrongpass"), []byte("true\n"), 0644)
	if err != nil {
		t.Fatal(err)
	}

	server := webdavd.Server{
		Dir:  dir,
		User: "hello",
		Pass: "letmein",
	}

	shutdown := server.Start()
	defer shutdown()

	udata := &users.Users{
		Version: 1,
		Accounts: []users.Account{
			{
				// for bad url test
				Owner:    "userwithbadurl",
				Domain:   domain,
				URL:      "notaurl",
				Login:    server.User,
				Password: server.Pass,
			},
			{
				// for problem connecting
				Owner:    "userwithdownserver",
				Domain:   domain,
				URL:      "http://127.0.0.1:1", // well, _probably_ will reset
				Login:    server.User,
				Password: server.Pass,
			},
			{
				// test for bad user/pass
				Owner:    "userwithwrongpass",
				Domain:   domain,
				URL:      server.Addr,
				Login:    server.User,
				Password: "not" + server.Pass,
			},
		},
	}

	path := filepath.Join(dir, "users.json")

	if err = udata.Save(path); err != nil {
		t.Fatal(err)
	}

	for _, test := range tests {
		args := []string{
			"--db", path,
			"--handler", "testdata/dummy-handler.sh",
			test.owner, domain,
		}
		exit := app.Run(args)
		if exit != 1 {
			t.Errorf("%v: exit %d, want 1", test, exit)
		}
	}
}

func TestInstructions(t *testing.T) {
	owner := "owner"
	domain := "example.com"

	var tests = []struct {
		address      string
		instructions string
		exit         int
	}{
		{owner + `-missing`, ``, 100},
		{owner + `-1`, `sh -c "exit 1"`, 1},
		{owner + `-99`, `sh -c "exit 99"`, 0},
		{owner + `-100`, `sh -c "exit 100"`, 100},
		{owner + `-111`, `sh -c "exit 111"`, 111},
		{owner + `-0`, `sh -c "exit 0"`, 0},
		{owner + `-sleep`, `sh -c "sleep 15"`, 1},
	}

	dir, err := ioutil.TempDir("", "TestInstructions")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	server := webdavd.Server{
		Dir:  dir,
		User: "hello",
		Pass: "letmein",
	}

	shutdown := server.Start()
	defer shutdown()

	udata := &users.Users{
		Version: 1,
		Accounts: []users.Account{
			{
				Owner:    owner,
				Domain:   domain,
				URL:      server.Addr,
				Login:    server.User,
				Password: server.Pass,
				Notify:   true,
			},
		},
	}

	dbpath := filepath.Join(dir, "users.json")

	if err = udata.Save(dbpath); err != nil {
		t.Fatal(err)
	}

	for _, test := range tests {
		if test.instructions != "" {
			err = ioutil.WriteFile(filepath.Join(dir, test.address)+".txt", []byte(test.instructions), 0644)
			if err != nil {
				t.Errorf("%s: can't create delivery instructions", test.address)
				continue
			}
		}

		args := []string{
			"--db", dbpath,
			"--handler", "testdata/handler.sh",
			test.address, domain,
		}

		exit := app.Run(args)
		if exit != test.exit {
			t.Errorf("%s: test.address: exit %d, want %d", test.address, exit, test.exit)
		}
	}

	// Test that missing file created if default exists
	//
	err = ioutil.WriteFile(filepath.Join(dir, "default.txt"), []byte(`sh -c "exit 0"`), 0644)
	if err != nil {
		t.Fatal(err)
	}

	args := []string{
		"--db", dbpath,
		"--handler", "testdata/handler.sh",
		"--notify", "testdata/notify.sh",
		owner + "-missing", domain,
	}

	os.Setenv("TESTDIR", dir) // for notify.sh
	exit := app.Run(args)
	if exit != 0 {
		t.Fatalf("%s-missing: exit %d, want 0", owner, exit)
	}

	// verify missing file was created
	//
	_, err = ioutil.ReadFile(filepath.Join(dir, owner+"-missing") + ".txt")
	if err != nil {
		t.Fatalf("%s-missing: expected instructions file to be created", owner)
	}

	// verify notify.sh ran
	//
	notifyMsg, err := ioutil.ReadFile(filepath.Join(dir, "notify.out"))
	if err != nil {
		t.Fatal(err)
	}
	want := fmt.Sprintf("recipient: %s@%s\nnewaddress: %s-missing@%s\n", owner, domain, owner, domain)
	if string(notifyMsg) != want {
		t.Fatalf("notify output: %q, want %q", notifyMsg, want)
	}
}
