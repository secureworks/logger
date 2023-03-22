package log

import (
	"fmt"
	"reflect"
)

// Option is a function type that accepts an interface value and returns
// an error. Use it to handle applying settings directly to the logger
// implementation that are not covered by the Config.
//
// See the example for the basic steps to implement an Option.
type Option func(any) error

// CustomOption is used to pass settings directly to the logger
// implementation that are not covered by the Config.
//
// It takes the case-sensitive name of a method on the logger
// implementation as well as a value and returns an Option. This will
// only work in the case that the logger implementation has a method
// that accepts a single value; if several values are needed as input
// then val should be a function that accepts no input and returns
// values to be used as input to the named method. A nil value is valid
// so long as the named method expects nil input or no input.
//
// If the given value is a function and returns an error as its only or
// last value it will be checked and returned when it is not nil,
// otherwise remaining values are fed to the named method.
//
// If the named method returns an instance of itself it will be set back
// as the new UnderlyingLogger.
//
// If the named method returns an error that error will be checked and
// returned.
func CustomOption(name string, val any) Option {
	if name == "" {
		return noopOption
	}

	valFunc, err := getReflectVals(val)
	if err != nil {
		return errOption(err)
	}

	return func(topLogger any) (err error) {
		ul, ok := topLogger.(UnderlyingLogger)
		if !ok {
			return fmt.Errorf("log: Logger type (%T) does not support the UnderlyingLogger interface", topLogger)
		}

		// Handle panics during CustomOption application.
		defer func() {
			pv := recover()

			if pv == nil || err != nil {
				return
			}
			if e, ok := pv.(error); ok {
				err = e
			} else {
				err = fmt.Errorf("log: Panic caught in CustomOption for %s: %v", name, pv)
			}
		}()

		logger := ul.GetLogger()
		if logger == nil {
			return
		}
		logVal := reflect.ValueOf(logger)
		if !logVal.IsValid() {
			return
		}
		methodVal := logVal.MethodByName(name)
		if !methodVal.IsValid() {
			return
		}

		var wasError bool
		vals := valFunc()

		// Check if last val is error or error that is nil.
		if l := len(vals); l > 0 {
			err, wasError = valueToError(vals[l-1])
			if err != nil {
				return
			}

			// Remove value.
			if wasError {
				vals = vals[:l-1]
			}
		}

		// From this point we have the method we want to call and the values
		// with which to call it. We could check if each input matches what
		// is expected but instead we'll just call the method and rely on
		// the defer recover above to stop us from calling something wrong.
		out := methodVal.Call(vals)
		le := len(out)
		if le == 0 {
			return
		}

		err, wasError = valueToError(out[le-1])
		if err != nil {
			return
		}
		if wasError {
			out = out[:le-1]
			le = len(out)
		}
		if le == 0 {
			return
		}

		// If one of the remaining types in the output is the same as the
		// underlying logger, set it back as the underlying logger. This is
		// common in methods/funcs that chain configuration.
		logType := logVal.Type()
		for _, val := range out {
			if !val.IsValid() {
				continue
			}

			//	Since logType is an interface this never seems correct, but
			//	according to https://golang.org/pkg/reflect/#Type:
			//
			//	  Type values are comparable, such as with the == operator, so
			//	  they can be used as map keys. Two Type values are equal if
			//	  they represent identical types.
			if val.Type() == logType {
				ul.SetLogger(val.Interface())
				break
			}

			// With the expectation that logger implementations may type check
			// and allow a value they can reference when setting it.
			if val.Type() == logType.Elem() {
				ul.SetLogger(val.Interface())
				break
			}
		}

		return
	}
}

func noopOption(_ any) error { return nil }

func errOption(err error) Option {
	return func(_ any) error {
		return err
	}
}

// Identify and validate the CustomOption value input, and normalize
// it into a function that returns a slice of reflect.Values.
func getReflectVals(val any) (func() []reflect.Value, error) {
	reflval := reflect.ValueOf(val)
	if !reflval.IsValid() {
		return func() []reflect.Value { return []reflect.Value{} }, nil
	}

	typ := reflval.Type()
	if typ.Kind() != reflect.Func {
		return func() []reflect.Value { return []reflect.Value{reflval} }, nil
	}

	// We know it's a func, check to make sure it doesn't take args.
	if typ.NumIn() > 0 {
		name := typ.Name()
		if name == "" {
			name = "anon func"
		}
		return nil, fmt.Errorf("log: Function value (%s) expects inputs", name)
	}

	return func() []reflect.Value { return reflval.Call([]reflect.Value{}) }, nil
}

var errInterface = reflect.TypeOf((*error)(nil)).Elem()

// Check and convert a reflect.Value to an error if appropriate.
// nolint
func valueToError(val reflect.Value) (err error, wasError bool) {
	if !val.IsValid() {
		return
	}

	if val.Kind() != reflect.Interface || !val.Type().Implements(errInterface) {
		return
	}

	wasError = true
	if val.IsNil() {
		return
	}

	err = val.Elem().Interface().(error)
	return
}
