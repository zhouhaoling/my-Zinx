package znet

import (
	"fmt"
	"github.com/dokidokikoi/my-zinx/utils"
	"github.com/dokidokikoi/my-zinx/ziface"
	"strconv"
)

type MsgHandler struct {
	Apis map[uint32]ziface.IRouter
	// 业务 Worker 池的数量
	WorkerPoolSize uint32
	// Worker 负责任务的消息队列
	TaskQueue []chan ziface.IRequest
}

func (mh *MsgHandler) DoMsgHandler(request ziface.IRequest) {
	handler, ok := mh.Apis[request.GetMsgID()]
	if !ok {
		fmt.Printf("api msgID=%d is not FOUND!\n", request.GetMsgID())
		return
	}

	// 执行对应处理方法
	handler.PreHandle(request)
	handler.Handle(request)
	handler.PostHandle(request)
}

func (mh *MsgHandler) AddRouter(msgID uint32, router ziface.IRouter) {
	// 1.判断当前 msg 绑定的 API 处理方式是否已经存在
	if _, ok := mh.Apis[msgID]; ok {
		panic("repeated api, msgID = " + strconv.Itoa(int(msgID)))
	}
	// 2.添加 msg 与 api 的绑定关系
	mh.Apis[msgID] = router
	fmt.Println("Add api msgID = ", msgID)
}

func (mh *MsgHandler) StartOneWorker(workerID int, taskQueue chan ziface.IRequest) {
	fmt.Println("Worker ID = ", workerID, " is started")
	// 不断等待消息队列中的消息
	for {
		select {
		// 如果有消息，则取出队列的 Request，并执行绑定的业务方法
		case req := <-taskQueue:
			mh.DoMsgHandler(req)
		}
	}
}

func (mh *MsgHandler) StartWorkerPool() {
	// 遍历需要启动的 worker 数量，依次启动
	for i := 0; i < int(mh.WorkerPoolSize); i++ {
		mh.TaskQueue[i] = make(chan ziface.IRequest, utils.GlobalObject.MaxWorkerTaskLen)
		// 启动当前 worker，阻塞的等待对应消息队列是否有消息传递进来
		go mh.StartOneWorker(i, mh.TaskQueue[i])
	}
}

// 根据 ConnID 来分配当前的连接应该由哪个 worker 负责处理
// 轮询的平均分配法则
func (mh *MsgHandler) SendMsg2TaskQueue(request ziface.IRequest) {
	// 得到需要处理此条连接的 workerID
	workerID := request.GetConnection().GetConnID() % mh.WorkerPoolSize

	fmt.Printf("Add ConnID %d request msgID = %d to workerID = %d\n",
		request.GetConnection().GetConnID(), request.GetMsgID(), workerID)
	// 将请求消息发送给任务队列
	mh.TaskQueue[workerID] <- request
}

func NewMsgHandler() *MsgHandler {
	return &MsgHandler{
		Apis:           make(map[uint32]ziface.IRouter),
		WorkerPoolSize: utils.GlobalObject.WorkerPoolSize,
		TaskQueue:      make([]chan ziface.IRequest, utils.GlobalObject.WorkerPoolSize),
	}
}
