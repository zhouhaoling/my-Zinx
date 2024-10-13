package ziface

import "net"

// 连接接口
type IConnection interface {
	// 启动连接，让当前连接开始工作
	Start()
	// 停止连接，结束当前连接状态
	Stop()
	// 从当前连接获取原始的 socket TCPConn
	GetTCPConnection() *net.TCPConn
	// 获取当前连接 ID
	GetConnID() uint32
	// 获取远程客户端地址信息
	RemoteAddr() net.Addr
	// 直接将 Message 数据发送给远程的 TCP 客户端
	SendMsg(msgID uint32, data []byte) error
	// 直接将 Message 数据发送给远程的 TCP 客户端(带缓冲)
	SendBuffMsg(msgID uint32, data []byte) error

	// 设置连接属性
	SetProperty(key string, value interface{})
	// 获取连接属性
	GetProperty(key string) (interface{}, error)
	// 移除连接属性
	RemoveProperty(key string)
}

// 定义一个统一处理连接业务的接口
type HandFunc func(*net.TCPConn, []byte, int) error
