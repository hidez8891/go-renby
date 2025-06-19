//go:build windows

package ostime

import (
	"os"
	"syscall"
	"time"
)

func GetOsTime(info os.FileInfo) OsTime {
	stat, ok := info.Sys().(*syscall.Win32FileAttributeData)
	if !ok {
		panic("GetOsTime: os.FileInfo.Sys() is not a Win32FileAttributeData")
	}

	return OsTime{
		CreationTime:     time.Unix(0, stat.CreationTime.Nanoseconds()),
		ModificationTime: time.Unix(0, stat.LastWriteTime.Nanoseconds()),
		AccessTime:       time.Unix(0, stat.LastAccessTime.Nanoseconds()),
	}
}
