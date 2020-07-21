package notify

import (
	"context"
	"log"
	"net/mail"
	"os"
	"os/exec"
	"sync"
)

func Notify(ctx context.Context, wg *sync.WaitGroup, script, recipient, newaddress string) {
	defer wg.Done()

	e, err := mail.ParseAddress(recipient)
	if err != nil {
		log.Println(err)
		return
	}

	cmd := exec.CommandContext(ctx, script, e.Address, newaddress)
	cmd.Stdin = nil
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		log.Println(err)
	}
}
