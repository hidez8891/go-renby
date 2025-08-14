package main

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime/debug"

	"github.com/hidez8891/go-renby"
	"github.com/spf13/pflag"
)

const (
	defaultPattern = "000000"
	exitSuccess    = 0
	exitFailure    = 1
)

var (
	// Version is set at build time using -X flag
	Version = ""
)

type config struct {
	reverse        bool
	pattern        string
	pre            string
	post           string
	forceOverwrite bool
	help           bool
	version        bool
	init           int
	filePatterns   []string
}

func main() {
	if err := run(os.Args[:]); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(exitFailure)
	}
}

func run(args []string) error {
	if len(args) < 2 {
		showHelp()
		os.Exit(exitSuccess)
		return nil
	}

	// Handle global flags first
	if args[1] == "--help" {
		showHelp()
		os.Exit(exitSuccess)
		return nil
	}
	if args[1] == "--version" {
		showVersion()
		os.Exit(exitSuccess)
		return nil
	}

	// Validate subcommand
	subCmd := args[1]
	if !isValidSubCmd(subCmd) {
		return fmt.Errorf("invalid subcommand '%s'", subCmd)
	}

	// Parse configuration
	cfg, err := parseFlags(args[0], args[2:])
	if err != nil {
		return err
	}

	if cfg.help {
		showHelp()
		os.Exit(exitSuccess)
		return nil
	}

	if cfg.version {
		showVersion()
		os.Exit(exitSuccess)
		return nil
	}

	// Process files
	files, err := processFilePatterns(cfg.filePatterns)
	if err != nil {
		return err
	}

	// Execute renaming
	opts := renby.Options{
		Pre:            cfg.pre,
		Post:           cfg.post,
		Pattern:        cfg.pattern,
		Reverse:        cfg.reverse,
		FileMode:       parseSortMode(subCmd),
		Init:           cfg.init,
		ForceOverwrite: cfg.forceOverwrite,
	}

	return renby.RenameFiles(files, opts)
}

func parseFlags(name string, args []string) (*config, error) {
	flags := pflag.NewFlagSet(name, pflag.ExitOnError)

	cfg := &config{}
	flags.BoolVarP(&cfg.reverse, "reverse", "r", false, "reverse order")
	flags.StringVarP(&cfg.pattern, "pattern", "p", defaultPattern, "rename pattern (0: decimal, x: hexadecimal)")
	flags.IntVar(&cfg.init, "init", 1, "initial number (non-negative)")
	flags.StringVar(&cfg.pre, "pre", "", "prefix string")
	flags.StringVar(&cfg.post, "post", "", "postfix string")
	flags.BoolVar(&cfg.forceOverwrite, "force", false, "allow overwriting existing destination files (performs a safe two-phase rename)")
	flags.BoolVar(&cfg.help, "help", false, "show help")
	flags.BoolVar(&cfg.version, "version", false, "show version")

	if err := flags.Parse(args); err != nil {
		return nil, err
	}

	// Store remaining args as file patterns
	cfg.filePatterns = flags.Args()
	if len(cfg.filePatterns) == 0 {
		return nil, fmt.Errorf("file pattern required")
	}

	return cfg, nil
}

func processFilePatterns(patterns []string) ([]string, error) {
	var files []string
	for _, pat := range patterns {
		matches, err := filepath.Glob(pat)
		if err != nil {
			return nil, fmt.Errorf("invalid file pattern '%s'", pat)
		}
		if len(matches) == 0 {
			fmt.Fprintf(os.Stderr, "Warning: no files match pattern '%s'\n", pat)
			continue
		}
		files = append(files, matches...)
	}
	if len(files) == 0 {
		return nil, fmt.Errorf("no files found")
	}

	return removeDuplicates(files)
}

func removeDuplicates(files []string) ([]string, error) {
	savedFiles := make([]string, 0, len(files))
	uniqueFiles := make(map[string]struct{})

	for _, file := range files {
		absPath, err := filepath.Abs(file)
		if err != nil {
			return nil, fmt.Errorf("could not get absolute path for '%s': %v", file, err)
		}
		if _, exists := uniqueFiles[absPath]; !exists {
			uniqueFiles[absPath] = struct{}{}
			savedFiles = append(savedFiles, absPath)
		}
	}

	return savedFiles, nil
}

func isValidSubCmd(cmd string) bool {
	validCmds := []string{"ctime", "mtime", "atime", "size"}
	for _, valid := range validCmds {
		if cmd == valid {
			return true
		}
	}
	return false
}

func parseSortMode(cmd string) renby.SortMode {
	switch cmd {
	case "ctime":
		return renby.SortByCreationTime
	case "mtime":
		return renby.SortByModificationTime
	case "atime":
		return renby.SortByAccessTime
	case "size":
		return renby.SortBySize
	default:
		return renby.SortByCreationTime
	}
}

func showHelp() {
	fmt.Println(`Usage: renby SUBCOMMAND [OPTIONS] FILES...

SUBCOMMAND:
  ctime     sort by creation time
  mtime     sort by modification time
  atime     sort by access time
  size      sort by file size

OPTIONS:
  -r, --reverse         reverse sort order
  -p, --pattern=STRING  rename pattern (0: decimal, x: hexadecimal)
                        default: 000000
  --init=NUMBER         initial number (non-negative)
                        default: 1
  --pre=STRING          prefix string
  --post=STRING         postfix string
  --force               allow overwriting existing destination files (performs a safe two-phase rename)
  --help                show this help
  --version             show version

Example:
  renby ctime *.png
  renby size -r --pre=img --post=test *.jpg
  renby size -p=xxx *.txt
  renby size --init=100 *.txt`)
}

func showVersion() {
	// go-build -ldflags="-X main.Version=XXX" to set Version at build time
	if Version != "" {
		fmt.Printf("renby version %s\n", Version)
		return
	}

	// go-install is not set Version, so we try to read build info
	version := "(unknown)"
	if buildInfo, ok := debug.ReadBuildInfo(); ok {
		version = buildInfo.Main.Version
	}
	fmt.Printf("renby version %s\n", version)
}
