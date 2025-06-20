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
	Pre      string
	Post     string
	Pattern  string
	Reverse  bool
	FileMode SortMode
}

// Validate checks if the options are valid
func (o Options) Validate() error {
	if o.Pattern == "" {
		return fmt.Errorf("pattern cannot be empty")
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
		newName = fmt.Sprintf("%s%0*x%s%s", opts.Pre, len(opts.Pattern), index+1, opts.Post, ext)
	} else {
		newName = fmt.Sprintf("%s%0*d%s%s", opts.Pre, len(opts.Pattern), index+1, opts.Post, ext)
	}

	return filepath.Join(dir, newName)
}

// RenameFiles renames files according to the specified options
func RenameFiles(files []string, opts Options) error {
	if err := opts.Validate(); err != nil {
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

	for i, fi := range fileInfos {
		newPath := generateNewName(fi, i, opts)
		if fi.Path != newPath {
			if err := os.Rename(fi.Path, newPath); err != nil {
				return fmt.Errorf("failed to rename file: %w", err)
			}
		}
	}

	return nil
}
