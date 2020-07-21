package deliver

import (
	"context"
	"sync"
	"testing"
	"time"
)

func TestRun(t *testing.T) {
	var tests = []struct {
		line    string
		timeout int
		status  int
	}{
		{`"`, 0, 1}, // deliberate syntax error
		{"", 0, 0},
		{"/noexist", 0, -1}, // -1 means any non-zero; shells are different
		{"false", 0, 1},
		{"true", 0, 0},
		{"./exit99.sh", 0, 99},
		{"./signal.sh", 0, 1},
		{"sh -c true", 0, 0},
		{"sh -c false", 0, 1},
		{"sh -c ./exit99.sh", 0, 99},
		{"./sleep.sh", 2, -1}, // -1 means any non-zero; shells are different
		{`sh -c "exit 0"`, 0, 0},
		{`sh -c "exit 1"`, 0, 1},
		{`sh -c "exit 99"`, 0, 99},
	}

	for _, test := range tests {
		if test.timeout == 0 {
			test.timeout = 100 // anything longer than the usual test
		}
		ctx, cancel := context.WithTimeout(context.Background(), time.Duration(test.timeout)*time.Second)

		status := run(ctx, "testdata/deliver.sh", test.line)
		if test.status == -1 {
			if status == 0 {
				t.Errorf("%q: exit 0, wanted non-zero", test.line)
			}
		} else if status != test.status {
			t.Errorf("%q: %d, want %d", test.line, status, test.status)
		}
		cancel()
	}
}

func TestDeliver(t *testing.T) {
	var tests = []struct {
		instructions string
		timeout      int
		status       int
	}{
		{"", 0, 0},
		{"\n", 0, 0},
		{"#", 0, 0},
		{"#\n", 0, 0},
		{"/noexist", 0, -1}, // -1 means any non-zero; shells are different
		{"true", 0, 0},
		{"false", 0, 1},
		{"./testdata/exit99.sh", 0, 0},
		{"./testdata/signal.sh", 0, 1},
		{"./testdata/exit99.sh\nfalse", 0, 0}, // test short circuit 99 success
		{"false\ntrue", 0, 1},                 // test stops on error
		{"true\nfalse", 0, 1},                 // test executes each line
		{"true\n#\nfalse\n", 0, 1},            // test skips comments
		{"true\n\nfalse\n", 0, 1},             // skips blank lines
		{"./testdata/sleep.sh", 2, -1},        // -1 means any non-zero; shells are different
	}

	for _, test := range tests {
		if test.timeout == 0 {
			test.timeout = 100 // anything longer than the usual test
		}
		ctx, cancel := context.WithTimeout(context.Background(), time.Duration(test.timeout)*time.Second)

		var wg sync.WaitGroup
		wg.Add(1)
		status := Deliver(ctx, &wg, "testdata/deliver.sh", test.instructions)
		wg.Wait()

		if test.status == -1 {
			if status == 0 {
				t.Errorf("%q: exit 0, wanted non-zero", test.instructions)
			}
		} else if status != test.status {
			t.Errorf("%q: %d, want %d", test.instructions, status, test.status)
		}
		cancel()
	}
}
