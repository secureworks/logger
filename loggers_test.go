package logger

import (
	"bytes"
	"os"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/pkg/errors"

	"github.com/secureworks/logger/internal/testutils"
	"github.com/secureworks/logger/log"
	_ "github.com/secureworks/logger/logrus"
	_ "github.com/secureworks/logger/testlogger"
	_ "github.com/secureworks/logger/zerolog"
)

func TestLoggers(t *testing.T) {
	tests := []struct {
		driverName string
	}{
		{"testlogger"},
		{"zerolog"},
		{"logrus"},
	}
	for _, tc := range tests {
		t.Run(tc.driverName, func(t *testing.T) {
			LoggerOutput(t, tc.driverName)
			CallerOutput(t, tc.driverName)
			// ErrorStacks(t, tc.driverName) FIXME
			// DurationFormatting(t, tc.driverName) FIXME
			// TimeFormatting(t, tc.driverName) FIXME
			// CallSendNTimes(t, tc.driverName) FIXME
			// WithNilValues(t, tc.driverName) FIXME
		})
	}
}

func initLogger(t *testing.T, driverName string) (*bytes.Buffer, log.Logger) {
	t.Helper()

	buf := new(bytes.Buffer)
	config := log.DefaultConfig(os.Getenv)
	config.Level = log.INFO
	// config.EnableErrStack = true
	config.Output = buf
	logger, err := log.Open(driverName, config)
	testutils.AssertNil(t, err)
	testutils.AssertNotNil(t, logger)

	return buf, logger
}

func LoggerOutput(t *testing.T, driverName string) {
	buf, logger := initLogger(t, driverName)

	dur := 2*time.Hour + 22*time.Minute + 4*time.Second + 531*time.Millisecond + 22*time.Nanosecond
	tim := time.Date(2011, 9, 1, 21, 3, 40, 4123412, time.UTC)
	sct := struct {
		Example string
		Value   int
	}{"value of 20", 20}
	err1 := errors.WithStack(errors.New("err with stack"))

	entry := logger.Entry(log.ERROR)
	entry.
		WithStr("withStr", "s1", "s2").
		WithBool("with-bool", true, false).
		WithDur("with_dur", 22*time.Millisecond, dur).
		WithInt("WithInt", -35).
		WithUint("WITH_UINT", 402, 8).
		WithTime("with time", tim).
		WithError(errors.New("basic error"), err1).
		WithField("example-field-one", "str").
		WithField("example-field-two", 421.644).
		WithField("example-field-three", "overwrite me!").
		WithField("example-field-four", tim).
		WithField("example-field-five", err1).
		WithField("example-field-six", "overwrite me!").
		WithFields(map[string]interface{}{
			"example-field-three": dur,
			"example-field-six":   sct,
		}).
		// Caller(0). FIXME: add to test
		Msgf("this is the message value %v", sct)

	testutils.AssertEqual(t, `{
    "WITH_UINT": [402,8],
    "WithInt": -35,
    "error": ["basic error","err with stack"],
    "example-field-five": "err with stack",
    "example-field-four": "2011-09-01T21:03:40.004123412Z",
    "example-field-one": "str",
    "example-field-six": {"Example":"value of 20","Value":20},
    "example-field-three": 8524531000022,
    "example-field-two": 421.644,
    "level": "info",
    "message": "this is the message value {value of 20 20}",
    "with time": "2011-09-01T21:03:40Z.004123412Z",
    "with-bool": [true,false],
    "withStr": ["s1","s2"],
    "with_dur": [22000000,8524531000022]
}`,
		// Make the output stable and diffable:
		reformatJSON(buf.String()))
}

func CallerOutput(t *testing.T, driverName string) {
	buf, logger := initLogger(t, driverName)

	entry := logger.Entry(log.ERROR)
	entry.Caller(0)
	entry.Caller(1)
	entry.Send()

	if out := reformatJSON(buf.String()); !strings.Contains(out, "loggers_test.go:*108") {
		t.Fatalf("bad output: entry %q does not include \"loggers_test.go:108\"", out)
	}
	if out := reformatJSON(buf.String()); !strings.Contains(out, "loggers_test.go:31") {
		t.Fatalf("bad output: entry %q does not include \"loggers_test.go:31\"", out)
	}
}

func reformatJSON(in string) string {
	out := new(bytes.Buffer)
	out.Grow(len(in))

	var isField, isValue, isListValue, isObjectValue, isEscape bool
	var indent, ldepth, odepth int
	for _, char := range in {
		if isEscape {
			isEscape = false
			out.WriteRune(char)
			continue
		}
		if '\\' == char {
			isEscape = true
			out.WriteRune(char)
			continue
		}
		if isField {
			out.WriteRune(char)
			if ':' == char {
				isField = false
				isValue = true
				out.WriteString(" ")
			}
			continue
		}
		if isValue && '}' == char && odepth == 0 {
			isValue = false
		}
		if isValue {
			out.WriteRune(char)
			if '[' == char {
				ldepth++
				isListValue = true
			}
			if ']' == char {
				ldepth--
				if ldepth == 0 {
					isListValue = false
				}
			}
			if '{' == char {
				odepth++
				isObjectValue = true
			}
			if '}' == char {
				odepth--
				if odepth == 0 {
					isObjectValue = false
				}
			}
			if ',' == char && !isListValue && !isObjectValue {
				isValue = false
				out.WriteString("\n")
				continue
			}
			continue
		}
		if '{' == char {
			indent++
			out.WriteRune(char)
			out.WriteString("\n")
			continue
		}
		if '}' == char {
			indent--
			out.WriteString("\n" + strings.Repeat("    ", indent))
			out.WriteRune(char)
			continue
		}
		if '"' == char {
			isField = true
			out.WriteString(strings.Repeat("    ", indent))
			out.WriteRune(char)
		}
	}
	formatted := out.String()
	lines := strings.Split(formatted, "\n")
	lines = lines[1 : len(lines)-1]
	for i := 0; i < len(lines); i++ {
		lines[i] = strings.TrimRight(lines[i], "\n ,")
	}
	sort.Strings(lines)
	return "{\n" + strings.Join(lines, ",\n") + "\n}"
}
