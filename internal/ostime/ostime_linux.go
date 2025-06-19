//go:build linux

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
		CreationTime:     time.Unix(int64(stat.Ctim.Sec), int64(stat.Ctim.Nsec)),
		ModificationTime: time.Unix(int64(stat.Mtim.Sec), int64(stat.Mtim.Nsec)),
		AccessTime:       time.Unix(int64(stat.Atim.Sec), int64(stat.Atim.Nsec)),
	}
}
