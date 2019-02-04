package cli

import (
	"flag"
)

type General struct {
	Dryrun bool
	Silent bool
}

func (flag *General) Register(f *flag.FlagSet) {
	f.BoolVar(&flag.Dryrun, "n", true, "dry run, see specific command description")
	f.BoolVar(&flag.Silent, "silent", false, "suppress output")
}
