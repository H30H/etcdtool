package command

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/codegangsta/cli"
	"github.com/mickep76/iodatafmt"
)

// NewImportCommand sets data from input.
func NewEditCommand() cli.Command {
	return cli.Command{
		Name:  "edit",
		Usage: "edit a directory",
		Flags: []cli.Flag{
			cli.BoolFlag{Name: "sort, s", Usage: "returns result in sorted order"},
			cli.BoolFlag{Name: "yes, y", Usage: "Answer yes to any questions"},
			cli.BoolFlag{Name: "replace, r", Usage: "Replace data"},
			cli.StringFlag{Name: "format, f", Value: "JSON", EnvVar: "ETCDTOOL_FORMAT", Usage: "Data serialization format YAML, TOML or JSON"},
			cli.StringFlag{Name: "editor, e", Value: "vim", Usage: "Editor", EnvVar: "EDITOR"},
			cli.StringFlag{Name: "tmp-file, t", Value: ".etcdtool.swp", Usage: "Temporary file"},
		},
		Action: func(c *cli.Context) {
			editCommandFunc(c)
		},
	}
}

func editFile(editor string, file string) error {
	_, err := exec.LookPath(editor)
	if err != nil {
		fatalf("Editor doesn't exist: %s", editor)
	}

	cmd := exec.Command(editor, file)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	if err := cmd.Run(); err != nil {
		fatal(err.Error())
	}
	return nil
}

// editCommandFunc edit data as either JSON, YAML or TOML.
func editCommandFunc(c *cli.Context) {
	if len(c.Args()) == 0 {
		fatal("You need to specify directory")
	}
	dir := c.Args()[0]

	// Remove trailing slash.
	if dir != "/" {
		dir = strings.TrimRight(dir, "/")
	}
	infof("Using dir: %s", dir)

	// Load configuration file.
	e := loadConfig(c)

	// New dir API.
	ki := newKeyAPI(e)

	sort := c.Bool("sort")

	// Get data format.
	f, err := iodatafmt.Format(c.String("format"))
	if err != nil {
		fatal(err.Error())
	}

	// Export to file.
	exportFunc(dir, sort, c.String("tmp-file"), f, c, ki)

	// Get modified time stamp.
	before, err := os.Stat(c.String("tmp-file"))
	if err != nil {
		fatal(err.Error())
	}

	// Edit file.
	editFile(c.String("editor"), c.String("tmp-file"))

	// Check modified time stamp.
	after, err := os.Stat(c.String("tmp-file"))
	if err != nil {
		fatal(err.Error())
	}

	// Import from file if it has changed.
	if before.ModTime() != after.ModTime() {
		importFunc(dir, c.String("tmp-file"), f, c.Bool("replace"), c.Bool("yes"), c, ki)
	} else {
		fmt.Printf("File wasn't modified, skipping import\n")
	}

	// Unlink file.
	if err := os.Remove(c.String("tmp-file")); err != nil {
		fatal(err.Error())
	}
}
