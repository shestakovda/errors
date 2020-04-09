package errors

import (
	stderr "errors"
	"fmt"
	"path"
	"runtime"
	"strings"

	"github.com/kr/pretty"
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
	detail string
	stack  []string
	debug  map[string]string
	proto  *v1Error
	reason error
}

func (e *v1Error) Error() string    { return e.text }
func (e *v1Error) Unwrap() error    { return e.reason }
func (e *v1Error) WithStack() Error { return e.withStack() }

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

func (e *v1Error) WithDetail(tpl string, args ...interface{}) Error {
	err := e.withStack()
	err.detail = fmt.Sprintf(tpl, args...)
	return err
}

func (e *v1Error) WithDebug(items map[string]interface{}) Error {
	err := e.withStack()
	err.debug = make(map[string]string, len(items))
	for key := range items {
		err.debug[key] = fmt.Sprintf("%#v", pretty.Formatter(items[key]))
	}
	return err
}

func (e *v1Error) Format(f fmt.State, r rune) {
	// Сначала всегда на той же строке основное сообщение
	fmt.Fprintf(f, "> %s", e.text)

	// Затем в скобках детализация для пользователя
	if e.detail != "" {
		fmt.Fprintf(f, " (%s)", e.detail)
	}

	// Если не нужна детальная инфа, на этом все
	if r != 'v' {
		return
	}

	// Затем, на каждой строчке со сдвигом и кареткой, отладка (если есть)
	for key := range e.debug {
		fmt.Fprintf(f, "\n|   %s: %s", key, e.debug[key])
	}

	// Затем, если нужны подробности, выводим стек
	if f.Flag('+') && len(e.stack) > 0 {
		fmt.Fprintf(f, "\n|       %s", strings.Join(e.stack, "\n|       "))
	}

	// Затем, если есть кто-то в цепочке, выводим его со след. строки
	if e.reason != nil && e.deep < 10 {
		e.deep++

		if next, ok := e.reason.(Error); ok {
			if f.Flag('+') {
				fmt.Fprintf(f, "\n|-%+v", next)
			} else {
				fmt.Fprintf(f, "\n|-%v", next)
			}
		} else {
			fmt.Fprintf(f, "\n|-> %s", e.reason)
		}
	}
}

func (e *v1Error) Export() *View {
	v := &View{
		Text:   e.text,
		Detail: e.detail,
		Stack:  e.stack,
		Debug:  e.debug,
	}

	if e.reason != nil {
		if next, ok := e.reason.(Error); ok {
			v.Next = next.Export()
		} else {
			v.Next = &View{Text: e.reason.Error()}
		}
	}

	return v
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
