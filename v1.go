package errors

import (
	stderr "errors"
	"fmt"
	"path"
	"runtime"
	"strings"
)

const stackTpl = "%s:%d -> %s()"

func newErrorV1(text string) Error {
	return &v1Error{
		text: text,
	}
}

type v1Error struct {
	deep   int
	text   string
	item   interface{}
	args   []interface{}
	stack  []string
	proto  *v1Error
	reason error
}

func (e *v1Error) Item() interface{}   { return e.item }
func (e *v1Error) Args() []interface{} { return e.args }
func (e *v1Error) Error() string       { return e.text }
func (e *v1Error) Unwrap() error       { return e.reason }
func (e *v1Error) WithStack() Error    { return e.withStack() }

func (e *v1Error) Is(err error) bool {
	if err == nil {
		return false
	}

	if e == err {
		return true
	}

	if e.text == err.Error() {
		return true
	}

	if e.proto != nil && stderr.Is(e.proto, err) {
		return true
	}

	if e.reason != nil && stderr.Is(e.reason, err) {
		return true
	}

	return false
}

func (e *v1Error) WithReason(reason error) Error {
	err := e.withStack()
	err.reason = reason
	return err
}

func (e *v1Error) WithArgs(args ...interface{}) Error {
	err := e.withStack()
	err.args = args
	err.text = fmt.Sprintf(err.text, args...)
	return err
}

func (e *v1Error) WithItem(item interface{}) Error {
	err := e.withStack()
	err.item = item
	return err
}

func (e *v1Error) Format(f fmt.State, r rune) {
	f.Write([]byte(e.text))

	if r != 'v' {
		return
	}

	if e.item != nil {
		fmt.Fprintf(f, " (%+v)", e.item)
	}

	if e.reason != nil && e.deep < 10 {
		e.deep++
		fmt.Fprintf(f, "\n\t<- %v", e.reason)
	}

	if f.Flag('+') && len(e.stack) > 0 {
		fmt.Fprintf(f, "\n\t%s", strings.Join(e.stack, "\n\t"))
	}
}

func (e *v1Error) withStack() *v1Error {
	var frame int

	if e.proto != nil {
		return e
	}

	err := &v1Error{
		text:  e.text,
		proto: e,
		stack: make([]string, 0, 16),
	}

	for {
		frame++
		if pc, file, line, ok := runtime.Caller(frame); ok {
			err.stack = append(err.stack, fmt.Sprintf(
				stackTpl, path.Base(file), line, path.Base(runtime.FuncForPC(pc).Name()),
			))
		} else {
			break
		}
	}

	return err
}
