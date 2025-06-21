# go-renby (rename by)

A CLI tool to rename files sequentially based on file information (file size,
creation time, etc.)

## Installation

```bash
go install github.com/hidez8891/go-renby/cmd/renby@latest
```

## Usage

```bash
renby SUBCOMMAND [OPTIONS] FILES
```

### Subcommands

- `ctime`: Sort files by creation time
- `mtime`: Sort files by modification time
- `atime`: Sort files by last access time
- `size`: Sort files by file size

### Options

- `-r, --reverse`: Sort in descending order (default: ascending)
- `-p, --pattern=STRING`: Specify renaming pattern
  - '0': Zero-padded decimal numbers
  - 'x': Zero-padded lowercase hexadecimal numbers
  - Default: '000000'
- `--init=NUMBER`: Initial number for renaming pattern (default: 1)
- `--pre=STRING`: Prefix string for renamed files (default: '')
- `--post=STRING`: Suffix string for renamed files (default: '')
- `--help`: Show help message
- `--version`: Show version number

### Examples

1. Rename PNG files in order of creation time:

```bash
$ renby ctime *.png
00001.png
00002.png
00003.png
```

2. Rename JPG files in reverse order of file size with prefix and suffix:

```bash
$ renby size -r --pre=img --post=test *.jpg
img00001test.jpg
img00002test.jpg
img00003test.jpg
```

3. Rename TXT files by size using 3-digit hexadecimal numbers:

```bash
$ renby size -p=xxx *.txt
001.txt
002.txt
003.txt
...
00b.txt
```

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file
for details.
