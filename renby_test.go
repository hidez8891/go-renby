package renby

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestRenameFiles(t *testing.T) {
	// Create temporary test directory
	tempDir, err := os.MkdirTemp("", "renby-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create test files with different sizes and times
	testFiles := []struct {
		name string
		size int64
	}{
		{"test1.txt", 100},
		{"test2.txt", 50},
		{"test3.txt", 200},
	}

	var files []string
	for _, tf := range testFiles {
		path := filepath.Join(tempDir, tf.name)
		if err := os.WriteFile(path, make([]byte, tf.size), 0644); err != nil {
			t.Fatalf("failed to create test file: %v", err)
		}
		files = append(files, path)
		// Add delay to ensure different timestamps
		time.Sleep(100 * time.Millisecond)
	}

	tests := []struct {
		name      string
		opts      Options
		wantOrder []string
	}{
		{
			name: "sort by size asc",
			opts: Options{
				Pattern:  "000",
				FileMode: SortBySize,
				Reverse:  false,
			},
			wantOrder: []string{"001.txt", "002.txt", "003.txt"},
		},
		{
			name: "sort by size desc",
			opts: Options{
				Pattern:  "000",
				FileMode: SortBySize,
				Reverse:  true,
			},
			wantOrder: []string{"001.txt", "002.txt", "003.txt"},
		},
		{
			name: "sort with prefix and postfix",
			opts: Options{
				Pre:      "img",
				Post:     "test",
				Pattern:  "00",
				FileMode: SortBySize,
				Reverse:  false,
			},
			wantOrder: []string{"img01test.txt", "img02test.txt", "img03test.txt"},
		},
		{
			name: "sort with hex pattern",
			opts: Options{
				Pattern:  "xxx",
				FileMode: SortBySize,
				Reverse:  false,
			},
			wantOrder: []string{"001.txt", "002.txt", "003.txt"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Execute rename
			if err := RenameFiles(files, tt.opts); err != nil {
				t.Fatalf("RenameFiles() error = %v", err)
			}

			// Verify renamed files
			entries, err := os.ReadDir(tempDir)
			if err != nil {
				t.Fatalf("failed to read dir: %v", err)
			}

			if len(entries) != len(tt.wantOrder) {
				t.Errorf("got %d files, want %d", len(entries), len(tt.wantOrder))
			}

			for i, want := range tt.wantOrder {
				if i >= len(entries) {
					break
				}
				if got := entries[i].Name(); got != want {
					t.Errorf("file[%d] = %s, want %s", i, got, want)
				}
			}

			// Reset file names for next test
			for i, entry := range entries {
				oldPath := filepath.Join(tempDir, entry.Name())
				newPath := filepath.Join(tempDir, testFiles[i].name)
				if err := os.Rename(oldPath, newPath); err != nil {
					t.Fatalf("failed to reset file name: %v", err)
				}
				files[i] = newPath
			}
		})
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
