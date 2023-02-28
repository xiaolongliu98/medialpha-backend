package utils

import (
	"fmt"
	"medialpha-backend/types"
	"strings"
)

type StringWrapper struct {
	Data string
}

func (w *StringWrapper) In(s ...string) bool {
	for _, e := range s {
		if e == w.Data {
			return true
		}
	}
	return false
}

func (w *StringWrapper) NotIn(s ...string) bool {
	return !w.In(s...)
}

func (w *StringWrapper) EndsWithAny(s ...string) bool {
	for _, e := range s {
		if strings.HasSuffix(w.Data, e) {
			return true
		}
	}
	return false
}
func (w *StringWrapper) StartsWithAny(s ...string) bool {
	for _, e := range s {
		if strings.HasPrefix(w.Data, e) {
			return true
		}
	}
	return false
}

func (w *StringWrapper) ContainsAny(s ...string) bool {
	for _, e := range s {
		if strings.Contains(w.Data, e) {
			return true
		}
	}
	return false
}

func S(s string) *StringWrapper {
	return &StringWrapper{
		Data: s,
	}
}

type IntWrapper struct {
	Data int
}

func (w *IntWrapper) In(s ...int) bool {
	for _, e := range s {
		if e == w.Data {
			return true
		}
	}
	return false
}

func (w *IntWrapper) NotIn(s ...int) bool {
	return !w.In(s...)
}

func I(s int) *IntWrapper {
	return &IntWrapper{
		Data: s,
	}
}

// WithDefault
// @Description   将指定的返回值替换成默认值y
// @Author        xiaolong
// @Date          2022/11/29 16:11(create);
// @Param         x       	T     value
// @Param         undesired T     想要替换的值
// @Param         y 		T     默认值
// @Return        T
func WithDefault[T types.BaseType](x, undesired, y T) T {
	if x == undesired {
		return y
	}
	return x
}

func ErrorsOf(errs ...error) error {
	for _, err := range errs {
		if err != nil {
			return err
		}
	}
	return nil
}

func ErrorNil() error {
	return fmt.Errorf("空指针异常")
}
