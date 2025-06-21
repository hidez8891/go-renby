package main

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestParseFlags(t *testing.T) {
	tests := []struct {
		name        string
		args        []string
		want        *config
		wantErr     bool
		errContains string
	}{
		{
			name: "basic file pattern",
			args: []string{"*.png"},
			want: &config{
				reverse:      false,
				pattern:      defaultPattern,
				pre:          "",
				post:         "",
				help:         false,
				version:      false,
				init:         1,
				filePatterns: []string{"*.png"},
			},
			wantErr: false,
		},
		{
			name: "reverse with pre/post",
			args: []string{"-r", "--pre=img", "--post=test", "*.jpg"},
			want: &config{
				reverse:      true,
				pattern:      defaultPattern,
				pre:          "img",
				post:         "test",
				help:         false,
				version:      false,
				init:         1,
				filePatterns: []string{"*.jpg"},
			},
			wantErr: false,
		},
		{
			name: "hex pattern",
			args: []string{"-p=xxx", "*.txt"},
			want: &config{
				reverse:      false,
				pattern:      "xxx",
				pre:          "",
				post:         "",
				help:         false,
				version:      false,
				init:         1,
				filePatterns: []string{"*.txt"},
			},
			wantErr: false,
		},
		{
			name: "decimal pattern",
			args: []string{"-p=000", "*.txt"},
			want: &config{
				reverse:      false,
				pattern:      "000",
				pre:          "",
				post:         "",
				help:         false,
				version:      false,
				init:         1,
				filePatterns: []string{"*.txt"},
			},
			wantErr: false,
		},
		{
			name: "custom init number",
			args: []string{"--init=100", "*.txt"},
			want: &config{
				reverse:      false,
				pattern:      defaultPattern,
				pre:          "",
				post:         "",
				help:         false,
				version:      false,
				init:         100,
				filePatterns: []string{"*.txt"},
			},
			wantErr: false,
		},
		{
			name: "multiple patterns",
			args: []string{"*.jpg", "*.png"},
			want: &config{
				reverse:      false,
				pattern:      defaultPattern,
				pre:          "",
				post:         "",
				help:         false,
				version:      false,
				init:         1,
				filePatterns: []string{"*.jpg", "*.png"},
			},
			wantErr: false,
		},
		{
			name:        "no file patterns",
			args:        []string{},
			want:        nil,
			wantErr:     true,
			errContains: "file pattern required",
		},
		{
			name: "help flag",
			args: []string{"--help", "*.txt"},
			want: &config{
				reverse:      false,
				pattern:      defaultPattern,
				pre:          "",
				post:         "",
				help:         true,
				version:      false,
				init:         1,
				filePatterns: []string{"*.txt"},
			},
			wantErr: false,
		},
		{
			name: "version flag",
			args: []string{"--version", "*.txt"},
			want: &config{
				reverse:      false,
				pattern:      defaultPattern,
				pre:          "",
				post:         "",
				help:         false,
				version:      true,
				init:         1,
				filePatterns: []string{"*.txt"},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseFlags("renby", tt.args)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseFlags() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				if err == nil || !strings.HasPrefix(err.Error(), tt.errContains) {
					t.Errorf("parseFlags() error = %v, want error containing %v", err, tt.errContains)
				}
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("parseFlags() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestProcessFilePatterns(t *testing.T) {
	// Create temporary test files
	tmpDir := t.TempDir()

	files := []string{"test1.txt", "test2.txt", "test.jpg"}
	for _, f := range files {
		path := filepath.Join(tmpDir, f)
		if err := os.WriteFile(path, []byte("test"), 0644); err != nil {
			t.Fatal(err)
		}
	}

	// Save current working directory and change to temp dir
	pwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(pwd)

	tests := []struct {
		name        string
		patterns    []string
		want        int
		wantErr     bool
		errContains string
	}{
		{
			name:     "match txt files",
			patterns: []string{"*.txt"},
			want:     2,
			wantErr:  false,
		},
		{
			name:     "match jpg files",
			patterns: []string{"*.jpg"},
			want:     1,
			wantErr:  false,
		},
		{
			name:     "multiple patterns",
			patterns: []string{"*.txt", "*.jpg"},
			want:     3,
			wantErr:  false,
		},
		{
			name:        "invalid pattern",
			patterns:    []string{"["},
			want:        0,
			wantErr:     true,
			errContains: "invalid file pattern",
		},
		{
			name:        "no matches",
			patterns:    []string{"*.nonexistent"},
			want:        0,
			wantErr:     true,
			errContains: "no files found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := processFilePatterns(tt.patterns)
			if (err != nil) != tt.wantErr {
				t.Errorf("processFilePatterns() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				if err == nil || !strings.HasPrefix(err.Error(), tt.errContains) {
					t.Errorf("processFilePatterns() error = %v, want error containing %v", err, tt.errContains)
				}
				return
			}
			if len(got) != tt.want {
				t.Errorf("processFilePatterns() got %d files, want %d", len(got), tt.want)
			}
		})
	}
}
