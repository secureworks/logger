//go:build !safe
// +build !safe

package zerolog

import (
	"reflect"
	"unsafe"

	"github.com/rs/zerolog"
)

// Since we want the zerolog logger implementation to support both
// changing levels and "Disabled", this is needed in order to retain
// Zerolog's performance for high event counts: if we don't write it
// won't be reused. See also:
//   - https://github.com/rs/zerolog/pull/255
//   - https://github.com/rs/zerolog/blob/7825d863376faee2723fc99c061c538bd80812c8/event.go#L79
//
//go:linkname putEvent github.com/rs/zerolog.putEvent
func putEvent(ent *zerolog.Event)

// // ptrSize is the size of a pointer: it identifies 32bit vs 64bit systems.
// const ptrSize = 4 << (^uintptr(0) >> 63)

var (
	// sliceHSize = reflect.TypeOf([]byte{}).Size() // reflect.SliceHeader{}
	// lwSize     = 2*ptrSize

	// This method is "safer" if level is moved. No need to check bool,
	// tests will fail if this changes.
	lvlField, _ = reflect.TypeOf(zerolog.Event{}).FieldByName("level")
)

// This and the above code is for changing the value of the private
// 'level' field in the zerolog.Event struct. Please see the below issue
// for more information:
//   - https://github.com/rs/zerolog/issues/408
//
// The hope is that this code is temporary and remains optional.
//
// Note: this is the zerolog.Event as of 1.26.1:
//
//	type Event struct {
//	    // Size of slice header.
//	    buf []byte
//
//	    // Size of pointer/interface we'll use reflect to be safe(er).
//	    w  LevelWriter
//
//	    // The field we want to change.
//	    level Level
//
//	    done      func(msg string)
//	    stack     bool
//	    ch        []Hook
//	    skipFrame int
//	}
func changeEventLevel(ent *zerolog.Event, lvl zerolog.Level) {
	levelField := (*zerolog.Level)(unsafe.Pointer((uintptr(unsafe.Pointer(ent)) + lvlField.Offset)))
	*levelField = lvl
}
