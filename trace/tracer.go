package trace

import (
	"fmt"
	"io"
)

// Tracer is the interface that describes an object capable of
// tracing events throughout code
type Tracer interface {
	Trace(...interface{})
}

// taking io.Writer as the input, we can be sure that the user of the lib
// can decide where the output should be written, network, file, stdout etc.
func New(w io.Writer) Tracer {
	return &tracer{out: w}
}

type tracer struct {
	out io.Writer
}

func (t *tracer) Trace(a ...interface{}) {
	fmt.Fprintln(t.out, a...)
	fmt.Println(t.out)
}

type nilTracer struct{}

func (t *nilTracer) Trace(a ...interface{}) {}

// Off creates a Tracer that will ignore calls to Trace.
func Off() Tracer {
	return &nilTracer{}
}
