package znet

import "github.com/dokidokikoi/my-zinx/ziface"

// 实现 IRouter 时，先嵌入这个基类，
// 然后根据需要对这个基类的方法进行重写
// 好处是，嵌入基类的实现类不需要实现 PreHandle 等方法也可以实例化
type BaseRouter struct{}

func (r *BaseRouter) PreHandle(req ziface.IRequest)  {}
func (r *BaseRouter) Handle(req ziface.IRequest)     {}
func (r *BaseRouter) PostHandle(req ziface.IRequest) {}
