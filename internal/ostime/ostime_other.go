//go:build !windows && !linux

package ostime

import "os"

func GetOsTime(info os.FileInfo) OsTime {
	return OsTime{
		CreationTime:     info.ModTime(), // Fallback for systems without specific creation time
		ModificationTime: info.ModTime(),
		AccessTime:       info.ModTime(), // Fallback for systems without specific access time
	}
}
