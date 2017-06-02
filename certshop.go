package main

import (
	"flag"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"
)

var (
	// Version is populated using `-ldflags -X `git describe --tags`` build option
	// build using the Makefile to inject this value
	Version string
	// Build is populated using `-ldflags -X `date +%FT%T%z`` build option
	// build using the Makefile to inject this value
	Build string
	// InfoLog logs informational messages to stderr
	InfoLog = log.New(os.Stderr, ``, 0)
	// DebugLog logs additional debugging information to stderr when the -debug
	// flag is used
	DebugLog = log.New(ioutil.Discard, ``, 0)
	// ErrorLog logs error messages (which are typically fatal)
	// The philosophy of this program is to fail fast on an error and not try
	// to do any recovery (so the user knows something is wrong and can
	// explicitely fix it)
	ErrorLog = log.New(os.Stderr, `Error: `, 0)
	// Root is the working directory (defaults to "./")
	Root string
	// Overwrite specifies not to abort if the output directory already exists
	Overwrite bool
	// RunTime stores the time the program started running (used to determine
	// certificate NotBefore and NotAfter values)
	RunTime time.Time
	// NilString is a string used to determine whether user input to a flag
	// was provided. This is done by setting the default flag value to
	// NilString.
	NilString = `\x00` // default for string flags (used to detect if user supplied value)
	// Commands is a map from the string name of a command to meta information required to execute the command
	Commands = map[string]*Command{}
)

func init() {
	RunTime = time.Now().UTC()
	Commands[`version`] = &Command{
		Description: `display certshop version and build date and exit`,
		Function: func(fs *GlobalFlags) {
			InfoLog.Printf("certshop %s\nBuilt: %s\nCopyright (c) 2017 VARASYS Limited", Version, Build)
			os.Exit(0)
		},
	}
}

func main() {
	fs := ParseGlobalFlags()
	fs.Command.Function(fs)
}

// Command holds meta information about each command
type Command struct {
	Command     string
	Description string
	HelpString  string
	Function    func(*GlobalFlags)
}

// Name does a reverse lookup in the Commands map and returns the key
func (command *Command) Name() string {
	for key, value := range Commands {
		if command == value {
			return key
		}
	}
	ErrorLog.Fatalf("Failed to lookup command name")
	return ""
}

// GlobalFlags holds the global command line flags
type GlobalFlags struct {
	flag.FlagSet
	Root    string
	Debug   bool
	Command *Command
	Args    []string
}

// ParseGlobalFlags parses the global command line flags
func ParseGlobalFlags() *GlobalFlags {
	DebugLog.Println(`Parsing global flags`)
	fs := GlobalFlags{FlagSet: *flag.NewFlagSet(`certshop`, flag.ContinueOnError)}
	fs.StringVar(&fs.Root, `root`, `./`, `certificate tree root directory`)
	fs.BoolVar(&Overwrite, `overwrite`, false, `don't abort if output directory already exists`)
	fs.BoolVar(&fs.Debug, `debug`, false, `output extra debugging information`)
	if err := fs.Parse(os.Args[1:]); err != nil {
		printGlobalHelp(os.Stderr, &fs)
		ErrorLog.Fatalf(`Failed to parse global flags: %s`, strings.Join(os.Args[1:], ` `))
	}
	if debug {
		DebugLog = log.New(os.Stderr, ``, log.Lshortfile)
		ErrorLog.SetFlags(log.Lshortfile)
	}
	if root, err := filepath.Abs(fs.Root); err != nil {
		ErrorLog.Fatalf("Failed to parse root path %s: %s", fs.Root, err)
	} else {
		fs.Root = root
		SetRootDir(fs.Root)
	}
	fs.Args = fs.FlagSet.Args()
	if len(fs.Args) > 0 {
		fs.Command = Commands[fs.Args[0]]
	} else {
		fs.Command = Commands[`help`]
	}
	return &fs
}

// SetRootDir sets the applications working directory
func SetRootDir(root string) {
	DebugLog.Printf("Using root directory: %s", root)
	if err := os.MkdirAll(root, os.FileMode(0755)); err != nil {
		ErrorLog.Fatalf("Failed to create root directory %s: %s", root, err)
	}
	if err := os.Chdir(root); err != nil {
		ErrorLog.Fatalf("Failed to set root directory to %s: %s", root, err)
	}
}
