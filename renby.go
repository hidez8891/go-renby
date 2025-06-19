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

// Options represents configuration options for file renaming
type Options struct {
	Pre      string
	Post     string
	Pattern  string
	Reverse  bool
	FileMode SortMode
}

// FileInfo represents file information used for sorting
type FileInfo struct {
	Path       string
	Size       int64
	CreateTime time.Time
	ModTime    time.Time
	AccessTime time.Time
}

// getFileInfo returns FileInfo for the given file path
func getFileInfo(path string) (FileInfo, error) {
	info, err := os.Stat(path)
	if err != nil {
		return FileInfo{}, fmt.Errorf("failed to get file info: %w", err)
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

// RenameFiles renames files according to the specified options
func RenameFiles(files []string, opts Options) error {
	if len(files) == 0 {
		return nil
	}

	// Collect file information
	fileInfos := make([]FileInfo, 0, len(files))
	for _, file := range files {
		fi, err := getFileInfo(file)
		if err != nil {
			return err
		}

		info, err := os.Stat(file)
		if err != nil {
			return fmt.Errorf("failed to get file info: %w", err)
		}
		if info.IsDir() {
			continue
		}

		fileInfos = append(fileInfos, fi)
	}

	// Sort files
	sort.Slice(fileInfos, func(i, j int) bool {
		var result bool
		switch opts.FileMode {
		case SortByCreationTime:
			result = fileInfos[i].CreateTime.Before(fileInfos[j].CreateTime)
		case SortByModificationTime:
			result = fileInfos[i].ModTime.Before(fileInfos[j].ModTime)
		case SortByAccessTime:
			result = fileInfos[i].AccessTime.Before(fileInfos[j].AccessTime)
		case SortBySize:
			result = fileInfos[i].Size < fileInfos[j].Size
		}
		if opts.Reverse {
			return !result
		}
		return result
	})

	// Generate new names
	for i, fi := range fileInfos {
		ext := filepath.Ext(fi.Path)
		basePath := fi.Path[:len(fi.Path)-len(ext)]
		dir := filepath.Dir(basePath)

		// Generate new filename
		var newName string
		if strings.Contains(opts.Pattern, "x") {
			newName = fmt.Sprintf("%s%0*x%s%s", opts.Pre, len(opts.Pattern), i+1, opts.Post, ext)
		} else {
			newName = fmt.Sprintf("%s%0*d%s%s", opts.Pre, len(opts.Pattern), i+1, opts.Post, ext)
		}

		newPath := filepath.Join(dir, newName)
		if fi.Path != newPath {
			if err := os.Rename(fi.Path, newPath); err != nil {
				return fmt.Errorf("failed to rename file: %w", err)
			}
		}
	}

	return nil
}
