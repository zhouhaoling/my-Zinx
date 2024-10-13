package ziface

// 路由接口
type IRouter interface {
	// 在处理 conn 业务之前的钩子方法
	PreHandle(request IRequest)
	// 处理 conn 业务方法
	Handle(request IRequest)
	// 处理 conn 业务之后的钩子方法
	PostHandle(request IRequest)
}
