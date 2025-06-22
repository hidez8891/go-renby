//go:build darwin

package ostime

import (
	"os"
	"syscall"
	"time"
)

func GetOsTime(info os.FileInfo) OsTime {
	stat, ok := info.Sys().(*syscall.Stat_t)
	if !ok {
		panic("GetOsTime: os.FileInfo.Sys() is not a Stat_t")
	}

	return OsTime{
		CreationTime:     time.Unix(int64(stat.Ctimespec.Sec), int64(stat.Ctimespec.Nsec)),
		ModificationTime: time.Unix(int64(stat.Mtimespec.Sec), int64(stat.Mtimespec.Nsec)),
		AccessTime:       time.Unix(int64(stat.Atimespec.Sec), int64(stat.Atimespec.Nsec)),
	}
}
