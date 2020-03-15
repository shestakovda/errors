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

func (s *InterfaceSuite) TestItem() {
	item := map[string]int{"ololo": 42}
	err := errors.New("some msg")
	err = err.WithItem(item)
	s.Equal(item, err.Item())
}

func (s *InterfaceSuite) TestArgs() {
	args := []interface{}{42, "msg"}
	err := errors.New("some %d %s")
	err = err.WithArgs(args...)
	s.Equal("some 42 msg", err.Error())
	s.Equal(args, err.Args())
}

func (s *InterfaceSuite) TestReason() {
	err1 := io.EOF
	err2 := errors.New("some reason")
	err3 := errors.New("some %s")
	err4 := errors.New("some arg")

	err := err3.WithArgs("arg").WithReason(err2.WithReason(err1))

	s.False(err.Is(nil))
	s.True(err.Is(err))
	s.True(errors.Is(err, err4))
	s.True(errors.Is(err, err3))
	s.True(errors.Is(err, err2))
	s.True(errors.Is(err, err1))
	s.Equal(err4.Error(), err.Error())
	s.True(errors.Is(errors.Unwrap(err), err2))
}

func (s *InterfaceSuite) TestFormat() {
	err1 := io.EOF
	err2 := errors.New("some reason").WithItem("test")
	err3 := errors.New("some %s").WithArgs("error").WithItem([]string{"ololo", "test"})

	err := err3.WithReason(err2.WithReason(err1))

	log := fmt.Sprintf("\n%s\n", err)
	s.Equal(`
some error
`, log)

	log = fmt.Sprintf("\n%v\n", err)
	s.Equal(`
some error ([ololo test])
	<- some reason (test)
	<- EOF
`, log)

	log = fmt.Sprintf("\n%+v\n", err)
	s.Equal(`
some error ([ololo test])
	<- some reason (test)
	<- EOF
	v1.go:66 -> errors.(*v1Error).WithArgs()
	interface_test.go:63 -> errors_test.(*InterfaceSuite).TestFormat()
	value.go:460 -> reflect.Value.call()
	value.go:321 -> reflect.Value.Call()
	suite.go:137 -> suite.Run.func2()
	testing.go:909 -> testing.tRunner()
	asm_amd64.s:1357 -> runtime.goexit()
`, log)

}
