package logger_test

import (
	"errors"
	"testing"

	"github.com/secureworks/logger/internal/testutils"
	"github.com/secureworks/logger/log"
)

func TestOptions_CustomOption(t *testing.T) {
	t.Run("single value", func(t *testing.T) {
		w := newTestLogger()
		const a = "A set"

		err := log.CustomOption("SetA", a)(w)
		testutils.AssertNil(t, err)
		testutils.AssertEqual(t, a, w.ul.A)
	})

	t.Run("function with single value", func(t *testing.T) {
		w := newTestLogger()
		const a = "A set"

		err := log.CustomOption("SetA", func() string { return a })(w)
		testutils.AssertNil(t, err)
		testutils.AssertEqual(t, a, w.ul.A)
	})

	t.Run("function with multiple values including nil error", func(t *testing.T) {
		w := newTestLogger()
		orig := w.ul
		a, b := "A", "B"

		opt := log.CustomOption("WithAB", func() (string, string, error) { return a, b, nil })
		err := opt(w)
		testutils.AssertNil(t, err)
		testutils.AssertNotSame(t, orig, w.ul)
		testutils.AssertEqual(t, a, w.ul.A)
		testutils.AssertEqual(t, b, w.ul.B)
		testutils.AssertEqual(t, "", orig.A)
		testutils.AssertEqual(t, "", orig.B)
	})

	t.Run("malformed function causes error", func(t *testing.T) {
		w := newTestLogger()
		b := "B"

		// Pass a func that accepts a value, which we don't support.
		err := log.CustomOption("SetB", func(i int) string { return b })(w)
		testutils.AssertNotNil(t, err)
		testutils.AssertNotEqual(t, b, w.ul.B)
	})

	t.Run("method returns nil error value", func(t *testing.T) {
		w := newTestLogger()
		w.ul.C = "reflect me"
		orig := w.ul

		err := log.CustomOption("ChainClearCNil", nil)(w)
		testutils.AssertNil(t, err)
		testutils.AssertSame(t, orig, w.ul)
		testutils.AssertEqual(t, "", w.ul.C)
	})

	t.Run("method returns non-nil error value", func(t *testing.T) {
		w := newTestLogger()
		orig := w.ul
		b := "B"

		err := log.CustomOption("ChainBFailure", func() string { return b })(w)
		testutils.AssertEqual(t, errTheSentinel, err)
		testutils.AssertSame(t, orig, w.ul)
		testutils.AssertEqual(t, "", w.ul.B)
	})

	t.Run("method that is chainable updates logger", func(t *testing.T) {
		w := newTestLogger()
		w.ul.A = "old a"
		orig := w.ul

		err := log.CustomOption("WithA", "new a")(w)
		testutils.AssertNil(t, err)
		testutils.AssertNotSame(t, orig, w.ul)
		testutils.AssertEqual(t, "old a", orig.A)
		testutils.AssertEqual(t, "new a", w.ul.A)
	})

	t.Run("method that is chainable (non-pointer) updates logger", func(t *testing.T) {
		w := newTestLogger()
		w.ul.A = "old a"
		orig := w.ul

		err := log.CustomOption("WithAVal", "new a")(w)
		testutils.AssertNil(t, err)
		testutils.AssertNotSame(t, orig, w.ul)
		testutils.AssertEqual(t, "old a", orig.A)
		testutils.AssertEqual(t, "new a", w.ul.A)
	})

	t.Run("recover from panic", func(t *testing.T) {
		w := newTestLogger()
		orig := w.ul

		// Pass func that returns value that is not appropriate for the
		// reflected method.
		err := log.CustomOption("WithA", func() int { return 42 })(w)
		testutils.AssertNotNil(t, err)
		testutils.AssertSame(t, orig, w.ul)
	})
}

// A simple UnderlyingLogger implementation.

type wrapper struct {
	ul *ulLogger
}

func newTestLogger() *wrapper {
	return &wrapper{new(ulLogger)}
}

func (w *wrapper) GetLogger() interface{} {
	return w.ul
}

func (w *wrapper) SetLogger(iface interface{}) {
	if ul, ok := iface.(*ulLogger); ok {
		w.ul = ul
	}
	if ul, ok := iface.(ulLogger); ok {
		w.ul = &ul
	}
}

type ulLogger struct {
	A, B, C string
}

func (ul *ulLogger) SetA(a string) {
	ul.A = a
}

func (ul *ulLogger) SetB(b string) string {
	ul.B = b
	return b
}

func (ul *ulLogger) WithAB(a, b string) (*ulLogger, error) {
	cpy := *ul
	cpy.SetA(a)
	cpy.SetB(b)

	return &cpy, nil
}

func (ul *ulLogger) WithA(a string) *ulLogger {
	cpy := *ul
	cpy.SetA(a)
	return &cpy
}

func (ul *ulLogger) WithAVal(a string) ulLogger {
	cpy := *ul
	cpy.SetA(a)
	return cpy
}

func (ul *ulLogger) ChainClearCNil() (*ulLogger, error) {
	ul.C = ""
	return ul, nil
}

var errTheSentinel = errors.New("Oh noooo")

func (ul *ulLogger) ChainBFailure(_ string) (*ulLogger, error) {
	return nil, errTheSentinel
}
