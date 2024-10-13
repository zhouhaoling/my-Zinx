package ziface

type IConnManager interface {
	// 添加连接
	Add(conn IConnection)
	// 移除连接
	Remove(conn IConnection)
	// 使用 Conn 获取连接
	Get(connID uint32) (IConnection, error)
	// 获取当前连接管理模块的总连接个数
	Len() int
	// 删除并停止所有连接
	ClearConn()
}
