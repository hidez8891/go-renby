package renby

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/hidez8891/go-renby/internal/ostime"
)

// SortMode represents the file sorting mode
type SortMode int

const (
	SortByCreationTime SortMode = iota
	SortByModificationTime
	SortByAccessTime
	SortBySize
)

// FileInfo represents file information used for sorting
type FileInfo struct {
	Path       string
	Size       int64
	CreateTime time.Time
	ModTime    time.Time
	AccessTime time.Time
}

// Options represents configuration options for file renaming
type Options struct {
	Pre            string
	Post           string
	Pattern        string
	Reverse        bool
	FileMode       SortMode
	Init           int // default: 1
	ForceOverwrite bool
}

// Validate checks if the options are valid
func (o *Options) Validate() error {
	if o.Pattern == "" {
		return fmt.Errorf("pattern cannot be empty")
	}
	if o.Init < 0 {
		return fmt.Errorf("init value must be non-negative")
	}
	return nil
}

// getFileInfo returns FileInfo for the given file path
func getFileInfo(path string) (FileInfo, error) {
	info, err := os.Stat(path)
	if err != nil {
		return FileInfo{}, fmt.Errorf("failed to get file info: %w", err)
	}

	if info.IsDir() {
		// Skip directories silently as per original behavior
		return FileInfo{}, nil
	}

	fi := FileInfo{
		Path: path,
		Size: info.Size(),
	}

	// Get system-specific file times
	ostime := ostime.GetOsTime(info)
	fi.CreateTime = ostime.CreationTime
	fi.ModTime = ostime.ModificationTime
	fi.AccessTime = ostime.AccessTime

	return fi, nil
}

// collectFileInfo gathers FileInfo for all input files
func collectFileInfo(files []string) ([]FileInfo, error) {
	if len(files) == 0 {
		return nil, nil
	}

	fileInfos := make([]FileInfo, 0, len(files))
	for _, file := range files {
		fi, err := getFileInfo(file)
		if err != nil {
			return nil, err
		}
		if fi != (FileInfo{}) { // Skip empty FileInfo (directories)
			fileInfos = append(fileInfos, fi)
		}
	}
	return fileInfos, nil
}

// sortFiles sorts FileInfo slice based on the specified mode
func sortFiles(files []FileInfo, mode SortMode, reverse bool) {
	sort.Slice(files, func(i, j int) bool {
		result := compareFiles(files[i], files[j], mode)
		if reverse {
			return !result
		}
		return result
	})
}

// compareFiles compares two files based on the sort mode
func compareFiles(a, b FileInfo, mode SortMode) bool {
	switch mode {
	case SortByCreationTime:
		return a.CreateTime.Before(b.CreateTime)
	case SortByModificationTime:
		return a.ModTime.Before(b.ModTime)
	case SortByAccessTime:
		return a.AccessTime.Before(b.AccessTime)
	case SortBySize:
		return a.Size < b.Size
	default:
		return a.Path < b.Path
	}
}

// generateNewName creates a new filename based on the pattern
func generateNewName(fi FileInfo, index int, opts Options) string {
	ext := filepath.Ext(fi.Path)
	basePath := fi.Path[:len(fi.Path)-len(ext)]
	dir := filepath.Dir(basePath)

	var newName string
	if strings.Contains(opts.Pattern, "x") {
		newName = fmt.Sprintf("%s%0*x%s%s", opts.Pre, len(opts.Pattern), index+opts.Init, opts.Post, ext)
	} else {
		newName = fmt.Sprintf("%s%0*d%s%s", opts.Pre, len(opts.Pattern), index+opts.Init, opts.Post, ext)
	}

	return filepath.Join(dir, newName)
}

// RenameFiles renames files according to the specified options
func RenameFiles(files []string, opts Options) error {
	if err := (&opts).Validate(); err != nil {
		return fmt.Errorf("invalid options: %w", err)
	}

	fileInfos, err := collectFileInfo(files)
	if err != nil {
		return err
	}

	if len(fileInfos) == 0 {
		return nil
	}

	sortFiles(fileInfos, opts.FileMode, opts.Reverse)

	// Build planned mapping: src -> dst
	plan := make(map[string]string, len(fileInfos))
	dstToSrc := make(map[string][]string)
	srcSet := make(map[string]struct{}, len(fileInfos))
	for i, fi := range fileInfos {
		dst := generateNewName(fi, i, opts)
		plan[fi.Path] = dst
		dstToSrc[dst] = append(dstToSrc[dst], fi.Path)
		srcSet[fi.Path] = struct{}{}
	}

	// Detect conflicts:
	// - Multiple sources mapping to the same destination
	// - Destination already exists on filesystem and is not one of the sources
	var conflicts []string
	for dst, srcs := range dstToSrc {
		if len(srcs) > 1 {
			conflicts = append(conflicts, fmt.Sprintf("multiple sources %v -> same destination %q", srcs, dst))
		}
		if _, ok := srcSet[dst]; !ok {
			if _, err := os.Stat(dst); err == nil {
				conflicts = append(conflicts, fmt.Sprintf("destination already exists: %q", dst))
			}
		}
	}

	if len(conflicts) > 0 && !opts.ForceOverwrite {
		return fmt.Errorf("conflicts detected, aborting: %s", strings.Join(conflicts, "; "))
	}

	// Perform renames.
	// If not ForceOverwrite, fail renames when any destination already exists.
	// If the existing destination is itself one of the sources in this batch,
	// report it as a conflict so callers receive "conflicts detected".
	if !opts.ForceOverwrite {
		for src, dst := range plan {
			if src == dst {
				continue
			}
			if _, err := os.Stat(dst); err == nil {
				if _, isSource := srcSet[dst]; isSource {
					return fmt.Errorf("conflicts detected: destination %q is also a source", dst)
				}
				return fmt.Errorf("destination already exists before renaming: %q", dst)
			}
			if err := os.Rename(src, dst); err != nil {
				return fmt.Errorf("failed to rename file: %w", err)
			}
		}
		return nil
	}

	// Force mode: use safe two-phase renaming to avoid overwrites/cycles:
	// 1) rename each src -> unique temp
	// 2) rename each temp -> final dst
	pid := os.Getpid()
	temps := make(map[string]string, len(plan)) // src -> temp
	counter := 0
	for src, dst := range plan {
		if src == dst {
			continue
		}
		// build a unique temp name in same dir as dst
		dir := filepath.Dir(dst)
		ext := filepath.Ext(dst)
		base := strings.TrimSuffix(filepath.Base(dst), ext)
		var temp string
		for {
			temp = filepath.Join(dir, fmt.Sprintf("%s.renby.tmp.%d.%d%s", base, pid, counter, ext))
			counter++
			if _, err := os.Stat(temp); os.IsNotExist(err) {
				break
			}
		}
		if err := os.Rename(src, temp); err != nil {
			return fmt.Errorf("failed to move source %q to temp %q: %w", src, temp, err)
		}
		temps[src] = temp
	}

	// move temps to final destinations
	for src, dst := range plan {
		if src == dst {
			continue
		}
		temp := temps[src]
		if temp == "" {
			// shouldn't happen
			return fmt.Errorf("missing temp for source %q", src)
		}
		// ensure dst does not exist (remove if present)
		if _, err := os.Stat(dst); err == nil {
			if err := os.Remove(dst); err != nil {
				return fmt.Errorf("failed to remove existing destination %q: %w", dst, err)
			}
		}
		if err := os.Rename(temp, dst); err != nil {
			return fmt.Errorf("failed to rename temp %q to dst %q: %w", temp, dst, err)
		}
	}

	return nil
}
