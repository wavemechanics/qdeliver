package app

import (
	"context"
	"errors"
	"flag"
	"log"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/wavemechanics/qdeliver/deliver"
	"github.com/wavemechanics/qdeliver/lookup"
	"github.com/wavemechanics/qdeliver/notify"
	"github.com/wavemechanics/qdeliver/store/webdav"
	"github.com/wavemechanics/qdeliver/users"
)

// Run is a more testable main
//
func Run(args []string) int {
	var dbpath string
	var handler string
	var notifyscript string

	flags := flag.NewFlagSet("main", flag.ContinueOnError)
	flags.StringVar(&dbpath, "db", "users.json", "path to user database")
	flags.StringVar(&handler, "handler", "./qdeliver-handler.sh", "delivery handler script")
	flags.StringVar(&notifyscript, "notify", "./qdeliver-notify.sh", "new address notification script")

	u := usage{
		Flags: flags,
	}
	flags.Usage = u.Usage

	if err := flags.Parse(args); err != nil {
		log.Println(err)
		return 2
	}
	if flags.NArg() != 2 {
		flags.Usage()
		return 2
	}
	if handler == "" {
		flags.Usage()
		return 2
	}

	localpart := strings.ToLower(flags.Arg(0))
	domain := flags.Arg(1)
	owner := strings.Split(localpart, "-")[0]
	if owner == "" || localpart == "" || domain == "" {
		flags.Usage()
		return 2
	}

	db, err := users.Load(dbpath)
	if err != nil {
		log.Println(err)
		return 1
	}

	account, err := db.Lookup(owner, domain)
	if err != nil {
		log.Println(err)
		if errors.Is(err, os.ErrNotExist) {
			return 100 // permanent; owner/domain not in userdb
		}
		return 1
	}

	storage, err := webdav.New(account.URL, account.Login, account.Password)
	if err != nil {
		log.Println(err)
		return 1
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	instructions, created, err := lookup.Lookup(ctx, storage, localpart)
	if errors.Is(err, os.ErrNotExist) {
		log.Printf("%s@%s: localpart not found, and no default\n", localpart, domain)
		return 100 // permanent; address file doesn't exist
	}
	if err != nil {
		log.Println(err)
		return 1
	}

	var wg sync.WaitGroup
	var status int
	wg.Add(1)
	go func() {
		status = deliver.Deliver(ctx, &wg, handler, instructions)
	}()
	if created && account.Notify {
		wg.Add(1)
		go notify.Notify(ctx, &wg, notifyscript, owner+"@"+domain, localpart+"@"+domain)
	}
	wg.Wait()

	return status
}
