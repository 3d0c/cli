package cli

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"sort"
	"text/tabwriter"
)

type Command interface {
	Register(f *flag.FlagSet)
	Process() error
	Run(f *flag.FlagSet) ([]byte, error)
}

func generalHelp() {
	fmt.Printf("Usage: %s <COMMAND>\nAvailable commands:\n", os.Args[0])

	sorted := []string{}

	for name, _ := range commands {
		sorted = append(sorted, name)
	}

	sort.Strings(sorted)

	for _, name := range sorted {
		fmt.Printf("  %s\n", name)
	}
}

func commandHelp(name string, cmd Command, f *flag.FlagSet, err error) {
	type HasUsage interface {
		Usage() string
	}

	if err != nil && err != flag.ErrHelp {
		fmt.Printf("\n%s\n", err)
	}

	fmt.Printf("Usage: %s %s [OPTIONS]", os.Args[0], name)

	if u, ok := cmd.(HasUsage); ok {
		fmt.Printf(" %s\n", u.Usage())
	}

	type HasDescription interface {
		Description() string
	}

	if u, ok := cmd.(HasDescription); ok {
		fmt.Printf("\n%s\n", u.Description())
	}

	n := 0
	f.VisitAll(func(_ *flag.Flag) {
		n += 1
	})

	if n > 0 {
		fmt.Printf("\nOptions:\n")
		tw := tabwriter.NewWriter(os.Stdout, 2, 0, 2, ' ', 0)
		f.VisitAll(func(f *flag.Flag) {
			fmt.Fprintf(tw, "\t-%s=%s\t%s\n", f.Name, f.DefValue, f.Usage)
		})
		tw.Flush()
	}
}

func Run(args []string) int {
	if len(args) == 0 {
		generalHelp()
		return 1
	}

	if args[0] == "-h" || args[0] == "--help" {
		generalHelp()
		return 0
	}

	cmd, ok := commands[args[0]]
	if !ok {
		generalHelp()
		return 1
	}

	fs := flag.NewFlagSet("", flag.ContinueOnError)
	fs.SetOutput(ioutil.Discard)

	cmd.Register(fs)

	var (
		err    error
		output []byte
		silent bool = true
		retval int  = 0
	)

	defer func() {
		if err != nil && err == flag.ErrHelp {
			commandHelp(args[0], cmd, fs, err)
			return
		}
	}()

	if err = fs.Parse(args[1:]); err != nil {
		if err != flag.ErrHelp {
			fmt.Println(err)
		}

		return 1
	}

	if silentFlag := fs.Lookup("silent"); silentFlag != nil {
		if silentFlag.Value.String() == "false" {
			silent = false
		}
	}

	if err = cmd.Process(); err != nil {
		fmt.Println(err)
		return 1
	}

	if output, err = cmd.Run(fs); err != nil {
		retval = 1
	}

	if !silent {
		fmt.Print(string(output))
	}
	if err != nil {
		fmt.Println("!", err)
	}

	return retval
}
