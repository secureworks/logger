package wrappers

import (
	"sync"
	"time"

	"github.com/secureworks/logger/log"
)

// TODO(PH)
type ConcurrentEntry struct {
	mux   *sync.Mutex
	entry log.Entry
}

// TODO(PH)
func ConcurrencySafe(entry log.Entry) log.Entry {
	return &ConcurrentEntry{mux: new(sync.Mutex), entry: entry}
}

func (c *ConcurrentEntry) Async() log.Entry {
	c.mux.Lock()
	defer c.mux.Unlock()
	return c.entry.Async()
}

func (c *ConcurrentEntry) IsLevelEnabled(level log.Level) bool {
	c.mux.Lock()
	defer c.mux.Unlock()
	return c.entry.IsLevelEnabled(level)
}

func (c *ConcurrentEntry) Send() {
	c.mux.Lock()
	defer c.mux.Unlock()
	c.entry.Send()
}

func (c *ConcurrentEntry) Msgf(format string, v ...any) {
	c.mux.Lock()
	defer c.mux.Unlock()
	c.entry.Msgf(format, v...)
}

func (c *ConcurrentEntry) Msg(v any) {
	c.mux.Lock()
	defer c.mux.Unlock()
	c.entry.Msg(v)
}

func (c *ConcurrentEntry) Caller(skip ...int) log.Entry {
	c.mux.Lock()
	defer c.mux.Unlock()
	return c.entry.Caller(skip...)
}

func (c *ConcurrentEntry) WithErrors(errs ...error) log.Entry {
	c.mux.Lock()
	defer c.mux.Unlock()
	return c.entry.WithErrors(errs...)
}

func (c *ConcurrentEntry) WithField(key string, value any) log.Entry {
	c.mux.Lock()
	defer c.mux.Unlock()
	return c.entry.WithField(key, value)
}

func (c *ConcurrentEntry) WithFields(fields map[string]any) log.Entry {
	c.mux.Lock()
	defer c.mux.Unlock()
	return c.entry.WithFields(fields)
}

func (c *ConcurrentEntry) WithStr(key string, strings ...string) log.Entry {
	c.mux.Lock()
	defer c.mux.Unlock()
	return c.entry.WithStr(key, strings...)
}

func (c *ConcurrentEntry) WithBool(key string, bools ...bool) log.Entry {
	c.mux.Lock()
	defer c.mux.Unlock()
	return c.entry.WithBool(key, bools...)
}

func (c *ConcurrentEntry) WithTime(key string, times ...time.Time) log.Entry {
	c.mux.Lock()
	defer c.mux.Unlock()
	return c.entry.WithTime(key)
}

func (c *ConcurrentEntry) WithDur(key string, durations ...time.Duration) log.Entry {
	c.mux.Lock()
	defer c.mux.Unlock()
	return c.entry.WithDur(key, durations...)
}

func (c *ConcurrentEntry) WithInt(key string, ints ...int) log.Entry {
	c.mux.Lock()
	defer c.mux.Unlock()
	return c.entry.WithInt(key, ints...)
}

func (c *ConcurrentEntry) WithUint(key string, uints ...uint) log.Entry {
	c.mux.Lock()
	defer c.mux.Unlock()
	return c.entry.WithUint(key, uints...)
}

func (c *ConcurrentEntry) Trace() log.Entry {
	c.mux.Lock()
	defer c.mux.Unlock()
	return c.entry.Trace()
}

func (c *ConcurrentEntry) Debug() log.Entry {
	c.mux.Lock()
	defer c.mux.Unlock()
	return c.entry.Debug()
}

func (c *ConcurrentEntry) Info() log.Entry {
	c.mux.Lock()
	defer c.mux.Unlock()
	return c.entry.Info()
}

func (c *ConcurrentEntry) Warn() log.Entry {
	c.mux.Lock()
	defer c.mux.Unlock()
	return c.entry.Warn()
}

func (c *ConcurrentEntry) Error() log.Entry {
	c.mux.Lock()
	defer c.mux.Unlock()
	return c.entry.Error()
}

func (c *ConcurrentEntry) Panic() log.Entry {
	c.mux.Lock()
	defer c.mux.Unlock()
	return c.entry.Panic()
}

func (c *ConcurrentEntry) UnwrapEntry() log.Entry {
	c.mux.Lock()
	defer c.mux.Unlock()
	return c.entry
}
