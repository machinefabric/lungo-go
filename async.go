package lungo

/*
#include "lungo.h"
*/
import "C"

import (
	"context"
	"fmt"
)

// Kinds of steps of an async program.
const (
	stepDone = 0
	stepCall = 1
)

// DriveAsync runs an async program to its end. status and out are what the program's entry
// point returned: its first step. Each step either ends the program with a value of value, or
// asks the host to perform an operation of op: perform performs it (with ctx) and writes the
// answer into w, and the program resumes with it.
//
// When perform fails, panics, or ctx ends, the program is abandoned: the runtime releases it,
// and DriveAsync returns the error (or panics again). A resumption is never used twice.
// Generated packages use it.
func DriveAsync[O, T any](
	ctx context.Context,
	p *Program,
	status int32,
	out []byte,
	op Type[O],
	value Type[T],
	perform func(ctx context.Context, op O, w *Writer) error,
) (T, error) {
	var zero T
	switch status {
	case statusOK:
	case statusMalformed:
		return zero, &MalformedError{Message: string(out)}
	default:
		panic(fmt.Sprintf("lungo: a generated entry point returned status %d", status))
	}
	for {
		r := NewReader(p, out)
		kind, err := r.U8()
		if err != nil {
			panic("lungo: the runtime produced a malformed step: " + err.Error())
		}
		switch kind {
		case stepDone:
			v, err := value.Decode(r)
			if err == nil {
				err = r.Finish()
			}
			if err != nil {
				panic("lungo: the runtime produced a malformed step: " + err.Error())
			}
			return v, nil
		case stepCall:
		default:
			panic(fmt.Sprintf("lungo: the runtime produced a step of kind %d", kind))
		}
		o, err := op.Decode(r)
		if err != nil {
			panic("lungo: the runtime produced a malformed operation: " + err.Error())
		}
		id, err := r.U64()
		if err == nil {
			err = r.Finish()
		}
		if err != nil {
			panic("lungo: the runtime produced a malformed step: " + err.Error())
		}
		out, err = performAndResume(ctx, p, id, o, perform)
		if err != nil {
			return zero, err
		}
	}
}

// performAndResume performs operation o of the program waiting on resumption id and resumes it,
// returning its next step; or cancels it, when the operation fails, panics or ctx ends.
func performAndResume[O any](
	ctx context.Context,
	p *Program,
	id uint64,
	o O,
	perform func(ctx context.Context, op O, w *Writer) error,
) (next []byte, err error) {
	resumed := false
	defer func() {
		if !resumed {
			cancelResumption(id)
		}
	}()
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	w := &Writer{program: p, result: true}
	if err := perform(ctx, o, w); err != nil {
		return nil, err
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	var buf C.lungo_buffer
	status := C.lungo_async_resume(C.uint64_t(id), bytesPtr(w.buf), C.size_t(len(w.buf)), &buf)
	resumed = true
	bytes := takeBuffer(&buf)
	switch status {
	case statusOK:
		return bytes, nil
	case statusMalformed:
		// The answer was encoded at the operation's answer type: the runtime refusing it is a
		// defect, not a failure of the host.
		panic("lungo: the runtime refused an answer: " + string(bytes))
	default:
		panic(fmt.Sprintf("lungo: resuming an async program returned status %d: %s", status, string(bytes)))
	}
}

// cancelResumption gives up the program waiting on resumption id.
func cancelResumption(id uint64) {
	if status := C.lungo_async_cancel(C.uint64_t(id)); status != statusOK {
		panic(fmt.Sprintf("lungo: resumption %d was resumed or cancelled already", id))
	}
}

// Outstanding is the number of async programs waiting for an answer: zero once every async call
// has returned.
func Outstanding() int { return int(C.lungo_async_outstanding()) }
