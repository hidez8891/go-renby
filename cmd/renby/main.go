package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/hidez8891/go-renby"
	"github.com/spf13/pflag"
)

const (
	version = "0.1.0"
)

func main() {
	args := os.Args[:]
	flags := pflag.NewFlagSet(args[0], pflag.ExitOnError)

	if len(args) < 2 || args[1] == "--help" {
		showHelp()
		os.Exit(0)
	}
	if args[1] == "--version" {
		showVersion()
		os.Exit(0)
	}

	// Get subcommand
	subCmd := args[1]
	if !isValidSubCmd(subCmd) {
		fmt.Fprintf(os.Stderr, "Error: invalid subcommand '%s'\n", subCmd)
		showHelp()
		os.Exit(1)
	}

	// Parse flags
	var (
		reverse = flags.BoolP("reverse", "r", false, "reverse order")
		pattern = flags.StringP("pattern", "p", "000000", "rename pattern (0: decimal, x: hexadecimal)")
		pre     = flags.String("pre", "", "prefix string")
		post    = flags.String("post", "", "postfix string")
		help    = flags.Bool("help", false, "show help")
		ver     = flags.Bool("version", false, "show version")
	)

	if err := flags.Parse(args[2:]); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		showHelp()
		os.Exit(1)
	}

	if *help {
		showHelp()
		os.Exit(0)
	}

	if *ver {
		showVersion()
		os.Exit(0)
	}

	// Get file patterns
	filePatterns := flags.Args()
	if len(filePatterns) == 0 {
		fmt.Fprintf(os.Stderr, "Error: file pattern required\n")
		showHelp()
		os.Exit(1)
	}

	// Get matching files from all patterns
	var files []string
	for _, pat := range filePatterns {
		matches, err := filepath.Glob(pat)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: invalid file pattern '%s'\n", pat)
			os.Exit(1)
		}
		if len(matches) == 0 {
			fmt.Fprintf(os.Stderr, "Warning: no files match pattern '%s'\n", pat)
			continue
		}
		files = append(files, matches...)
	}
	if len(files) == 0 {
		fmt.Fprintf(os.Stderr, "Error: no files found\n")
		os.Exit(1)
	}

	// Remove duplicates while preserving order
	savedFiles := make([]string, 0, len(files))
	uniqueFiles := make(map[string]struct{})
	for _, file := range files {
		absPath, err := filepath.Abs(file)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: could not get absolute path for '%s': %v\n", file, err)
			os.Exit(1)
		}
		if _, exists := uniqueFiles[absPath]; !exists {
			uniqueFiles[absPath] = struct{}{}
			savedFiles = append(savedFiles, absPath)
		}
	}
	files = savedFiles

	// Execute renaming
	opts := renby.Options{
		Pre:      *pre,
		Post:     *post,
		Pattern:  *pattern,
		Reverse:  *reverse,
		FileMode: parseSortMode(subCmd),
	}

	if err := renby.RenameFiles(files, opts); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
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
  --pre=STRING          prefix string
  --post=STRING         postfix string
  --help                show this help
  --version             show version

Example:
  renby ctime *.png
  renby size -r --pre=img --post=test *.jpg
  renby size -p=xxx *.txt`)
}

func showVersion() {
	fmt.Printf("renby version %s\n", version)
}
