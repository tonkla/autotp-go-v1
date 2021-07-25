package helper

import "time"

func Now13() int64 {
	return time.Now().UnixNano() / 1e6
}
