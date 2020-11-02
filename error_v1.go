package errx

import (
	"errors"
	"fmt"
	"path"
	"runtime"
	"strings"
	"sync"

	"github.com/kr/pretty"

	fbs "github.com/google/flatbuffers/go"
)

const stackTpl = "%s:%d -> %s()"

var fbsPool = sync.Pool{New: func() interface{} { return fbs.NewBuilder(128) }}

func newErrorV1(text string) Error {
	return &v1Error{
		text: text,
	}
}

func unpackV1(buf []byte) Error {
	return new(v1Error).importModel(GetRootAsErrorModel(buf, 0).UnPack())
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

	if e.proto != nil && errors.Is(e.proto, err) {
		return true
	}

	if e.reason != nil && errors.Is(e.reason, err) {
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

func (e *v1Error) WithDebug(items Debug) Error {
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
	e.deep = 0
}

func (e *v1Error) Export() *View {
	v := &View{
		Text:   e.text,
		Detail: e.detail,
		Stack:  e.stack,
		Debug:  e.debug,
	}

	if e.reason != nil && e.deep < 10 {
		e.deep++
		if next, ok := e.reason.(Error); ok {
			v.Next = next.Export()
		} else {
			v.Next = &View{Text: e.reason.Error()}
		}
	}
	e.deep = 0
	return v
}

func (e *v1Error) Pack() []byte {
	buf := fbsPool.Get().(*fbs.Builder)
	buf.Finish(e.exportModel().Pack(buf))
	res := buf.FinishedBytes()
	buf.Reset()
	fbsPool.Put(buf)
	return res
}

func (e *v1Error) withStack() *v1Error {
	var frame int

	err := &v1Error{
		text:   e.text,
		detail: e.detail,
		debug:  e.debug,
		reason: e.reason,
		proto:  e,
		stack:  make([]string, 0, 16),
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

func (e *v1Error) exportModel() *ErrorModelT {
	m := &ErrorModelT{
		Text:   e.text,
		Detail: e.detail,
		Stack:  e.stack,
		Debug:  make([]*KeyValueT, 0, len(e.debug)),
	}

	for k, v := range e.debug {
		m.Debug = append(m.Debug, &KeyValueT{
			Key:   k,
			Value: v,
		})
	}

	if e.reason != nil && e.deep < 10 {
		e.deep++
		if next, ok := e.reason.(*v1Error); ok {
			m.Next = next.exportModel()
		} else {
			m.Next = &ErrorModelT{Text: e.reason.Error()}
		}
	}

	e.deep = 0
	return m
}

func (e *v1Error) importModel(m *ErrorModelT) *v1Error {
	e.text = m.Text
	e.detail = m.Detail
	e.stack = m.Stack
	e.debug = make(map[string]string, len(m.Debug))

	for i := range m.Debug {
		e.debug[m.Debug[i].Key] = m.Debug[i].Value
	}

	if m.Next != nil {
		e.reason = new(v1Error).importModel(m.Next)
	}

	return e
}
