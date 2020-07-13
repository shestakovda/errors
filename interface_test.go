package errors_test

import (
	"fmt"
	"io"
	"testing"

	"github.com/shestakovda/errors"
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
	err := errors.New(msg).WithStack()
	s.Equal(msg, err.Error())
}

func (s *InterfaceSuite) TestDebug() {
	msg := "some msg"
	err := errors.New(msg).WithDebug(map[string]interface{}{
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
	err := errors.New(msg).WithDetail("some %d %s", 42, "msg")
	s.Equal(msg, err.Error())
	s.Equal("some 42 msg", err.Export().Detail)
}

func (s *InterfaceSuite) TestReason() {
	err1 := io.EOF
	err2 := errors.New("some reason")
	err3 := errors.New("some other")

	err := err3.WithDetail("some arg").WithReason(err2.WithReason(err1))

	s.False(err.Is(nil))
	s.True(err.Is(err))
	s.True(errors.Is(err, err3))
	s.True(errors.Is(err, err2))
	s.True(errors.Is(err, err1))
	s.Equal(err3.Error(), err.Error())
	s.True(errors.Is(errors.Unwrap(err), err2))
}

func (s *InterfaceSuite) TestFormat() {
	err1 := io.EOF
	err2 := errors.New("error 2").WithDebug(map[string]interface{}{
		"list": []string{"some", "test"},
	})
	err3 := errors.New("error 3").WithDetail("some %d %s", 42, "msg").WithDebug(map[string]interface{}{
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
|       v1.go:60 -> errors.(*v1Error).WithReason()
|       interface_test.go:73 -> errors_test.(*InterfaceSuite).TestFormat()
|       value.go:460 -> reflect.Value.call()
|       value.go:321 -> reflect.Value.Call()
|       suite.go:137 -> suite.Run.func2()
|       testing.go:991 -> testing.tRunner()
|       asm_amd64.s:1373 -> runtime.goexit()
|-> error 2
|   list: []string{"some", "test"}
|       v1.go:60 -> errors.(*v1Error).WithReason()
|       interface_test.go:73 -> errors_test.(*InterfaceSuite).TestFormat()
|       value.go:460 -> reflect.Value.call()
|       value.go:321 -> reflect.Value.Call()
|       suite.go:137 -> suite.Run.func2()
|       testing.go:991 -> testing.tRunner()
|       asm_amd64.s:1373 -> runtime.goexit()
|-> EOF
`, fmt.Sprintf("\n%+v\n", err))

	v := err.Export()

	s.Equal("error 3", v.Text)
	s.Equal("some 42 msg", v.Detail)
	s.Len(v.Stack, 7)
	s.Equal(`v1.go:60 -> errors.(*v1Error).WithReason()`, v.Stack[0])
	s.Equal(`&errors.errorString{s:"EOF"}`, v.Debug["err1"])

	if v = v.Next; s.NotNil(v) {

		s.Equal("error 2", v.Text)
		s.Empty(v.Detail)
		s.Len(v.Stack, 7)
		s.Equal(`v1.go:60 -> errors.(*v1Error).WithReason()`, v.Stack[0])
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
