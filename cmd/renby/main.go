package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/hidez8891/go-renby"
)

const (
	version = "0.1.0"
)

func main() {
	var (
		reverse = flag.Bool("r", false, "reverse order")
		pattern = flag.String("p", "000000", "rename pattern (0: decimal, x: hexadecimal)")
		pre     = flag.String("pre", "", "prefix string")
		post    = flag.String("post", "", "postfix string")
		help    = flag.Bool("help", false, "show help")
		ver     = flag.Bool("version", false, "show version")
	)

	// Add long option aliases
	flag.Bool("reverse", false, "reverse order")
	flag.String("pattern", "000000", "rename pattern (0: decimal, x: hexadecimal)")

	flag.Parse()

	if *help {
		showHelp()
		os.Exit(0)
	}

	if *ver {
		fmt.Printf("renby version %s\n", version)
		os.Exit(0)
	}

	args := flag.Args()
	if len(args) < 1 {
		fmt.Fprintf(os.Stderr, "Error: subcommand required\n")
		showHelp()
		os.Exit(1)
	}

	subCmd := args[0]
	if !isValidSubCmd(subCmd) {
		fmt.Fprintf(os.Stderr, "Error: invalid subcommand '%s'\n", subCmd)
		showHelp()
		os.Exit(1)
	}

	// Get file patterns after the subcommand, excluding any flags
	var filePatterns []string
	for _, arg := range args[1:] {
		// Skip if arg starts with - or -- (flags)
		if len(arg) > 0 && arg[0] == '-' {
			continue
		}
		filePatterns = append(filePatterns, arg)
	}

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
  --pre=STRING         prefix string
  --post=STRING        postfix string
  --help              show this help
  --version           show version

Example:
  renby ctime *.png
  renby size -r --pre=img --post=test *.jpg
  renby size -p=xxx *.txt`)
}
