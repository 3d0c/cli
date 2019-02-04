package command

import (
	"flag"
	"fmt"
	"go/build"
	"go/parser"
	"go/printer"
	"go/token"
	"log"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"golang.org/x/tools/go/ast/astutil"

	"github.com/3d0c/cli/pkg"
)

type command struct {
	*cli.General

	Name       string
	Subcommand string
	force      bool
	fullname   string
	fullpath   string
	gopath     string
	maingo     string
	pkgname    string
}

var commandTemplate = template.Must(template.New("").Parse(`// Code generated by go generate;
package {{.Name}}

import (
	"flag"

	"github.com/3d0c/cli/pkg"
)

type {{.Subcommand}} struct {
	*cli.General
}

func init() {
	cli.Register("{{.Name}}.{{.Subcommand}}", &{{.Subcommand}}{General: &cli.General{}})
}

func (cmd *{{.Subcommand}}) Register(f *flag.FlagSet) {
	cmd.General.Register(f)
}

func (cmd *{{.Subcommand}}) Description() string {
	return ""
}

func (cmd *{{.Subcommand}}) Process() error {
	return nil
}

func (cmd *{{.Subcommand}}) Run(f *flag.FlagSet) ([]byte, error) {
	return nil, nil
}
`))

var defaultMainGo = template.Must(template.New("").Parse(`// Code generated by go generate;
package main

import (
	"os"

	"github.com/3d0c/cli/pkg"
)

func main() {
	os.Exit(cli.Run(os.Args[1:]))
}
`))

func init() {
	cli.Register("command.add", &command{General: &cli.General{}})
}

func (cmd *command) Register(f *flag.FlagSet) {
	cmd.General.Register(f)
	f.StringVar(&cmd.Name, "name", "", "command/subcommand")
	f.StringVar(&cmd.pkgname, "package", "", "package name")
	f.BoolVar(&cmd.force, "force", false, "override existing subcommand")
}

func (cmd *command) Description() string {
	return `
	Add "cli" command into package.".
	
Examples:
	; create "push" subcommand for "foo" project commands set
	; "foo" package will be created if not exists
	cligen command.add -name="origin/push"
	`
}

func (cmd *command) Process() error {
	if cmd.Name == "" {
		return cli.ErrFlagRequired("name")
	}
	if cmd.pkgname == "" {
		return cli.ErrFlagRequired("package")
	}

	tmp := strings.Split(cmd.Name, "/")
	if len(tmp) != 2 {
		return cli.ErrWrongFormat("name")
	}

	cmd.Name = tmp[0]
	cmd.Subcommand = tmp[1]

	if cmd.gopath = os.Getenv("GOPATH"); cmd.gopath == "" {
		cmd.gopath = build.Default.GOPATH
	}

	cmd.fullpath = filepath.Join(cmd.gopath, "src", cmd.pkgname, cmd.Name)
	cmd.fullname = filepath.Join(cmd.fullpath, cmd.Subcommand+".go")
	cmd.maingo = filepath.Join(cmd.gopath, "src", cmd.pkgname, "main.go")

	if _, err := os.Stat(cmd.fullname); err == nil {
		if !cmd.force {
			return fmt.Errorf("subcommand '%s' already exists at '%s'", cmd.Subcommand, cmd.fullpath)
		}
	} else {
		if !os.IsNotExist(err) {
			return fmt.Errorf("error checking %s/%s - %s", cmd.Name, cmd.Subcommand, err)
		}
	}

	if _, err := os.Stat(cmd.maingo); err != nil {
		if os.IsNotExist(err) {
			wr, err := os.OpenFile(cmd.maingo, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0644)
			if err != nil {
				return fmt.Errorf("error creating output file - %s", err)
			}
			defer wr.Close()

			if err = defaultMainGo.Execute(wr, nil); err != nil {
				return err
			}
		}
	}

	return nil
}

func (cmd *command) Run(f *flag.FlagSet) ([]byte, error) {
	if cmd.fullpath == "" || cmd.fullname == "" {
		return nil, fmt.Errorf("uninitialized")
	}

	log.Printf("[+] creating %s\n", cmd.fullpath)

	if err := os.MkdirAll(cmd.fullpath, 0755); err != nil && !os.IsExist(err) {
		return nil, fmt.Errorf("error creating destination directory - %s", err)
	}

	wr, err := os.OpenFile(cmd.fullname, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0644)
	if err != nil {
		return nil, fmt.Errorf("error creating output file - %s", err)
	}
	defer wr.Close()

	if err = commandTemplate.Execute(wr, cmd); err != nil {
		return nil, err
	}

	fset := token.NewFileSet()

	log.Printf("\t[-] parsing '%s'\n", cmd.maingo)

	file, err := parser.ParseFile(fset, cmd.maingo, nil, 0)
	if err != nil {
		return nil, err
	}

	ipath := cmd.pkgname + "/" + cmd.Name

	log.Printf("\t[-] adding import '%s'\n", ipath)

	if !astutil.AddNamedImport(fset, file, "_", ipath) {
		return nil, fmt.Errorf("error adding import '%s' to '%s'", ipath, cmd.maingo)
	}

	wr, err = os.OpenFile(cmd.maingo, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0644)
	if err != nil {
		return nil, fmt.Errorf("error creating output file - %s", err)
	}
	defer wr.Close()

	if err = printer.Fprint(wr, fset, file); err != nil {
		return nil, fmt.Errorf("error saving file - %s", err)
	}

	return nil, nil
}