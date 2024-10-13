package ztimer

import (
	"fmt"
	"testing"
)

func SayHello(message ...interface{}) {
	fmt.Println(message[0].(string), " ", message[1].(string))
}

func TestDelayFunc_Call(t *testing.T) {
	df := NewDelayFunc(SayHello, []interface{}{"hello", "zinx!"})
	fmt.Println("df.String() = ", df.String())
	df.Call()
}
