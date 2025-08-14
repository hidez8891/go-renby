package renby

import (
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"slices"
	"sort"
	"strings"
	"testing"
	"time"
)

type fileInfo struct {
	name string
	size int64
}

func generateTestFiles(n int) (string, error) {
	// Create temporary test directory
	tempDir, err := os.MkdirTemp("", "renby-test-*")
	if err != nil {
		return "", fmt.Errorf("failed to create temp dir: %w", err)
	}

	// Generate random sizes for test files
	sizes := make([]int64, n)
	for i := 0; i < n; i++ {
		sizes[i] = int64(i+1) * 10
	}
	rand.Shuffle(len(sizes), func(i, j int) {
		sizes[i], sizes[j] = sizes[j], sizes[i]
	})

	// Create test files with different sizes and timestamps
	testFiles := make([]fileInfo, n)
	for i := 0; i < n; i++ {
		name := filepath.Join(tempDir, fmt.Sprintf("test%d.txt", i+1))
		size := sizes[i]
		if err := os.WriteFile(name, make([]byte, size), 0644); err != nil {
			return "", fmt.Errorf("failed to create test file %s: %w", name, err)
		}
		testFiles[i] = fileInfo{name: name, size: size}
		time.Sleep(10 * time.Millisecond) // Ensure different timestamps
	}

	return tempDir, nil
}

func generateResultNamesWithInit(format string, n int, init int) []string {
	results := make([]string, n)
	for i := 0; i < n; i++ {
		results[i] = fmt.Sprintf(format, i+init)
	}
	return results
}

func generateResultNames(format string, n int) []string {
	return generateResultNamesWithInit(format, n, 1)
}

func TestRenameFiles(t *testing.T) {
	testN := 16
	tests := []struct {
		name    string
		opts    Options
		results []string
	}{
		{
			name: "sort by creation time asc",
			opts: Options{
				Pattern:  "000",
				FileMode: SortByCreationTime,
				Reverse:  false,
				Init:     1,
			},
			results: generateResultNames("%03d.txt", testN),
		},
		{
			name: "sort by creation time desc",
			opts: Options{
				Pattern:  "00000",
				FileMode: SortByCreationTime,
				Reverse:  true,
				Init:     1,
			},
			results: generateResultNames("%05d.txt", testN),
		},
		{
			name: "sort by modification time asc",
			opts: Options{
				Pattern:  "0000",
				FileMode: SortByModificationTime,
				Reverse:  false,
				Init:     1,
			},
			results: generateResultNames("%04d.txt", testN),
		},
		{
			name: "sort by modification time desc",
			opts: Options{
				Pattern:  "000",
				FileMode: SortByModificationTime,
				Reverse:  true,
				Init:     1,
			},
			results: generateResultNames("%03d.txt", testN),
		},
		{
			name: "sort by size asc",
			opts: Options{
				Pattern:  "000",
				FileMode: SortBySize,
				Reverse:  false,
				Init:     1,
			},
			results: generateResultNames("%03d.txt", testN),
		},
		{
			name: "sort by size desc",
			opts: Options{
				Pattern:  "000",
				FileMode: SortBySize,
				Reverse:  true,
				Init:     1,
			},
			results: generateResultNames("%03d.txt", testN),
		},
		{
			name: "sort with prefix and postfix",
			opts: Options{
				Pre:      "img",
				Post:     "test",
				Pattern:  "00",
				FileMode: SortBySize,
				Reverse:  false,
				Init:     1,
			},
			results: generateResultNames("img%02dtest.txt", testN),
		},
		{
			name: "sort with hex pattern",
			opts: Options{
				Pattern:  "xxx",
				FileMode: SortBySize,
				Reverse:  false,
				Init:     1,
			},
			results: generateResultNames("%03x.txt", testN),
		},
		{
			name: "sort with init number 100",
			opts: Options{
				Pattern:  "000",
				FileMode: SortBySize,
				Reverse:  false,
				Init:     100,
			},
			results: generateResultNamesWithInit("%03d.txt", testN, 100),
		},
		{
			name: "sort with init hex",
			opts: Options{
				Pattern:  "xxx",
				FileMode: SortBySize,
				Reverse:  false,
				Init:     10,
			},
			results: generateResultNamesWithInit("%03x.txt", testN, 10),
		},
		{
			name: "sort with init zero",
			opts: Options{
				Pattern:  "000",
				FileMode: SortBySize,
				Reverse:  false,
				Init:     0,
			},
			results: generateResultNamesWithInit("%03d.txt", testN, 0),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Generate test files
			tempDir, err := generateTestFiles(testN)
			if err != nil {
				t.Fatalf("failed to generate test files: %v", err)
			}
			defer os.RemoveAll(tempDir) // Clean up after test

			// Prepare file paths for renaming
			preentries, err := os.ReadDir(tempDir)
			if err != nil {
				t.Fatalf("failed to read dir: %v", err)
			}

			files := make([]string, len(preentries))
			for i, entry := range preentries {
				files[i] = filepath.Join(tempDir, entry.Name())
			}

			// Prepare file entries for sorting
			switch tt.opts.FileMode {
			case SortByCreationTime:
				sort.Slice(preentries, func(i, j int) bool {
					info1, _ := getFileInfo(filepath.Join(tempDir, preentries[i].Name()))
					info2, _ := getFileInfo(filepath.Join(tempDir, preentries[j].Name()))
					return info1.CreateTime.Before(info2.CreateTime)
				})
			case SortByModificationTime:
				sort.Slice(preentries, func(i, j int) bool {
					info1, _ := getFileInfo(filepath.Join(tempDir, preentries[i].Name()))
					info2, _ := getFileInfo(filepath.Join(tempDir, preentries[j].Name()))
					return info1.ModTime.Before(info2.ModTime)
				})
			case SortByAccessTime:
				sort.Slice(preentries, func(i, j int) bool {
					info1, _ := getFileInfo(filepath.Join(tempDir, preentries[i].Name()))
					info2, _ := getFileInfo(filepath.Join(tempDir, preentries[j].Name()))
					return info1.AccessTime.Before(info2.AccessTime)
				})
			case SortBySize:
				sort.Slice(preentries, func(i, j int) bool {
					info1, _ := getFileInfo(filepath.Join(tempDir, preentries[i].Name()))
					info2, _ := getFileInfo(filepath.Join(tempDir, preentries[j].Name()))
					return info1.Size < info2.Size
				})
			}
			if tt.opts.Reverse {
				slices.Reverse(preentries)
			}

			resultFileOrder := make([]int64, len(preentries))
			for i := range preentries {
				info, _ := preentries[i].Info()
				resultFileOrder[i] = info.Size()
			}

			// Execute rename
			if err := RenameFiles(files, tt.opts); err != nil {
				t.Fatalf("RenameFiles() error = %v", err)
			}

			// Verify renamed files and their order
			entries, err := os.ReadDir(tempDir)
			if err != nil {
				t.Fatalf("failed to read dir: %v", err)
			}
			if len(entries) != len(tt.results) {
				t.Errorf("got %d files, want %d", len(entries), len(tt.results))
			}
			sort.Slice(entries, func(i, j int) bool {
				return entries[i].Name() < entries[j].Name()
			})

			// Verify file names
			for i, want := range tt.results {
				if got := entries[i].Name(); got != want {
					t.Errorf("file[%d] = %s, want %s", i, got, want)
				}
			}

			// Verify sort order based on file properties
			for i, entry := range entries {
				info, err := entry.Info()
				if err != nil {
					t.Fatalf("failed to get file info for %s: %v", entry.Name(), err)
				}

				if info.Size() != resultFileOrder[i] {
					t.Errorf("sort is maybe failed: file[%d] ID = %d, want %d", i, info.Size(), resultFileOrder[i])
				}
			}
		})
	}
}

func TestRenameFiles_NegativeInit(t *testing.T) {
	opts := Options{
		Pattern:  "000",
		FileMode: SortBySize,
		Init:     -1,
	}

	err := RenameFiles([]string{"test.txt"}, opts)
	if err == nil {
		t.Error("RenameFiles() error = nil, want error for negative init")
	}
	if err.Error() != "invalid options: init value must be non-negative" {
		t.Errorf("RenameFiles() error = %v, want 'invalid options: init value must be non-negative'", err)
	}
}

func TestRenameFiles_Empty(t *testing.T) {
	opts := Options{
		Pattern:  "000",
		FileMode: SortBySize,
	}

	if err := RenameFiles(nil, opts); err != nil {
		t.Errorf("RenameFiles() error = %v, want nil", err)
	}

	if err := RenameFiles([]string{}, opts); err != nil {
		t.Errorf("RenameFiles() error = %v, want nil", err)
	}
}

func TestRenameFiles_CollisionDetected(t *testing.T) {
	dir := t.TempDir()
	a := filepath.Join(dir, "A.txt")
	b := filepath.Join(dir, "B.txt")
	if err := os.WriteFile(a, []byte("a"), 0644); err != nil {
		t.Fatalf("write a: %v", err)
	}
	if err := os.WriteFile(b, []byte("b"), 0644); err != nil {
		t.Fatalf("write b: %v", err)
	}

	files := []string{a, b}
	opts := Options{
		Pattern:  "00000",
		Init:     1,
		FileMode: SortByCreationTime,
	}

	if err := RenameFiles(files, opts); err != nil {
		t.Fatalf("initial rename failed: %v", err)
	}

	got, err := filepath.Glob(filepath.Join(dir, "*"))
	if err != nil {
		t.Fatalf("glob failed: %v", err)
	}
	var names []string
	for _, p := range got {
		names = append(names, filepath.Base(p))
	}
	sort.Strings(names)

	want := []string{"00001.txt", "00002.txt"}
	if !slices.Equal(names, want) {
		t.Fatalf("unexpected names after first rename: got=%v want=%v", names, want)
	}

	// Attempt reverse without Force: should detect conflicts and return error
	reverseOpts := opts
	reverseOpts.Reverse = true
	if err := RenameFiles(got, reverseOpts); err == nil {
		t.Fatalf("expected conflict error on reverse rename, got nil")
	} else if !strings.Contains(err.Error(), "conflicts detected") {
		t.Fatalf("unexpected error message: %v", err)
	}

	// Now allow Force: should succeed
	reverseForce := reverseOpts
	reverseForce.ForceOverwrite = true
	if err := RenameFiles(got, reverseForce); err != nil {
		t.Fatalf("force rename failed: %v", err)
	}

	final, err := filepath.Glob(filepath.Join(dir, "*"))
	if err != nil {
		t.Fatalf("glob failed: %v", err)
	}
	if len(final) != 2 {
		t.Fatalf("expected 2 files after force rename, got %d", len(final))
	}
}
