package notify_test

import (
	"context"
	"io/ioutil"
	"path/filepath"
	"fmt"
	"testing"
	"sync"
	"os"

	"github.com/wavemechanics/deliver/notify"
)

func TestNotify(t *testing.T) {

	var wg sync.WaitGroup
	ctx := context.TODO()

	dir, err := ioutil.TempDir("", "NotifyTest")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	recipient := "recipient@example.com"
	newaddress := "newaddress@example.com"

	os.Setenv("TESTDIR", dir)
	wg.Add(1)
	go notify.Notify(ctx, &wg, "testdata/notify.sh", recipient, newaddress)
	wg.Wait()

	results, err := ioutil.ReadFile(filepath.Join(dir, "notify.out"))
	if err != nil {
		t.Fatal(err)
	}

	expected := fmt.Sprintf("recipient: %s\nnewaddress: %s\n", recipient, newaddress)

	if string(results) != expected {
		t.Fatalf("expected %v, got %v", expected, results)
	}
}