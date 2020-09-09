package errx_test

import (
	"fmt"
	"io"
	"regexp"
	"testing"

	"github.com/shestakovda/errx"
	"github.com/stretchr/testify/suite"
)

// TestErrors - индивидуальные тесты драйверов
func TestErrors(t *testing.T) {
	suite.Run(t, new(InterfaceSuite))
}

type InterfaceSuite struct {
	suite.Suite
}

func (s *InterfaceSuite) TestError() {
	msg := "some msg"
	err := errx.New(msg).WithStack()
	s.Equal(msg, err.Error())
}

func (s *InterfaceSuite) TestDebug() {
	msg := "some msg"
	err := errx.New(msg).WithDebug(errx.Debug{
		"int":  42,
		"list": []string{"some", "test"},
	})
	s.Equal(msg, err.Error())

	s.Equal(map[string]string{
		"int":  "42",
		"list": `[]string{"some", "test"}`,
	}, err.Export().Debug)
}

func (s *InterfaceSuite) TestDetail() {
	msg := "some msg"
	err := errx.New(msg).WithDetail("some %d %s", 42, "msg")
	s.Equal(msg, err.Error())
	s.Equal("some 42 msg", err.Export().Detail)
}

func (s *InterfaceSuite) TestReason() {
	err1 := io.EOF
	err2 := errx.New("some reason")
	err3 := errx.New("some other")

	err := err3.WithDetail("some arg").WithReason(err2.WithReason(err1))

	s.False(err.Is(nil))
	s.True(err.Is(err))
	s.True(errx.Is(err, err3))
	s.True(errx.Is(err, err2))
	s.True(errx.Is(err, err1))
	s.Equal(err3.Error(), err.Error())
	s.True(errx.Is(errx.Unwrap(err), err2))
}

func (s *InterfaceSuite) TestFormat() {
	const line = ".go:123"
	const lineS = ".s:1373"

	err1 := io.EOF
	err2 := errx.New("error 2").WithDebug(errx.Debug{
		"list": []string{"some", "test"},
	})
	err3 := errx.New("error 3").WithDetail("some %d %s", 42, "msg").WithDebug(errx.Debug{
		"err1": err1,
	})

	err := err3.WithReason(err2.WithReason(err1))

	s.Equal(`
> error 2
`, fmt.Sprintf("\n%s\n", err2))

	s.Equal(`
> error 3 (some 42 msg)
`, fmt.Sprintf("\n%s\n", err))

	s.Equal(`
> error 3 (some 42 msg)
|   err1: &errors.errorString{s:"EOF"}
|-> error 2
|   list: []string{"some", "test"}
|-> EOF
`, fmt.Sprintf("\n%v\n", err))

	s.Equal(`
> error 3 (some 42 msg)
|   err1: &errors.errorString{s:"EOF"}
|       error_v1.go:123 -> errx.(*v1Error).WithReason()
|       interface_test.go:123 -> errx_test.(*InterfaceSuite).TestFormat()
|       value.go:123 -> reflect.Value.call()
|       value.go:123 -> reflect.Value.Call()
|       suite.go:123 -> suite.Run.func2()
|       testing.go:123 -> testing.tRunner()
|       asm_amd64.s:1373 -> runtime.goexit()
|-> error 2
|   list: []string{"some", "test"}
|       error_v1.go:123 -> errx.(*v1Error).WithReason()
|       interface_test.go:123 -> errx_test.(*InterfaceSuite).TestFormat()
|       value.go:123 -> reflect.Value.call()
|       value.go:123 -> reflect.Value.Call()
|       suite.go:123 -> suite.Run.func2()
|       testing.go:123 -> testing.tRunner()
|       asm_amd64.s:1373 -> runtime.goexit()
|-> EOF
`, lineSRx.ReplaceAllString(lineRx.ReplaceAllString(fmt.Sprintf("\n%+v\n", err), line), lineS))

	v := err.Export()

	s.Equal("error 3", v.Text)
	s.Equal("some 42 msg", v.Detail)
	s.Len(v.Stack, 7)
	s.Equal(`error_v1.go:123 -> errx.(*v1Error).WithReason()`, lineRx.ReplaceAllString(v.Stack[0], line))
	s.Equal(`&errors.errorString{s:"EOF"}`, v.Debug["err1"])

	if v = v.Next; s.NotNil(v) {

		s.Equal("error 2", v.Text)
		s.Empty(v.Detail)
		s.Len(v.Stack, 7)
		s.Equal(`error_v1.go:123 -> errx.(*v1Error).WithReason()`, lineRx.ReplaceAllString(v.Stack[0], line))
		s.Equal(`[]string{"some", "test"}`, v.Debug["list"])

		if v = v.Next; s.NotNil(v) {
			s.Equal("EOF", v.Text)
			s.Empty(v.Detail)
			s.Empty(v.Stack)
			s.Empty(v.Debug)
			s.Nil(v.Next)
		}
	}
}

var lineRx = regexp.MustCompile(`\.go:\d+`)
var lineSRx = regexp.MustCompile(`\.s:\d+`)
