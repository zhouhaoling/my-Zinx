package ztimer

import (
	"fmt"
	"github.com/dokidokikoi/my-zinx/zlog"
	"reflect"
)

// DelayFunc 延迟调用函数对象
type DelayFunc struct {
	// 延迟调用函数对象
	f func(...interface{})
	// 延迟调用函数传递的形参
	args []interface{}
}

// 打印当前延迟函数信息，用于日志记录
func (df *DelayFunc) String() string {
	return fmt.Sprintf("{DelayFunc:%s, args:%v}", reflect.TypeOf(df.f).Name(), df.args)
}

// 执行延迟函数，如果执行失败则抛出异常
func (df *DelayFunc) Call() {
	defer func() {
		if err := recover(); err != nil {
			zlog.Error(df.String(), "Call err: ", err)
		}
	}()

	df.f(df.args...)
}

func NewDelayFunc(f func(...interface{}), args []interface{}) *DelayFunc {
	return &DelayFunc{
		f:    f,
		args: args,
	}
}
