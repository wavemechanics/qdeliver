package app

import (
	"flag"
	"fmt"
	"os"
)

// This struct and Usage method is only to create a Usage method with the
// right signature to be used as flag.Usage, but being able to use
// PrintDefaults.
//
type usage struct {
	Flags *flag.FlagSet
}

func (u *usage) Usage() {
	fmt.Fprintf(u.Flags.Output(), "usage: %s [options] localpart domain\n", os.Args[0])
	u.Flags.PrintDefaults()
}
