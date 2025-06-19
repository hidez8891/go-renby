package ostime

import "time"

type OsTime struct {
	CreationTime     time.Time
	ModificationTime time.Time
	AccessTime       time.Time
}
