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
	// Define flags
	reverse := flag.Bool("r", false, "reverse order")
	reverseLong := flag.Bool("reverse", false, "reverse order")
	pattern := flag.String("p", "000000", "rename pattern (0: decimal, x: hexadecimal)")
	patternLong := flag.String("pattern", "000000", "rename pattern (0: decimal, x: hexadecimal)")
	pre := flag.String("pre", "", "prefix string")
	post := flag.String("post", "", "postfix string")
	help := flag.Bool("help", false, "show help")
	ver := flag.Bool("version", false, "show version")

	flag.Parse()

	// Show help
	if *help {
		showHelp()
		os.Exit(0)
	}

	// Show version
	if *ver {
		fmt.Printf("renby version %s\n", version)
		os.Exit(0)
	}

	// Get subcommand and pattern
	args := flag.Args()
	if len(args) < 2 {
		fmt.Fprintf(os.Stderr, "Error: insufficient arguments\n")
		showHelp()
		os.Exit(1)
	}

	subCmd := args[0]
	filePattern := args[1]

	// Validate subcommand
	if !isValidSubCmd(subCmd) {
		fmt.Fprintf(os.Stderr, "Error: invalid subcommand '%s'\n", subCmd)
		showHelp()
		os.Exit(1)
	}

	// Get effective pattern and reverse flag
	effectivePattern := *pattern
	if *patternLong != "000000" {
		effectivePattern = *patternLong
	}
	isReverse := *reverse || *reverseLong

	// Get matching files
	files, err := filepath.Glob(filePattern)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: invalid file pattern '%s'\n", filePattern)
		os.Exit(1)
	}

	// Create options
	opts := renby.Options{
		Pre:      *pre,
		Post:     *post,
		Pattern:  effectivePattern,
		Reverse:  isReverse,
		FileMode: parseSortMode(subCmd),
	}

	// Execute renaming
	err = renby.RenameFiles(files, opts)
	if err != nil {
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
	fmt.Println(`Usage: renby SUBCOMMAND [OPTIONS] PATTERN

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
