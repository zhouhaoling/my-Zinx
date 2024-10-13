package ziface

// 服务接口
type IServer interface {
	// 启动服务器
	Start()
	// 停止服务
	Stop()
	// 开启业务服务
	Serve()
	// 路由功能，给当前的服务注册一个路由业务方法，
	// 供客户端连接处理使用
	AddRouter(msgID uint32, router IRouter)
	// 得到连接管理器
	GetConnMgr() IConnManager
	// 设置该 server 连接创建时的 hook 函数
	SetOnConnStart(func(IConnection))
	// 设置该 server 连接断开时的 hook 函数
	SetOnConnStop(func(IConnection))
	// 调用连接创建时的 hook 函数
	CallOnConnStart(IConnection)
	// 调用连接断开时的 hook 函数
	CallOnConnStop(IConnection)
	Packet() IPacket
}
