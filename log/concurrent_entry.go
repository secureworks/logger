package log

import (
	"sync"
	"time"
)

// TODO(PH)
type ConcurrentEntry struct {
	mux   *sync.Mutex
	entry Entry
}

// TODO(PH)
func ConcurrencySafe(entry Entry) Entry {
	return &ConcurrentEntry{mux: new(sync.Mutex), entry: entry}
}

func (c *ConcurrentEntry) Async() Entry {
	c.mux.Lock()
	defer c.mux.Unlock()
	return c.entry.Async()
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

func (c *ConcurrentEntry) Caller(skip ...int) Entry {
	c.mux.Lock()
	defer c.mux.Unlock()
	return c.entry.Caller(skip...)
}

func (c *ConcurrentEntry) WithErrors(errs ...error) Entry {
	c.mux.Lock()
	defer c.mux.Unlock()
	return c.entry.WithErrors(errs...)
}

func (c *ConcurrentEntry) WithField(key string, value any) Entry {
	c.mux.Lock()
	defer c.mux.Unlock()
	return c.entry.WithField(key, value)
}

func (c *ConcurrentEntry) WithFields(fields map[string]any) Entry {
	c.mux.Lock()
	defer c.mux.Unlock()
	return c.entry.WithFields(fields)
}

func (c *ConcurrentEntry) WithStr(key string, strings ...string) Entry {
	c.mux.Lock()
	defer c.mux.Unlock()
	return c.entry.WithStr(key, strings...)
}

func (c *ConcurrentEntry) WithBool(key string, bools ...bool) Entry {
	c.mux.Lock()
	defer c.mux.Unlock()
	return c.entry.WithBool(key, bools...)
}

func (c *ConcurrentEntry) WithTime(key string, times ...time.Time) Entry {
	c.mux.Lock()
	defer c.mux.Unlock()
	return c.entry.WithTime(key)
}

func (c *ConcurrentEntry) WithDur(key string, durations ...time.Duration) Entry {
	c.mux.Lock()
	defer c.mux.Unlock()
	return c.entry.WithDur(key, durations...)
}

func (c *ConcurrentEntry) WithInt(key string, ints ...int) Entry {
	c.mux.Lock()
	defer c.mux.Unlock()
	return c.entry.WithInt(key, ints...)
}

func (c *ConcurrentEntry) WithUint(key string, uints ...uint) Entry {
	c.mux.Lock()
	defer c.mux.Unlock()
	return c.entry.WithUint(key, uints...)
}

func (c *ConcurrentEntry) Trace() Entry {
	c.mux.Lock()
	defer c.mux.Unlock()
	return c.entry.Trace()
}

func (c *ConcurrentEntry) Debug() Entry {
	c.mux.Lock()
	defer c.mux.Unlock()
	return c.entry.Debug()
}

func (c *ConcurrentEntry) Info() Entry {
	c.mux.Lock()
	defer c.mux.Unlock()
	return c.entry.Info()
}

func (c *ConcurrentEntry) Warn() Entry {
	c.mux.Lock()
	defer c.mux.Unlock()
	return c.entry.Warn()
}

func (c *ConcurrentEntry) Error() Entry {
	c.mux.Lock()
	defer c.mux.Unlock()
	return c.entry.Error()
}

func (c *ConcurrentEntry) Panic() Entry {
	c.mux.Lock()
	defer c.mux.Unlock()
	return c.entry.Panic()
}

func (c *ConcurrentEntry) UnwrapEntry() Entry {
	c.mux.Lock()
	defer c.mux.Unlock()
	return c.entry
}
