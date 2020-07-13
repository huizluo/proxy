package utils


import (
	"strconv"
	"sync/atomic"
	"time"
)

var (
	// lastTime is used to cache time
	lastTime atomic.Value
)

// timeCache is used to reduce format
type timeCache struct {
	t int64
	s string
}

// CacheTime returns a time cache in seconds.
// we use a cache to reduce the format
func CacheTime() string {
	var s string
	t := time.Now()
	nano := t.UnixNano()
	now := nano / 1e9
	value := lastTime.Load()
	if value != nil {
		last := value.(*timeCache)
		if now <= last.t {
			s = last.s
		}
	}
	if s == "" {
		s = t.Format("2006-01-02 15:04:05")
		lastTime.Store(&timeCache{now, s})
	}
	mi := nano % 1e9 / 1e6
	s = s + "," + strconv.Itoa(int(mi))
	return s
}
