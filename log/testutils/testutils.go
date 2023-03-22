package testutils

import (
	"bytes"
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/secureworks/logger/log"
)

// NewConfigWithBuffer generates a default testing config (log.Level is
// set to log.INFO and EnableErrStack is true) and a linked output
// buffer that the logger writes too.
func NewConfigWithBuffer(t *testing.T, logLevel log.Level) (*log.Config, *bytes.Buffer) {
	t.Helper()

	buf := make([]byte, 0, 100)
	out := bytes.NewBuffer(buf)
	config := log.DefaultConfigWithEnvLookup(func(envVar string) string {
		return os.Getenv(envVar)
	})
	config.Level = logLevel
	config.Output = out
	config.EnableErrStack = true

	return config, out
}

// AssertTrue is a semantic test assertion for object truthiness.
func AssertTrue(t *testing.T, object bool) {
	t.Helper()
	if !object {
		t.Errorf("is not true")
	}
}

// AssertFalse is a semantic test assertion for object truthiness.
func AssertFalse(t *testing.T, object bool) {
	t.Helper()
	if object {
		t.Errorf("is not false")
	}
}

// AssertEqual is a semantic test assertion for object equality.
func AssertEqual(t *testing.T, expected interface{}, actual interface{}) {
	t.Helper()
	assertEquality(t, expected, actual, true)
}

// AssertNotEqual is a semantic test assertion for object equality.
func AssertNotEqual(t *testing.T, expected interface{}, actual interface{}) {
	t.Helper()
	assertEquality(t, expected, actual, false)
}

// AssertNearEqual is a semantic test assertion for numeric accuracy.
func AssertNearEqual(t *testing.T, expected int64, actual int64, delta int64) {
	t.Helper()

	diff := actual - expected
	if diff < -delta || diff > delta {
		t.Errorf(
			"is not within delta (%d):\nexpected: %d\nactual: %d\n",
			delta,
			expected,
			actual,
		)
	}
}

// AssertSame is a semantic test assertion for referential equality.
func AssertSame(t *testing.T, expected interface{}, actual interface{}) {
	t.Helper()
	assertSameness(t, expected, actual, true)
}

// AssertNotSame is a semantic test assertion for referential equality.
func AssertNotSame(t *testing.T, expected interface{}, actual interface{}) {
	t.Helper()
	assertSameness(t, expected, actual, false)
}

// AssertNil is a semantic test assertion for nility.
func AssertNil(t *testing.T, object interface{}) {
	t.Helper()
	assertNility(t, object, true)
}

// AssertNotNil is a semantic test assertion for nility.
func AssertNotNil(t *testing.T, object interface{}) {
	t.Helper()
	assertNility(t, object, false)
}

// AssertStringContains is a semantic test assertion for partial string
// matching.
func AssertStringContains(t *testing.T, expectedContained string, actualContaining string) {
	t.Helper()
	if !strings.Contains(actualContaining, expectedContained) {
		t.Errorf(
			"does not contain:\nexpected to contain: %s\nactual: %s\n",
			strings.Trim(expectedContained, "\n"),
			strings.Trim(actualContaining, "\n"),
		)
	}
}

// AssertNotPanics is a semantic test assertion that a function does not
// panic.
func AssertNotPanics(t *testing.T, fn func()) {
	t.Helper()

	didPanic := true
	defer func() {
		if didPanic {
			t.Errorf("did panic")
		}
	}()

	fn()
	didPanic = false
}

// NOTE(PH): does not handle bytes well, update if we need to check
// them.
func assertEquality(t *testing.T, expected interface{}, actual interface{}, wantEqual bool) {
	t.Helper()

	if expected == nil && actual == nil && !wantEqual {
		t.Errorf("is equal:\nexpected: %s\nactual: %s\n", expected, actual)
	}

	isEqual := reflect.DeepEqual(expected, actual)
	if wantEqual && !isEqual {
		t.Errorf("not equal:\nexpected: %s\nactual: %s\n", expected, actual)
	}
	if !wantEqual && isEqual {
		t.Errorf("is equal:\nexpected: %s\nactual: %s\n", expected, actual)
	}
}

func assertSameness(t *testing.T, expected interface{}, actual interface{}, wantSame bool) {
	t.Helper()

	isSame := false
	if expected == actual {
		isSame = true
		expectedPtr := reflect.ValueOf(expected)
		actualPtr := reflect.ValueOf(actual)
		if expectedPtr.Kind() != reflect.Ptr || actualPtr.Kind() != reflect.Ptr {
			isSame = false
		}

		expectedType := reflect.TypeOf(expected)
		actualType := reflect.TypeOf(actual)
		if isSame && (expectedType != actualType) {
			isSame = false
		}
	}
	if wantSame && !isSame {
		t.Errorf("not same:\nexpected: %s\nactual: %s\n", expected, actual)
	}
	if !wantSame && isSame {
		t.Errorf("is same:\nexpected: %s\nactual: %s\n", expected, actual)
	}
}

func assertNility(t *testing.T, object interface{}, wantNil bool) {
	t.Helper()

	isNil := object == nil
	if !isNil {
		value := reflect.ValueOf(object)
		isNilable := false
		switch value.Kind() {
		case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Ptr, reflect.Slice:
			isNilable = true
		default:
		}
		if isNilable && value.IsNil() {
			isNil = true
		}
	}
	if wantNil && !isNil {
		t.Errorf("not nil: %s\n", object)
	}
	if !wantNil && isNil {
		t.Errorf("is nil: %s\n", object)
	}
}
