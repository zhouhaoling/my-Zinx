package ziface

type IMsgHandler interface {
	// 马上以非阻塞的方式处理消息
	DoMsgHandler(request IRequest)
	// 为消息添加具体的处理逻辑
	AddRouter(msgID uint32, router IRouter)
	// 启动 Worker 工作池
	StartWorkerPool()
	// 将消息交给 TaskQueue，由 Worker 进行管理
	SendMsg2TaskQueue(request IRequest)
}
