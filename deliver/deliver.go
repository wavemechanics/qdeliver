package deliver

import (
	"context"
	"log"
	"os"
	"os/exec"
	"strings"
	"sync"

	"github.com/wavemechanics/qdeliver/token"
)

// Deliver runs delivery instructions in an address file.
//
func Deliver(ctx context.Context, wg *sync.WaitGroup, handler, instructions string) int {
	defer wg.Done()

	for _, line := range token.SplitFile(instructions) {
		line = strings.TrimSpace(line)
		if len(line) == 0 || strings.HasPrefix(line, "#") {
			continue
		}
		status := run(ctx, handler, line)
		if status == 99 {
			return 0
		}
		if status != 0 {
			return status
		}
	}
	return 0
}

func run(ctx context.Context, handler, line string) int {
	tokens, err := token.SplitLine(line)
	if err != nil {
		log.Printf("instruction line: %v", err)
		return 1
	}
	if len(tokens) == 0 {
		return 0
	}
	if _, err := os.Stdin.Seek(0, 0); err != nil {
		log.Printf("rewind: %v", err)
		return 1
	}
	cmd := exec.CommandContext(ctx, handler, tokens...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	if err == nil {
		return 0
	}
	err1, ok := err.(*exec.ExitError)
	if !ok {
		log.Println(err)
		return 1
	}
	if err1.ExitCode() == -1 {
		log.Println(err)
		return 1 // signal
	}
	return err1.ExitCode()
}
