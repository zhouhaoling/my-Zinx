package ziface

// 将客户端请求的连接信息和请求的数据包装到 Request 里
type IRequest interface {
	// 获取请求连接信息
	GetConnection() IConnection
	// 获取请求消息的数据
	GetData() []byte
	GetMsgID() uint32
}
